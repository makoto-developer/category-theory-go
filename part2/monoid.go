package part2

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Monoid は結合的な二項演算 Append と、その単位元 Empty の組。
// 対象がひとつだけの圏だと思ってもよい（射が T の値、合成が Append）。
type Monoid[T any] struct {
	Empty  T
	Append func(T, T) T
}

// Fold は左から順に畳み込む。
func Fold[T any](m Monoid[T], xs []T) T {
	acc := m.Empty
	for _, x := range xs {
		acc = m.Append(acc, x)
	}
	return acc
}

// FoldParallel はスライスを分割して並列に畳み込み、部分結果をさらに畳み込む。
// 分割の仕方によらず同じ答えになるのは結合律のおかげで、それ以外の保証は無い。
func FoldParallel[T any](m Monoid[T], xs []T, workers int) T {
	if workers <= 1 || len(xs) < workers {
		return Fold(m, xs)
	}

	partials := make([]T, workers)
	chunk := (len(xs) + workers - 1) / workers
	var wg sync.WaitGroup

	for w := range workers {
		lo := w * chunk
		hi := min(lo+chunk, len(xs))
		if lo >= hi {
			partials[w] = m.Empty
			continue
		}
		wg.Add(1)
		go func(w, lo, hi int) {
			defer wg.Done()
			partials[w] = Fold(m, xs[lo:hi])
		}(w, lo, hi)
	}
	wg.Wait()

	return Fold(m, partials)
}

// DefaultWorkers は並列畳み込みの既定の分割数。
func DefaultWorkers() int { return runtime.GOMAXPROCS(0) }

// 標準ライブラリの中に元からあるモノイドたち。
var (
	// SumInt は加算と 0。
	SumInt = Monoid[int]{Empty: 0, Append: func(a, b int) int { return a + b }}

	// ConcatString は連結と空文字列。strings.Builder が速いのは、この構造を使い回すため。
	ConcatString = Monoid[string]{Empty: "", Append: func(a, b string) string { return a + b }}

	// JoinErrors は errors.Join と nil。nil が単位元として振る舞う。
	JoinErrors = Monoid[error]{Empty: nil, Append: func(a, b error) error { return errors.Join(a, b) }}

	// MaxDuration は最大値と 0。SLO 集計でよく使う形。
	MaxDuration = Monoid[time.Duration]{Empty: 0, Append: func(a, b time.Duration) time.Duration { return max(a, b) }}

	// BuildString は strings.Builder を経由する連結。文字列結合と同じ答えを返す。
	BuildString = Monoid[string]{Empty: "", Append: func(a, b string) string {
		var sb strings.Builder
		sb.Grow(len(a) + len(b))
		sb.WriteString(a)
		sb.WriteString(b)
		return sb.String()
	}}
)

// CompareBy は比較関数のモノイド。単位元は「常に同順」を返す比較で、
// Append は「先に差がついたほうを採る」。多段ソートの比較関数はこの形をしている。
// 標準ライブラリの cmp.Or が値に対してやっていることと同じ構造。
func CompareBy[T any](cmps ...func(a, b T) int) func(a, b T) int {
	return func(a, b T) int {
		for _, cmp := range cmps {
			if r := cmp(a, b); r != 0 {
				return r
			}
		}
		return 0
	}
}

// Mean は平均。結合的でないのでモノイドにならず、分割して並列に畳み込むと答えが変わる。
func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// MeanParallel は Mean を「分割して部分平均の平均を取る」形に素朴に並列化したもの。
// 平均が結合的でないため、分割数によって答えが変わる。
func MeanParallel(xs []float64, workers int) float64 {
	if workers <= 1 || len(xs) < workers {
		return Mean(xs)
	}

	partials := make([]float64, workers)
	chunk := (len(xs) + workers - 1) / workers
	var wg sync.WaitGroup

	for w := range workers {
		lo := w * chunk
		hi := min(lo+chunk, len(xs))
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(w, lo, hi int) {
			defer wg.Done()
			partials[w] = Mean(xs[lo:hi])
		}(w, lo, hi)
	}
	wg.Wait()

	return Mean(partials)
}

// SumCount は「合計と個数」の組。平均をモノイドで扱うための正しい形。
type SumCount struct {
	Sum   float64
	Count int
}

// MeanMonoid は合計と個数を別々に畳み込む。両方とも結合的なので分割して並列化できる。
var MeanMonoid = Monoid[SumCount]{
	Empty: SumCount{},
	Append: func(a, b SumCount) SumCount {
		return SumCount{Sum: a.Sum + b.Sum, Count: a.Count + b.Count}
	},
}

// Value は畳み込んだ結果から平均を取り出す。
func (sc SumCount) Value() float64 {
	if sc.Count == 0 {
		return 0
	}
	return sc.Sum / float64(sc.Count)
}
