// B01b: integer multiply and remainder throughput (plain TypeScript numbers).
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var acc = 1;
for (var i = 0; i < n; i++) {
    acc = (acc * 31 + i) % 1000003;
}
process.stdout.write(acc + "\n");
