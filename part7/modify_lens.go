package part7

// LensM は Get/Set ではなく Get/Mod（焦点に関数を適用する）でレンズを表したもの。
// 圏論的な中身は Lens と同じだが、合成したときの複製回数が変わる。
//
// Lens の Compose は outer.Set(s, inner.Set(outer.Get(s), b)) なので、
// 中間の段が「渡された値は自分のものか分からない」まま Set を受け取る。
// LensM では中間の段が関数を内側へ渡すだけなので、複製する理由が無くなる。
type LensM[S, A any] struct {
	Get func(S) A
	Mod func(S, func(A) A) S
}

// SetM は Mod で書ける。差し替えは「元を無視して返す関数」を適用すること。
func SetM[S, A any](l LensM[S, A], s S, a A) S {
	return l.Mod(s, func(A) A { return a })
}

// ComposeM は関数を内側へ渡していくだけ。中間で値を組み直さない。
func ComposeM[S, A, B any](outer LensM[S, A], inner LensM[A, B]) LensM[S, B] {
	return LensM[S, B]{
		Get: func(s S) B { return inner.Get(outer.Get(s)) },
		Mod: func(s S, f func(B) B) S {
			return outer.Mod(s, func(a A) A { return inner.Mod(a, f) })
		},
	}
}
