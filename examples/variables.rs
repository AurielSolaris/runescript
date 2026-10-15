// Declarations, assignment, and compound assignment.
package examples.variables;

int main() {
    int count = 10;
    sint offset = -3;          // 'sint' is another spelling of 'int'
    uint total = 0;
    string label = "count";
    bool ready = true;

    count = count + offset;    // 7
    count *= 2;                // 14
    count++;                   // 15
    total += 40;

    trace(label, count);
    trace("total", total);
    trace("ready", ready);
    return 0;
}
