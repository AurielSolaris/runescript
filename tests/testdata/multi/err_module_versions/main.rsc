package main;
use "geom";

protected int main() {
    trace(geom.version); // expect: ASX-USE-011 main.rsc:5:11
    trace(geom.version['missing']); // expect: ASX-USE-011 main.rsc:6:24
    return 0;
}
