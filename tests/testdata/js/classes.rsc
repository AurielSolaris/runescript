// Classes (OQ-CLASS-01): reference semantics, constructors, methods,
// `this`, identity, and evaluation order, the same on both targets.
package tests.js;

public struct Vec {
    int x;
    int y;
}

public int noisy(string what, int v) {
    trace("eval", what);
    return v;
}

public class Counter {
    private int count = noisy("count initializer", 0);
    public string label = "clicks";
    public Vec pos;
    public void(int) onChange = (int n) => trace("changed to", n);

    public Counter(int start) {
        this.count = noisy("constructor body", start);
    }

    public void add(int n) {
        this.count += n;
        this.pos.x += 1;
        this.onChange(this.count);
    }

    public int value() {
        return this.count;
    }
}

public class Node {
    public int value;
    public Array<Node> children;

    public Node(int value) {
        this.value = value;
    }

    public Node add(int value) {
        Node child = new Node(value);
        this.children.push(child);
        return child;
    }

    public int total() {
        int sum = this.value;
        for each (Node c in this.children) {
            sum += c.total();
        }
        return sum;
    }
}

// A field without a zero value gets its first value at the start of the
// constructor, before the object is used.
public class Segment {
    public Counter from;
    public Counter to;
    public string name = "segment";

    public Segment(Counter a, Counter b) {
        this.to = b;
        this.from = a;
        trace("segment made", this.from.value(), this.to.value());
    }

    public int length() {
        return this.to.value() - this.from.value();
    }
}

public class Guard {
    public int n = 1;

    public Guard(int n) {
        if (n < 0) {
            return; // the object keeps n = 1
        }
        this.n = n;
    }

    public int twice() {
        private int double(int x) {
            return x * 2; // a nested function returns its own value
        }
        return double(this.n);
    }
}

public struct Pair {
    Counter c;
    int tag;
}

public void bump(Counter c) {
    c.add(10);
}

protected int main() {
    Counter a = new Counter(1);
    Counter b = a;
    b.add(2);
    bump(a);
    trace(a.value(), a.label, a.pos.x, a.pos, a == b, a != b, a);
    a.onChange = (int n) => trace("now", n);
    a.add(1);
    a.onChange(-1);

    Counter c = new Counter(5);
    trace(a == c);
    Array<Counter> list = [a, c];
    list[1].add(1);
    list[1].pos.y = 7;
    trace(list.indexOf(c), list.contains(new Counter(0)), c.pos);

    Node root = new Node(1);
    Node leaf = root.add(2).add(3);
    Node four = root.add(4);
    trace(root.total(), root.children.length, root.children[0].children[0].value, leaf == root.children[0].children[0], four.value);

    Segment s = new Segment(a, c);
    trace(s.length(), s.name);
    s.from = c;
    trace(s.length());

    trace(new Guard(-5).n, new Guard(21).twice());

    Pair p = Pair{c: a, tag: 1};
    Pair q = Pair{c: a, tag: 1};
    Pair r = Pair{c: c, tag: 1};
    trace(p == q, p == r, p);
    p.c.add(1);
    trace(q.c.value());
    return 0;
}
