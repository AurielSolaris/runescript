// B05: appends, sequential reads and writes, and random reads.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var a: number[] = [];
for (var i = 0; i < n; i++) {
    a.push(i % 1000);
}

var sum = 0;
for (var i = 0; i < n; i++) {
    sum = (sum + a[i]) % 1000003;
}

for (var i = 0; i < n; i++) {
    a[i] = (a[i] * 3 + 1) % 1000;
}

var seed = 7;
var picked = 0;
for (var i = 0; i < n; i++) {
    seed = seed * 16807 % 2147483647;
    picked = (picked + a[seed % n]) % 1000003;
}
process.stdout.write(a.length + " " + sum + " " + picked + "\n");
