// 連載「Goで書く実践圏論」より
// https://blog.makoto-developer.net/articles/2026-08-04-practical-category-theory-go-1
//
// 圏の法則が本当に成り立つこと、そして法則が無いと何が壊れるかを、
// 実際に走らせて確かめます。数字をいじって遊んでみてください。
package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

// ---- ここが圏のすべて --------------------------------------------------
// 対象 = 型、射 = func(A) B、あとは合成と恒等射だけ。

// Compose は f のあとに g を適用する射を返す。数学の記法では g∘f。
func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C { return g(f(a)) }
}

// Identity は恒等射。合成における単位元になる。
func Identity[A any](a A) A { return a }

// ---- 検証に使う射 ------------------------------------------------------

func itoa(n int) string   { return strconv.Itoa(n) }
func length(s string) int { return len(s) }
func isEven(n int) bool   { return n%2 == 0 }

func main() {
	checkAssociativity()
	checkIdentity()
	showMemoizeBreaks()
	showMeanBreaks()
	showFloatBreaks()

	fmt.Println()
	fmt.Println("──────────────────────────────────────────")
	fmt.Println("続きは記事で:")
	fmt.Println("  https://blog.makoto-developer.net/articles/2026-08-04-practical-category-theory-go-1")
	fmt.Println("全4回のコード（テスト・ベンチ込み）:")
	fmt.Println("  https://github.com/makoto-developer/category-theory-go")
}

// 結合律: (h∘g)∘f と h∘(g∘f) は、どんな入力でも等しい。
// これが「関数を切り出すリファクタリング」が安全な理由。
func checkAssociativity() {
	left := Compose(Compose(itoa, length), isEven)  // 前2つを先にまとめた
	right := Compose(itoa, Compose(length, isEven)) // 後2つを先にまとめた

	for i := 0; i < 10000; i++ {
		x := rand.Intn(2_000_000) - 1_000_000
		if left(x) != right(x) {
			fmt.Printf("結合律が破れた: x=%d\n", x)
			return
		}
	}
	fmt.Println("[結合律] 1万通りのランダム入力で一致 —— 括り方を変えても結果は同じ")
}

// 単位律: 恒等射を前後どちらに挟んでも、元の射と変わらない。
// 「何もしないステップ」を足したり消したりできる理由。
func checkIdentity() {
	for i := 0; i < 10000; i++ {
		x := rand.Intn(2_000_000) - 1_000_000
		if Compose(Identity[int], itoa)(x) != itoa(x) || Compose(itoa, Identity[string])(x) != itoa(x) {
			fmt.Printf("単位律が破れた: x=%d\n", x)
			return
		}
	}
	fmt.Println("[単位律] 恒等射を挟んでも変わらない —— 素通しのミドルウェアを入れてよい")
}

// 純粋でない射はメモ化すると答えが変わる。
// 「純粋関数にしろ」という経験則が守っていたものが、ここに出る。
func showMemoizeBreaks() {
	counter := newCounter()
	memoized := memoize(counter)

	first, second, direct := memoized(10), memoized(10), counter(10)

	fmt.Printf("\n[メモ化] 同じ counter(10) が… メモ化あり=%d, %d / メモ化なし=%d\n", first, second, direct)
	fmt.Println("         純粋でない関数は Go の関数ではあっても「圏の射」ではない")
}

func newCounter() func(int) int {
	n := 0
	return func(x int) int { n++; return x + n }
}

func memoize[A comparable, B any](f func(A) B) func(A) B {
	cache := map[A]B{}
	return func(a A) B {
		if b, ok := cache[a]; ok {
			return b
		}
		b := f(a)
		cache[a] = b
		return b
	}
}

// 平均は結合的でない = モノイドでない。だから分割して並列に計算すると答えが変わる。
// 「合計と個数」に分ければ両方とも結合的になり、並列化しても一致する。
func showMeanBreaks() {
	xs := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	fmt.Printf("\n[平均] 逐次=%.4f / 3分割して部分平均の平均=%.4f\n", mean(xs), meanInParallel(xs, 3))
	fmt.Printf("       合計と個数に分けてから並列化すると=%.4f （直った）\n", meanViaSumCount(xs, 3))
	fmt.Println("       MapReduce で「合計と件数」を持ち回るのは、平均をモノイドにするための変形だった")
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// 素朴に分割して部分平均の平均を取る。結合的でないので壊れる。
func meanInParallel(xs []float64, workers int) float64 {
	chunk := (len(xs) + workers - 1) / workers
	partials := make([]float64, 0, workers)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for lo := 0; lo < len(xs); lo += chunk {
		hi := lo + chunk
		if hi > len(xs) {
			hi = len(xs)
		}
		wg.Add(1)
		go func(part []float64) {
			defer wg.Done()
			m := mean(part)
			mu.Lock()
			partials = append(partials, m)
			mu.Unlock()
		}(xs[lo:hi])
	}
	wg.Wait()
	return mean(partials)
}

// 合計と個数はどちらも結合的なので、分割しても答えが変わらない。
func meanViaSumCount(xs []float64, workers int) float64 {
	type sumCount struct {
		sum   float64
		count int
	}

	chunk := (len(xs) + workers - 1) / workers
	results := make(chan sumCount, workers)
	var wg sync.WaitGroup

	for lo := 0; lo < len(xs); lo += chunk {
		hi := lo + chunk
		if hi > len(xs) {
			hi = len(xs)
		}
		wg.Add(1)
		go func(part []float64) {
			defer wg.Done()
			s := 0.0
			for _, x := range part {
				s += x
			}
			results <- sumCount{sum: s, count: len(part)}
		}(xs[lo:hi])
	}
	wg.Wait()
	close(results)

	total, n := 0.0, 0
	for r := range results {
		total += r.sum
		n += r.count
	}
	if n == 0 {
		return 0
	}
	return total / float64(n)
}

// float64 の加算はそもそも結合的でない。分割数を変えると集計結果がずれうる。
func showFloatBreaks() {
	a, b, c := 1e16, -1e16, 1.0

	fmt.Printf("\n[浮動小数点] (a+b)+c = %v / a+(b+c) = %v\n", (a+b)+c, a+(b+c))
	fmt.Println("             金額を扱うなら整数（最小単位）にせよ、の根拠がこれ")
}
