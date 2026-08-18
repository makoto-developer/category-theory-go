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
		// 余代数はループの外で作る。中で作ると、要素ごとの費用に
		// クロージャ生成という定数コストが混ざる。
		co := CountTo(n)

		b.Run(fmt.Sprintf("n=%06d/1_via_slice", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloVia(co, add, 0, 1)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/2_fused", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloFused(co, add, 0, 1)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/3_seq", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloSeq(co, add, 0, 1)
			}
		})
		// 余代数を呼び出し位置で組み立てた場合。変数に入れた上と何が違うかを見る
		// （コンパイラが中身を見通せるかどうかが変わる）。
		b.Run(fmt.Sprintf("n=%06d/5_fused_literal_coalgebra", n), func(b *testing.B) {
			for b.Loop() {
				sinkInt = HyloFused(CountTo(n), add, 0, 1)
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
//
// MergeSortHylo と MergeSortFused は入力を破壊しない（merge が新しい出力を確保し、
// 入力側はスライスヘッダを進めるだけ）。よって両者には毎回のコピーが要らない。
// slices.Sort だけが破壊的なので、そちらにはコピーが要る。
// この非対称を隠さないよう、コピーを含めない比較と、含めた比較を分けて測る。
func BenchmarkMergeSort(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		src := make([]int, n)
		r := rand.New(rand.NewPCG(1, 2))
		for i := range src {
			src[i] = r.IntN(1 << 20)
		}

		// 木を作る／作らないの差だけを見る。どちらも非破壊なのでコピー不要。
		b.Run(fmt.Sprintf("n=%06d/1_hylo_with_tree", n), func(b *testing.B) {
			for b.Loop() {
				sinkSlice = MergeSortHylo(src)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/2_fused", n), func(b *testing.B) {
			for b.Loop() {
				sinkSlice = MergeSortFused(src)
			}
		})
		// 実務の比較。元の入力を残したままソートする、という同じ条件に揃える。
		// 確保も計測区間に入れる（slices.Sort の 0 B/op はコピー用バッファを含まないため）。
		b.Run(fmt.Sprintf("n=%06d/3_slices_sort_with_clone", n), func(b *testing.B) {
			for b.Loop() {
				buf := slices.Clone(src)
				slices.Sort(buf)
				sinkSlice = buf
			}
		})
	}
}
