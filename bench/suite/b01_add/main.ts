// B01a: integer addition throughput (plain TypeScript numbers).
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var acc = 0;
for (var i = 0; i < n; i++) {
    acc = acc + i;
    if (acc > 1000000000) {
        acc = acc - 1000000000;
    }
}
process.stdout.write(acc + "\n");
