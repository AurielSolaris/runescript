package main;

public struct Point {
    int x;
}

public struct Segment {
    Point start;
    public void move() {
        this.start.x = 2; // expect: ASX-NAME-006 main.rsc:10:9
    }
}

protected int main() { return 0; }
