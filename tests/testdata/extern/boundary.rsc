package main;

// Every value returned by JavaScript is checked against its declared type.
// The helper, boundary.mjs, returns wrong values on purpose.

extern "./boundary.mjs" int intOf(int x);
extern "./boundary.mjs" int halfInt(int n);
extern "./boundary.mjs" int textThree();
extern "./boundary.mjs" int tooBig();
extern "./boundary.mjs" int negZero();
extern "./boundary.mjs" uint negUint();
extern "./boundary.mjs" float64 notANumber();
extern "./boundary.mjs" float32 tenth();
extern "./boundary.mjs" bool oneAsBool();
extern "./boundary.mjs" string greet(string name);
extern "./boundary.mjs" string loneSurrogate();
extern "./boundary.mjs" void voidOk();
extern "./boundary.mjs" void voidFive();
extern "./boundary.mjs" string nothing();
extern "./boundary.mjs" RangeError rangeError(string m);
extern "./boundary.mjs" RangeError typeErrorValue();
extern "./boundary.mjs" RangeError jsRangeError();
extern "./boundary.mjs" RangeError jsPlainError();
extern "./boundary.mjs" Error badName();
extern "./boundary.mjs" string describeError(Error e);
extern "./boundary.mjs" void throwsError();
extern "./boundary.mjs" void throwsArgumentError();
extern "./boundary.mjs" void throwsString();
extern "./boundary.mjs" void throwsBack(Error e);
extern "./boundary.mjs" void missing();

public int main() {
    // Values of the right type cross unchanged; float32 is rounded.
    trace(intOf(-7), halfInt(8), negZero(), notANumber(), tenth(), greet("rune"));
    voidOk();

    // Values of the wrong type throw a TypeError at the call.
    try { trace(halfInt(3)); } catch (TypeError e) { trace(e.message, e.errorID); }
    try { trace(textThree()); } catch (TypeError e) { trace(e.message); }
    try { trace(tooBig()); } catch (TypeError e) { trace(e.message); }
    try { trace(negUint()); } catch (TypeError e) { trace(e.message); }
    try { trace(oneAsBool()); } catch (TypeError e) { trace(e.message); }
    try { trace(loneSurrogate()); } catch (TypeError e) { trace(e.message); }
    try { voidFive(); } catch (TypeError e) { trace(e.message); }
    try { trace(nothing()); } catch (TypeError e) { trace(e.message); }

    // Errors cross both ways as RSError objects.
    RangeError r = rangeError("too far");
    trace(r.name, r.message, r.errorID);
    trace(describeError(new ArgumentError("arg", 3)));
    RangeError fromJS = jsRangeError();
    trace(fromJS.name, fromJS.message, fromJS.errorID);
    try { trace(typeErrorValue().message); } catch (TypeError e) { trace(e.message); }
    try { trace(jsPlainError().message); } catch (TypeError e) { trace(e.message); }
    try { trace(badName().message); } catch (TypeError e) { trace(e.message); }

    // JavaScript exceptions become RuneScript errors.
    try { throwsError(); } catch (Error e) { trace(e.name, e.message, e.errorID); }
    try { throwsArgumentError(); } catch (ArgumentError e) { trace(e.name, e.message, e.errorID); }
    try { throwsString(); } catch (Error e) { trace(e.message); }
    try { throwsBack(new RangeError("round trip", 9)); } catch (RangeError e) { trace(e.message, e.errorID); }
    try { missing(); } catch (Error e) { trace(e.message, e.errorID); }
    return 0;
}
