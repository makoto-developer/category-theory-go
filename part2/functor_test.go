package part2

import (
	"errors"
	"slices"
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

func itoa(n int) string   { return strconv.Itoa(n) }
func length(s string) int { return len(s) }

// Functor 則1（恒等の保存）: 恒等射で写しても何も変わらない。
// Functor 則2（合成の保存）: 2回写すのと、合成した射で1回写すのは等しい。
// 則2 は「ループを1本にまとめてよい」という許可証そのものである。
func TestSliceFunctorLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOf(rapid.Int()).Draw(t, "xs")

		if got := MapSlice(xs, Identity[int]); !slices.Equal(got, xs) {
			t.Fatalf("恒等の保存が破れた: got=%v, want=%v", got, xs)
		}

		twice := MapSlice(MapSlice(xs, itoa), length)
		fused := MapSlice(xs, Compose(itoa, length))

		if !slices.Equal(twice, fused) {
			t.Fatalf("合成の保存が破れた: 2回=%v, 合成=%v", twice, fused)
		}
	})
}

func TestPtrFunctorLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")
		p := &n

		if got := MapPtr(p, Identity[int]); *got != n {
			t.Fatalf("恒等の保存が破れた: got=%d, want=%d", *got, n)
		}

		twice := MapPtr(MapPtr(p, itoa), length)
		fused := MapPtr(p, Compose(itoa, length))

		if *twice != *fused {
			t.Fatalf("合成の保存が破れた: 2回=%d, 合成=%d", *twice, *fused)
		}
	})

	// 「値が無い」という構造も保たれる。
	if MapPtr[int, string](nil, itoa) != nil {
		t.Fatal("nil が nil に写っていない")
	}
}

func TestSeqFunctorLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOf(rapid.Int()).Draw(t, "xs")
		seq := slices.Values(xs)

		if got := slices.Collect(MapSeq(seq, Identity[int])); !slices.Equal(got, xs) {
			t.Fatalf("恒等の保存が破れた: got=%v, want=%v", got, xs)
		}

		twice := slices.Collect(MapSeq(MapSeq(seq, itoa), length))
		fused := slices.Collect(MapSeq(seq, Compose(itoa, length)))

		if !slices.Equal(twice, fused) {
			t.Fatalf("合成の保存が破れた: 2回=%v, 合成=%v", twice, fused)
		}
	})
}

// iter.Seq は遅延なので、途中で打ち切れば残りの要素に射は適用されない。
// 同じ Functor 則を満たしながら、評価回数はスライス版と違う。
func TestSeqFunctorIsLazy(t *testing.T) {
	applied := 0
	counted := func(n int) int {
		applied++
		return n * 2
	}

	for range MapSeq(slices.Values([]int{1, 2, 3, 4, 5}), counted) {
		break
	}

	if applied != 1 {
		t.Fatalf("遅延評価されていない: 適用回数=%d, want=1", applied)
	}
}

func TestErrFunctorLaws(t *testing.T) {
	sentinel := errors.New("失敗")

	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")

		if got, err := MapErr(n, nil, Identity[int]); got != n || err != nil {
			t.Fatalf("恒等の保存が破れた: got=(%d,%v), want=(%d,nil)", got, err, n)
		}

		s, err := MapErr(n, nil, itoa)
		twiceVal, twiceErr := MapErr(s, err, length)
		fusedVal, fusedErr := MapErr(n, nil, Compose(itoa, length))

		if twiceVal != fusedVal || !errors.Is(twiceErr, fusedErr) {
			t.Fatalf("合成の保存が破れた: 2回=(%d,%v), 合成=(%d,%v)", twiceVal, twiceErr, fusedVal, fusedErr)
		}
	})

	// エラーは素通しされ、射は適用されない。
	if _, err := MapErr(0, sentinel, itoa); !errors.Is(err, sentinel) {
		t.Fatalf("エラーが素通しされていない: %v", err)
	}
}
