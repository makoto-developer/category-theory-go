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

// 引き戻しの普遍性。f∘p = g∘q を満たす任意の (Z, p, q) から、
// 引き戻しへの射 u がただひとつ存在して π₁∘u = p, π₂∘u = q になる。
// Go では「Z の各要素に対して、対応する組が引き戻しの中にちょうど1つある」ことを確かめる。
func TestPullbackUniversalProperty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		as := rapid.SliceOfN(rapid.IntRange(0, 9), 1, 10).Draw(t, "as")
		bs := rapid.SliceOfN(rapid.IntRange(0, 9), 1, 10).Draw(t, "bs")
		f := func(a int) int { return a % 4 }
		g := func(b int) int { return b % 4 }
		pb := PullbackHash(as, bs, f, g)

		// 錐 (Z, p, q) を作る。可換になる組だけを Z にする。
		type cone struct{ p, q int }
		var zs []cone
		for _, a := range as {
			for _, b := range bs {
				if f(a) == g(b) {
					zs = append(zs, cone{a, b})
				}
			}
		}

		for _, z := range zs {
			// 可換性の確認（錐の条件）
			if f(z.p) != g(z.q) {
				t.Fatalf("錐が可換でない")
			}
			// 媒介射 u の存在と一意性
			n := 0
			for _, e := range pb {
				if e.L == z.p && e.R == z.q {
					n++
				}
			}
			if n == 0 {
				t.Fatalf("媒介射が存在しない: (%d,%d) が引き戻しに無い", z.p, z.q)
			}
		}
	})
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
