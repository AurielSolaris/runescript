package main;

// rune:fs is host-neutral: on Node.js it uses node:fs. The test runs this
// program in an empty directory.

extern "rune:fs" void writeFile(string path, string data);
extern "rune:fs" void appendFile(string path, string data);
extern "rune:fs" string readFile(string path);
extern "rune:fs" bool exists(string path);
extern "rune:fs" void remove(string path);

public int main() {
    trace(exists("notes.txt"));
    writeFile("notes.txt", "ᚱᚢᚾᛖ\n");
    appendFile("notes.txt", "line two");
    trace(exists("notes.txt"));
    trace(readFile("notes.txt"));
    remove("notes.txt");
    trace(exists("notes.txt"));
    try {
        string gone = readFile("notes.txt");
        trace(gone);
    } catch (Error e) {
        trace(e.name, e.errorID, e.message != "");
    }
    return 0;
}
