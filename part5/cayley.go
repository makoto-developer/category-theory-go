package part5

import "strings"

// Cayley 表現とは、モノイド M を End(M)（M 上の自己射のモノイド）へ
//
//	m ↦ (m · -)
//
// で埋め込むこと。単射で、合成は関数合成そのものになる。
//
// 圏論が保証するのはここまで、つまり「埋め込んで畳んで戻しても意味は変わらない」ことだけで、
// 速さについては何も言っていない。速さを決めるのは Append のコストがどちら側に付いているか。
// このパッケージは同じ変換を3つのモノイドに掛けて、その差だけを見る。

// Monoid は単位元と結合的な二項演算の組。法則は cayley_test.go で確かめる。
type Monoid[M any] struct {
	Empty  M
	Append func(M, M) M
}

// Cayley は m を「左から m を足す関数」に持ち上げる。
func Cayley[M any](mo Monoid[M], m M) func(M) M {
	return func(rest M) M { return mo.Append(m, rest) }
}

// FoldNaive は左から順に畳む。((x1·x2)·x3)·… の形になる。
func FoldNaive[M any](mo Monoid[M], xs []M) M {
	acc := mo.Empty
	for _, x := range xs {
		acc = mo.Append(acc, x)
	}
	return acc
}

// FoldCayley は End(M) の上で畳んでから単位元に適用する。
// 合成の時点では Append を1回も呼ばず、適用時に x1·(x2·(x3·…)) の形で払う。
func FoldCayley[M any](mo Monoid[M], xs []M) M {
	f := func(rest M) M { return rest }
	for _, x := range xs {
		prev, g := f, Cayley(mo, x)
		f = func(rest M) M { return prev(g(rest)) }
	}
	return f(mo.Empty)
}

// --- 対象となる3つのモノイド -------------------------------------------

// StringMonoid の Append は両辺の長さぶんコピーする。左右どちらに寄せても総コピー量は変わらない。
var StringMonoid = Monoid[string]{
	Empty:  "",
	Append: func(a, b string) string { return a + b },
}

// --- Go で普通に書く形。Cayley 表現の比較対象になる -----------------------

// BuildString は可変バッファに書き足す。Append の結合の向きとは無関係に O(n) で済む。
func BuildString(xs []string) string {
	var b strings.Builder
	for _, x := range xs {
		b.WriteString(x)
	}
	return b.String()
}

// BuildStringGrow は最終長が分かっている場合。再確保が消える。
func BuildStringGrow(xs []string, total int) string {
	var b strings.Builder
	b.Grow(total)
	for _, x := range xs {
		b.WriteString(x)
	}
	return b.String()
}
