package checker

import (
	"strings"
	"testing"

	"github.com/AurielSolaris/runescript/internal/diagnostics"
)

// classProgram is a file with the class declarations decls and main's
// body (which ends with `return 0;`).
func classProgram(decls, body string) string {
	return "package p;\n" + decls + "\nprotected int main() {\n" + body + "\n    return 0;\n}\n"
}

const pointClass = `public class Point {
    public int x;
    public int y = 1;
    private string tag = "p";
    public Point(int x, int y) {
        this.x = x;
        this.y = y;
    }
    public int sum() {
        return this.x + this.y;
    }
    public void move(int dx) {
        this.x += dx;
        this.log();
    }
    private void log() {
        trace(this.tag);
    }
}
`

func TestClassesAccepted(t *testing.T) {
	cases := map[string]string{
		"basics": classProgram(pointClass, `    Point p = new Point(1, 2);
    p.move(3);
    trace(p.sum(), p.x, p);`),
		"implicit constructor": classProgram("public class Bag {\n    public Array<int> items;\n    public int count() { return this.items.length; }\n}\n",
			"    Bag b = new Bag();\n    b.items.push(1);\n    trace(b.count());"),
		"reference parameter": classProgram(pointClass+"public void reset(Point p) {\n    p.x = 0;\n    p.y += 1;\n}\n",
			"    Point p = new Point(1, 2);\n    reset(p);\n    trace(p.x);"),
		"identity": classProgram(pointClass, "    Point a = new Point(1, 2);\n    Point b = a;\n    trace(a == b, a != new Point(1, 2));"),
		"arrays of objects": classProgram(pointClass, `    Array<Point> ps = [new Point(1, 2)];
    ps.push(new Point(3, 4));
    ps[1].x = 9;
    ps[0].move(1);
    trace(ps.indexOf(ps[1]), ps.contains(ps[0]));
    for each (Point q in ps) {
        trace(q.sum());
    }`),
		"field of a returned object": classProgram(pointClass+"public Point origin() { return new Point(0, 0); }\n",
			"    origin().x = 5;\n    trace(origin().sum());"),
		"struct field in an object": classProgram("public struct V { int x; int y; }\npublic class Body {\n    public V pos;\n    public void step() { this.pos.x += 1; }\n}\n",
			"    Body b = new Body();\n    b.step();\n    b.pos.y = 2;\n    trace(b.pos);"),
		"object field in a struct parameter": classProgram(pointClass+"public struct Holder { Point p; }\npublic void bump(Holder h) { h.p.x += 1; }\n",
			"    Holder h = Holder{p: new Point(1, 1)};\n    bump(h);\n    trace(h.p.x, h == h);"),
		"function field": classProgram("public class Btn {\n    public void() onClick = () => trace(\"hi\");\n    public void click() { this.onClick(); }\n}\n",
			"    Btn b = new Btn();\n    b.click();\n    b.onClick();\n    b.onClick = () => trace(\"bye\");"),
		"object field": classProgram("public class Node {\n    public int v;\n    public Array<Node> next;\n    public Node(int v) { this.v = v; }\n}\n",
			"    Node n = new Node(1);\n    n.next.push(new Node(2));\n    trace(n.next[0].v);"),
		"constructor returns early": classProgram("public class C {\n    public int n;\n    public C(int n) {\n        if (n < 0) {\n            return;\n        }\n        this.n = n;\n    }\n}\n",
			"    trace(new C(-1).n);"),
		"method calls method on another object": classProgram(pointClass+"public class Line {\n    public Point a;\n    public Point b;\n    public Line(Point a, Point b) { this.a = a; this.b = b; }\n    public int len() { return this.b.sum() - this.a.sum(); }\n}\n",
			"    Line l = new Line(new Point(0, 0), new Point(1, 1));\n    trace(l.len());"),
		"this as an argument": classProgram("public class C {\n    public int n;\n    public void register(Array<C> all) { all.push(this); }\n}\n",
			"    Array<C> all = [];\n    new C().register(all);\n    trace(all.length);"),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := check(t, src); len(diags) != 0 {
				t.Errorf("unexpected diagnostics: %v", diags)
			}
		})
	}
}

func TestClassesRejected(t *testing.T) {
	cases := []struct {
		name, src string
		code      diagnostics.Code
		msg       string // in the message or help
	}{
		{"this outside a class", classProgram("", "    trace(this);"), diagnostics.NameUndefined, "methods and the constructor"},
		{"this in a lambda", classProgram("public class C {\n    public int n;\n    public void f() {\n        int() g = () => this.n;\n    }\n}\n", ""), diagnostics.NameAcrossBoundary, "pass the object"},
		{"this in a nested function", classProgram("public class C {\n    public int n;\n    public void f() {\n        private int g() { return this.n; }\n    }\n}\n", ""), diagnostics.NameAcrossBoundary, "nested function"},
		{"this in a field initializer", classProgram("public class C {\n    public int n = 1;\n    public int m = this.n;\n}\n", ""), diagnostics.NameUndefined, "'this'"},
		{"bare field name", classProgram("public class C {\n    public int n;\n    public int f() { return n; }\n}\n", ""), diagnostics.NameUndefined, "'n'"},
		{"bare method name", classProgram("public class C {\n    public int n;\n    public int f() { return 1; }\n    public int g() { return f(); }\n}\n", ""), diagnostics.NameUndefined, "'f'"},
		{"no zero value", classProgram("public class C {\n    public Error e;\n}\n", ""), diagnostics.TypeNoZeroValue, "initial value"},
		{"object without initializer", classProgram(pointClass, "    Point p;"), diagnostics.TypeNoZeroValue, "new Point"},
		{"field of object type without initializer", classProgram(pointClass+"public class Seg {\n    public Point a;\n}\n", ""), diagnostics.TypeNoZeroValue, "no zero value"},
		{"first value after a use of this", classProgram(pointClass+"public class Seg {\n    public Point a;\n    public Seg(Point p) {\n        trace(this);\n        this.a = p;\n    }\n}\n", ""), diagnostics.TypeNoZeroValue, "before any other use of this"},
		{"first value in an if", classProgram(pointClass+"public class Seg {\n    public Point a;\n    public Seg(Point p) {\n        if (true) {\n            this.a = p;\n        }\n    }\n}\n", ""), diagnostics.TypeNoZeroValue, "field 'a'"},
		{"first value uses this", classProgram(pointClass+"public Point copy(Seg s) { return new Point(0, 0); }\npublic class Seg {\n    public Point a;\n    public Seg() {\n        this.a = copy(this);\n    }\n}\n", ""), diagnostics.TypeNoZeroValue, "field 'a'"},
		{"first value by compound assignment", classProgram("public class Seg {\n    public Error e;\n    public Seg() {\n        this.e += 1;\n    }\n}\n", ""), diagnostics.TypeNoZeroValue, "field 'e'"},
		{"unknown field", classProgram(pointClass, "    trace(new Point(1, 2).z);"), diagnostics.TypeNoField, "its fields are x, y, tag"},
		{"unknown method", classProgram(pointClass, "    new Point(1, 2).jump();"), diagnostics.TypeNoField, "sum()"},
		{"method as a value", classProgram(pointClass, "    int() f = new Point(1, 2).sum;"), diagnostics.NameNotAValue, "can only be called"},
		{"wrong constructor arguments", classProgram(pointClass, "    Point p = new Point(1);"), diagnostics.TypeArgCount, "expects 2"},
		{"constructor argument type", classProgram(pointClass, "    Point p = new Point(1, \"2\");"), diagnostics.TypeMismatch, ""},
		{"implicit constructor takes nothing", classProgram("public class C {\n    public int n;\n}\n", "    C c = new C(1);"), diagnostics.TypeArgCount, "expects 0"},
		{"two constructors", classProgram("public class C {\n    public C() {}\n    public C(int n) {}\n}\n", ""), diagnostics.NameRedeclared, "at most one constructor"},
		{"duplicate member", classProgram("public class C {\n    public int n;\n    public void n() {}\n}\n", ""), diagnostics.NameRedeclared, "already has a member 'n'"},
		{"overloaded method", classProgram("public class C {\n    public void f() {}\n    public void f(int x) {}\n}\n", ""), diagnostics.NameRedeclared, "cannot be overloaded"},
		{"method named like the class", classProgram("public class C {\n    public int C() { return 1; }\n}\n", ""), diagnostics.NameRedeclared, "constructor has no result type"},
		{"constructor returns a value", classProgram("public class C {\n    public C() { return 1; }\n}\n", ""), diagnostics.TypeBadReturn, "constructor"},
		{"class as a value", classProgram(pointClass, "    trace(Point);"), diagnostics.NameNotAValue, "new Point(...)"},
		{"class called like a function", classProgram(pointClass, "    Point p = Point(1, 2);"), diagnostics.NameNotCallable, "new Point(...)"},
		{"class as a struct literal", classProgram(pointClass, "    Point p = Point{x: 1, y: 2};"), diagnostics.TypeStructLit, "is a class"},
		{"new on a struct", classProgram("public struct S { int x; }\n", "    S s = new S();"), diagnostics.TypeBadNew, "S{x: ...}"},
		{"compare different classes", classProgram(pointClass+"public class Q {\n    public int n;\n}\n", "    trace(new Point(1, 2) == new Q());"), diagnostics.TypeBadOperand, "compares"},
		{"order objects", classProgram(pointClass, "    trace(new Point(1, 2) < new Point(1, 2));"), diagnostics.TypeBadOperand, ""},
		{"object in a string", classProgram(pointClass, "    string s = $\"{new Point(1, 2)}\";"), diagnostics.TypeBadConversion, "cannot be shown"},
		{"assign to this", classProgram("public class C {\n    public int n;\n    public void f(C other) { this.n = other.n; }\n}\n", ""), "", ""},
		{"void field", classProgram("public class C {\n    public void v;\n}\n", ""), diagnostics.TypeVoidValue, "void"},
		{"rune field", "package p;\nuse \"rune:dom\";\npublic class C {\n    public Signal<int> s = signal(0);\n}\nprotected int main() { return 0; }\n", diagnostics.TypeMismatch, "cannot hold"},
		{"redeclared class", classProgram("public class C {}\npublic struct C { int x; }\n", ""), diagnostics.NameRedeclared, "already declared"},
		{"function named like a class", classProgram("public class C {}\npublic void C() {}\n", ""), diagnostics.NameRedeclared, "as a class"},
		{"sort objects without a function", classProgram(pointClass, "    Array<Point> ps = [];\n    ps.sort();"), diagnostics.TypeBadOperand, "pass a function"},
		{"field of a non-function called", classProgram(pointClass, "    new Point(1, 2).x();"), diagnostics.TypeNotCallable, "not a function type"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := check(t, c.src)
			if c.code == "" {
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics: %v", diags)
				}
				return
			}
			for _, d := range diags {
				if d.Code == c.code && strings.Contains(d.Message+" "+d.Help, c.msg) {
					return
				}
			}
			t.Errorf("got %v, want %s with %q", diags, c.code, c.msg)
		})
	}
}

// Member visibility means what it means at the top level (owner decision,
// 2026-11-01): private is this file, protected is this package, public is
// everywhere. Other files use classes through the multi-file tests in
// tests/testdata/multi; here, a private member is usable anywhere in its
// own file, including outside the class.
func TestClassPrivateIsFileScope(t *testing.T) {
	src := classProgram(pointClass+"public string peek(Point p) { p.log(); return p.tag; }\n", "    trace(peek(new Point(1, 2)));")
	if diags := check(t, src); len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diags)
	}
}
