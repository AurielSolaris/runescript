// Each file may have its own private function with the same name, in one
// package (both files) and across packages (lib).
package app;

use "./other.rsc";
use "./lib/lib.rsc";

protected int main() {
    trace(helper(), fromOther(), lib.fromLib());
    return 0;
}

private int helper() {
    return 1;
}
