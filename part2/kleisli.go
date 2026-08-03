package part2

import "context"

// Kleisli は「失敗しうる射」。Go のコードのほとんどはこの形をしている。
type Kleisli[A, B any] func(A) (B, error)

// Then は Kleisli 射の合成。メソッドにしたいところだが、Go はメソッドに型パラメータを
// 追加できない（型 C を導入できない）ため、関数として書くしかない。
func Then[A, B, C any](f Kleisli[A, B], g Kleisli[B, C]) Kleisli[A, C] {
	return func(a A) (C, error) {
		b, err := f(a)
		if err != nil {
			var zero C
			return zero, err
		}
		return g(b)
	}
}

// Pure は普通の射を Kleisli 射に持ち上げる。失敗しない処理を混ぜるときに使う。
func Pure[A, B any](f func(A) B) Kleisli[A, B] {
	return func(a A) (B, error) { return f(a), nil }
}

// KleisliIdentity は Kleisli 圏の恒等射。
func KleisliIdentity[A any](a A) (A, error) { return a, nil }

// Step は context を取る Kleisli 射。実務のパイプラインはたいていこの形になる。
type Step[A, B any] func(context.Context, A) (B, error)

// ChainStep は Step の合成。context のキャンセルを各段で確認する。
func ChainStep[A, B, C any](f Step[A, B], g Step[B, C]) Step[A, C] {
	return func(ctx context.Context, a A) (C, error) {
		var zero C
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		b, err := f(ctx, a)
		if err != nil {
			return zero, err
		}
		return g(ctx, b)
	}
}
