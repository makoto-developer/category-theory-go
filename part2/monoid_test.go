package part2

import (
	"errors"
	"math"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// checkMonoidLaws は結合律と単位律を確かめる。等価性の判定は呼び出し側から渡す
// （error のように == で比べられない型があるため）。
func checkMonoidLaws[T any](t *rapid.T, m Monoid[T], gen *rapid.Generator[T], eq func(T, T) bool) {
	a := gen.Draw(t, "a")
	b := gen.Draw(t, "b")
	c := gen.Draw(t, "c")

	left := m.Append(m.Append(a, b), c)
	right := m.Append(a, m.Append(b, c))
	if !eq(left, right) {
		t.Fatalf("結合律が破れた: (a·b)·c=%v, a·(b·c)=%v", left, right)
	}

	if got := m.Append(m.Empty, a); !eq(got, a) {
		t.Fatalf("左単位律が破れた: e·a=%v, a=%v", got, a)
	}
	if got := m.Append(a, m.Empty); !eq(got, a) {
		t.Fatalf("右単位律が破れた: a·e=%v, a=%v", got, a)
	}
}

func eqComparable[T comparable](a, b T) bool { return a == b }

func eqError(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Error() == b.Error()
}

func TestSumIntIsMonoid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkMonoidLaws(t, SumInt, rapid.Int(), eqComparable[int])
	})
}

func TestConcatStringIsMonoid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkMonoidLaws(t, ConcatString, rapid.String(), eqComparable[string])
	})
}

// strings.Builder を経由しても、文字列連結と同じモノイドである。
func TestBuildStringIsSameMonoid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkMonoidLaws(t, BuildString, rapid.String(), eqComparable[string])

		a := rapid.String().Draw(t, "a")
		b := rapid.String().Draw(t, "b")
		if BuildString.Append(a, b) != ConcatString.Append(a, b) {
			t.Fatal("Builder 版と連結版で結果が違う")
		}
	})
}

func TestMaxDurationIsMonoid(t *testing.T) {
	gen := rapid.Custom(func(t *rapid.T) time.Duration {
		return time.Duration(rapid.Int64Range(0, int64(time.Hour)).Draw(t, "d"))
	})

	rapid.Check(t, func(t *rapid.T) {
		checkMonoidLaws(t, MaxDuration, gen, eqComparable[time.Duration])
	})
}

// errors.Join も nil を単位元とするモノイド。nil を渡しても増えないのが効いている。
func TestJoinErrorsIsMonoid(t *testing.T) {
	gen := rapid.Custom(func(t *rapid.T) error {
		msg := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "msg")
		if rapid.Bool().Draw(t, "isNil") {
			return nil
		}
		return errors.New(msg)
	})

	rapid.Check(t, func(t *rapid.T) {
		checkMonoidLaws(t, JoinErrors, gen, eqError)
	})
}

// 結合律があるので、どう分割して並列に畳み込んでも答えは変わらない。
func TestFoldParallelMatchesSequential(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOfN(rapid.IntRange(-1000, 1000), 0, 500).Draw(t, "xs")
		workers := rapid.IntRange(1, 8).Draw(t, "workers")

		if got, want := FoldParallel(SumInt, xs, workers), Fold(SumInt, xs); got != want {
			t.Fatalf("並列と逐次で答えが違う: workers=%d, got=%d, want=%d", workers, got, want)
		}
	})
}

// 平均は結合的でない。素朴に分割並列化すると答えが変わる。
func TestMeanBreaksUnderParallelSplit(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	sequential := Mean(xs)
	parallel := MeanParallel(xs, 3)

	if sequential == parallel {
		t.Fatalf("平均が並列化で壊れなかった: 逐次=%v, 並列=%v", sequential, parallel)
	}
	t.Logf("逐次=%v / 3分割=%v （差=%v）", sequential, parallel, math.Abs(sequential-parallel))
}

// 合計と個数に分ければ両方とも結合的になり、並列化しても答えが合う。
func TestSumCountMonoidFixesParallelMean(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOfN(rapid.Float64Range(-1e6, 1e6), 1, 200).Draw(t, "xs")
		workers := rapid.IntRange(1, 8).Draw(t, "workers")

		pairs := MapSlice(xs, func(x float64) SumCount { return SumCount{Sum: x, Count: 1} })

		got := FoldParallel(MeanMonoid, pairs, workers).Value()
		want := Fold(MeanMonoid, pairs).Value()

		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("並列と逐次で答えが違う: workers=%d, got=%v, want=%v", workers, got, want)
		}
	})
}

// 浮動小数点の加算は結合的でないため、分割数を変えると答えがずれることがある。
func TestFloatAdditionIsNotAssociative(t *testing.T) {
	a, b, c := 1e16, -1e16, 1.0

	left := (a + b) + c
	right := a + (b + c)

	if left == right {
		t.Fatalf("結合律が破れなかった: 両方とも %v", left)
	}
	t.Logf("(a+b)+c = %v / a+(b+c) = %v", left, right)
}
