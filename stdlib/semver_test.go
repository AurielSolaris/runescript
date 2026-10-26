package stdlib

import "testing"

// TestSemverCases checks std:semver on cases from the SemVer 2.0.0 spec
// and node-semver's documentation. TestSemverMatchesNodeSemver compares
// many more with node-semver itself.
func TestSemverCases(t *testing.T) {
	valid := map[string]bool{
		"0.0.0": true, "1.2.3": true, "10.20.30": true, "1.0.0-alpha": true, "1.0.0-alpha.1": true,
		"1.0.0-0.3.7": true, "1.0.0-x.7.z.92": true, "1.0.0-x-y-z.--": true, "1.0.0+20130313144700": true,
		"1.0.0-beta+exp.sha.5114f85": true, "1.0.0+21AF26D3----117B344092BD": true, "1.2.3-rc-01": true,
		"99999999999999999999999.999999999999999999.99999999999999999": true,
		"1.2": false, "1.2.3.4": false, "01.2.3": false, "1.02.3": false, "1.2.03": false, "v1.2.3": false,
		" 1.2.3": false, "1.2.3 ": false, "1.2.3-": false, "1.2.3+": false, "1.2.3-01": false, "1.2.3-rc.01": false,
		"1.2.3-a..b": false, "1.2.3-a_b": false, "1.2.3+a..b": false, "-1.2.3": false, "": false, "1.2.3-é": false,
	}
	for v, want := range valid {
		if _, ok := svParse(v); ok != want {
			t.Errorf("valid(%q) = %v, want %v", v, ok, want)
		}
	}
	// The SemVer 2.0.0 precedence example, lowest first.
	order := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2",
		"1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.9.0", "1.10.0", "2.0.0", "10.0.0", "99999999999999999999.0.0"}
	for i := range order {
		for j := range order {
			want := svSign(i - j)
			if got := svCompare(svMust(order[i]), svMust(order[j])); got != want {
				t.Errorf("compare(%s, %s) = %d, want %d", order[i], order[j], got, want)
			}
		}
	}
	if svCompare(svMust("1.0.0+a"), svMust("1.0.0+b")) != 0 {
		t.Error("build metadata must not affect precedence")
	}

	ranges := []struct {
		r, v string
		want bool
	}{
		{"^1.2.3", "1.9.9", true}, {"^1.2.3", "2.0.0", false}, {"^1.2.3", "1.2.2", false},
		{"^0.2.3", "0.2.9", true}, {"^0.2.3", "0.3.0", false}, {"^0.0.3", "0.0.4", false},
		{"^1.2.3-beta.2", "1.2.3-beta.4", true}, {"^1.2.3-beta.2", "1.2.4-beta.2", false},
		{"~1.2.3", "1.2.9", true}, {"~1.2.3", "1.3.0", false}, {"~1", "1.9.0", true}, {"~>1.2", "1.2.5", true},
		{"1.2.x", "1.2.7", true}, {"1.x", "1.0.0", true}, {"1.x", "2.0.0", false}, {"*", "0.0.1", true},
		{"", "3.0.0", true}, {"*", "1.0.0-rc.1", false}, {">1.2", "1.2.9", false}, {">1.2", "1.3.0", true},
		{"<=1.2", "1.2.99", true}, {"<1.2", "1.1.9", true}, {"<1.2", "1.2.0-rc", false},
		{"1.2.3 - 2.3.4", "2.3.4", true}, {"1.2.3 - 2.3", "2.3.9", true}, {"1.2.3 - 2", "2.9.9", true},
		{"1.2 - 2.3.4", "1.2.0", true}, {">=1.2.3 <2.0.0", "1.5.0", true}, {">= 1.2.3", "1.2.3", true},
		{"<1.0.0 || >=2.0.0", "1.5.0", false}, {"<1.0.0 || >=2.0.0", "2.1.0", true},
		{">1.2.3-alpha.3", "1.2.3-alpha.7", true}, {">1.2.3-alpha.3", "3.4.5-alpha.9", false},
		{"=1.2.3", "1.2.3+build", true}, {"1.2.3", "1.2.4", false}, {"<*", "0.0.0", false},
		{"^2.0.0-alpha || *", "2.0.0-beta", false}, {"^2.0.0-alpha || 3.x", "2.0.0-beta", true},
	}
	for _, c := range ranges {
		r, err := svParseRange(c.r)
		if err != "" {
			t.Errorf("range %q: %s", c.r, err)
			continue
		}
		if got := svSatisfies(svMust(c.v), r); got != c.want {
			t.Errorf("satisfies(%q, %q) = %v, want %v", c.v, c.r, got, c.want)
		}
	}
	for _, bad := range []string{"1.2.3 -", ">=", "v1.2.3", "1.2.3.4", ">=a", "~01", "1 - 2 - 3", "^1.2.3-01"} {
		if _, err := svParseRange(bad); err == "" {
			t.Errorf("range %q is accepted, want an error", bad)
		}
	}
	if f := fault(func() { svMust("1.2") }); f.Type != "ArgumentError" ||
		f.Message != `"1.2" is not a valid version: SemVer 2.0.0 is MAJOR.MINOR.PATCH, e.g. 1.2.3 or 1.2.3-rc.1` {
		t.Errorf("invalid version: %v", f)
	}
	if f := fault(func() { svInt("2147483648", "major") }); f.Type != "RangeError" {
		t.Errorf("major 2147483648: %v", f)
	}
	if svInt("2147483647", "major") != 2147483647 {
		t.Error("svInt(2147483647)")
	}
	if svInc("199") != "200" || svInc("99") != "100" || svInc("0") != "1" {
		t.Error("svInc")
	}
}
