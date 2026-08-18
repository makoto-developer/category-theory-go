package part6

// hylomorphism は unfold してから fold すること。間に立つ構造は、
// 作られてすぐ畳まれるので、理屈のうえでは要らない（deforestation）。
// 要らないものが Go で本当に消えるのかを、ここで測る。
//
// ここの Hylo* は一般の関手に対する hylomorphism ではなく、リスト関手の
// anamorphism と左畳み込み（step は func(B, A) B なので foldl 型）を繋いだ限定版。
// 下のマージソートは別の関手（F(X) = 葉 + X×X）で、SplitAna / MergeCata を直接書いている。

// HyloVia は素直に書いた版。中間のスライスを実際に作ってから畳む。
func HyloVia[S, A, B any](co Coalgebra[S, A], step func(B, A) B, zero B, seed S) B {
	acc := zero
	for _, a := range Unfold(co, seed) {
		acc = step(acc, a)
	}
	return acc
}

// HyloFused は中間を作らない版。展開しながらその場で畳む。
// 手で書く deforestation にあたる。
func HyloFused[S, A, B any](co Coalgebra[S, A], step func(B, A) B, zero B, seed S) B {
	acc := zero
	for {
		a, next, ok := co(seed)
		if !ok {
			return acc
		}
		acc = step(acc, a)
		seed = next
	}
}

// HyloSeq は iter.Seq を挟む版。スライスは作らないが、継続渡しの往復が挟まる。
func HyloSeq[S, A, B any](co Coalgebra[S, A], step func(B, A) B, zero B, seed S) B {
	acc := zero
	for a := range UnfoldSeq(co, seed) {
		acc = step(acc, a)
	}
	return acc
}

// --- マージソート —— 教科書的な hylomorphism ---------------------------

// SplitTree は分割の木。unfold で作り、fold でマージして畳む。
type SplitTree struct {
	Leaf []int
	L, R *SplitTree
}

// SplitAna は木を生やす（anamorphism）。要素1個以下になったら葉。
func SplitAna(xs []int) *SplitTree {
	if len(xs) <= 1 {
		return &SplitTree{Leaf: xs}
	}
	mid := len(xs) / 2
	return &SplitTree{L: SplitAna(xs[:mid]), R: SplitAna(xs[mid:])}
}

// MergeCata は木を畳む（catamorphism）。葉はそのまま、節はマージ。
// SplitAna が作った木だけを受け取る前提（nil や片側だけの節は想定しない）。
func MergeCata(t *SplitTree) []int {
	if t == nil {
		return nil
	}
	if t.L == nil {
		return t.Leaf
	}
	return merge(MergeCata(t.L), MergeCata(t.R))
}

// MergeSortHylo は「木を作ってから畳む」。中間の木がヒープに載る。
func MergeSortHylo(xs []int) []int { return MergeCata(SplitAna(xs)) }

// MergeSortFused は木を作らない。普通に書くマージソートは、
// hylomorphism を手で融合したものだった、というのがこの節の眼目。
func MergeSortFused(xs []int) []int {
	if len(xs) <= 1 {
		return xs
	}
	mid := len(xs) / 2
	return merge(MergeSortFused(xs[:mid]), MergeSortFused(xs[mid:]))
}

func merge(a, b []int) []int {
	out := make([]int, 0, len(a)+len(b))
	for len(a) > 0 && len(b) > 0 {
		if a[0] <= b[0] {
			out = append(out, a[0])
			a = a[1:]
		} else {
			out = append(out, b[0])
			b = b[1:]
		}
	}
	out = append(out, a...)
	out = append(out, b...)
	return out
}
