package part3

import "iter"

// Cont は継続渡し形式の値。「A を持っている」ことを「A を受け取る関数を受け取れる」で表す。
// iter.Seq[V] = func(yield func(V) bool) も、打ち切り可能な継続を取るこの形をしている。
type Cont[R, A any] func(func(A) R) R

// ToCont は値を継続渡しに変える。この向きの変換はいつでもできる。
func ToCont[R, A any](a A) Cont[R, A] {
	return func(k func(A) R) R { return k(a) }
}

// FromCont は継続渡しから値を取り出す。継続に恒等射を渡せばよい。
// 行き先の型 R を A に固定できるときだけ書ける（米田の補題が言っているのはこの往復のこと）。
func FromCont[A any](c Cont[A, A]) A {
	return c(func(a A) A { return a })
}

// MapCont は継続渡しの上の Functor。継続の側を先に加工する。
func MapCont[R, A, B any](c Cont[R, A], f func(A) B) Cont[R, B] {
	return func(k func(B) R) R {
		return c(func(a A) R { return k(f(a)) })
	}
}

// SeqOf はスライスを push 型の列にする（slices.Values と同じ）。
func SeqOf[A any](xs []A) iter.Seq[A] {
	return func(yield func(A) bool) {
		for _, x := range xs {
			if !yield(x) {
				return
			}
		}
	}
}

// SumSeq は push 型の列を畳み込む。呼び出し側が制御を握る形。
func SumSeq(seq iter.Seq[int]) int {
	total := 0
	for v := range seq {
		total += v
	}
	return total
}

// SumPull は iter.Pull で pull 型に変換してから畳み込む。
// push と pull は相互に変換できるが、変換にはコルーチンの切り替えが要る。
func SumPull(seq iter.Seq[int]) int {
	next, stop := iter.Pull(seq)
	defer stop()

	total := 0
	for {
		v, ok := next()
		if !ok {
			return total
		}
		total += v
	}
}
