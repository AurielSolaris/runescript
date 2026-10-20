// D05: a function declared to return number returns a string.
declare var process: any;

function label(n: number): number {
    return "item " + n;
}

process.stdout.write(label(1) + "\n");
