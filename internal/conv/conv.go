// Package conv holds the number formatting and parsing rules shared by the
// interpreter and the checker. The JavaScript runtime implements the same
// rules, and tests compare the two.
package conv

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// FormatFloat formats a finite float the way ActionScript 3 (ECMAScript's
// Number::toString) does: the shortest decimal that reads back as the same
// value, in plain notation for magnitudes from 1e-6 to below 1e21 and in
// exponent notation otherwise. Integral values have no fraction: 64.0
// formats as "64". -0 formats as "0", and the non-finite values as "NaN",
// "Infinity", and "-Infinity".
//
// bits is 32 for float32 values, whose digits are chosen at float32
// precision by shortest32, and 64 otherwise.
func FormatFloat(f float64, bits int) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	if f == 0 {
		return "0"
	}
	if f < 0 {
		return "-" + FormatFloat(-f, bits)
	}
	var digits string
	var exp int // f ≈ 0.digits × 10^exp
	if bits == 32 {
		digits, exp = shortest32(f)
	} else {
		// d.ddddde±XX, with the shortest digits that round-trip.
		mant, e, _ := strings.Cut(strconv.FormatFloat(f, 'e', -1, 64), "e")
		x, _ := strconv.Atoi(e)
		digits, exp = strings.Replace(mant, ".", "", 1), x+1
	}
	return layout(digits, exp)
}

// layout writes digits × 10^(n-len(digits)) as ECMAScript Number::toString
// does.
func layout(digits string, n int) string {
	k := len(digits)
	switch {
	case k <= n && n <= 21:
		return digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return "0." + strings.Repeat("0", -n) + digits
	}
	sign := "+"
	if n-1 < 0 {
		sign = "-"
	}
	exponent := "e" + sign + strconv.Itoa(abs(n-1))
	if k == 1 {
		return digits + exponent
	}
	return digits[:1] + "." + digits[1:] + exponent
}

// shortest32 returns the significant digits and decimal exponent of the
// decimal that float32 y (> 0) is printed as: for the fewest digits p that
// work, the p-digit decimal that reads back as y (decimal → nearest float64
// → nearest float32) and is closest to y, preferring an even last digit on
// an exact tie. Only the correctly rounded p-digit decimal m and its
// neighbours m±1 can qualify. The JavaScript runtime implements the same
// algorithm (Float32JS).
func shortest32(y float64) (string, int) {
	yNum, yDen := exact(y)
	for p := 1; p <= 9; p++ {
		mant, e, _ := strings.Cut(strconv.FormatFloat(y, 'e', p-1, 64), "e")
		x, _ := strconv.Atoi(e)
		m, _ := new(big.Int).SetString(strings.Replace(mant, ".", "", 1), 10)
		e10 := x - (p - 1) // the decimal is m × 10^e10
		var best, bestDist *big.Int
		for _, d := range []int64{-1, 0, 1} {
			k := new(big.Int).Add(m, big.NewInt(d))
			if k.Sign() <= 0 {
				continue
			}
			v, _ := strconv.ParseFloat(k.String()+"e"+strconv.Itoa(e10), 64)
			if ToFloat32(v) != y {
				continue
			}
			dist := distance(k, e10, yNum, yDen)
			if c := cmpDist(dist, bestDist); best == nil || c < 0 || c == 0 && k.Bit(0) == 0 {
				best, bestDist = k, dist
			}
		}
		if best != nil {
			digits := strings.TrimRight(best.String(), "0")
			return digits, e10 + len(best.String())
		}
	}
	panic("conv: no float32 representation") // 9 digits always suffice
}

func cmpDist(a, b *big.Int) int {
	if b == nil {
		return -1
	}
	return a.Cmp(b)
}

// exact returns y as the exact fraction num/den.
func exact(y float64) (num, den *big.Int) {
	mant, exp := math.Frexp(y) // y = mant × 2^exp, 0.5 <= mant < 1
	m := new(big.Int).SetUint64(uint64(mant * (1 << 53)))
	exp -= 53
	if exp >= 0 {
		return m.Lsh(m, uint(exp)), big.NewInt(1)
	}
	return m, new(big.Int).Lsh(big.NewInt(1), uint(-exp))
}

// distance is |k × 10^e10 − num/den|, scaled by a positive factor that
// depends only on e10 and den, so distances for one e10 compare correctly.
func distance(k *big.Int, e10 int, num, den *big.Int) *big.Int {
	pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(e10))), nil)
	var a, b *big.Int
	if e10 >= 0 {
		// k·10^e − n/d  ∝  k·10^e·d − n
		a = new(big.Int).Mul(new(big.Int).Mul(k, pow), den)
		b = num
	} else {
		// k/10^-e − n/d  ∝  k·d − n·10^-e
		a = new(big.Int).Mul(k, den)
		b = new(big.Int).Mul(num, pow)
	}
	return a.Sub(a, b).Abs(a)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

var (
	intRE   = regexp.MustCompile(`^[+-]?[0-9]+$`)
	floatRE = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
)

// ParseInt parses a decimal integer: an optional sign and digits, with no
// spaces. ok is false if s is not of that form; the value is exact, so it
// can be range-checked by the caller (fits is false if it doesn't even fit
// in int64).
func ParseInt(s string) (v int64, fits, ok bool) {
	if !intRE.MatchString(s) {
		return 0, false, false
	}
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err == nil, true
}

// ParseFloat parses a decimal number: an optional sign, digits, an optional
// fraction, and an optional exponent, with no spaces ("64.7", "-1e5",
// "42"), or one of "NaN", "Infinity", "+Infinity", and "-Infinity". The
// result is the nearest float64 (infinite if s is too large); a float32
// conversion rounds that value again. ok is false if s is not of that form.
func ParseFloat(s string) (v float64, ok bool) {
	switch s {
	case "NaN":
		return math.NaN(), true
	case "Infinity", "+Infinity":
		return math.Inf(1), true
	case "-Infinity":
		return math.Inf(-1), true
	}
	if !floatRE.MatchString(s) {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil && !math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// ToFloat32 rounds f to the nearest float32.
func ToFloat32(f float64) float64 { return float64(float32(f)) }

// Float32JS is the JavaScript runtime's float32 formatting, f32str(y),
// where y is a float32 value held in a JavaScript number. It implements
// shortest32 and layout, and is inlined into the runtime by internal/jsgen.
const Float32JS = `function f32str(y) {
  if (!isFinite(y)) return String(y);
  if (y === 0) return "0";
  if (y < 0) return "-" + f32str(-y);
  const f = new Float32Array([y]), bits = new Uint32Array(f.buffer)[0];
  const be = (bits >>> 23) & 0xff, frac = BigInt(bits & 0x7fffff);
  // y = yNum / yDen exactly.
  const m2 = be === 0 ? frac : frac | 0x800000n, e2 = (be === 0 ? 1 : be) - 150;
  const yNum = e2 >= 0 ? m2 << BigInt(e2) : m2, yDen = e2 >= 0 ? 1n : 1n << BigInt(-e2);
  for (let p = 1; p <= 9; p++) {
    const [mant, ex] = y.toExponential(p - 1).split("e");
    const m = BigInt(mant.replace(".", "")), e10 = Number(ex) - (p - 1);
    const pow = 10n ** BigInt(Math.abs(e10));
    let best = null, bestDist = null;
    for (const k of [m - 1n, m, m + 1n]) {
      if (k <= 0n || Math.fround(Number(k + "e" + e10)) !== y) continue;
      let d = e10 >= 0 ? k * pow * yDen - yNum : k * yDen - yNum * pow;
      if (d < 0n) d = -d;
      if (best === null || d < bestDist || (d === bestDist && k % 2n === 0n)) { best = k; bestDist = d; }
    }
    if (best !== null) {
      const s = best.toString(), digits = s.replace(/0+$/, ""), n = e10 + s.length, k = digits.length;
      if (k <= n && n <= 21) return digits + "0".repeat(n - k);
      if (0 < n && n <= 21) return digits.slice(0, n) + "." + digits.slice(n);
      if (-6 < n && n <= 0) return "0." + "0".repeat(-n) + digits;
      const exp = "e" + (n - 1 < 0 ? "-" : "+") + Math.abs(n - 1);
      return k === 1 ? digits + exp : digits[0] + "." + digits.slice(1) + exp;
    }
  }
}`
