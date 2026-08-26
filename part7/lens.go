// Package part7 は連載「Goで書く実践圏論」発展編3本目の検証コード。
//
// レンズ（Lens）は「大きい構造 S の中の小さい部分 A」への焦点を、
// get と set の組で表したもの。合成できるので、入れ子の奥へ届く。
//
// このパッケージが測るのは、レンズ則が Go で何を守っているか。
// Go の struct は代入でコピーされるが、その中の slice と map はコピーされない。
// だから「素直に書いた set」は、法則を静かに破る。
package part7

// Lens は S の中の A に焦点を合わせる。get で取り出し、set で差し替えた S を返す。
// set が S を返す（*S を取らない）のは、非破壊であることを型で示すため。
type Lens[S, A any] struct {
	Get func(S) A
	Set func(S, A) S
}

// Compose は S→A と A→B のレンズを繋いで S→B にする。
// これがあるからレンズは入れ子の奥に届く。射の合成そのもの。
func Compose[S, A, B any](outer Lens[S, A], inner Lens[A, B]) Lens[S, B] {
	return Lens[S, B]{
		Get: func(s S) B { return inner.Get(outer.Get(s)) },
		Set: func(s S, b B) S { return outer.Set(s, inner.Set(outer.Get(s), b)) },
	}
}

// Modify は焦点だけに関数を適用する。get して f を通して set するだけ。
func Modify[S, A any](l Lens[S, A], s S, f func(A) A) S {
	return l.Set(s, f(l.Get(s)))
}

// Identity は S 全体に焦点を合わせるレンズ。合成の単位射にあたる。
func Identity[S any]() Lens[S, S] {
	return Lens[S, S]{
		Get: func(s S) S { return s },
		Set: func(_ S, s S) S { return s },
	}
}
