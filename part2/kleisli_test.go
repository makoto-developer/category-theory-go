package part2

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

var errOdd = errors.New("奇数は半分にできない")

func parseInt(s string) (int, error) { return strconv.Atoi(s) }

// halve は偶数だけを受け付ける。失敗しうる射の例として使う。
func halve(n int) (int, error) {
	if n%2 != 0 {
		return 0, errOdd
	}
	return n / 2, nil
}

func stringify(n int) (string, error) { return strconv.Itoa(n), nil }

func sameErr(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Error() == b.Error()
}

// Kleisli 圏の結合律。失敗する経路も含めて確かめる。
func TestKleisliCompositionIsAssociative(t *testing.T) {
	left := Then(Then(Kleisli[string, int](parseInt), halve), stringify)
	right := Then(Kleisli[string, int](parseInt), Then(Kleisli[int, int](halve), stringify))

	rapid.Check(t, func(t *rapid.T) {
		s := rapid.OneOf(
			rapid.Custom(func(t *rapid.T) string { return strconv.Itoa(rapid.Int().Draw(t, "n")) }),
			rapid.String(),
		).Draw(t, "s")

		gotL, errL := left(s)
		gotR, errR := right(s)

		if gotL != gotR || !sameErr(errL, errR) {
			t.Fatalf("結合律が破れた: s=%q, 左=(%q,%v), 右=(%q,%v)", s, gotL, errL, gotR, errR)
		}
	})
}

// Kleisli 圏の単位律。恒等射は「値をそのまま返し、エラーは nil」。
func TestKleisliIdentityIsUnit(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "s")
		want, wantErr := parseInt(s)

		if got, err := Then(Kleisli[string, string](KleisliIdentity[string]), parseInt)(s); got != want || !sameErr(err, wantErr) {
			t.Fatalf("左単位律が破れた: s=%q", s)
		}
		if got, err := Then(Kleisli[string, int](parseInt), Kleisli[int, int](KleisliIdentity[int]))(s); got != want || !sameErr(err, wantErr) {
			t.Fatalf("右単位律が破れた: s=%q", s)
		}
	})
}

// Pure で持ち上げた射は、失敗しない射として合成に混ぜられる。
func TestPureLiftsPlainMorphism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 1_000_000).Draw(t, "n")

		got, err := Then(Pure(itoa), Kleisli[string, int](parseInt))(n)

		if err != nil || got != n {
			t.Fatalf("持ち上げた射の合成が壊れた: n=%d, got=(%d,%v)", n, got, err)
		}
	})
}

// context 付きの合成では、キャンセル済みなら最初の段すら実行されない。
func TestChainStepStopsOnCanceledContext(t *testing.T) {
	called := false
	first := Step[int, int](func(ctx context.Context, n int) (int, error) {
		called = true
		return n, nil
	})
	second := Step[int, string](func(ctx context.Context, n int) (string, error) {
		return strconv.Itoa(n), nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ChainStep(first, second)(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("キャンセルが伝わっていない: %v", err)
	}
	if called {
		t.Fatal("キャンセル済みなのに最初の段が実行された")
	}
}
