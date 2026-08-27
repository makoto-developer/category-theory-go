package part8

import (
	"fmt"
	"slices"
	"testing"

	"pgregory.net/rapid"
)

func key(p Pair[int, int]) string { return fmt.Sprintf("%d/%d", p.L, p.R) }

func sortedKeys(ps []Pair[int, int]) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, key(p))
	}
	slices.Sort(out)
	return out
}

// 作り方が3つあっても、できる対象は同じ。極限は一意なので当然だが、
// 「当然」で済ませずに確かめる。
func TestAllPullbackConstructionsAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		as := rapid.SliceOfN(rapid.IntRange(0, 5), 0, 12).Draw(t, "as")
		bs := rapid.SliceOfN(rapid.IntRange(0, 5), 0, 12).Draw(t, "bs")
		f := func(a int) int { return a % 3 }
		g := func(b int) int { return b % 3 }

		nested := sortedKeys(PullbackNested(as, bs, f, g))
		hash := sortedKeys(PullbackHash(as, bs, f, g))
		viaProd := sortedKeys(PullbackViaProductAndEqualizer(as, bs, f, g))

		if !slices.Equal(nested, hash) {
			t.Fatalf("Nested と Hash が違う: %v vs %v", nested, hash)
		}
		if !slices.Equal(nested, viaProd) {
			t.Fatalf("Nested と 積+イコライザ が違う: %v vs %v", nested, viaProd)
		}
	})
}

// 引き戻しの普遍性を、存在と一意性の両方について検査する。
//
// 前の版は存在しか見ておらず、しかも錐を「引き戻しに入る組そのもの」から
// 作っていたので、実質は包含検査だった。要素を ID で区別し、任意の Z と
// p: Z→A, q: Z→B を生成して可換なものだけを錐として使い、媒介射が
// ちょうど1つであることを数で確かめる形に直した。
type Elem struct {
	ID  int
	Key int
}

func elemKey(e Elem) int { return e.Key }

// 値が重複しても要素として区別できるよう、ID を振った集合を作る。
func genElems(prefix int, keys int) *rapid.Generator[[]Elem] {
	return rapid.Custom(func(t *rapid.T) []Elem {
		ks := rapid.SliceOfN(rapid.IntRange(0, keys-1), 0, 8).Draw(t, "keys")
		out := make([]Elem, len(ks))
		for i, k := range ks {
			out[i] = Elem{ID: prefix*1000 + i, Key: k}
		}
		return out
	})
}

func TestPullbackUniversalProperty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		as := genElems(1, 4).Draw(t, "as")
		bs := genElems(2, 4).Draw(t, "bs")
		if len(as) == 0 || len(bs) == 0 {
			return
		}
		pb := PullbackHash(as, bs, elemKey, elemKey)

		// 任意の錐 (Z, p, q) を作る。Z の各点で p と q を独立に選び、
		// 可換な点だけを残す（残った部分が錐になる）。
		n := rapid.IntRange(0, 10).Draw(t, "z")
		type cone struct{ p, q Elem }
		var zs []cone
		for i := 0; i < n; i++ {
			a := as[rapid.IntRange(0, len(as)-1).Draw(t, "pi")]
			b := bs[rapid.IntRange(0, len(bs)-1).Draw(t, "qi")]
			if elemKey(a) == elemKey(b) {
				zs = append(zs, cone{a, b})
			}
		}

		for _, z := range zs {
			matches := 0
			for _, e := range pb {
				if e.L.ID == z.p.ID && e.R.ID == z.q.ID {
					matches++
				}
			}
			if matches == 0 {
				t.Fatalf("媒介射が存在しない: (ID %d, ID %d) が引き戻しに無い", z.p.ID, z.q.ID)
			}
			if matches > 1 {
				t.Fatalf("媒介射が一意でない: (ID %d, ID %d) が %d 個ある", z.p.ID, z.q.ID, matches)
			}
		}
	})
}

// 要素を ID で区別しないと、引き戻しの一意性は崩れる。
// SQL の JOIN が多重集合を返すのはこちら側で、集合の圏の引き戻しとはズレる。
func TestBagJoinIsNotASetPullback(t *testing.T) {
	// 値としては同じ 1 が2件ずつ。ID を持たない int で引き戻すと……
	as := []int{1, 1}
	bs := []int{1, 1}
	id := func(x int) int { return x }
	got := PullbackHash(as, bs, id, id)

	if len(got) != 4 {
		t.Fatalf("多重集合としては 2×2 = 4 件になるはず: %d 件", len(got))
	}
	// 集合 {(a,b) | f(a)=g(b)} としては (1,1) の1点しかない。
	distinct := map[Pair[int, int]]bool{}
	for _, p := range got {
		distinct[p] = true
	}
	if len(distinct) != 1 {
		t.Fatalf("集合としては1点のはず: %d 点", len(distinct))
	}
	t.Logf("多重集合では %d 件、集合としては %d 点。SQL の JOIN は前者", len(got), len(distinct))
}

// イコライザは、f と g が一致するところだけを取り出す。
// 引き戻しは「積を作ってイコライザで絞る」に分解できる。
func TestPullbackDecomposesIntoProductAndEqualizer(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		as := rapid.SliceOfN(rapid.IntRange(0, 6), 0, 8).Draw(t, "as")
		bs := rapid.SliceOfN(rapid.IntRange(0, 6), 0, 8).Draw(t, "bs")
		f := func(a int) int { return a % 3 }
		g := func(b int) int { return b % 3 }

		direct := sortedKeys(PullbackHash(as, bs, f, g))
		decomposed := sortedKeys(PullbackViaProductAndEqualizer(as, bs, f, g))
		if !slices.Equal(direct, decomposed) {
			t.Fatalf("分解した結果が違う")
		}
	})
}
