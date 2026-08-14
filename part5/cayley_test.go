package part5

import (
	"slices"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func drawPieces(t *rapid.T) []string {
	return rapid.SliceOfN(rapid.StringMatching(`[a-z]{1,4}`), 0, 12).Draw(t, "pieces")
}

// checkMonoid は結合律と単位律を、生成された例について検査する。証明ではない。
// build は任意長の値を作る（1要素だけで検査すると、長さが関わる破れを見逃す）。
// 3つのモノイドは全部これを通る。通ってしまうからこそ、要素と演算結果しか見ない
// この抽象からは3つを区別できない。
func checkMonoid[M any](t *rapid.T, mo Monoid[M], build func([]string) M, eq func(M, M) bool) {
	a := build(drawPieces(t))
	b := build(drawPieces(t))
	c := build(drawPieces(t))

	left := mo.Append(mo.Append(a, b), c)
	right := mo.Append(a, mo.Append(b, c))
	if !eq(left, right) {
		t.Fatalf("結合律が破れた")
	}
	if !eq(mo.Append(mo.Empty, a), a) || !eq(mo.Append(a, mo.Empty), a) {
		t.Fatalf("単位律が破れた")
	}
}

// buildCons / buildSnoc は断片の列から任意長の値を組み立てる。
func buildCons(xs []string) *List {
	cells := make([]*List, len(xs))
	for i, x := range xs {
		cells[i] = SingleCons(x)
	}
	return FoldNaive(ConsMonoid, cells)
}

func buildSnoc(xs []string) *RList {
	cells := make([]*RList, len(xs))
	for i, x := range xs {
		cells[i] = SingleSnoc(x)
	}
	return FoldNaive(SnocMonoid, cells)
}

func TestStringMonoidLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkMonoid(t, StringMonoid, func(xs []string) string { return strings.Join(xs, "") },
			func(a, b string) bool { return a == b })
	})
}

func TestConsMonoidLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkMonoid(t, ConsMonoid, buildCons,
			func(a, b *List) bool { return slices.Equal(a.Slice(), b.Slice()) })
	})
}

func TestSnocMonoidLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		checkMonoid(t, SnocMonoid, buildSnoc,
			func(a, b *RList) bool { return slices.Equal(a.Slice(), b.Slice()) })
	})
}

// Cayley 埋め込みはモノイド準同型。lift(a·b) と lift(a)∘lift(b) が一致する。
// これが成り立つから「畳み方を変えても意味は変わらない」と言える。
func TestCayleyIsMonoidHomomorphism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := drawPieces(t)
		if len(xs) < 3 {
			return
		}
		a, b, rest := xs[0], xs[1], xs[2]

		combined := Cayley(StringMonoid, StringMonoid.Append(a, b))(rest)
		composed := Cayley(StringMonoid, a)(Cayley(StringMonoid, b)(rest))
		if combined != composed {
			t.Fatalf("準同型でない: %q != %q", combined, composed)
		}
		if got := Cayley(StringMonoid, StringMonoid.Empty)(rest); got != rest {
			t.Fatalf("単位元が恒等射に移っていない: %q", got)
		}
	})
}

// 記事の主張の前提。左結合で畳んでも Cayley 表現を経由しても、答えは同じ。
// 変わるのはコストだけで、意味は変わらない。
func TestFoldsAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := drawPieces(t)

		want := strings.Join(xs, "")
		if got := FoldNaive(StringMonoid, xs); got != want {
			t.Fatalf("FoldNaive(string) = %q, want %q", got, want)
		}
		if got := FoldCayley(StringMonoid, xs); got != want {
			t.Fatalf("FoldCayley(string) = %q, want %q", got, want)
		}
		if got := BuildString(xs); got != want {
			t.Fatalf("BuildString = %q, want %q", got, want)
		}
		if got := BuildStringGrow(xs, len(want)); got != want {
			t.Fatalf("BuildStringGrow = %q, want %q", got, want)
		}

		cons := make([]*List, len(xs))
		for i, x := range xs {
			cons[i] = SingleCons(x)
		}
		for name, got := range map[string][]string{
			"FoldNaive(cons)":  FoldNaive(ConsMonoid, cons).Slice(),
			"FoldCayley(cons)": FoldCayley(ConsMonoid, cons).Slice(),
			"PrependReverse":   PrependReverse(xs).Slice(),
		} {
			if !slices.Equal(got, orNil(xs)) {
				t.Fatalf("%s = %v, want %v", name, got, xs)
			}
		}

		snoc := make([]*RList, len(xs))
		for i, x := range xs {
			snoc[i] = SingleSnoc(x)
		}
		for name, got := range map[string][]string{
			"FoldNaive(snoc)":  FoldNaive(SnocMonoid, snoc).Slice(),
			"FoldCayley(snoc)": FoldCayley(SnocMonoid, snoc).Slice(),
		} {
			if !slices.Equal(got, orNil(xs)) {
				t.Fatalf("%s = %v, want %v", name, got, xs)
			}
		}
	})
}

// 空リストの Slice() は nil を返すので、期待値側も揃える。
func orNil(xs []string) []string {
	if len(xs) == 0 {
		return nil
	}
	return xs
}
