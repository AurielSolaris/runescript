// for each (T x in a): nested arrays, break and continue, pushes during the
// loop, an array expression evaluated once, and the loop variable as a copy.
package tests.js;

public Array<int> numbers(Array<int> calls) {
    calls.push(1);
    return [10, 20, 30];
}

protected int main() {
    Array<Array<int>> grid = [[1, 2, 3], [4, 5, 6], [], [7]];
    int total = 0;
    for each (Array<int> row in grid) {
        for each (int v in row) {
            if (v == 2) {
                continue;
            }
            if (v == 6) {
                break;
            }
            total += v;
        }
    }
    trace("total", total);

    // Elements pushed during the loop are visited; popped ones are not.
    Array<string> words = ["a", "b"];
    for each (string w in words) {
        if (words.length < 4) {
            words.push(w + "!");
        }
        trace(w);
    }
    Array<int> shrink = [1, 2, 3, 4];
    for each (int v in shrink) {
        trace("saw", v, "popped", shrink.pop());
    }

    // The array expression is evaluated once.
    Array<int> calls = [];
    for each (int v in numbers(calls)) {
        trace(v);
    }
    trace("calls", calls.length);

    // Assigning the loop variable does not change the array, but the
    // elements are shared references when they are arrays.
    Array<int> xs = [1, 2];
    for each (int x in xs) {
        x = x * 100;
    }
    trace(xs[0], xs[1]);
    for each (Array<int> row in grid) {
        row.push(0);
    }
    trace(grid[2].length, grid[3][1]);

    // Building a nested array row by row.
    Array<Array<string>> table = [];
    for (int r = 0; r < 3; r++) {
        Array<string> row = [];
        for (int c = 0; c < 2; c++) {
            row.push($"{r}{c}");
        }
        table.push(row);
    }
    table[1][0] = "x";
    for each (Array<string> row in table) {
        string line = "";
        for each (string cell in row) {
            line += cell + " ";
        }
        trace(line);
    }
    return 0;
}
