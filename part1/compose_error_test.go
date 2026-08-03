package part1

import (
	"errors"
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

var errDivideByZero = errors.New("0 では割れない")

// 射の等価性を判定するには error も比較しなければならない。strconv は呼び出しごとに
// 別の *NumError を返すため、errors.Is ではなくメッセージで突き合わせる。
func sameError(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Error() == b.Error()
}

func parseInt(s string) (int, error) { return strconv.Atoi(s) }

func reciprocal(n int) (float64, error) {
	if n == 0 {
		return 0, errDivideByZero
	}
	return 1 / float64(n), nil
}

func formatFloat(f float64) (string, error) {
	return strconv.FormatFloat(f, 'f', 4, 64), nil
}

// error を返す射の合成でも結合律は成り立つ。エラーが出る入力も含めて確かめる。
func TestComposeEIsAssociative(t *testing.T) {
	left := ComposeE(ComposeE(parseInt, reciprocal), formatFloat)
	right := ComposeE(parseInt, ComposeE(reciprocal, formatFloat))

	rapid.Check(t, func(t *rapid.T) {
		// 数字文字列を多めに引くため、int を文字列化したものと任意文字列を混ぜる。
		s := rapid.OneOf(
			rapid.Custom(func(t *rapid.T) string { return strconv.Itoa(rapid.Int().Draw(t, "n")) }),
			rapid.String(),
		).Draw(t, "s")

		gotL, errL := left(s)
		gotR, errR := right(s)

		if gotL != gotR || !sameError(errL, errR) {
			t.Fatalf("結合律が破れた: s=%q, 左=(%q,%v), 右=(%q,%v)", s, gotL, errL, gotR, errR)
		}
	})
}

func TestIdentityEIsUnit(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "s")

		want, wantErr := parseInt(s)

		if got, err := ComposeE(IdentityE[string], parseInt)(s); got != want || !sameError(err, wantErr) {
			t.Fatalf("左単位律が破れた: s=%q, got=(%d,%v), want=(%d,%v)", s, got, err, want, wantErr)
		}
		if got, err := ComposeE(parseInt, IdentityE[int])(s); got != want || !sameError(err, wantErr) {
			t.Fatalf("右単位律が破れた: s=%q, got=(%d,%v), want=(%d,%v)", s, got, err, want, wantErr)
		}
	})
}

// エラーが出た時点で後続の射は呼ばれない。if err != nil { return } を合成器に閉じ込めた効果。
func TestComposeEShortCircuits(t *testing.T) {
	called := false
	spy := func(f float64) (string, error) {
		called = true
		return formatFloat(f)
	}

	if _, err := ComposeE(ComposeE(parseInt, reciprocal), spy)("0"); !errors.Is(err, errDivideByZero) {
		t.Fatalf("想定したエラーが返っていない: %v", err)
	}
	if called {
		t.Fatal("エラーが出たのに後続の射が呼ばれた")
	}
}
