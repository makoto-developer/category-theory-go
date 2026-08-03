package part2

import "net/http"

// Middleware は Handler を Handler に写す。対象も射も Handler の世界に閉じているので、
// 圏から自分自身への関手（エンドファンクタ）にあたる。
type Middleware func(http.Handler) http.Handler

// Chain は middleware を左から順に適用する。ms[0] が最も外側になる。
// 引数ゼロなら素通し（恒等）になり、これが単位元として振る舞う。
func Chain(ms ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(ms) - 1; i >= 0; i-- {
			next = ms[i](next)
		}
		return next
	}
}

// Tap は通過した印を order に追記するだけの middleware。実行順の検証に使う。
func Tap(label string, order *[]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, label)
			next.ServeHTTP(w, r)
			*order = append(*order, label+":after")
		})
	}
}

// Passthrough は何もしない middleware。Chain における恒等射。
func Passthrough(next http.Handler) http.Handler { return next }
