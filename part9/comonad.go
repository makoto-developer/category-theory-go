// Package part9 は連載「Goで書く実践圏論」発展編5本目の検証コード。
//
// 第3回で Functor・Applicative・Monad を積み上げた。Monad の矢印を裏返すと
// Comonad（余モナド）になる。Monad が「値を文脈に入れる」なら、
// Comonad は「文脈から値を取り出す」。
//
// このパッケージが測るのは、Comonad 則が要求する duplicate を素直に実体化すると
// 何が起きるか、そしてレンズが実は Store 余モナドの余代数だったこと。
package part9

// Zipper は焦点つきのリスト。Items の Pos 番目を「いま見ている場所」とする。
// Comonad の典型例で、移動平均のような窓処理がそのまま extend になる。
type Zipper[A any] struct {
	Items []A
	Pos   int
}

// Extract は焦点の値を取り出す。Monad の return を裏返したもの（counit）。
func Extract[A any](w Zipper[A]) A { return w.Items[w.Pos] }

// Duplicate は「各位置に、そこへ焦点を移した自分自身を置く」。
// Monad の join を裏返したもの。定義どおり実体化すると n 個の Zipper ができる。
func Duplicate[A any](w Zipper[A]) Zipper[Zipper[A]] {
	inner := make([]Zipper[A], len(w.Items))
	for i := range w.Items {
		inner[i] = Zipper[A]{Items: w.Items, Pos: i}
	}
	return Zipper[Zipper[A]]{Items: inner, Pos: w.Pos}
}

// FMap は Zipper の Functor 部分。
func FMap[A, B any](f func(A) B, w Zipper[A]) Zipper[B] {
	out := make([]B, len(w.Items))
	for i, a := range w.Items {
		out[i] = f(a)
	}
	return Zipper[B]{Items: out, Pos: w.Pos}
}

// Extend は「各位置で、その位置から見た文脈全体に f を適用する」。
// 中間の Zipper[Zipper[A]] を作らずに直接回す。
func Extend[A, B any](f func(Zipper[A]) B, w Zipper[A]) Zipper[B] {
	out := make([]B, len(w.Items))
	for i := range w.Items {
		out[i] = f(Zipper[A]{Items: w.Items, Pos: i})
	}
	return Zipper[B]{Items: out, Pos: w.Pos}
}

// ExtendViaDuplicate は法則が言う定義そのもの: extend f = fmap f ∘ duplicate。
// 上の Extend と同じ答えを返すが、中間構造を実体化する。
func ExtendViaDuplicate[A, B any](f func(Zipper[A]) B, w Zipper[A]) Zipper[B] {
	return FMap(f, Duplicate(w))
}

// MovingAverage は窓幅 n の移動平均。焦点の周りだけを見るので、
// Zipper を受け取って1つの値を返す形——つまり extend に渡せる形になる。
func MovingAverage(n int) func(Zipper[float64]) float64 {
	half := n / 2
	return func(w Zipper[float64]) float64 {
		lo, hi := max(0, w.Pos-half), min(len(w.Items)-1, w.Pos+half)
		sum := 0.0
		for i := lo; i <= hi; i++ {
			sum += w.Items[i]
		}
		return sum / float64(hi-lo+1)
	}
}

// MovingAverageLoop は同じ計算を手書きのループで。抽象化の代金の基準線。
func MovingAverageLoop(xs []float64, n int) []float64 {
	half := n / 2
	out := make([]float64, len(xs))
	for p := range xs {
		lo, hi := max(0, p-half), min(len(xs)-1, p+half)
		sum := 0.0
		for i := lo; i <= hi; i++ {
			sum += xs[i]
		}
		out[p] = sum / float64(hi-lo+1)
	}
	return out
}
