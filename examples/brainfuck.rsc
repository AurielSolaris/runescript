// A Brainfuck interpreter with 32,768 cells.
//
//   runescript run examples/brainfuck.rsc examples/bf/hello.bf
//   echo "Hello, World!" | runescript run examples/brainfuck.rsc examples/bf/rot13.bf
//   runescript run examples/brainfuck.rsc -e "++++++++[>++++++++<-]>+."
//
// Without arguments it runs a built-in Hello World.
//
// The tape has 32,768 cells of 8 bits, which wrap around (255 + 1 is 0).
// Moving off either end of the tape is an error. ',' reads one byte of
// standard input, and leaves the cell unchanged at the end of the input.
// '.' writes one byte; bytes that form UTF-8 characters print as those
// characters.
//
// The program is compiled first: runs of + and - (and of > and <) become
// one operation, [-] and [+] become "set to zero", and every bracket
// learns where its partner is, so an unmatched one is reported, with its
// line and column, before anything runs.
//
// Shows std:fs, std:io, and std:process, structs, Array<T>, uint8,
// char, and errors.
package examples.brainfuck;

use "std:fs";
use "std:io";
use "std:process";
use "std:string";

// One compiled operation.
public struct Op {
    char kind;    // '+', '>', '.', ',', '[', ']', or 'z' (set to zero)
    int arg;      // + : the amount (0 to 255); > : the distance, possibly
                  // negative; [ and ] : the index of the partner bracket
    int line;     // where it starts in the source
    int col;
}

// compile turns source into operations. An unmatched bracket throws an
// ArgumentError whose message starts with its line and column.
public Array<Op> compile(string source) {
    Array<Op> ops;
    Array<int> open;               // indexes of the [ not yet closed
    Array<string> chars = string.split(source, "");
    int line = 1;
    int col = 0;
    int i = 0;
    while (i < chars.length) {
        char c = (char) chars[i];
        col++;
        if (c == '\n') {
            line++;
            col = 0;
            i++;
            continue;
        }
        if (c == '+' || c == '-' || c == '>' || c == '<') {
            // A run of + and - (or > and <) is one operation.
            bool cells = c == '+' || c == '-';
            int startCol = col;
            int sum = 0;
            while (i < chars.length) {
                char d = (char) chars[i];
                if (cells && d == '+' || !cells && d == '>') {
                    sum++;
                } else if (cells && d == '-' || !cells && d == '<') {
                    sum--;
                } else {
                    break;
                }
                i++;
                col++;
            }
            col--;
            if (cells) {
                ops.push(Op{kind: '+', arg: (sum % 256 + 256) % 256, line: line, col: startCol});
            } else {
                ops.push(Op{kind: '>', arg: sum, line: line, col: startCol});
            }
            continue;
        }
        if (c == '[' && i + 2 < chars.length && (chars[i + 1] == "-" || chars[i + 1] == "+") && chars[i + 2] == "]") {
            ops.push(Op{kind: 'z', arg: 0, line: line, col: col});
            i += 3;
            col += 2;
            continue;
        }
        if (c == '[') {
            open.push(ops.length);
            ops.push(Op{kind: '[', arg: -1, line: line, col: col});
        } else if (c == ']') {
            if (open.length == 0) {
                throw new ArgumentError($"{line}:{col}: this ']' has no matching '['");
            }
            int start = open.pop();
            ops[start].arg = ops.length;
            ops.push(Op{kind: ']', arg: start, line: line, col: col});
        } else if (c == '.' || c == ',') {
            ops.push(Op{kind: c, arg: 0, line: line, col: col});
        }
        // Every other character is a comment.
        i++;
    }
    if (open.length > 0) {
        Op unclosed = ops[open.pop()];
        throw new ArgumentError($"{unclosed.line}:{unclosed.col}: this '[' has no matching ']'");
    }
    return ops;
}

// ---- output ----

// flush prints the bytes in out that form whole UTF-8 characters, and
// returns the bytes of a character that is not finished yet (none if done
// is set). Bytes that are not UTF-8 print as U+FFFD.
public Array<int> flush(Array<int> out, bool done) {
    string text = "";
    int i = 0;
    while (i < out.length) {
        int b = out[i];
        if (b < 0x80) {
            text += (string) (char) b;
            i++;
            continue;
        }
        int need = 0;
        int cp = 0;
        int least = 0;             // the smallest code point of that length
        if (b >= 0xC2 && b <= 0xDF) {
            need = 1;
            cp = b & 0x1F;
            least = 0x80;
        } else if (b >= 0xE0 && b <= 0xEF) {
            need = 2;
            cp = b & 0x0F;
            least = 0x800;
        } else if (b >= 0xF0 && b <= 0xF4) {
            need = 3;
            cp = b & 0x07;
            least = 0x10000;
        } else {
            text += "\u{FFFD}";
            i++;
            continue;
        }
        if (i + need >= out.length && !done) {
            // Maybe an unfinished character: wait for the rest of it,
            // unless what is there already cannot be one.
            bool possible = true;
            for (int j = i + 1; j < out.length; j++) {
                if (out[j] < 0x80 || out[j] > 0xBF) {
                    possible = false;
                }
            }
            if (possible) {
                break;
            }
        }
        int j = i + 1;
        while (j <= i + need && j < out.length && out[j] >= 0x80 && out[j] <= 0xBF) {
            cp = cp << 6 | out[j] & 0x3F;
            j++;
        }
        if (j == i + need + 1 && cp >= least && cp <= 0x10FFFF && (cp < 0xD800 || cp > 0xDFFF)) {
            text += (string) (char) cp;
            i = j;
        } else {
            text += "\u{FFFD}";
            i++;
        }
    }
    io.print(text);
    Array<int> rest;
    for (int k = i; k < out.length; k++) {
        rest.push(out[k]);
    }
    return rest;
}

// ---- running ----

// run executes ops on a tape of 32,768 cells, and returns the exit code.
public int run(Array<Op> ops, string name) {
    int cells = 32768;
    Array<uint8> tape;
    for (int i = 0; i < cells; i++) {
        tape.push(0);
    }
    Array<int> out;            // bytes written but not printed yet
    int ptr = 0;
    int pc = 0;
    while (pc < ops.length) {
        Op op = ops[pc];
        if (op.kind == '+') {
            tape[ptr] = (uint8) (((int) tape[ptr] + op.arg) % 256);
        } else if (op.kind == '>') {
            ptr += op.arg;
            if (ptr < 0 || ptr >= cells) {
                out = flush(out, true);
                io.error($"{name}:{op.line}:{op.col}: the data pointer moved off the tape (to cell {ptr}; the cells are 0 to {cells - 1})");
                return 1;
            }
        } else if (op.kind == 'z') {
            tape[ptr] = 0;
        } else if (op.kind == '[') {
            if (tape[ptr] == 0) {
                pc = op.arg;
            }
        } else if (op.kind == ']') {
            if (tape[ptr] != 0) {
                pc = op.arg;
            }
        } else if (op.kind == '.') {
            out.push((int) tape[ptr]);
            if (tape[ptr] == 10 || out.length >= 4096) {
                out = flush(out, false);
            }
        } else {
            // ',': show any prompt first, then read a byte.
            out = flush(out, false);
            int b = io.readByte();
            if (b >= 0) {
                tape[ptr] = (uint8) b;
            }
        }
        pc++;
    }
    out = flush(out, true);
    return 0;
}

protected int main() {
    Array<string> args = process.args();
    string name = "hello";
    string source = "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.";
    if (args.length == 2 && args[0] == "-e") {
        name = "-e";
        source = args[1];
    } else if (args.length == 1 && args[0] != "-e") {
        name = args[0];
        try {
            source = fs.readFile(name);
        } catch (Error e) {
            io.error(e.message);
            return 1;
        }
    } else if (args.length != 0) {
        io.error("usage: brainfuck.rsc [<file.bf> | -e <code>]");
        return 2;
    }
    try {
        return run(compile(source), name);
    } catch (ArgumentError e) {
        io.error(name + ":" + e.message);
        return 1;
    }
}
