# 24. Classes

A class describes objects: values with named fields that methods can read
and change. Unlike a [struct](19-structs.md), an object is a reference.
Assigning it or passing it to a function shares the same object, so a
change made through one name is seen through every other.

```actionscript
package main;

public class Counter {
    private int count;

    public Counter(int start) {
        this.count = start;
    }

    public void add(int n) {
        this.count += n;
    }

    public int value() {
        return this.count;
    }
}

protected int main() {
    Counter a = new Counter(1);
    Counter b = a;          // the same object
    b.add(2);
    trace(a.value());       // 3
    trace(a == b, a);       // true [object Counter]
    return 0;
}
```

This chapter covers the first version of classes: fields, methods, and one
constructor. Inheritance, interfaces, `static` members, and classes inside
classes are not part of it yet (see [the end of this chapter](#not-yet-available)).

## Declaring a class

A class is declared at the top level of a file with an access modifier,
like a struct or a function. Every member also starts with one:

- a **field**, `public int x;` or, with an initial value, `public int x = 1;`
- a **method**, written like a function: `public int area() { ... }`
- the **constructor**, named after the class and without a result type:
  `public Point(int x, int y) { ... }`

```actionscript
package main;

public class Rect {
    public int width;
    public int height = 1;
    private string label = "rect";

    public Rect(int width) {
        this.width = width;
    }

    public int area() {
        return this.width * this.height;
    }

    public string describe() {
        return $"{this.label} {this.width}x{this.height}";
    }
}

protected int main() {
    Rect r = new Rect(4);
    r.height = 3;
    trace(r.area(), r.describe());   // 12 rect 4x3
    return 0;
}
```

The modifiers mean what they mean everywhere else: `public` members can be
used everywhere, `protected` ones only in the class's package, and
`private` ones only in the class's file. So a function in the same file as
a class may use the class's private members; a file of another package
sees only the public ones.

A class has at most one constructor; methods and constructors cannot be
overloaded. A class without a constructor gets one that takes no
arguments and is as visible as the class. Fields and methods share one
set of names, so a field and a method can't have the same name.

## `this`

Inside a method or the constructor, `this` is the object the method was
called on. It is required: a field is always `this.x`, and a method of the
same object is always `this.m()`. A bare `x` is only ever a local variable
or a parameter, so a parameter may have the same name as a field:

```actionscript
package main;

public class Player {
    public string name;
    public int score;

    public Player(string name) {
        this.name = name;
    }

    public void win(int points) {
        this.score += points;
        this.announce();
    }

    private void announce() {
        trace(this.name, "has", this.score);
    }
}

protected int main() {
    Player p = new Player("Ada");
    p.win(10);   // Ada has 10
    p.win(5);    // Ada has 15
    return 0;
}
```

Nested functions and lambdas don't capture anything from the function
around them, and that includes `this`: pass the object as a parameter
instead.

## Creating objects

`new Name(arguments)` creates an object. First every field gets its
initial value, in the order the fields are declared: its initializer, or
its type's zero value (0, `""`, `false`, an empty array, ...). Then the
constructor's body runs. A field's initializer can't use `this`, and a
bare `return;` in the constructor ends it early, with the object as it is.

RuneScript has no null (see [chapter 17](17-no-null-any-generics.md)), so
every field always has a value. A field whose type has no zero value,
such as another class, an error, or a function type, needs an initializer,
or it must be set at the very start of the constructor, before `this` is
used for anything else:

```actionscript
package main;

public class Point {
    public int x;
    public int y;

    public Point(int x, int y) {
        this.x = x;
        this.y = y;
    }
}

public class Segment {
    public Point from;            // set at the start of the constructor
    public Point to;
    public string name = "segment";

    public Segment(Point a, Point b) {
        this.from = a;
        this.to = b;
        trace(this.name, this.from.x, this.to.x);
    }
}

protected int main() {
    Segment s = new Segment(new Point(0, 0), new Point(5, 2));   // segment 0 5
    s.to.x = 7;
    trace(s.to.x);                                               // 7
    return 0;
}
```

A local variable of a class type must be initialized too: `Point p;` is a
compile error, because there is no object to start with.

## References

Since objects are references, a function can change the object it is
given, and arrays of objects hold the objects themselves, not copies:

```actionscript
package main;

public class Item {
    public string name;
    public int count;

    public Item(string name) {
        this.name = name;
    }
}

public void restock(Item item) {
    item.count += 10;   // changes the caller's object
}

protected int main() {
    Item apples = new Item("apples");
    Array<Item> shelf = [apples, new Item("pears")];
    restock(apples);
    shelf[1].count = 3;
    for each (Item i in shelf) {
        trace(i.name, i.count);   // apples 10, then pears 3
    }
    trace(shelf.indexOf(apples));  // 0
    return 0;
}
```

`==` and `!=` compare identity: two objects are equal only if they are the
same object, even when every field matches. `indexOf` and `contains` on an
array of objects look for that same object. Objects can't be ordered with
`<`, so `sort()` on an array of objects needs a function:
`items.sort((Item a, Item b) => a.count < b.count)`.

A struct may hold an object, and an object may hold a struct. The struct
is still a value and the object still a reference: copying the struct
copies its fields, and the copy holds the same object.

## Showing objects

`trace` shows an object as `[object Name]`, as ActionScript 3 does. To
show its contents, write a method that returns a string. An object can't
be put in an interpolated string or converted with `(string)`.

## Not yet available

- Inheritance (`extends`), interfaces, and abstract classes.
- `static` fields and methods.
- Classes declared inside other classes or inside functions.
- Methods used as values (`p.area` without a call): wrap the call in a
  lambda, such as `() => p.area()`.
- Getters and setters, operator overloading, and generic classes.
- Objects crossing the JavaScript boundary of extern functions.

Next: back to the [guide index](README.md)
