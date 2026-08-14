package part5

import (
	"fmt"
	"strings"
	"testing"
)

var (
	sinkString string
	sinkList   *List
	sinkRList  *RList
)

// makePieces は同じ長さの断片を n 個作る。
func makePieces(n, size int) []string {
	xs := make([]string, n)
	chunk := strings.Repeat("x", size)
	for i := range xs {
		xs[i] = chunk
	}
	return xs
}

// 文字列の Append は両辺のぶんコピーする。Cayley 表現で結合を右に寄せても
// 総コピー量は変わらないはず、という予想を確かめる。
func BenchmarkStringMonoid(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		xs := makePieces(n, 8)
		total := n * 8

		b.Run(fmt.Sprintf("n=%05d/naive", n), func(b *testing.B) {
			for b.Loop() {
				sinkString = FoldNaive(StringMonoid, xs)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/cayley", n), func(b *testing.B) {
			for b.Loop() {
				sinkString = FoldCayley(StringMonoid, xs)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/builder", n), func(b *testing.B) {
			for b.Loop() {
				sinkString = BuildString(xs)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/builder_grow", n), func(b *testing.B) {
			for b.Loop() {
				sinkString = BuildStringGrow(xs, total)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/join", n), func(b *testing.B) {
			for b.Loop() {
				sinkString = strings.Join(xs, "")
			}
		})
	}
}

// Append のコストが左辺に付くモノイド。左結合が O(n^2)、Cayley 表現が O(n) になるはず。
func BenchmarkConsMonoid(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		xs := makePieces(n, 8)
		cells := make([]*List, n)
		for i, x := range xs {
			cells[i] = SingleCons(x)
		}

		b.Run(fmt.Sprintf("n=%05d/naive", n), func(b *testing.B) {
			for b.Loop() {
				sinkList = FoldNaive(ConsMonoid, cells)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/cayley", n), func(b *testing.B) {
			for b.Loop() {
				sinkList = FoldCayley(ConsMonoid, cells)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/prepend_reverse", n), func(b *testing.B) {
			for b.Loop() {
				sinkList = PrependReverse(xs)
			}
		})
	}
}

// 鏡像。Append のコストを右辺へ移すと、上と正確に逆の結果になるはず。
func BenchmarkSnocMonoid(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		cells := make([]*RList, n)
		for i, x := range makePieces(n, 8) {
			cells[i] = SingleSnoc(x)
		}

		b.Run(fmt.Sprintf("n=%05d/naive", n), func(b *testing.B) {
			for b.Loop() {
				sinkRList = FoldNaive(SnocMonoid, cells)
			}
		})
		b.Run(fmt.Sprintf("n=%05d/cayley", n), func(b *testing.B) {
			for b.Loop() {
				sinkRList = FoldCayley(SnocMonoid, cells)
			}
		})
	}
}
