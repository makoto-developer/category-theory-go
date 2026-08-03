package part3

import (
	"slices"
	"strconv"
	"testing"
)

var (
	sinkFloat   float64
	sinkInt     int
	sinkStrings []string
	sinkPair    Pair[float64, int]
	sinkAny     AnyFunctor
)

// deepExpr は指定した深さの完全二分木を作る。
func deepExpr(depth int) Expr {
	if depth <= 1 {
		return Add{L: Var{Name: "x"}, R: Num{V: 2}}
	}
	return Mul{L: deepExpr(depth - 1), R: deepExpr(depth - 1)}
}

// 代数を経由する fold と、手書きの再帰の差。抽象化の代金を測る。
func BenchmarkFoldVsDirect(b *testing.B) {
	e := deepExpr(12)
	alg := EvalAlgebra(env)

	b.Run("fold", func(b *testing.B) {
		for b.Loop() {
			sinkFloat = Fold(e, alg)
		}
	})
	b.Run("direct", func(b *testing.B) {
		for b.Loop() {
			sinkFloat = EvalDirect(e, env)
		}
	})
}

// 2回たどるのと、代数の積で1回たどるのの差。
func BenchmarkTwoFoldsVsProduct(b *testing.B) {
	e := deepExpr(12)
	alg := EvalAlgebra(env)
	prod := ProductAlgebra(alg, CountAlgebra)

	b.Run("two_folds", func(b *testing.B) {
		for b.Loop() {
			sinkPair = Pair[float64, int]{Fold(e, alg), Fold(e, CountAlgebra)}
		}
	})
	b.Run("product_algebra", func(b *testing.B) {
		for b.Loop() {
			sinkPair = Fold(e, prod)
		}
	})
}

// push 型（range-over-func）と pull 型（iter.Pull）とスライス直走査の差。
func BenchmarkSeqStyles(b *testing.B) {
	for _, n := range []int{10, 1_000, 100_000} {
		xs := make([]int, n)
		for i := range xs {
			xs[i] = i
		}

		b.Run("slice/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				total := 0
				for _, x := range xs {
					total += x
				}
				sinkInt = total
			}
		})
		b.Run("push_seq/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = SumSeq(slices.Values(xs))
			}
		})
		b.Run("pull_seq/"+strconv.Itoa(n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = SumPull(slices.Values(xs))
			}
		})
	}
}

// 高階カインドの回避策3種のコスト。型を消すと何を払うことになるか。
func BenchmarkHKTWorkarounds(b *testing.B) {
	const n = 1_000
	xs := make([]int, n)
	boxed := make(AnySlice, n)
	for i := range xs {
		xs[i] = i
		boxed[i] = i
	}

	b.Run("direct_loop", func(b *testing.B) {
		for b.Loop() {
			out := make([]string, len(xs))
			for i, x := range xs {
				out[i] = strconv.Itoa(x)
			}
			sinkStrings = out
		}
	})
	b.Run("generic", func(b *testing.B) {
		for b.Loop() {
			sinkStrings = MapSliceGeneric(xs, strconv.Itoa)
		}
	})
	b.Run("type_erased", func(b *testing.B) {
		for b.Loop() {
			sinkAny = boxed.MapAny(func(a any) any { return strconv.Itoa(a.(int)) })
		}
	})
	b.Run("dictionary", func(b *testing.B) {
		d := SliceFunctor[int, string]()
		for b.Loop() {
			sinkStrings = d.Map(xs, strconv.Itoa)
		}
	})
}
