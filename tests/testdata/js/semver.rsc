// std:semver: SemVer 2.0.0 versions, npm-style precedence and ranges, and
// the errors each function throws, the same on both targets.
package tests.js;

use "std:semver";

public void attempt(string what, void() f) {
    try {
        f();
        trace(what, "ok");
    } catch (Error e) {
        trace(what, "->", e.name, e.message);
    }
}

public void check(string version, string range) {
    trace(version, "in", range, "->", semver.satisfies(version, range));
}

protected int main() {
    Array<string> versions = ["1.2.3", "1.2.3-rc.1", "1.2.3+build.5", "1.2", "01.2.3", "v1.2.3", "1.2.3-rc.01", "1.2.3-rc-01"];
    for each (string v in versions) {
        trace(v, semver.valid(v));
    }

    // The SemVer 2.0.0 precedence example, in order.
    Array<string> order = ["1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0"];
    for (int i = 0; i + 1 < order.length; i++) {
        trace(order[i], "<", order[i + 1], semver.lt(order[i], order[i + 1]), semver.compare(order[i + 1], order[i]));
    }
    trace(semver.compare("1.10.0", "1.9.0"));
    trace(semver.eq("1.0.0+a", "1.0.0+b"), semver.neq("1.0.0", "1.0.1"), semver.gt("2.0.0", "10.0.0"), semver.gte("1.0.0", "1.0.0"), semver.lte("1.0.0-rc.1", "1.0.0"));
    trace(semver.major("12.34.56-rc.1+b.7"), semver.minor("12.34.56"), semver.patch("12.34.56"), semver.prerelease("12.34.56-rc.1+b.7"), semver.build("12.34.56-rc.1+b.7"), semver.prerelease("1.0.0") == "");

    check("1.4.0", "^1.2.0");
    check("2.0.0", "^1.2.0");
    check("0.2.9", "^0.2.3");
    check("0.3.0", "^0.2.3");
    check("1.2.9", "~1.2.3");
    check("1.3.0", "~1.2.3");
    check("1.9.9", "1.x");
    check("1.5.0", ">=1.2.3 <2.0.0");
    check("1.5.0", "<1.0.0 || >=2.0.0");
    check("2.3.9", "1.2.3 - 2.3");
    check("1.2.4-beta", "^1.2.3");
    check("1.2.3-beta.4", "^1.2.3-beta.2");
    check("1.0.0-rc.1", "*");
    check("3.0.0", "");
    trace(semver.validRange(">= 1.2.3 || ~2"), semver.validRange("1.2.3 -"), semver.validRange("=v1"));

    attempt("invalid version", () => { trace(semver.compare("1.2", "1.0.0")); });
    attempt("invalid satisfies version", () => { trace(semver.satisfies("latest", "*")); });
    attempt("invalid range", () => { trace(semver.satisfies("1.0.0", ">=1.0.0 <")); });
    attempt("bad hyphen range", () => { trace(semver.satisfies("1.0.0", "1.0 - a")); });
    attempt("major too big", () => { trace(semver.major("2147483648.0.0")); });
    trace(semver.lt("99999999999999999999.0.0", "100000000000000000000.0.0"));
    return 0;
}
