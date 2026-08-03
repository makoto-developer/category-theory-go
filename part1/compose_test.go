package part1

import (
	"testing"

	"pgregory.net/rapid"
)

// 結合律: (h∘g)∘f と h∘(g∘f) はどんな入力に対しても等しい。
func TestComposeIsAssociative(t *testing.T) {
	left := Compose(Compose(itoa, length), isEven)
	right := Compose(itoa, Compose(length, isEven))

	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Int().Draw(t, "x")

		if left(x) != right(x) {
			t.Fatalf("結合律が破れた: x=%d, (h∘g)∘f=%v, h∘(g∘f)=%v", x, left(x), right(x))
		}
	})
}

// 単位律: 恒等射を前後どちらに合成しても元の射と等しい。
func TestIdentityIsUnit(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Int().Draw(t, "x")

		if got := Compose(Identity[int], itoa)(x); got != itoa(x) {
			t.Fatalf("左単位律が破れた: x=%d, f∘id=%q, f=%q", x, got, itoa(x))
		}
		if got := Compose(itoa, Identity[string])(x); got != itoa(x) {
			t.Fatalf("右単位律が破れた: x=%d, id∘f=%q, f=%q", x, got, itoa(x))
		}
	})
}

// 結合律の実用上の帰結: 畳み込む向きを変えても Pipe の結果は変わらない。
func TestPipeFoldDirectionDoesNotMatter(t *testing.T) {
	fs := []func(int) int{double, increment, negate}

	leftFold := Pipe(fs...)
	rightFold := Compose(fs[0], Compose(fs[1], fs[2]))

	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Int().Draw(t, "x")

		if leftFold(x) != rightFold(x) {
			t.Fatalf("畳み込みの向きで結果が変わった: x=%d, 左=%d, 右=%d", x, leftFold(x), rightFold(x))
		}
	})
}

// 空の Pipe は恒等射になる（射の列の「単位元」）。
func TestEmptyPipeIsIdentity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Int().Draw(t, "x")

		if got := Pipe[int]()(x); got != x {
			t.Fatalf("空の Pipe が恒等射でない: x=%d, got=%d", x, got)
		}
	})
}
