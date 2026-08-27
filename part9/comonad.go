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
//
// **余モナドなのは「有効な焦点を持つ非空の Zipper」に限る。**
// Items が空、または Pos が範囲外だと Extract が panic するので、
// counit 則 Extract(Duplicate(w)) = w をそもそも評価できない。
// この型全体が余モナドなのではなく、Valid が真になる部分についてだけ成り立つ。
//
// もうひとつ前提がある。Items は不変として扱う。共有しているスライスを
// 呼び出し側が書き換えると、法則の両辺が値として一致していても
// エイリアスの違いを観測できてしまう。
type Zipper[A any] struct {
	Items []A
	Pos   int
}

// Valid は余モナドとして扱ってよい形かを見る。
func (w Zipper[A]) Valid() bool { return len(w.Items) > 0 && w.Pos >= 0 && w.Pos < len(w.Items) }

// Extract は焦点の値を取り出す。Monad の return を裏返したもの（counit）。
// Valid でない Zipper に対しては定義されない（panic する）。
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

// MovingAverage は焦点の前後 half 要素を見る移動平均。窓幅は 2*half+1 で、
// 端では窓が切り詰められる。焦点の周りだけを見るので、Zipper を受け取って
// 1つの値を返す形——つまり extend に渡せる局所規則になる。
//
// 引数を「窓幅 n」ではなく half にしてあるのは、以前 n/2 で half を作っていて
// 偶数の n が窓幅 n にならなかったため（n=2 で3要素を平均していた）。
func MovingAverage(half int) func(Zipper[float64]) float64 {
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
func MovingAverageLoop(xs []float64, half int) []float64 {
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
