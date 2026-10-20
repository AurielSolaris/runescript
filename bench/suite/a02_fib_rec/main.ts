// A02: naive recursive Fibonacci.
declare var process: any;

function fib(n: number): number {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

process.stdout.write(fib(parseInt(process.argv[2], 10)) + "\n");
