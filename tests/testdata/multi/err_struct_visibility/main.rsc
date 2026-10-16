package main;

use "shapes";

public void main() {
    // The type and the literal are each reported.
    shapes.Grid g = shapes.Grid{step: 1};
    // expect: ASX-NAME-011 main.rsc:7:5
    // expect: ASX-NAME-011 main.rsc:7:28
    shapes.Missing m = shapes.Missing{};
    // expect: ASX-NAME-008 main.rsc:10:5
    // expect: ASX-NAME-008 main.rsc:10:31
}
