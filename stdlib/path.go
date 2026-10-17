package stdlib

import (
	"strings"

	"github.com/AurielSolaris/runescript/internal/types"
)

// std:path: path manipulation with the semantics of Node.js's path
// module, implemented here (and, line for line, in the JavaScript
// runtime) rather than borrowed from either host, so both targets give
// the same results.
//
// path.join, path.resolve, ... follow the rules of the platform the
// program runs on: Windows rules on Windows, POSIX rules elsewhere.
// path.posixJoin, path.win32Join, ... choose one explicitly, for paths
// that must look the same everywhere.
//
// Strings are handled as bytes. Every character the rules look at
// (separators, '.', ':', drive letters) is ASCII, and UTF-8 never uses
// ASCII bytes inside other characters, so this matches Node, which works
// on UTF-16 code units. Two differences from Node, to keep both targets
// deterministic: the Windows rules compare paths ignoring ASCII case only,
// and resolving a path on another drive ("D:x" while the working
// directory is on C:) starts from that drive's root instead of consulting
// cmd.exe's per-drive "=D:" environment variables.

// pathFlavor is one platform's path rules.
type pathFlavor struct {
	win bool
	sep string
}

var (
	posixPath = pathFlavor{sep: "/"}
	win32Path = pathFlavor{win: true, sep: `\`}
)

func (f pathFlavor) isSep(c byte) bool { return c == '/' || (f.win && c == '\\') }

func isDriveLetter(c byte) bool { return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' }

// hasDrive reports whether p starts with a drive like "C:".
func hasDrive(p string) bool { return len(p) >= 2 && isDriveLetter(p[0]) && p[1] == ':' }

// normalizeString resolves "." and ".." segments and repeated separators
// in p, which has no root. allowAboveRoot keeps ".." segments that go
// above the start.
func (f pathFlavor) normalizeString(p string, allowAboveRoot bool) string {
	var res string
	lastSegmentLength, lastSlash, dots := 0, -1, 0
	var c byte
	for i := 0; i <= len(p); i++ {
		switch {
		case i < len(p):
			c = p[i]
		case f.isSep(c):
			return res
		default:
			c = '/'
		}
		if f.isSep(c) {
			switch {
			case lastSlash == i-1 || dots == 1:
				// an empty or "." segment
			case dots == 2:
				if len(res) < 2 || lastSegmentLength != 2 || res[len(res)-1] != '.' || res[len(res)-2] != '.' {
					if len(res) > 2 {
						if j := strings.LastIndex(res, f.sep); j == -1 {
							res, lastSegmentLength = "", 0
						} else {
							res = res[:j]
							lastSegmentLength = len(res) - 1 - strings.LastIndex(res, f.sep)
						}
						lastSlash, dots = i, 0
						continue
					} else if len(res) != 0 {
						res, lastSegmentLength = "", 0
						lastSlash, dots = i, 0
						continue
					}
				}
				if allowAboveRoot {
					if len(res) > 0 {
						res += f.sep + ".."
					} else {
						res = ".."
					}
					lastSegmentLength = 2
				}
			default:
				if len(res) > 0 {
					res += f.sep + p[lastSlash+1:i]
				} else {
					res = p[lastSlash+1 : i]
				}
				lastSegmentLength = i - lastSlash - 1
			}
			lastSlash, dots = i, 0
		} else if c == '.' && dots != -1 {
			dots++
		} else {
			dots = -1
		}
	}
	return res
}

// uncRoot parses the "\\server\share" at the start of p, which begins
// with two separators. It returns the device written with backslashes,
// the index after the share (len(p) if the share ends p), and whether p
// has both a server and a share.
func (f pathFlavor) uncRoot(p string) (device string, end int, ok bool) {
	j := 2
	for j < len(p) && !f.isSep(p[j]) {
		j++
	}
	if j >= len(p) || j == 2 {
		return "", 0, false
	}
	server := p[2:j]
	last := j
	for j < len(p) && f.isSep(p[j]) {
		j++
	}
	if j >= len(p) || j == last {
		return "", 0, false
	}
	last = j
	for j < len(p) && !f.isSep(p[j]) {
		j++
	}
	return `\\` + server + `\` + p[last:j], j, true
}

// winRoot splits a Windows path into its device ("C:" or
// "\\server\share", or "" for none), where the rest starts, and whether
// it is absolute.
func (f pathFlavor) winRoot(p string) (device string, rootEnd int, abs bool) {
	switch {
	case len(p) == 0:
	case len(p) == 1:
		if f.isSep(p[0]) {
			return "", 1, true
		}
	case f.isSep(p[0]):
		if f.isSep(p[1]) {
			if device, end, ok := f.uncRoot(p); ok {
				return device, end, true
			}
		}
		return "", 1, true
	case hasDrive(p):
		if len(p) > 2 && f.isSep(p[2]) {
			return p[:2], 3, true
		}
		return p[:2], 2, false
	}
	return "", 0, false
}

func (f pathFlavor) isAbsolute(p string) bool {
	if f.win {
		return len(p) > 0 && (f.isSep(p[0]) || len(p) > 2 && hasDrive(p) && f.isSep(p[2]))
	}
	return len(p) > 0 && p[0] == '/'
}

func (f pathFlavor) normalize(p string) string {
	if p == "" {
		return "."
	}
	if !f.win {
		abs := p[0] == '/'
		trailing := p[len(p)-1] == '/'
		p2 := f.normalizeString(p, !abs)
		switch {
		case p2 == "" && abs:
			return "/"
		case p2 == "" && trailing:
			return "./"
		case p2 == "":
			return "."
		}
		if trailing {
			p2 += "/"
		}
		if abs {
			return "/" + p2
		}
		return p2
	}
	if len(p) == 1 {
		if p[0] == '/' {
			return `\`
		}
		return p
	}
	var device string
	hasDevice, rootEnd, abs := false, 0, false
	if f.isSep(p[0]) {
		abs, rootEnd = true, 1
		if f.isSep(p[1]) {
			if d, end, ok := f.uncRoot(p); ok {
				if end == len(p) {
					// Only the server and share: it is a root, which ends
					// with a separator.
					return d + `\`
				}
				device, hasDevice, rootEnd = d, true, end
			}
		}
	} else if hasDrive(p) {
		device, hasDevice, rootEnd = p[:2], true, 2
		if len(p) > 2 && f.isSep(p[2]) {
			abs, rootEnd = true, 3
		}
	}
	tail := ""
	if rootEnd < len(p) {
		tail = f.normalizeString(p[rootEnd:], !abs)
	}
	if tail == "" && !abs {
		tail = "."
	}
	if tail != "" && f.isSep(p[len(p)-1]) {
		tail += `\`
	}
	if !abs && !hasDevice && strings.Contains(p, ":") {
		// A relative path must not turn into one that Windows reads as
		// absolute, such as "C:..." (Node.js, CVE-2024-36139).
		if hasDrive(tail) {
			return `.\` + tail
		}
		for i := strings.Index(p, ":"); i != -1; {
			if i == len(p)-1 || f.isSep(p[i+1]) {
				return `.\` + tail
			}
			next := strings.Index(p[i+1:], ":")
			if next == -1 {
				break
			}
			i += next + 1
		}
	}
	if !hasDevice {
		if abs {
			return `\` + tail
		}
		return tail
	}
	if abs {
		return device + `\` + tail
	}
	return device + tail
}

func (f pathFlavor) join(parts []string) string {
	var joined, first string
	for _, p := range parts {
		if p == "" {
			continue
		}
		if joined == "" {
			joined, first = p, p
		} else {
			joined += f.sep + p
		}
	}
	if joined == "" {
		return "."
	}
	if f.win {
		// Keep a leading "\\server" only when the first part starts it.
		needsReplace, slashes := true, 0
		if f.isSep(first[0]) {
			slashes++
			if len(first) > 1 && f.isSep(first[1]) {
				slashes++
				if len(first) > 2 {
					if f.isSep(first[2]) {
						slashes++
					} else {
						needsReplace = false
					}
				}
			}
		}
		if needsReplace {
			for slashes < len(joined) && f.isSep(joined[slashes]) {
				slashes++
			}
			if slashes >= 2 {
				joined = `\` + joined[slashes:]
			}
		}
	}
	return f.normalize(joined)
}

// cwd is the working directory as these rules see it: POSIX rules on
// Windows drop the drive and use '/' ("C:\Users\a" is "/Users/a").
func (f pathFlavor) cwd(env *Env) string {
	dir := env.getwd()
	if !f.win && env.Windows {
		dir = strings.ReplaceAll(dir, `\`, "/")
		if i := strings.Index(dir, "/"); i >= 0 {
			dir = dir[i:]
		}
	}
	return dir
}

func (f pathFlavor) resolve(env *Env, parts []string) string {
	if !f.win {
		resolved, abs := "", false
		for i := len(parts) - 1; i >= -1 && !abs; i-- {
			var p string
			if i >= 0 {
				p = parts[i]
			} else {
				p = f.cwd(env)
			}
			if p == "" {
				continue
			}
			resolved = p + "/" + resolved
			abs = p[0] == '/'
		}
		resolved = f.normalizeString(resolved, !abs)
		if abs {
			return "/" + resolved
		}
		if resolved == "" {
			return "."
		}
		return resolved
	}
	var resolvedDevice, resolvedTail string
	resolvedAbs := false
	for i := len(parts) - 1; i >= -1; i-- {
		var p string
		switch {
		case i >= 0:
			p = parts[i]
			if p == "" {
				continue
			}
		case resolvedDevice == "":
			p = f.cwd(env)
		default:
			p = f.cwd(env)
			if len(p) > 2 && !strings.EqualFold(p[:2], resolvedDevice) && p[2] == '\\' {
				// The working directory is on another drive: start at
				// this drive's root.
				p = resolvedDevice + `\`
			}
		}
		device, rootEnd, abs := f.winRoot(p)
		if device != "" {
			if resolvedDevice != "" {
				if !strings.EqualFold(device, resolvedDevice) {
					continue // another drive: not part of this path
				}
			} else {
				resolvedDevice = device
			}
		}
		if resolvedAbs {
			if resolvedDevice != "" {
				break
			}
		} else {
			resolvedTail = p[rootEnd:] + `\` + resolvedTail
			resolvedAbs = abs
			if abs && resolvedDevice != "" {
				break
			}
		}
	}
	resolvedTail = f.normalizeString(resolvedTail, !resolvedAbs)
	if resolvedAbs {
		return resolvedDevice + `\` + resolvedTail
	}
	if s := resolvedDevice + resolvedTail; s != "" {
		return s
	}
	return "."
}

func (f pathFlavor) relative(env *Env, from, to string) string {
	if from == to {
		return ""
	}
	fromOrig, toOrig := f.resolve(env, []string{from}), f.resolve(env, []string{to})
	if fromOrig == toOrig {
		return ""
	}
	if !f.win {
		fromStart, fromEnd := 1, len(fromOrig)
		fromLen := fromEnd - fromStart
		toStart := 1
		toLen := len(toOrig) - toStart
		length := min(fromLen, toLen)
		lastCommonSep, i := -1, 0
		for ; i < length; i++ {
			c := fromOrig[fromStart+i]
			if c != toOrig[toStart+i] {
				break
			} else if c == '/' {
				lastCommonSep = i
			}
		}
		if i == length {
			if toLen > length {
				if toOrig[toStart+i] == '/' {
					return toOrig[toStart+i+1:]
				}
				if i == 0 {
					return toOrig[toStart+i:]
				}
			} else if fromLen > length {
				if fromOrig[fromStart+i] == '/' {
					lastCommonSep = i
				} else if i == 0 {
					lastCommonSep = 0
				}
			}
		}
		out := ""
		for i = fromStart + lastCommonSep + 1; i <= fromEnd; i++ {
			if i == fromEnd || fromOrig[i] == '/' {
				if out == "" {
					out = ".."
				} else {
					out += "/.."
				}
			}
		}
		return out + toOrig[toStart+lastCommonSep:]
	}
	fromL, toL := asciiLower(fromOrig), asciiLower(toOrig)
	if fromL == toL {
		return ""
	}
	fromStart := 0
	for fromStart < len(fromL) && fromL[fromStart] == '\\' {
		fromStart++
	}
	fromEnd := len(fromL)
	for fromEnd-1 > fromStart && fromL[fromEnd-1] == '\\' {
		fromEnd--
	}
	fromLen := fromEnd - fromStart
	toStart := 0
	for toStart < len(toL) && toL[toStart] == '\\' {
		toStart++
	}
	toEnd := len(toL)
	for toEnd-1 > toStart && toL[toEnd-1] == '\\' {
		toEnd--
	}
	toLen := toEnd - toStart
	length := min(fromLen, toLen)
	lastCommonSep, i := -1, 0
	for ; i < length; i++ {
		c := fromL[fromStart+i]
		if c != toL[toStart+i] {
			break
		} else if c == '\\' {
			lastCommonSep = i
		}
	}
	if i != length {
		if lastCommonSep == -1 {
			return toOrig // different drives
		}
	} else {
		if toLen > length {
			if toL[toStart+i] == '\\' {
				return toOrig[toStart+i+1:]
			}
			if i == 2 {
				return toOrig[toStart+i:] // "C:" to "C:\foo"
			}
		}
		if fromLen > length {
			if fromL[fromStart+i] == '\\' {
				lastCommonSep = i
			} else if i == 2 {
				lastCommonSep = 3
			}
		}
		if lastCommonSep == -1 {
			lastCommonSep = 0
		}
	}
	out := ""
	for i = fromStart + lastCommonSep + 1; i <= fromEnd; i++ {
		if i == fromEnd || fromL[i] == '\\' {
			if out == "" {
				out = ".."
			} else {
				out += `\..`
			}
		}
	}
	toStart += lastCommonSep
	if out != "" {
		return out + toOrig[toStart:toEnd]
	}
	if toOrig[toStart] == '\\' {
		toStart++
	}
	return toOrig[toStart:toEnd]
}

func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

func (f pathFlavor) dirname(p string) string {
	if p == "" {
		return "."
	}
	if !f.win {
		hasRoot := p[0] == '/'
		end, matchedSlash := -1, true
		for i := len(p) - 1; i >= 1; i-- {
			if p[i] == '/' {
				if !matchedSlash {
					end = i
					break
				}
			} else {
				matchedSlash = false
			}
		}
		switch {
		case end == -1 && hasRoot:
			return "/"
		case end == -1:
			return "."
		case hasRoot && end == 1:
			return "//"
		}
		return p[:end]
	}
	rootEnd, offset := -1, 0
	if len(p) == 1 {
		if f.isSep(p[0]) {
			return p
		}
		return "."
	}
	if f.isSep(p[0]) {
		rootEnd, offset = 1, 1
		if f.isSep(p[1]) {
			if _, end, ok := f.uncRoot(p); ok {
				if end == len(p) {
					return p // only a server and share
				}
				rootEnd, offset = end+1, end+1
			}
		}
	} else if hasDrive(p) {
		rootEnd = 2
		if len(p) > 2 && f.isSep(p[2]) {
			rootEnd = 3
		}
		offset = rootEnd
	}
	end, matchedSlash := -1, true
	for i := len(p) - 1; i >= offset; i-- {
		if f.isSep(p[i]) {
			if !matchedSlash {
				end = i
				break
			}
		} else {
			matchedSlash = false
		}
	}
	if end == -1 {
		if rootEnd == -1 {
			return "."
		}
		end = rootEnd
	}
	return p[:end]
}

// basename is the last segment of p, without suffix if it ends with
// suffix (and is more than suffix). suffix may be "".
func (f pathFlavor) basename(p, suffix string) string {
	start, end, matchedSlash := 0, -1, true
	if f.win && hasDrive(p) {
		start = 2
	}
	if suffix != "" && len(suffix) <= len(p) {
		if suffix == p {
			return ""
		}
		extIdx, firstNonSlashEnd := len(suffix)-1, -1
		for i := len(p) - 1; i >= start; i-- {
			c := p[i]
			if f.isSep(c) {
				if !matchedSlash {
					start = i + 1
					break
				}
				continue
			}
			if firstNonSlashEnd == -1 {
				matchedSlash, firstNonSlashEnd = false, i+1
			}
			if extIdx >= 0 {
				if c == suffix[extIdx] {
					if extIdx--; extIdx == -1 {
						end = i
					}
				} else {
					extIdx, end = -1, firstNonSlashEnd
				}
			}
		}
		if start == end {
			end = firstNonSlashEnd
		} else if end == -1 {
			end = len(p)
		}
		return p[start:end]
	}
	for i := len(p) - 1; i >= start; i-- {
		if f.isSep(p[i]) {
			if !matchedSlash {
				start = i + 1
				break
			}
		} else if end == -1 {
			matchedSlash, end = false, i+1
		}
	}
	if end == -1 {
		return ""
	}
	return p[start:end]
}

func (f pathFlavor) extname(p string) string {
	start, startDot, startPart, end, matchedSlash, preDotState := 0, -1, 0, -1, true, 0
	if f.win && hasDrive(p) {
		start, startPart = 2, 2
	}
	for i := len(p) - 1; i >= start; i-- {
		c := p[i]
		if f.isSep(c) {
			if !matchedSlash {
				startPart = i + 1
				break
			}
			continue
		}
		if end == -1 {
			matchedSlash, end = false, i+1
		}
		if c == '.' {
			if startDot == -1 {
				startDot = i
			} else if preDotState != 1 {
				preDotState = 1
			}
		} else if startDot != -1 {
			preDotState = -1
		}
	}
	if startDot == -1 || end == -1 || preDotState == 0 ||
		preDotState == 1 && startDot == end-1 && startDot == startPart+1 {
		return ""
	}
	return p[startDot:end]
}

// root is the root of p: "/" or "" with POSIX rules; with Windows rules
// also "C:", "C:\", or "\\server\share\".
func (f pathFlavor) root(p string) string {
	if !f.win {
		if strings.HasPrefix(p, "/") {
			return "/"
		}
		return ""
	}
	switch {
	case p == "":
		return ""
	case len(p) == 1:
		if f.isSep(p[0]) {
			return p
		}
		return ""
	case f.isSep(p[0]):
		rootEnd := 1
		if f.isSep(p[1]) {
			if _, end, ok := f.uncRoot(p); ok {
				rootEnd = end
				if end < len(p) {
					rootEnd = end + 1
				}
			}
		}
		return p[:rootEnd]
	case hasDrive(p):
		if len(p) > 2 && f.isSep(p[2]) {
			return p[:3]
		}
		return p[:2]
	}
	return ""
}

func init() {
	S, B := types.StringT, types.BoolT
	SA := types.NewArray(S)
	p := &Package{Path: "path", Qualifier: "path"}
	// Each function exists three times: with the platform's rules (join),
	// and with POSIX (posixJoin) or Windows (win32Join) rules.
	flavors := []struct {
		prefix string
		pick   func(env *Env) pathFlavor
	}{
		{"", func(env *Env) pathFlavor {
			if env.Windows {
				return win32Path
			}
			return posixPath
		}},
		{"posix", func(*Env) pathFlavor { return posixPath }},
		{"win32", func(*Env) pathFlavor { return win32Path }},
	}
	for _, fl := range flavors {
		pick := fl.pick
		name := func(n string) string {
			if fl.prefix == "" {
				return n
			}
			return fl.prefix + strings.ToUpper(n[:1]) + n[1:]
		}
		p.add(
			envFn(name("sep"), S, func(env *Env, a []any) any { return pick(env).sep }),
			envFn(name("delimiter"), S, func(env *Env, a []any) any {
				if pick(env).win {
					return ";"
				}
				return ":"
			}),
			envFn(name("join"), S, func(env *Env, a []any) any { return pick(env).join([]string{str(a[0]), str(a[1])}) }, S, S),
			envFn(name("joinAll"), S, func(env *Env, a []any) any { return pick(env).join(a[0].([]string)) }, SA),
			envFn(name("resolve"), S, func(env *Env, a []any) any { return pick(env).resolve(env, []string{str(a[0])}) }, S),
			envFn(name("resolveAll"), S, func(env *Env, a []any) any { return pick(env).resolve(env, a[0].([]string)) }, SA),
			envFn(name("normalize"), S, func(env *Env, a []any) any { return pick(env).normalize(str(a[0])) }, S),
			envFn(name("isAbsolute"), B, func(env *Env, a []any) any { return pick(env).isAbsolute(str(a[0])) }, S),
			envFn(name("relative"), S, func(env *Env, a []any) any { return pick(env).relative(env, str(a[0]), str(a[1])) }, S, S),
			envFn(name("dirname"), S, func(env *Env, a []any) any { return pick(env).dirname(str(a[0])) }, S),
			envFn(name("basename"), S, func(env *Env, a []any) any { return pick(env).basename(str(a[0]), "") }, S),
			envFn(name("extname"), S, func(env *Env, a []any) any { return pick(env).extname(str(a[0])) }, S),
			envFn(name("root"), S, func(env *Env, a []any) any { return pick(env).root(str(a[0])) }, S),
		)
		// basename(p, suffix) also drops a suffix: an overload, with a
		// JavaScript name of its own.
		bs := envFn(name("basename"), S, func(env *Env, a []any) any { return pick(env).basename(str(a[0]), str(a[1])) }, S, S)
		bs.JSName = name("basenameSuffix")
		p.add(bs)
	}
	register(p)
}
