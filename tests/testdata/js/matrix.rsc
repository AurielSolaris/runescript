// std:matrix: building, arithmetic, products, elimination, formatting,
// and the errors each function throws, the same on both targets (results
// are bit-identical).
package tests.js;

use "std:matrix";

public void attempt(string what, void() f) {
    try {
        f();
        trace(what, "ok");
    } catch (Error e) {
        trace(what, "->", e.name, e.message);
    }
}

protected int main() {
    Array<Array<float64>> a = [[1, 2], [3, 4]];
    Array<Array<float64>> b = [[0.5, -1], [2, 0.25]];
    trace(matrix.format(matrix.zeros(2, 3)));
    trace(matrix.format(matrix.identity(3)));
    trace(matrix.rows(a), matrix.cols(a), matrix.rows(matrix.zeros(0, 0)), matrix.cols([[1.0, 2.0, 3.0]]));
    trace(matrix.format(matrix.add(a, b)));
    trace(matrix.format(matrix.sub(a, b)));
    trace(matrix.format(matrix.scale(a, 0.1)));
    trace(matrix.format(matrix.mul(a, b)));
    trace(matrix.format(matrix.mul([[1, 2, 3]], [[1], [2], [3]])));
    Array<float64> v = matrix.mulVec(a, [1.0, -1.0]);
    trace(v[0], v[1]);
    trace(matrix.format(matrix.transpose([[1, 2, 3], [4, 5, 6]])));

    Array<Array<float64>> m = [[2, -1, 0], [-1, 2, -1], [0, -1, 2]];
    trace("det", matrix.det(m), matrix.det(a), matrix.det([[1, 2], [2, 4]]), matrix.det(matrix.zeros(0, 0)));
    trace(matrix.format(matrix.inverse(m)));
    trace(matrix.format(matrix.mul(m, matrix.inverse(m))));
    Array<float64> x = matrix.solve(m, [1.0, 0.0, 1.0]);
    trace("solve", x[0], x[1], x[2]);
    Array<float64> y = matrix.solve([[0.1, 0.2], [0.3, 0.4]], [0.5, 0.6]);
    trace("solve", y[0], y[1]);

    // The arguments are never changed.
    trace(matrix.format(a));
    float64 tenth = 0.1;
    trace(matrix.format([[-0.0, 1e21, 1e-7, tenth + 0.2]]));

    attempt("not rectangular", () => { trace(matrix.det([[1, 2], [3]])); });
    attempt("add sizes", () => { trace(matrix.add([[1, 2]], [[1], [2]]).length); });
    attempt("sub sizes", () => { trace(matrix.sub([[1, 2]], [[1]]).length); });
    attempt("mul sizes", () => { trace(matrix.mul([[1, 2]], [[1, 2]]).length); });
    attempt("mulVec sizes", () => { trace(matrix.mulVec([[1, 2]], [1.0]).length); });
    attempt("not square", () => { trace(matrix.inverse([[1, 2, 3]]).length); });
    attempt("singular inverse", () => { trace(matrix.inverse([[1, 2], [2, 4]]).length); });
    attempt("singular solve", () => { trace(matrix.solve([[0, 0], [0, 0]], [1.0, 1.0]).length); });
    attempt("solve sizes", () => { trace(matrix.solve([[1]], [1.0, 2.0]).length); });
    attempt("negative size", () => { trace(matrix.zeros(-1, 2).length); });
    attempt("negative identity", () => { trace(matrix.identity(-3).length); });
    return 0;
}
