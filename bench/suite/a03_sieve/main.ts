// A03: Sieve of Eratosthenes; counts the primes up to n.
declare var process: any;

var n: number = parseInt(process.argv[2], 10);
var composite: boolean[] = [];
for (var i = 0; i <= n; i++) {
    composite.push(false);
}
var count = 0;
var last = 0;
for (var i = 2; i <= n; i++) {
    if (composite[i]) {
        continue;
    }
    count++;
    last = i;
    // The same guard as the RuneScript version (where i * i would overflow).
    if (i <= Math.floor(n / i)) {
        for (var j = i * i; j <= n; j += i) {
            composite[j] = true;
        }
    }
}
process.stdout.write(count + " " + last + "\n");
