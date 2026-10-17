// A package directory, a file with a non-.rsc extension, an alias, a file
// of the entry's own package, and a diamond (util is used twice).
package app;

use "geom";
use "./util/text.txt" as t;
use "./helpers.rsc";

protected int main() {
    trace(geom.area(3, 4), geom.perimeter(3, 4));
    trace(t.shout("hi"));
    trace(twice(21));
    return 0;
}
