package part9

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

var (
	sinkF []float64
	sinkZ Zipper[float64]
)

// 法則が言う定義（fmap f ∘ duplicate）と、直接書いた extend と、手書きループ。
// 3つとも同じ答えを返す。中間構造を作るかどうかだけが違う。
func BenchmarkMovingAverage(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		r := rand.New(rand.NewPCG(11, 12))
		xs := make([]float64, n)
		for i := range xs {
			xs[i] = r.Float64() * 100
		}
		w := Zipper[float64]{Items: xs, Pos: 0}
		f := MovingAverage(2) // 窓幅5

		b.Run(fmt.Sprintf("n=%06d/1_via_duplicate", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkZ = ExtendViaDuplicate(f, w)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/2_extend", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkZ = Extend(f, w)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/3_hand_loop", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkF = MovingAverageLoop(xs, 2)
			}
		})
	}
}

// duplicate が作る中間構造の大きさを、単独で測る。
// 2点だけでは計算量を言えないので、n を5桁ぶん振る。
func BenchmarkDuplicateAlone(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 10000, 100000} {
		xs := make([]float64, n)
		w := Zipper[float64]{Items: xs, Pos: 0}
		b.Run(fmt.Sprintf("n=%06d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkDup = Duplicate(w)
			}
		})
	}
}

var sinkDup Zipper[Zipper[float64]]
