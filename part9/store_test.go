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
// Store は関数を含むので直接比較できない。位置 a を振って値で確かめる。
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
		probes := rapid.SliceOfN(rapid.String(), 1, 3).Draw(t, "probes")
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
	l := brokenSetSet()
	d := Doc{Title: "x"}

	// 破れているのは set-set だけであることを確認する
	if !eqDoc(l.Set(d, l.Get(d)), d) {
		t.Fatalf("get-set は保たれているはず")
	}
	if l.Get(l.Set(d, "a")) != "a" {
		t.Fatalf("set-get は保たれているはず")
	}
	if eqDoc(l.Set(l.Set(d, "a"), "b"), l.Set(d, "b")) {
		t.Fatalf("set-set は破れているはず")
	}
	if coassocHolds(l, d, []string{"a", "b"}) {
		t.Fatalf("set-set が破れているのに coassociativity が成り立ってしまった")
	}
	t.Logf("set-set 則が破れると coassociativity 則も破れる")
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

// レンズ → 余代数 → レンズ で元に戻る。
func TestLensCoalgebraRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		d := Doc{Title: rapid.String().Draw(t, "title")}
		v := rapid.String().Draw(t, "v")
		l := titleLens()
		back := CoalgebraToLens(LensToCoalgebra(l))

		if back.Get(d) != l.Get(d) {
			t.Fatalf("Get が一致しない")
		}
		if !eqDoc(back.Set(d, v), l.Set(d, v)) {
			t.Fatalf("Set が一致しない")
		}
	})
}
