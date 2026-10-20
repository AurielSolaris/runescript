// D02: TypeScript has one number type; nothing stops a negative "unsigned".
declare var process: any;
var signedValue: number = -10;
var unsignedValue: number = signedValue;
process.stdout.write(unsignedValue + "\n");
