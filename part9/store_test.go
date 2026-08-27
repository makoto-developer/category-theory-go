package part9

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

type Doc struct {
	Title string
	Tags  []string
}

func eqDoc(a, b Doc) bool { return a.Title == b.Title && slices.Equal(a.Tags, b.Tags) }

// 法則を満たすレンズ。
func titleLens() Lens[Doc, string] {
	return Lens[Doc, string]{
		Get: func(d Doc) string { return d.Title },
		Set: func(d Doc, s string) Doc { d.Title = s; return d },
	}
}

// set-set だけを破るレンズ。差し替えのたびに Tags へ履歴を積むので、
// 2回入れたものと1回入れたものが Tags で食い違う。
// 同じ値なら何もしないので get-set は保たれ、Title は常に入れた値なので set-get も保たれる。
func brokenSetSet() Lens[Doc, string] {
	return Lens[Doc, string]{
		Get: func(d Doc) string { return d.Title },
		Set: func(d Doc, s string) Doc {
			if d.Title == s {
				return d
			}
			d.Title = s
			d.Tags = append(slices.Clone(d.Tags), s)
			return d
		},
	}
}

// set-get を破るレンズ。入れた値と取れる値が違う。
func brokenSetGet() Lens[Doc, string] {
	return Lens[Doc, string]{
		Get: func(d Doc) string { return d.Title },
		Set: func(d Doc, s string) Doc { d.Title = s + "_"; return d },
	}
}

// counit 則: extract ∘ coalg = id。これは get-set 則そのもの。
func counitHolds(l Lens[Doc, string], d Doc) bool {
	return eqDoc(StoreExtract(LensToCoalgebra(l)(d)), d)
}

// coassociativity 則: fmap coalg ∘ coalg = duplicate ∘ coalg。
//
// **これは反例を見つける道具であって、判定器ではない。** Store は関数を含むので
// 直接比較できず、有限個の probe で外延的に比べているだけ。A が無限集合なら、
// probe の外だけで法則を破るレンズは見逃す。
//
// とくに **probe が1個だと set-set 違反は原理的に検出できない**
// （set-set は2つの異なる値を続けて入れたときの話なので）。
// TestCoassocCheckMissesViolationsWithOneProbe がそれを実演している。
func coassocHolds(l Lens[Doc, string], d Doc, probes []string) bool {
	coalg := LensToCoalgebra(l)
	st := coalg(d)

	for _, a := range probes {
		// 左辺: coalg を内側に写してから a を覗く
		left := coalg(st.Peek(a))
		// 右辺: duplicate してから a を覗く
		right := StoreDuplicate(st).Peek(a)

		if left.Pos != right.Pos {
			return false // ここが崩れるのは set-get 則が破れているとき
		}
		for _, a2 := range probes {
			if !eqDoc(left.Peek(a2), right.Peek(a2)) {
				return false // ここが崩れるのは set-set 則が破れているとき
			}
		}
	}
	return true
}

// 法則を満たすレンズは、Store 余モナドの余代数則も満たす。
func TestLawfulLensIsAStoreCoalgebra(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		d := Doc{Title: rapid.String().Draw(t, "title")}
		// probe は2つ以上ないと set-set 側を突けない
		probes := rapid.SliceOfNDistinct(rapid.String(), 2, 4, rapid.ID[string]).Draw(t, "probes")
		l := titleLens()

		if !counitHolds(l, d) {
			t.Fatalf("counit 則が破れた（= get-set 則）")
		}
		if !coassocHolds(l, d, probes) {
			t.Fatalf("coassociativity 則が破れた（= set-get / set-set 則）")
		}
	})
}

// set-set を破るレンズは、coassociativity 則も破る。対応が本物であることの裏。
func TestBrokenSetSetBreaksCoassociativity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := brokenSetSet()
		d := Doc{
			Title: rapid.String().Draw(t, "title"),
			Tags:  rapid.SliceOfN(rapid.String(), 0, 3).Draw(t, "tags"),
		}
		a := rapid.String().Draw(t, "a")
		b := rapid.String().Draw(t, "b")

		// get-set と set-get は、どんな入力でも保たれる
		if !eqDoc(l.Set(d, l.Get(d)), d) {
			t.Fatalf("get-set は保たれているはず")
		}
		if l.Get(l.Set(d, a)) != a {
			t.Fatalf("set-get は保たれているはず")
		}

		// a と b が違い、どちらも元の Title と違うときだけ set-set が破れる
		if a == b || a == d.Title || b == d.Title {
			return
		}
		if eqDoc(l.Set(l.Set(d, a), b), l.Set(d, b)) {
			t.Fatalf("set-set は破れているはず")
		}
		if coassocHolds(l, d, []string{a, b}) {
			t.Fatalf("set-set が破れているのに coassociativity が成り立ってしまった")
		}
	})
}

// set-get を破るレンズも同様。
func TestBrokenSetGetBreaksCoassociativity(t *testing.T) {
	l := brokenSetGet()
	d := Doc{Title: "x"}

	if l.Get(l.Set(d, "a")) == "a" {
		t.Fatalf("このレンズは set-get を破るはず")
	}
	if coassocHolds(l, d, []string{"a", "b"}) {
		t.Fatalf("set-get が破れているのに coassociativity が成り立ってしまった")
	}
	t.Logf("set-get 則が破れると coassociativity 則も破れる")
}

// レンズ → 余代数 → レンズ と、余代数 → レンズ → 余代数 の両方向で戻る。
func TestLensCoalgebraRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		d := Doc{Title: rapid.String().Draw(t, "title")}
		probes := rapid.SliceOfN(rapid.String(), 1, 4).Draw(t, "probes")
		l := titleLens()

		// Lens → Coalgebra → Lens
		back := CoalgebraToLens(LensToCoalgebra(l))
		if back.Get(d) != l.Get(d) {
			t.Fatalf("Get が一致しない")
		}
		for _, v := range probes {
			if !eqDoc(back.Set(d, v), l.Set(d, v)) {
				t.Fatalf("Set が一致しない（%q）", v)
			}
		}

		// Coalgebra → Lens → Coalgebra。関数を含むので probe で外延的に比べる。
		coalg := LensToCoalgebra(l)
		round := LensToCoalgebra(CoalgebraToLens(coalg))
		if coalg(d).Pos != round(d).Pos {
			t.Fatalf("Pos が一致しない")
		}
		for _, v := range probes {
			if !eqDoc(coalg(d).Peek(v), round(d).Peek(v)) {
				t.Fatalf("Peek が一致しない（%q）", v)
			}
		}
	})
}

// coassocHolds の限界を実演する。probe が1個だと、set-set を破るレンズが通ってしまう。
// 検査が「反例探し」であって「証明」ではない、ということ。
func TestCoassocCheckMissesViolationsWithOneProbe(t *testing.T) {
	l := brokenSetSet()
	d := Doc{Title: "x"}

	if eqDoc(l.Set(l.Set(d, "a"), "b"), l.Set(d, "b")) {
		t.Fatalf("set-set は破れているはず")
	}
	if !coassocHolds(l, d, []string{"a"}) {
		t.Fatalf("probe 1個では通ってしまうはず（この検査の限界の実演）")
	}
	if coassocHolds(l, d, []string{"a", "b"}) {
		t.Fatalf("probe 2個なら検出できるはず")
	}
	t.Logf("probe 1個では見逃し、2個で検出。coassocHolds は反例探しであって判定器ではない")
}
