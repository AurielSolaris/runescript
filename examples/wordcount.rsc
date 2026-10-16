// A small text report: split lines and words, count them, and build
// output with interpolated strings. Shows Array<T>, char callbacks, +,
// $"…", and std:string's split, fields, map, trimFunc, and join.
package examples;

use "std:string";

public bool isPunct(char c) {
    return c == ',' || c == '.' || c == '!' || c == '?';
}

public char lower(char c) {
    if (c >= 'A' && c <= 'Z') {
        return (char) ((int) c + 32);
    }
    return c;
}

// indexOf returns the position of word in words, or -1.
public int indexOf(Array<string> words, string word) {
    for (int i = 0; i < words.length; i++) {
        if (words[i] == word) {
            return i;
        }
    }
    return -1;
}

public void main() {
    string text = "The cat sat.\nThe dog sat, too!\nA cat? A dog.";
    Array<string> lines = string.split(text, "\n");

    Array<string> seen = [];
    Array<int> counts = [];
    int total = 0;
    for (int i = 0; i < lines.length; i++) {
        Array<string> words = string.fields(lines[i]);
        trace($"line {i + 1}: {words.length} words");
        for (int j = 0; j < words.length; j++) {
            string word = string.map(lower, string.trimFunc(words[j], isPunct));
            int at = indexOf(seen, word);
            if (at < 0) {
                seen.push(word);
                counts.push(1);
            } else {
                counts[at] += 1;
            }
            total++;
        }
    }

    Array<string> report = [];
    for (int i = 0; i < seen.length; i++) {
        report.push(seen[i] + "=" + (string) counts[i]);
    }
    trace($"{total} words, {seen.length} different");
    trace(string.join(report, ", "));
}
