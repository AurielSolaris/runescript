// JavaScript side of boundary.rsc. Some functions return values that do
// not match their RuneScript declarations, on purpose.
import { RSError } from "./prog.mjs";

export const intOf = (x) => x;
export const halfInt = (n) => n / 2;
export const textThree = () => "3";
export const tooBig = () => 2147483648;
export const negZero = () => -0;
export const negUint = () => -1;
export const notANumber = () => NaN;
export const tenth = () => 0.1;
export const oneAsBool = () => 1;
export const greet = (name) => `hello, ${name}`;
export const loneSurrogate = () => "a\uD800b";
export const voidOk = () => undefined;
export const voidFive = () => 5;
export const nothing = () => null;

export const rangeError = (m) => new RSError("RangeError", m, 42);
export const typeErrorValue = () => new RSError("TypeError", "wrong", 1);
export const jsRangeError = () => new RangeError("from JavaScript");
export const jsPlainError = () => new Error("plain");
export const badName = () => new RSError("Oops", "x", 1);

export const describeError = (e) =>
  `${e instanceof RSError} ${e.name} ${e.message} ${e.errorID}`;

export const throwsError = () => { throw new Error("disk on fire"); };
export const throwsArgumentError = () => { throw new RSError("ArgumentError", "bad argument", 5); };
export const throwsString = () => { throw "just a string"; };
export const throwsBack = (e) => { throw e; };
