package app;
use "./missing.rsc";        // expect: ASX-USE-004 main.rsc:2:5
use "nopkg" as n;           // expect: ASX-USE-004 main.rsc:3:5
use "emptypkg";             // expect: ASX-USE-004 main.rsc:4:5
use "./emptypkg/";          // expect: ASX-USE-004 main.rsc:5:5

protected int main() {
    n.f();
    emptypkg.g();
    return 0;
}
