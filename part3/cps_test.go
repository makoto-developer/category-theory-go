package part3

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

// 値と継続渡しは行き来できる。往復すると元に戻る。
func TestContRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")

		if got := FromCont(ToCont[int](n)); got != n {
			t.Fatalf("往復で値が変わった: got=%d, want=%d", got, n)
		}
	})
}

// 継続渡しの上でも Functor 則が成り立つ。
func TestContFunctorLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")
		c := ToCont[int](n)

		if got := FromCont(MapCont(c, func(x int) int { return x })); got != n {
			t.Fatalf("恒等の保存が破れた: got=%d", got)
		}

		f := func(x int) int { return x * 3 }
		g := func(x int) int { return x - 1 }

		twice := FromCont(MapCont(MapCont(c, f), g))
		fused := FromCont(MapCont(c, func(x int) int { return g(f(x)) }))

		if twice != fused {
			t.Fatalf("合成の保存が破れた: 2回=%d, 合成=%d", twice, fused)
		}
	})
}

// push 型と pull 型は同じ列を表す。どちらから畳んでも答えは同じ。
func TestPushAndPullAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOfN(rapid.IntRange(-100, 100), 0, 100).Draw(t, "xs")
		seq := SeqOf(xs)

		push := SumSeq(seq)
		pull := SumPull(seq)

		want := 0
		for _, x := range xs {
			want += x
		}

		if push != want || pull != want {
			t.Fatalf("畳み込みの答えが違う: push=%d, pull=%d, want=%d", push, pull, want)
		}
	})
}

// push 型は「呼ばれる側が制御を持つ」ので、途中で止めるには打ち切りの合図が要る。
func TestSeqStopsOnBreak(t *testing.T) {
	produced := 0
	seq := func(yield func(int) bool) {
		for i := range 100 {
			produced++
			if !yield(i) {
				return
			}
		}
	}

	var got []int
	for v := range seq {
		got = append(got, v)
		if len(got) == 3 {
			break
		}
	}

	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Fatalf("取り出した要素が違う: %v", got)
	}
	if produced != 3 {
		t.Fatalf("打ち切りが効いていない: 生成数=%d, want=3", produced)
	}
}
