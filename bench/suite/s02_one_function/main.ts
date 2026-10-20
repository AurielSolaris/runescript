// S02: one function and one line of output.
declare var process: any;

function add(a: number, b: number): number {
    return a + b;
}

process.stdout.write(add(2, 3) + "\n");
