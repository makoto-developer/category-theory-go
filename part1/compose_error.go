package part1

// ComposeE は error を返す射どうしを繋ぐ。Compose では戻り値の型が合わないため、
// error の受け渡しを合成器の側に閉じ込める。第2回で扱う Kleisli 合成の先取り。
func ComposeE[A, B, C any](f func(A) (B, error), g func(B) (C, error)) func(A) (C, error) {
	return func(a A) (C, error) {
		b, err := f(a)
		if err != nil {
			var zero C
			return zero, err
		}
		return g(b)
	}
}

// IdentityE は ComposeE における恒等射。
func IdentityE[A any](a A) (A, error) { return a, nil }
