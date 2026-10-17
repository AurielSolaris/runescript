package app;
use "one.util";
use "two.util";                // expect: ASX-USE-008 main.rsc:3:5
use "two.util" as u2;          // an alias resolves the clash
use "one.util";                // expect: ASX-USE-002 main.rsc:5:1
use "./str.rsc";
use "std:string" as util;      // expect: ASX-USE-008 main.rsc:7:21

protected int main() {
    trace(util.f());
    return 0;
}
