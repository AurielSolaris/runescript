// Array methods: insert, removeAt, slice, indexOf, contains, reverse, and
// sort, with and without a function; bounds errors are RangeErrors.
package tests.js;

use "std:string";

public struct Card {
    int rank;
    string suit;
}

public string ints(Array<int> a) {
    string s = "";
    for each (int v in a) {
        s += $"{v} ";
    }
    return s;
}

protected int main() {
    Array<int> a = [5, 3, 9, 1, 3];
    a.insert(0, 42);
    a.insert(a.length, 7);
    a.insert(2, -1);
    trace(ints(a));
    trace(a.removeAt(1), a.removeAt(a.length - 1), ints(a));
    trace(a.indexOf(3), a.indexOf(100), a.contains(9), a.contains(8));

    // slice copies.
    Array<int> part = a.slice(1, 3);
    part[0] = 99;
    trace(ints(part), a[1], a.slice(2, 2).length, a.slice(0, a.length).length);

    a.reverse();
    trace(ints(a));
    a.sort();
    trace(ints(a));
    a.sort((int x, int y) => x > y);
    trace(ints(a));

    // Strings sort by UTF-8 bytes: upper case first, U+FFFF before U+10000.
    Array<string> words = ["pear", "Apple", "\u{10000}", "\u{FFFF}", "apple", ""];
    words.sort();
    trace(string.join(words, ","));
    Array<char> letters = ['c', 'a', 'b'];
    letters.sort();
    trace(letters[0], letters[2]);
    Array<float64> fs = [2.5, -1.0, 0.5];
    fs.sort();
    trace(fs[0], fs[1], fs[2]);
    Array<uint64> big = [18446744073709551615, 0, 9223372036854775808];
    big.sort();
    trace(big[0], big[2]);

    // The sort is stable, and the function is called the same way on
    // both targets.
    Array<Card> hand = [Card{rank: 2, suit: "s"}, Card{rank: 1, suit: "h"}, Card{rank: 2, suit: "h"}, Card{rank: 1, suit: "s"}];
    hand.sort((Card x, Card y) => {
        trace("less", $"{x.rank}{x.suit}", $"{y.rank}{y.suit}");
        return x.rank < y.rank;
    });
    string order = "";
    for each (Card c in hand) {
        order += $"{c.rank}{c.suit} ";
    }
    trace(order);
    trace(hand.indexOf(Card{rank: 2, suit: "h"}), hand.contains(Card{rank: 3, suit: "h"}));

    // Nested arrays: sort rows by length, reverse each one.
    Array<Array<int>> rows = [[3, 2, 1], [5], [], [4, 4]];
    rows.sort((Array<int> x, Array<int> y) => x.length < y.length);
    for each (Array<int> r in rows) {
        r.reverse();
        trace(r.length, ints(r));
    }

    // An error thrown by the sort function leaves the array unchanged.
    Array<int> keep = [3, 1, 2];
    try {
        keep.sort((int x, int y) => {
            if (x == 2 || y == 2) {
                throw new Error("no twos");
            }
            return x < y;
        });
    } catch (Error e) {
        trace(e.message, ints(keep));
    }

    try {
        a.insert(-1, 0);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        a.insert(a.length + 1, 0);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        trace(a.removeAt(50));
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        trace(a.slice(3, 2).length);
    } catch (RangeError e) {
        trace(e.message);
    }
    try {
        trace(a.slice(0, a.length + 1).length);
    } catch (RangeError e) {
        trace(e.message);
    }
    Array<int> empty = [];
    empty.sort();
    empty.reverse();
    trace(empty.length, empty.indexOf(0));
    return 0;
}
