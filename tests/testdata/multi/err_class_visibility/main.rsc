package main;

use "shapes";

protected int main() {
    shapes.Circle c = new shapes.Circle(2);
    trace(c.secret);
    // expect: ASX-NAME-011 main.rsc:7:13
    c.scale(2);
    // expect: ASX-NAME-011 main.rsc:9:7
    shapes.Hidden h = new shapes.Hidden();
    // expect: ASX-NAME-011 main.rsc:11:5
    // expect: ASX-NAME-011 main.rsc:11:27
    shapes.Locked l = new shapes.Locked();
    // expect: ASX-NAME-011 main.rsc:14:27
    shapes.Square s = new shapes.Square();
    // expect: ASX-NAME-011 main.rsc:16:5
    // expect: ASX-NAME-011 main.rsc:16:27
    return 0;
}
