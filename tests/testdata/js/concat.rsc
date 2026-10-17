// String concatenation (+, +=) and interpolated strings ($"…{x}…").
package tests.js;

public string join3(string a, string b, string c) {
    return a + ", " + b + ", " + c;
}

protected int main() {
    string name = "héllo";
    string s = name + " wörld";
    s += " 🎉";
    trace(s);
    trace(join3("a", "b", "c"));

    string line = "";
    for (int i = 0; i < 5; i++) {
        line += (string) i;
    }
    trace(line);

    int n = -42;
    uint u = 4294967295;
    float f = 0.1;
    float32 g = (float32) 0.1;
    float64 h = 1e21;
    bool ok = n < 0;
    trace($"n={n} u={u} f={f} g={g} h={h} ok={ok}");
    trace($"sum={n + 2} nan={0.0 / 0.0} inf={1.0 / 0.0} neg={-1.0 / 0.0}");
    trace($"{name}{name}");
    trace($"{{literal}} {"{"}x{"}"} \"quoted\"\ttab");
    trace($"outer {$"inner {n}"} {name + "!"} done");
    trace($"no holes" + $"" + "");
    string(int) show = (int x) => $"<{x}>";
    trace(show(7), $"{show(8)}");
    return 0;
}
