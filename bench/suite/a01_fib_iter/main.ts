// A01: iterative Fibonacci, modulo 1,000,003.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var a = 0;
var b = 1;
for (var i = 0; i < n; i++) {
    var next = (a + b) % 1000003;
    a = b;
    b = next;
}
process.stdout.write(a + "\n");
