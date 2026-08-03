package part2

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var (
	sinkInts    []int
	sinkStrings []string
	sinkInt     int
	sinkString  string
)

func intsOf(n int) []int {
	xs := make([]int, n)
	for i := range xs {
		xs[i] = i
	}
	return xs
}

// Functor 則2（合成の保存）が保証しているのは、この2つが同じ結果になることである。
// 同じ結果になるなら、速いほうを選んでよい。
func BenchmarkMapTwice(b *testing.B) {
	for _, n := range []int{100, 10_000, 1_000_000} {
		xs := intsOf(n)
		b.Run("separate/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				sinkInts = MapSlice(MapSlice(xs, itoa), length)
			}
		})
		b.Run("fused/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				sinkInts = MapSlice(xs, Compose(itoa, length))
			}
		})
	}
}

// 結合律があるので分割して並列に畳み込める。実際に速くなるのはどこからか。
func BenchmarkFold(b *testing.B) {
	workers := DefaultWorkers()

	for _, n := range []int{1_000, 10_000, 30_000, 100_000, 10_000_000} {
		xs := intsOf(n)
		b.Run("sequential/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = Fold(SumInt, xs)
			}
		})
		b.Run("parallel/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = FoldParallel(SumInt, xs, workers)
			}
		})
	}
}

// 同じモノイド（連結と空文字列）でも、実装によって計算量が変わる。
func BenchmarkStringMonoid(b *testing.B) {
	parts := make([]string, 1_000)
	for i := range parts {
		parts[i] = "abcdefgh"
	}

	b.Run("concat", func(b *testing.B) {
		for b.Loop() {
			sinkString = Fold(ConcatString, parts)
		}
	})
	b.Run("builder_per_append", func(b *testing.B) {
		for b.Loop() {
			sinkString = Fold(BuildString, parts)
		}
	})
	b.Run("builder_once", func(b *testing.B) {
		for b.Loop() {
			var sb strings.Builder
			for _, p := range parts {
				sb.WriteString(p)
			}
			sinkString = sb.String()
		}
	})
	b.Run("strings_join", func(b *testing.B) {
		for b.Loop() {
			sinkString = strings.Join(parts, "")
		}
	})
}

// middleware チェーンの深さごとのコスト。1リクエストあたりで見る。
func BenchmarkMiddlewareChain(b *testing.B) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	noop := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })
	})

	for _, depth := range []int{0, 1, 4, 16} {
		ms := make([]Middleware, depth)
		for i := range ms {
			ms[i] = noop
		}
		handler := Chain(ms...)(final)

		b.Run("depth/"+strconv.Itoa(depth), func(b *testing.B) {
			for b.Loop() {
				handler.ServeHTTP(rec, req)
			}
		})
	}
}

// チェーンをリクエストごとに組み直した場合のコスト。起動時に1回組むのとの差。
func BenchmarkMiddlewareChainConstruction(b *testing.B) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	noop := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { next.ServeHTTP(w, r) })
	})
	ms := []Middleware{noop, noop, noop, noop}

	var sink http.Handler
	for b.Loop() {
		sink = Chain(ms...)(final)
	}
	_ = sink
}
