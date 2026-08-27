package part9

import (
	"math"
	"slices"
	"testing"

	"pgregory.net/rapid"
)

func genZipper() *rapid.Generator[Zipper[int]] {
	return rapid.Custom(func(t *rapid.T) Zipper[int] {
		xs := rapid.SliceOfN(rapid.IntRange(-20, 20), 1, 12).Draw(t, "xs")
		return Zipper[int]{Items: xs, Pos: rapid.IntRange(0, len(xs)-1).Draw(t, "pos")}
	})
}

func eqZ[A comparable](a, b Zipper[A]) bool {
	return a.Pos == b.Pos && slices.Equal(a.Items, b.Items)
}

// Comonad 則を extend の形で3つ検査する。
//
//	extend extract        = id
//	extract ∘ extend f    = f
//	extend f ∘ extend g   = extend (f ∘ extend g)
func TestComonadLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		w := genZipper().Draw(t, "w")
		// 焦点だけを見る関数だと、周りの文脈が壊れていても検出できない。
		// 近傍・長さ・位置を見る関数を使う。
		f := func(z Zipper[int]) int {
			lo, hi := max(0, z.Pos-1), min(len(z.Items)-1, z.Pos+1)
			s := z.Pos * 31
			for i := lo; i <= hi; i++ {
				s = s*7 + z.Items[i]
			}
			return s
		}
		g := func(z Zipper[int]) int {
			s := len(z.Items)*13 + z.Pos
			for i, v := range z.Items {
				s += v * (i + 1)
			}
			return s
		}

		if got := Extend(Extract[int], w); !eqZ(got, w) {
			t.Fatalf("extend extract = id が破れた: %v vs %v", got, w)
		}
		if got, want := Extract(Extend(f, w)), f(w); got != want {
			t.Fatalf("extract ∘ extend f = f が破れた: %d vs %d", got, want)
		}
		left := Extend(f, Extend(g, w))
		right := Extend(func(z Zipper[int]) int { return f(Extend(g, z)) }, w)
		if !eqZ(left, right) {
			t.Fatalf("結合律が破れた: %v vs %v", left, right)
		}
	})
}

// 法則が言う定義（fmap f ∘ duplicate）と、直接書いた extend が同じ答えを返す。
// 中間構造を作るかどうかだけが違う。
func TestExtendAgreesWithDuplicateThenMap(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		w := genZipper().Draw(t, "w")
		f := func(z Zipper[int]) int {
			lo, hi := max(0, z.Pos-1), min(len(z.Items)-1, z.Pos+1)
			s := 0
			for i := lo; i <= hi; i++ {
				s += z.Items[i]
			}
			return s
		}
		if a, b := Extend(f, w), ExtendViaDuplicate(f, w); !eqZ(a, b) {
			t.Fatalf("直接版と duplicate 経由で答えが違う: %v vs %v", a, b)
		}
	})
}

// 移動平均は extend そのもの。手書きループと同じ答えになる。
func TestMovingAverageAgreesWithLoop(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOfN(rapid.Float64Range(-100, 100), 1, 30).Draw(t, "xs")
		half := rapid.IntRange(0, 3).Draw(t, "half")

		viaExtend := Extend(MovingAverage(half), Zipper[float64]{Items: xs, Pos: 0}).Items
		viaLoop := MovingAverageLoop(xs, half)
		for i := range xs {
			if math.Abs(viaExtend[i]-viaLoop[i]) > 1e-9 {
				t.Fatalf("%d 番目が違う: %v vs %v", i, viaExtend[i], viaLoop[i])
			}
		}
	})
}

// 余モナドとして扱えるのは Valid な Zipper だけ。空だと Extract が定義されない。
func TestExtractIsUndefinedOnInvalidZipper(t *testing.T) {
	for _, w := range []Zipper[int]{
		{Items: nil, Pos: 0},
		{Items: []int{1}, Pos: 1},
		{Items: []int{1}, Pos: -1},
	} {
		if w.Valid() {
			t.Fatalf("%v は Valid ではないはず", w)
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("%v で Extract が panic しなかった", w)
				}
			}()
			_ = Extract(w)
		}()
	}
	t.Logf("Extract は部分関数。余モナドを名乗れるのは Valid な Zipper に限る")
}
