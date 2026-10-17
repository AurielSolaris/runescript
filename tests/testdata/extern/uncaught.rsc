package main;

extern "./boundary.mjs" int halfInt(int n);

protected int main() {
    trace("before");
    trace(halfInt(5));
    trace("never printed");
    return 0;
}
