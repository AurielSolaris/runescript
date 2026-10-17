// Arrays (OQ-TYPE-02a): literals, indexing, length, push, pop, reference
// semantics, bounds errors, and std:string's split, fields, and join.
package tests.js;

use "std:string";

public int sum(Array<int> xs) {
    int total = 0;
    for (int i = 0; i < xs.length; i++) {
        total += xs[i];
    }
    return total;
}

public void addTwice(Array<string> list, string s) {
    list.push(s);
    list.push(s);
}

public Array<int> squares(int n) {
    Array<int> out = [];
    for (int i = 1; i <= n; i++) {
        out.push(i * i);
    }
    return out;
}

protected int main() {
    Array<string> names = ["ada", "grace"];
    names.push("linus");
    trace(names.length, names[0], names[2]);
    names[1] = "hopper";
    names[0] += "!";
    trace(string.join(names, ", "));

    Array<int> nums = [3, 4, 5];
    trace(sum(nums), sum(squares(4)), squares(3)[2]);
    nums[0] *= 10;
    nums[2] -= 1;
    trace(nums[0], nums[2], nums.pop(), nums.length);

    // Arrays are held by reference.
    Array<string> alias = names;
    addTwice(alias, "x");
    trace(names.length, names[4]);

    Array<uint> us = [1, 2];
    Array<float32> fs = [0.5, 1.5];
    Array<char> cs = ['a', 'é'];
    Array<bool> bs = [true];
    Array<Array<int>> grid = [[1, 2], [3]];
    grid[1].push(4);
    trace(us[1], fs[0] + fs[1], cs[1], bs[0], grid[1][1], grid.length);
    Array<int(int)> fns = [(int x) => x + 1, (int x) => x * 2];
    trace(fns[1](21), [10, 20, 30][1], ["a", "b"].length);

    Array<string> parts = string.split("a,b,,c", ",");
    trace(parts.length, $"[{string.join(parts, "|")}]");
    trace(string.join(string.fields("  one two\tthree\n "), "+"));
    trace(string.join(string.splitN("a-b-c-d", "-", 2), " "), string.splitN("x", "-", 0).length);
    trace(string.join(string.splitAfter("a,b,c", ","), " "), string.join(string.splitAfterN("a,b,c", ",", 2), " "));
    trace(string.join(string.split("héllo", ""), "."));
    trace(string.join(string.fieldsFunc("a1b22c333", (char c) => c >= '0' && c <= '9'), " "));
    trace(string.join([], "-") == "", string.split("", ",").length);

    try {
        trace(names[9]);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        names[-1] = "x";
    } catch (RangeError e) {
        trace(e.message);
    }
    Array<int> empty = [];
    try {
        trace(empty.pop());
    } catch (RangeError e) {
        trace(e.message);
    }
    // The value of a compound assignment may shorten the array.
    Array<int> shrink = [1];
    try {
        shrink[0] += shrink.pop();
    } catch (RangeError e) {
        trace(e.message, shrink.length);
    }
    return 0;
}
