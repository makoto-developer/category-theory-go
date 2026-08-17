package part6

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

var (
	sinkInt   int
	sinkSlice []int
)

// 中間構造を作る版と作らない版を並べる。理屈のうえでは消えてよいものが、
// Go で実際に消えるのかを allocs/op で見る。
func BenchmarkHylomorphism(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		add := func(acc, x int) int { return acc + x }

		b.Run(fmt.Sprintf("n=%06d/1_via_slice", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloVia(CountTo(n), add, 0, 1)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/2_fused", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloFused(CountTo(n), add, 0, 1)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/3_seq", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloSeq(CountTo(n), add, 0, 1)
			}
		})
		// 余代数も代数も使わず、手で書いたループ。抽象化の代金の基準線。
		b.Run(fmt.Sprintf("n=%06d/4_hand_loop", n), func(b *testing.B) {
			for b.Loop() {
				acc := 0
				for i := 1; i <= n; i++ {
					acc += i
				}
				sinkInt = acc
			}
		})
	}
}

// マージソート。分割の木を実際に作る版と、作らない版。
// 「普通に書いたマージソート」は後者にあたる。
func BenchmarkMergeSort(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		src := make([]int, n)
		r := rand.New(rand.NewPCG(1, 2))
		for i := range src {
			src[i] = r.IntN(1 << 20)
		}

		b.Run(fmt.Sprintf("n=%06d/1_hylo_with_tree", n), func(b *testing.B) {
			buf := make([]int, n)
			for b.Loop() {
				copy(buf, src)
				sinkSlice = MergeSortHylo(buf)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/2_fused", n), func(b *testing.B) {
			buf := make([]int, n)
			for b.Loop() {
				copy(buf, src)
				sinkSlice = MergeSortFused(buf)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/3_slices_sort", n), func(b *testing.B) {
			buf := make([]int, n)
			for b.Loop() {
				copy(buf, src)
				slices.Sort(buf)
				sinkSlice = buf
			}
		})
	}
}
