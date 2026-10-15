package app;
use "./second.rsc";
use "./lib/start.rsc";

public int twice(int n); // expect: ASX-NAME-010 main.rsc:5:12

public int main() {
    return twice(1);
}
