// A runtime error inside a method is reported where it happens.
package tests.runtime;

public class Account {
    private int balance;

    public Account(int balance) {
        this.balance = balance;
    }

    public int share(int people) {
        return this.balance / people; // expect runtime error: ASX-RUN-003 12:29
    }
}

protected int main() {
    Account a = new Account(10);
    trace(a.share(2));
    trace(a.share(0));
    return 0;
}
