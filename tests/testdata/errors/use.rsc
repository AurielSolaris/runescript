// `use` declarations: only std: packages exist so far.
package tests.errors;

use "std:string";
use "std:string";          // expect: ASX-USE-002 5:1
use "std:nope";            // expect: ASX-USE-001 6:5
use "lib/helpers.rsc";      // expect: ASX-USE-003 7:5
use "go:strings";          // expect: ASX-USE-003 8:5

public int main() {
    trace(string.ToUpper("a"));        // expect: ASX-NAME-008 11:18
    trace(math.abs(1));                // expect: ASX-NAME-007 12:11
    uint n = string.count("aa", "a");  // expect: ASX-TYPE-004 13:14
    trace(string.repeat("a"));         // expect: ASX-TYPE-003 14:11
    return 0;
}
