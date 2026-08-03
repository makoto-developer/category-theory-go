package part1

import "sync/atomic"

// Memoize は入力ごとに結果を覚える。射が純粋なときに限り、元の射と等価な射になる。
// 記事の議論に必要な最小実装なので並行安全ではない。
func Memoize[A comparable, B any](f func(A) B) func(A) B {
	cache := make(map[A]B)
	return func(a A) B {
		if b, ok := cache[a]; ok {
			return b
		}
		b := f(a)
		cache[a] = b
		return b
	}
}

// Counter は呼ばれた回数を足して返す射を作る。同じ入力に違う結果を返すので純粋ではない。
func Counter() func(int) int {
	var n int64
	return func(x int) int { return x + int(atomic.AddInt64(&n, 1)) }
}

// Retry は f を最大 attempts 回まで呼び直す。純粋な射に対しては何度呼んでも結果が変わらない。
func Retry[A, B any](f func(A) (B, error), attempts int) func(A) (B, error) {
	return func(a A) (B, error) {
		var b B
		var err error
		for range attempts {
			if b, err = f(a); err == nil {
				return b, nil
			}
		}
		return b, err
	}
}
