package app;
use "lib";
use "./other.rsc";

protected int main() {
    trace(lib.open());
    trace(lib.shared());       // expect: ASX-NAME-011 main.rsc:7:15
    trace(lib.secret());       // expect: ASX-NAME-011 main.rsc:8:15
    trace(hidden(1));          // expect: ASX-NAME-011 main.rsc:9:11
    int(int) f = hidden;       // expect: ASX-NAME-011 main.rsc:10:18
    trace(lib.nothing());      // expect: ASX-NAME-008 main.rsc:11:15
    trace(near());
    return 0;
}

// A private function of another file may have the same name.
private int helper() {
    return 1;
}
