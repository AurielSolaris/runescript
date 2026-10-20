// D04: a call with too few arguments, then one with a wrong type.
declare var process: any;

function area(width: number, height: number): number {
    return width * height;
}

process.stdout.write(area(3) + "\n");
process.stdout.write(area(3, "4") + "\n");
