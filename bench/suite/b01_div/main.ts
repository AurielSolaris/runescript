// B01c: integer division throughput (explicit truncation).
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var acc = 0;
for (var i = 0; i < n; i++) {
    acc = (acc + Math.floor((i * 7 + 3) / (i % 13 + 1))) % 1000000007;
}
process.stdout.write(acc + "\n");
