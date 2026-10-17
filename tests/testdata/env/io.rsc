// std:io: output without a newline, error and warning lines on standard
// error, and reading standard input by line, by byte, and all at once.
// stdin: first line\nsecond line\r\nAB\nrest\nof it
package tests.env;

use "std:io";

protected int main() {
    io.print("no newline, ");
    io.print("then ");
    trace("a trace");
    trace(io.atEnd());
    trace("[" + io.readLine() + "]");
    trace("[" + io.readLine() + "]");      // "\r\n" ends a line too
    trace(io.readByte(), io.readByte());   // 65 66
    trace("[" + io.readLine() + "]");      // what is left of the line
    trace("[" + io.readAll() + "]");
    trace(io.atEnd(), io.readByte(), "[" + io.readLine() + "]", "[" + io.readAll() + "]");
    io.error("an error line");
    io.warn("a warning line");
    io.print("ends without a newline");
    return 3;
}
