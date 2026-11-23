package main;
use "geom" as shapes;
use "./single.rsm";
use "std:semver";

protected int main() {
    trace(shapes.version['area'], shapes.version["p"]);
    trace(shapes.version['p'] == "2.1.0-rc.1+build.005");
    trace(single.version, semver.lt(shapes.version['area'], shapes.version['p']));
    return 0;
}
