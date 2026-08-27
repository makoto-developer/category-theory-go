package part9

// Store は「位置 S と、位置から値を引く関数」の組。余モナドになる。
// 第3回で見た Reader (S → A) の親戚だが、位置を一緒に持ち歩くところが違う。
type Store[S, A any] struct {
	Peek func(S) A
	Pos  S
}

// StoreExtract は counit。いまの位置の値を引く。
func StoreExtract[S, A any](st Store[S, A]) A { return st.Peek(st.Pos) }

// StoreDuplicate は「各位置に、そこへ移した自分自身を置く」。
func StoreDuplicate[S, A any](st Store[S, A]) Store[S, Store[S, A]] {
	return Store[S, Store[S, A]]{
		Peek: func(s S) Store[S, A] { return Store[S, A]{Peek: st.Peek, Pos: s} },
		Pos:  st.Pos,
	}
}

// Lens は part7 と同じ形。ここでは余代数との対応を見るために再掲する。
type Lens[S, A any] struct {
	Get func(S) A
	Set func(S, A) S
}

// LensToCoalgebra はレンズを Store 余モナドの余代数 S → Store[A,S] に移す。
//
// この対応のもとで、余代数の counit 則が get-set 則に、
// coassociativity 則が set-get 則と set-set 則に対応する。
func LensToCoalgebra[S, A any](l Lens[S, A]) func(S) Store[A, S] {
	return func(s S) Store[A, S] {
		return Store[A, S]{
			Peek: func(a A) S { return l.Set(s, a) },
			Pos:  l.Get(s),
		}
	}
}

// CoalgebraToLens は逆向き。余代数からレンズを復元する。
func CoalgebraToLens[S, A any](coalg func(S) Store[A, S]) Lens[S, A] {
	return Lens[S, A]{
		Get: func(s S) A { return coalg(s).Pos },
		Set: func(s S, a A) S { return coalg(s).Peek(a) },
	}
}
