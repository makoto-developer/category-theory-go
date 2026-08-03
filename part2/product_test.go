package part2

import (
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

// 積の普遍性: Fanout で作った射を射影で戻すと、元の2本の射に一致する。
func TestFanoutSatisfiesUniversalProperty(t *testing.T) {
	f := func(n int) string { return strconv.Itoa(n) }
	g := func(n int) bool { return n%2 == 0 }
	both := Fanout(f, g)

	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")
		p := both(n)

		if p.First != f(n) || p.Second != g(n) {
			t.Fatalf("積の普遍性が破れた: n=%d, pair=%+v", n, p)
		}
	})
}

// 余積の普遍性: Fanin はどちらの側から来ても、対応する射を1回だけ適用する。
func TestFaninCoversBothSides(t *testing.T) {
	onLeft := func(n int) string { return "int:" + strconv.Itoa(n) }
	onRight := func(s string) string { return "str:" + s }
	fold := Fanin(onLeft, onRight)

	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")
		s := rapid.String().Draw(t, "s")

		if got := fold(Left[int, string](n)); got != onLeft(n) {
			t.Fatalf("左側が正しく畳まれていない: got=%q", got)
		}
		if got := fold(Right[int, string](s)); got != onRight(s) {
			t.Fatalf("右側が正しく畳まれていない: got=%q", got)
		}
	})
}

// Either[A, B] と Either[B, A] は同型。入れ替えを2回すると元に戻る。
func TestSwapIsInvolution(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")
		back := Swap(Swap(Left[int, string](n)))

		if got := Fanin(func(x int) int { return x }, func(string) int { return -1 })(back); got != n {
			t.Fatalf("2回入れ替えても元に戻らない: n=%d, got=%d", n, got)
		}
	})
}

// カリー化と非カリー化は互いに逆。積と冪の随伴の片鱗がここに出ている。
func TestCurryUncurryRoundTrip(t *testing.T) {
	area := func(p Pair[int, int]) int { return p.First * p.Second }
	roundTripped := Uncurry(Curry(area))

	rapid.Check(t, func(t *rapid.T) {
		w := rapid.IntRange(-1000, 1000).Draw(t, "w")
		h := rapid.IntRange(-1000, 1000).Draw(t, "h")
		p := Pair[int, int]{First: w, Second: h}

		if roundTripped(p) != area(p) {
			t.Fatalf("往復で射が変わった: %+v", p)
		}
	})
}
