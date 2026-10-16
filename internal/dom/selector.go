package dom

import "unicode/utf8"

// IsIDSelector reports whether sel is a single CSS id selector, such as
// "#app": '#' followed by a CSS identifier, with nothing else. renderApp
// accepts any selector, but warns about others, which may match several
// elements or none. A selector with CSS escapes counts as "other": the
// warning is conservative.
func IsIDSelector(sel string) bool {
	if len(sel) < 2 || sel[0] != '#' {
		return false
	}
	ident := sel[1:]
	for i, c := range ident {
		switch {
		case c == utf8.RuneError:
			return false
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_', c >= 0x80:
		case c == '-':
			// "-" alone, and "-" followed by a digit, are not identifiers.
			if i == 0 && (len(ident) == 1 || ident[1] >= '0' && ident[1] <= '9') {
				return false
			}
		case c >= '0' && c <= '9':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}
