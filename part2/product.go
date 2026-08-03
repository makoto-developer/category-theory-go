package part2

// Pair は積。struct がこれにあたる。
type Pair[A, B any] struct {
	First  A
	Second B
}

// Fanout は積の普遍性そのもの。C から A への射と C から B への射があれば、
// C から Pair[A, B] への射がただ一つ決まる。
func Fanout[C, A, B any](f func(C) A, g func(C) B) func(C) Pair[A, B] {
	return func(c C) Pair[A, B] { return Pair[A, B]{First: f(c), Second: g(c)} }
}

// Either は余積（直和）。Go に直和型が無いのでタグ付きの構造体で表す。
type Either[A, B any] struct {
	isRight bool
	left    A
	right   B
}

// Left は余積の左側への入射。
func Left[A, B any](a A) Either[A, B] { return Either[A, B]{left: a} }

// Right は余積の右側への入射。
func Right[A, B any](b B) Either[A, B] { return Either[A, B]{isRight: true, right: b} }

// Fanin は余積の普遍性。A から C への射と B から C への射があれば、
// Either[A, B] から C への射がただ一つ決まる。
// 分岐が構造体の中に閉じているので、case の書き漏らしが起こりえない。
func Fanin[A, B, C any](onLeft func(A) C, onRight func(B) C) func(Either[A, B]) C {
	return func(e Either[A, B]) C {
		if e.isRight {
			return onRight(e.right)
		}
		return onLeft(e.left)
	}
}

// Swap は Either[A, B] と Either[B, A] を行き来する。往復すると元に戻る（同型）。
func Swap[A, B any](e Either[A, B]) Either[B, A] {
	if e.isRight {
		return Left[B, A](e.right)
	}
	return Right[B, A](e.left)
}

// Curry は f(Pair[A, B]) C を f(A)(B) C に変える。積と冪の随伴の片側。
func Curry[A, B, C any](f func(Pair[A, B]) C) func(A) func(B) C {
	return func(a A) func(B) C {
		return func(b B) C { return f(Pair[A, B]{First: a, Second: b}) }
	}
}

// Uncurry は Curry の逆向き。往復すると元の射に戻る。
func Uncurry[A, B, C any](f func(A) func(B) C) func(Pair[A, B]) C {
	return func(p Pair[A, B]) C { return f(p.First)(p.Second) }
}
