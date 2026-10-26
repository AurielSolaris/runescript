package stdlib

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSemverMatchesNodeSemver compares std:semver with npm's semver
// package on random versions and ranges. It needs Node.js and a copy of
// the package: set RUNESCRIPT_NODE_SEMVER to its directory (the one with
// package.json), e.g. after `npm install semver@7` somewhere. Otherwise it
// is skipped; TestSemverCases covers the same rules without Node.
func TestSemverMatchesNodeSemver(t *testing.T) {
	dir := os.Getenv("RUNESCRIPT_NODE_SEMVER")
	if dir == "" {
		t.Skip("RUNESCRIPT_NODE_SEMVER is not set")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not on PATH")
	}
	r := rand.New(rand.NewSource(11))
	var versions, ranges []string
	for i := 0; i < 600; i++ {
		versions = append(versions, randVersion(r))
	}
	for i := 0; i < 3000; i++ {
		ranges = append(ranges, randRange(r))
	}
	type result struct {
		Valid     []bool   `json:"valid"`
		Compare   []int    `json:"compare"`
		ValidR    []bool   `json:"validRange"`
		Satisfies []string `json:"satisfies"`
	}
	var want result
	var good []semVersion
	var goodText []string
	for _, v := range versions {
		p, ok := svParse(v)
		want.Valid = append(want.Valid, ok)
		if ok {
			good = append(good, p)
			goodText = append(goodText, v)
		}
	}
	for i := 0; i+1 < len(good); i++ {
		want.Compare = append(want.Compare, svCompare(good[i], good[i+1]))
	}
	for _, rg := range ranges {
		sets, e := svParseRange(rg)
		want.ValidR = append(want.ValidR, e == "")
		var b strings.Builder
		if e == "" {
			for _, v := range good {
				if svSatisfies(v, sets) {
					b.WriteByte('1')
				} else {
					b.WriteByte('0')
				}
			}
		}
		want.Satisfies = append(want.Satisfies, b.String())
	}

	input, _ := json.Marshal(map[string][]string{"versions": versions, "good": goodText, "ranges": ranges})
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.json")
	script := filepath.Join(tmp, "check.cjs")
	os.WriteFile(in, input, 0o644)
	os.WriteFile(script, []byte(fmt.Sprintf(`
const semver = require(%q);
const d = JSON.parse(require("fs").readFileSync(process.argv[2], "utf8"));
const out = { valid: [], compare: [], validRange: [], satisfies: [] };
for (const v of d.versions) out.valid.push(semver.parse(v) !== null);
for (let i = 0; i + 1 < d.good.length; i++) out.compare.push(semver.compare(d.good[i], d.good[i + 1]));
for (const r of d.ranges) {
  const ok = semver.validRange(r) !== null;
  out.validRange.push(ok);
  out.satisfies.push(ok ? d.good.map((v) => semver.satisfies(v, r) ? "1" : "0").join("") : "");
}
process.stdout.write(JSON.stringify(out));
`, dir)), 0o644)
	raw, err := exec.Command(node, script, in).Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var got result
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	fails := 0
	report := func(format string, args ...any) {
		fails++
		if fails <= 15 {
			t.Errorf(format, args...)
		}
	}
	for i, v := range versions {
		if got.Valid[i] != want.Valid[i] {
			report("valid(%q): node %v, Go %v", v, got.Valid[i], want.Valid[i])
		}
	}
	for i := range want.Compare {
		if got.Compare[i] != want.Compare[i] {
			report("compare(%q, %q): node %d, Go %d", goodText[i], goodText[i+1], got.Compare[i], want.Compare[i])
		}
	}
	for i, rg := range ranges {
		if got.ValidR[i] != want.ValidR[i] {
			report("validRange(%q): node %v, Go %v", rg, got.ValidR[i], want.ValidR[i])
			continue
		}
		if got.Satisfies[i] != want.Satisfies[i] {
			for j := range goodText {
				if got.Satisfies[i][j] != want.Satisfies[i][j] {
					report("satisfies(%q, %q): node %c, Go %c", goodText[j], rg, got.Satisfies[i][j], want.Satisfies[i][j])
					break
				}
			}
		}
	}
	if fails > 0 {
		t.Errorf("%d differences", fails)
	}
}

func randNum(r *rand.Rand) string {
	switch r.Intn(12) {
	case 0:
		return "01" // invalid: a leading zero
	case 1:
		return "10"
	}
	return fmt.Sprint(r.Intn(4))
}

var preIDs = []string{"alpha", "beta", "rc", "0", "1", "2", "11", "x-y", "-", "01", "a1", ""}

func randPre(r *rand.Rand) string {
	n := 1 + r.Intn(3)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = preIDs[r.Intn(len(preIDs))]
		if ids[i] == "" && r.Intn(3) > 0 {
			ids[i] = "alpha"
		}
	}
	return strings.Join(ids, ".")
}

func randVersion(r *rand.Rand) string {
	v := randNum(r) + "." + randNum(r) + "." + randNum(r)
	if r.Intn(10) == 0 {
		v = randNum(r) + "." + randNum(r) // too short
	}
	if r.Intn(2) == 0 {
		v += "-" + randPre(r)
	}
	if r.Intn(5) == 0 {
		v += "+build." + fmt.Sprint(r.Intn(9))
	}
	return v
}

var rangeOps = []string{"", "", "=", "<", "<=", ">", ">=", "~", "^", "~>"}

func randPartial(r *rand.Rand) string {
	part := func() string {
		switch r.Intn(6) {
		case 0:
			return []string{"x", "X", "*"}[r.Intn(3)]
		}
		return fmt.Sprint(r.Intn(3))
	}
	n := 1 + r.Intn(3)
	if r.Intn(2) == 0 {
		n = 3
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = part()
	}
	s := strings.Join(parts, ".")
	if n == 3 && r.Intn(3) == 0 {
		s += "-" + []string{"alpha", "beta.1", "0", "rc.2"}[r.Intn(4)]
	}
	return s
}

func randRange(r *rand.Rand) string {
	var sets []string
	for n := 1 + r.Intn(2); n > 0; n-- {
		if r.Intn(6) == 0 {
			sets = append(sets, randPartial(r)+" - "+randPartial(r))
			continue
		}
		var cs []string
		for k := 1 + r.Intn(2); k > 0; k-- {
			op := rangeOps[r.Intn(len(rangeOps))]
			sp := ""
			if op != "" && r.Intn(5) == 0 {
				sp = " "
			}
			cs = append(cs, op+sp+randPartial(r))
		}
		sets = append(sets, strings.Join(cs, " "))
	}
	return strings.Join(sets, " || ")
}
