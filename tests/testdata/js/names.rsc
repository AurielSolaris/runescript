// RuneScript identifiers that are reserved words or runtime names in
// JavaScript must still work on the JavaScript target.
package tests.js;

public int function(int this) {
    return this + 1;
}

public int Math(int x) {
    return x * 2;
}

public void console(string message) {
    trace("console:", message);
}

public int run(int(int) let) {
    return let(20);
}

public int RuneScriptError() {
    return 99;
}

public int main() {
    int var = 1;
    int new = function(var);
    int typeof = Math(new);
    int arguments = 3;
    int eval = arguments + typeof;
    int undefined = 0;
    int NaN = eval - undefined;
    console("still the RuneScript function");
    trace(var, new, typeof, eval, undefined, NaN);
    trace(run((int yield) => yield + 1), RuneScriptError());
    bool process = true;
    string globalThis = "ok";
    trace(process, globalThis);
    return 0;
}
