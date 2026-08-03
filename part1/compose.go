// Package part1 は連載「Goで書く実践圏論」第1回の検証コード。
// 型を対象、func(A) B を射とみなしたとき、圏の公理（結合律・単位律）が
// 本当に成り立つのかをテストで確かめる。
package part1

// Compose は f のあとに g を適用する射を返す。数学の記法では g∘f にあたる。
func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C { return g(f(a)) }
}

// Identity は恒等射。合成における単位元になる。
func Identity[A any](a A) A { return a }

// Pipe は同じ型の上の射を左から順に合成する。
// 左畳み込みで書けるのは合成が結合的だから（右畳み込みでも結果は変わらない）。
func Pipe[A any](fs ...func(A) A) func(A) A {
	out := Identity[A]
	for _, f := range fs {
		out = Compose(out, f)
	}
	return out
}
