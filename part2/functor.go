// Package part2 は連載「Goで書く実践圏論」第2回の検証コード。
// Go の標準ライブラリと日常のコードに現れる Functor・Monoid・Kleisli 合成・
// 自然変換を取り出し、それぞれの法則が成り立つことを確かめる。
package part2

import "iter"

// MapSlice はスライス上の Functor。要素に射を適用し、長さと順序という構造は保つ。
func MapSlice[A, B any](xs []A, f func(A) B) []B {
	if xs == nil {
		return nil
	}
	out := make([]B, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}

// MapPtr はポインタ上の Functor。値が無いという構造を保つため nil は nil に写す。
func MapPtr[A, B any](p *A, f func(A) B) *B {
	if p == nil {
		return nil
	}
	b := f(*p)
	return &b
}

// MapSeq は iter.Seq 上の Functor。遅延のまま写すので、途中で break されれば残りは評価されない。
func MapSeq[A, B any](seq iter.Seq[A], f func(A) B) iter.Seq[B] {
	return func(yield func(B) bool) {
		for a := range seq {
			if !yield(f(a)) {
				return
			}
		}
	}
}

// MapErr は (T, error) 上の Functor。エラーはそのまま素通しする。
func MapErr[A, B any](a A, err error, f func(A) B) (B, error) {
	if err != nil {
		var zero B
		return zero, err
	}
	return f(a), nil
}

// Compose は第1回と同じ合成。Functor 則の検証に使う。
func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C { return g(f(a)) }
}

// Identity は恒等射。
func Identity[A any](a A) A { return a }
