package app;
use "1abc";                    // expect: ASX-USE-006 main.rsc:2:5
use "a..b";                    // expect: ASX-USE-006 main.rsc:3:5
use "if.then";                 // expect: ASX-USE-006 main.rsc:4:5
use "lib\\x.rsc";             // expect: ASX-USE-006 main.rsc:5:5
use "go:strings";              // expect: ASX-USE-003 main.rsc:6:5

public int main() {
    return 0;
}
