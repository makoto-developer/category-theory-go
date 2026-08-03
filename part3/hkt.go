package part3

// Go には高階カインド多相が無い。つまり「型を取って型を返す型」F を、
// F のまま型パラメータにできない。書きたくても書けないのはこういう宣言:
//
//	type Functor[F[_] any] interface {
//		Map[A, B any](fa F[A], f func(A) B) F[B]
//	}
//
// 以下は、それを諦めたうえでの3通りの凌ぎ方である。

// --- 回避策1: 型を消してインターフェースに載せる ---

// AnyFunctor は要素の型を any に潰して「map できるもの」を表す。
// F を抽象化できない代わりに、A と B を抽象化しないことで辻褄を合わせる。
type AnyFunctor interface {
	MapAny(f func(any) any) AnyFunctor
}

// AnySlice は AnyFunctor を満たすスライス。
type AnySlice []any

func (s AnySlice) MapAny(f func(any) any) AnyFunctor {
	out := make(AnySlice, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}

// AnyOption は AnyFunctor を満たす「あるかもしれない値」。
type AnyOption struct {
	Value any
	Some  bool
}

func (o AnyOption) MapAny(f func(any) any) AnyFunctor {
	if !o.Some {
		return o
	}
	return AnyOption{Value: f(o.Value), Some: true}
}

// --- 回避策2: 具体的な型ごとに同じ関数を書く ---

// MapSliceGeneric は要素型だけをジェネリックにする。実用上いちばん素直な方法で、
// 型安全もアロケーションも問題ない。ただし F の種類だけ関数が増える。
func MapSliceGeneric[A, B any](xs []A, f func(A) B) []B {
	out := make([]B, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}

// Option はジェネリックな「あるかもしれない値」。
type Option[T any] struct {
	Value T
	Some  bool
}

// MapOption は Option 用の map。MapSliceGeneric と同じことを、型ごとに書き直している。
func MapOption[A, B any](o Option[A], f func(A) B) Option[B] {
	if !o.Some {
		return Option[B]{}
	}
	return Option[B]{Value: f(o.Value), Some: true}
}

// --- 回避策3: map の実装を値として持ち回る（辞書渡し） ---

// FunctorDict は F[A] と F[B] を別々の型パラメータに分けることで、
// 高階カインドなしに「どの Functor で写すか」を引数にする。
type FunctorDict[FA, FB, A, B any] struct {
	Map func(FA, func(A) B) FB
}

// SliceFunctor はスライス用の辞書。
func SliceFunctor[A, B any]() FunctorDict[[]A, []B, A, B] {
	return FunctorDict[[]A, []B, A, B]{Map: MapSliceGeneric[A, B]}
}

// OptionFunctor は Option 用の辞書。
func OptionFunctor[A, B any]() FunctorDict[Option[A], Option[B], A, B] {
	return FunctorDict[Option[A], Option[B], A, B]{Map: MapOption[A, B]}
}

// MapTwiceWith は辞書を受け取って「2回写す」処理を Functor によらず書く。
// これが高階カインドの代わりにできる範囲の上限である。
func MapTwiceWith[FA, FB, FC, A, B, C any](
	d1 FunctorDict[FA, FB, A, B],
	d2 FunctorDict[FB, FC, B, C],
	fa FA, f func(A) B, g func(B) C,
) FC {
	return d2.Map(d1.Map(fa, f), g)
}

// メソッドに型パラメータを付けられない、という制約もある。書けないのはこの形:
//
//	func (o Option[A]) Map[B any](f func(A) B) Option[B]
//
// これが書ければ o.Map(f).Map(g) と繋げられるが、Go では関数のネストになる。
// 理由は golang/go#49085 にある通り、インターフェース充足の判定が難しくなるため。
