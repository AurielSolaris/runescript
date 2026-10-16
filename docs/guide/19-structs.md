# 19. Structs

A struct groups named values into one value of a new type:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public void main() {
    Point p = Point{x: 3, y: 4};
    trace(p, p.x + p.y);   // Point{x: 3, y: 4} 7
    p.y = 10;
    trace(p);              // Point{x: 3, y: 10}
}
```

## Declaring a struct

A struct is declared at the top level of a file, with an access modifier
like a function's, and a list of typed fields, each ending in `;`:

```actionscript
package main;

public struct Player {
    string name;
    int score;
    Array<string> badges;
}

private struct Settings {
    bool sound;
    float volume;
}

public void main() {
    Player p = Player{name: "Ada", score: 0, badges: []};
    Settings s = Settings{sound: true, volume: 0.5};
    trace(p.name, s.volume);   // Ada 0.5
}
```

- `public` structs can be used everywhere, `protected` ones only in their
  package, and `private` ones only in their file.
- A field can have any value type: numbers, `bool`, `string`, `char`,
  arrays, functions, errors, and other structs.
- Fields have no default values. That's why every literal must name them
  all.

## Building a struct

A **struct literal** names the struct, then gives every field a value,
with `name: value`:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public void main() {
    Point a = Point{x: 1, y: 2};
    Point b = Point{y: 2, x: 1};   // any order
    Point c = Point{x: 1};         // error: ASX-TYPE-026
}
```

Leaving a field out, giving one twice, or naming one the struct doesn't
have is `ASX-TYPE-026`. The values are computed in the order you write
them.

## Structs are values

Assigning a struct, passing it to a function, returning it, or putting it
in an array **copies** it. Changing the copy never changes the original:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public Point moved(Point p, int dx) {
    Point q = p;
    q.x += dx;
    return q;
}

public void main() {
    Point a = Point{x: 1, y: 1};
    Point b = moved(a, 5);
    trace(a, b);   // Point{x: 1, y: 1} Point{x: 6, y: 1}
}
```

This is the opposite of arrays, which are shared
([chapter 16](16-text-and-arrays.md)). An array *inside* a struct is still
an array, though: copying the struct shares the array.

A function's parameters are read-only, and so are a struct parameter's
fields: `p.x = 1` inside `moved` would be `ASX-NAME-006`. Copy it into a
local variable first, as above.

## Fields inside fields, and arrays of structs

Assignments reach through fields and array elements:

```actionscript
package main;

public struct Point {
    int x;
    int y;
}

public struct Box {
    Point min;
    Point max;
}

public void main() {
    Box box = Box{min: Point{x: 0, y: 0}, max: Point{x: 2, y: 2}};
    box.max.x = 5;

    Array<Box> boxes = [box, box];
    boxes[1].min.y = -3;
    trace(boxes[0].min, boxes[1].min, box.max.x);   // Point{x: 0, y: 0} Point{x: 0, y: -3} 5
}
```

## Comparing and printing

`==` and `!=` compare every field. A struct with an array, function, or
error field can't be compared; compare the fields you need instead.

`trace` shows a struct with its type and fields, with strings and chars in
quotes: `Player{name: "Ada", score: 3}`.

## Structs of other packages

Name another package's struct through its qualifier
([chapter 12](12-packages-and-files.md)): `geom.Point p =
geom.Point{x: 1, y: 2};`. The struct must be `public`.

## Using a struct for "no value"

A struct is the cleanest way to keep a value and its "is it there?" flag
together, the `false`-as-`null` pattern from
[chapter 17](17-no-null-any-generics.md):

```actionscript
package main;

use "std:string";

public struct MaybeInt {
    bool present;   // false means "no value"
    int value;
}

public MaybeInt parseAge(string text) {
    string digits = string.trimSpace(text);
    if (digits == "" || string.indexFunc(digits, (char c) => c < '0' || c > '9') >= 0) {
        return MaybeInt{present: false, value: 0};
    }
    return MaybeInt{present: true, value: (int) digits};
}

public void main() {
    MaybeInt age = parseAge(" 42 ");
    if (age.present) {
        trace("age", age.value);   // age 42
    }
    trace(parseAge("forty").present);   // false
}
```

## Not yet available

Methods, constructors, private fields, and passing structs to JavaScript
`extern` functions.

Next: back to the [guide index](README.md)
