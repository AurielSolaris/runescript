// std:process: arguments, environment variables, the working directory,
// and the platform.
// args: alpha beta --gamma=1
package tests.env;

use "std:process";
use "std:string";

protected int main() {
    Array<string> args = process.args();
    trace(args.length, string.join(args, "|"));
    args.push("changed");                  // a copy: the next call is unchanged
    trace(process.args().length);
    trace(process.env("RUNESCRIPT_TEST"), "[" + process.env("RUNESCRIPT_SURELY_UNSET_VARIABLE") + "]");
    trace(process.hasEnv("RUNESCRIPT_TEST"), process.hasEnv("RUNESCRIPT_SURELY_UNSET_VARIABLE"));
    string p = process.platform();
    trace(p == "windows" || p == "linux" || p == "macos" || string.len(p) > 0);
    trace(string.len(process.cwd()) > 0);
    return 0;
}
