// Package part8 は連載「Goで書く実践圏論」発展編4本目の検証コード。
//
// 第2回で積と余積を見た。その一般化が極限と余極限で、有限極限は
// 「積 + イコライザ」に分解できる。分解の具体形のひとつが引き戻し（pullback）で、
// これは集合の圏では INNER JOIN そのものになる。
//
// このパッケージが測るのは、圏論が「同じ答えになる」ことを保証する範囲と、
// そこから先をコストモデルが決める幅。
package part8

// Pullback は f: A→C と g: B→C の引き戻し。
// 集合の圏では { (a,b) | f(a) = g(b) } で、これは結合キーでの INNER JOIN そのもの。
type Pair[A, B any] struct {
	L A
	R B
}

// PullbackNested は定義をそのまま書いた版。全組を試すので O(n×m)。
func PullbackNested[A, B any, K comparable](as []A, bs []B, f func(A) K, g func(B) K) []Pair[A, B] {
	var out []Pair[A, B]
	for _, a := range as {
		for _, b := range bs {
			if f(a) == g(b) {
				out = append(out, Pair[A, B]{a, b})
			}
		}
	}
	return out
}

// PullbackHash は右側をキーで索引してから舐める版。O(n+m+出力)。
// 引き戻しとしては上とまったく同じ対象を作る。作り方だけが違う。
func PullbackHash[A, B any, K comparable](as []A, bs []B, f func(A) K, g func(B) K) []Pair[A, B] {
	idx := make(map[K][]B, len(bs))
	for _, b := range bs {
		k := g(b)
		idx[k] = append(idx[k], b)
	}
	var out []Pair[A, B]
	for _, a := range as {
		for _, b := range idx[f(a)] {
			out = append(out, Pair[A, B]{a, b})
		}
	}
	return out
}

// Equalizer は f と g が一致する部分だけを取り出す。
// 有限極限が「積 + イコライザ」に分解できる、のイコライザ側。
func Equalizer[A any, K comparable](as []A, f, g func(A) K) []A {
	var out []A
	for _, a := range as {
		if f(a) == g(a) {
			out = append(out, a)
		}
	}
	return out
}

// PullbackViaProductAndEqualizer は「積を作ってからイコライザで絞る」形。
// 引き戻しの定義そのものだが、積を実体化するので O(n×m) の中間データが要る。
// PullbackNested との違いは、中間の積を本当にスライスに置くかどうか。
func PullbackViaProductAndEqualizer[A, B any, K comparable](as []A, bs []B, f func(A) K, g func(B) K) []Pair[A, B] {
	product := make([]Pair[A, B], 0, len(as)*len(bs))
	for _, a := range as {
		for _, b := range bs {
			product = append(product, Pair[A, B]{a, b})
		}
	}
	return Equalizer(product,
		func(p Pair[A, B]) K { return f(p.L) },
		func(p Pair[A, B]) K { return g(p.R) })
}
