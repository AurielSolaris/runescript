// Integer semantics that JavaScript numbers do not have on their own.
package tests.js;

public int main() {
    int max = 2147483647;
    int min = -2147483648;
    trace("max - 1 + 1 =", max - 1 + 1);
    trace("min + 0 =", min + 0);

    // Truncating division and the sign of %.
    trace(-7 / 2, 7 / -2, -7 % 2, 7 % -2);
    int a = -7;
    int b = 2;
    trace(a / b, a % b, b / a, b % a);

    // Results that are -0 in JavaScript must print, compare, and store as 0.
    int zero = 0;
    int minusOne = -1;
    int negZeroProduct = zero * minusOne;
    int negZeroRemainder = -4 % b;
    int negZeroQuotient = minusOne / 5;
    int negated = -zero;
    trace(negZeroProduct, negZeroRemainder, negZeroQuotient, negated);
    trace(negZeroProduct == 0, negated == zero);
    trace(min % minusOne);

    // Products near the limit, too large to be exact as doubles once squared.
    int r = 46340;
    trace("46340 * 46340 =", r * r);
    trace("46341 * -46340 =", (r + 1) * -r);

    // Unsigned arithmetic uses the whole 0 .. 4294967295 range.
    uint u = 4294967294;
    u += 1;
    trace("uint max =", u);
    uint half = 2147483648;
    trace(half / 3, half % 7, half - 1, half + (half - 1));

    // Compound assignment and increments.
    int n = 10;
    n *= 3;
    n -= 4;
    n /= 5;
    n %= 4;
    n++;
    n--;
    n--;
    trace("n =", n);

    // Comparisons, booleans, and short-circuit evaluation.
    trace(min < max, u >= half, !(a == b), a != b && b > 0, a > 0 || b > 0);
    return 7;
}
