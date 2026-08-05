// 連載「Goで書く実践圏論」第2回のスニペット。
//
// 日常の道具が Functor とモノイドだった、という話を実際に走らせて確かめます。
// 「このループ1回にまとめていい？」「この集計、並列に分割して大丈夫？」に
// 根拠を出すのがここの目的です。
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

// 案内は先頭に出す。記事に埋め込むと末尾しか見えないので、末尾は検証結果のために空けておく。
func main() {
	fmt.Println("連載「Goで書く実践圏論」第2回")
	fmt.Println("  記事: https://blog.makoto-developer.net/articles/2026-08-05-practical-category-theory-go-2")
	fmt.Println("  全4回のコード（テスト・ベンチ込み）: https://github.com/makoto-developer/category-theory-go")
	fmt.Println("──────────────────────────────────────────")

	checkFunctorLaws()
	showLoopFusion()
	showParallelSumWorks()
	showParallelMeanBreaks()
	showMiddlewareIsAssociativeButNotCommutative()
}

// ---- Functor -----------------------------------------------------------

// Map は []A を []B に写す。中身を見ずに、渡された射を各要素に当てるだけ。
func Map[A, B any](xs []A, f func(A) B) []B {
	out := make([]B, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}

func Compose[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C { return g(f(a)) }
}

func Identity[A any](a A) A { return a }

// Functor 則は2つだけ。
//
//	恒等保存: Map(xs, Identity) == xs
//	合成保存: Map(Map(xs, f), g) == Map(xs, Compose(f, g))
//
// 後者が「2回まわしているループを1回にまとめていい」の許可証になる。
func checkFunctorLaws() {
	for i := 0; i < 10000; i++ {
		xs := randomInts(rand.Intn(8))

		double := func(n int) int { return n * 2 }
		plusOne := func(n int) int { return n + 1 }

		if !sameInts(Map(xs, Identity[int]), xs) {
			fmt.Printf("恒等保存が破れた: %v\n", xs)
			return
		}
		twice := Map(Map(xs, double), plusOne)
		once := Map(xs, Compose(double, plusOne))
		if !sameInts(twice, once) {
			fmt.Printf("合成保存が破れた: %v\n", xs)
			return
		}
	}
	fmt.Println("[Functor則] 1万通りのランダムな配列で一致 —— 2周を1周にまとめてよい")
}

// 合成保存はメモリにも効く。中間スライスが1本消えるので、実際に測ると差が出る。
func showLoopFusion() {
	xs := randomInts(100000)
	double := func(n int) int { return n * 2 }
	plusOne := func(n int) int { return n + 1 }

	twice := allocatedBytes(func() { _ = Map(Map(xs, double), plusOne) })
	once := allocatedBytes(func() { _ = Map(xs, Compose(double, plusOne)) })

	fmt.Printf("\n[ループ融合] 10万要素で 2周=%dKB / 1周にまとめると=%dKB\n", twice/1024, once/1024)
	fmt.Println("            Map(Map(xs, f), g) を Map(xs, g∘f) に書き換えてよい根拠が合成保存")
}

// f の実行中に確保されたバイト数を返す。
func allocatedBytes(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// ---- モノイド ----------------------------------------------------------

// 合計は結合的（(a+b)+c == a+(b+c)）なので、どこで区切って並列化しても答えが変わらない。
func showParallelSumWorks() {
	xs := randomFloats(10000)

	fmt.Printf("\n[合計] 逐次=%.1f / 4分割して並列=%.1f （一致）\n", sum(xs), sumInParallel(xs, 4))
	fmt.Println("       結合的な演算は分割してよい。errors.Join も strings.Builder も同じ形")
}

// 平均は結合的でない = モノイドでない。だから分割すると答えが変わる。
// 「合計と個数」に分ければ両方とも結合的になり、並列化しても一致する。
func showParallelMeanBreaks() {
	xs := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	fmt.Printf("\n[平均] 逐次=%.4f / 3分割して部分平均の平均=%.4f （ずれた）\n", mean(xs), meanInParallel(xs, 3))
	fmt.Printf("       合計と個数に分けてから並列化すると=%.4f （直った）\n", meanViaSumCount(xs, 3))
	fmt.Println("       MapReduce で「合計と件数」を持ち回るのは、平均をモノイドにするための変形だった")
}

func sum(xs []float64) float64 {
	total := 0.0
	for _, x := range xs {
		total += x
	}
	return total
}

func sumInParallel(xs []float64, workers int) float64 {
	results := make(chan float64, workers)
	var wg sync.WaitGroup

	for _, part := range chunks(xs, workers) {
		wg.Add(1)
		go func(part []float64) {
			defer wg.Done()
			results <- sum(part)
		}(part)
	}
	wg.Wait()
	close(results)

	total := 0.0
	for r := range results {
		total += r
	}
	return total
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	return sum(xs) / float64(len(xs))
}

// 素朴に分割して部分平均の平均を取る。結合的でないので壊れる。
func meanInParallel(xs []float64, workers int) float64 {
	partials := make([]float64, 0, workers)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, part := range chunks(xs, workers) {
		wg.Add(1)
		go func(part []float64) {
			defer wg.Done()
			m := mean(part)
			mu.Lock()
			partials = append(partials, m)
			mu.Unlock()
		}(part)
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

	results := make(chan sumCount, workers)
	var wg sync.WaitGroup

	for _, part := range chunks(xs, workers) {
		wg.Add(1)
		go func(part []float64) {
			defer wg.Done()
			results <- sumCount{sum: sum(part), count: len(part)}
		}(part)
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

// ---- middleware（自己射のモノイド）-------------------------------------

// Middleware は「文字列を返す処理」を包んで別の処理にする。
// http.Handler を包む middleware と同じ形。
type Middleware func(func() string) func() string

func tag(name string) Middleware {
	return func(next func() string) func() string {
		return func() string { return "<" + name + ">" + next() + "</" + name + ">" }
	}
}

func chain(ms ...Middleware) Middleware {
	return func(next func() string) func() string {
		// 後ろから巻くので、書いた順に外側から適用される
		for i := len(ms) - 1; i >= 0; i-- {
			next = ms[i](next)
		}
		return next
	}
}

// middleware の合成には結合律があるが可換性はない。
// だから「括り方を変えるリファクタリング」は安全で、「順序の入れ替え」は危ない。
func showMiddlewareIsAssociativeButNotCommutative() {
	body := func() string { return "本文" }
	a, b, c := tag("auth"), tag("log"), tag("gzip")

	left := chain(chain(a, b), c)(body)()  // 前2つを先にまとめた
	right := chain(a, chain(b, c))(body)() // 後2つを先にまとめた
	swapped := chain(b, a, c)(body)()      // 順序を入れ替えた

	fmt.Printf("\n[middleware] (a∘b)∘c = %s\n", left)
	fmt.Printf("             a∘(b∘c) = %s （括り方を変えても同じ）\n", right)
	fmt.Printf("             順序を入れ替えると = %s （%s）\n", swapped, verdict(left == swapped))
	fmt.Println("             結合律はあるが可換性はない。だから切り出しは安全で、並べ替えは危ない")
}

func verdict(same bool) string {
	if same {
		return "変わらない"
	}
	return "変わってしまう"
}

// ---- 小道具 ------------------------------------------------------------

func chunks(xs []float64, workers int) [][]float64 {
	if workers < 1 {
		workers = 1
	}
	size := (len(xs) + workers - 1) / workers
	var out [][]float64
	for lo := 0; lo < len(xs); lo += size {
		hi := lo + size
		if hi > len(xs) {
			hi = len(xs)
		}
		out = append(out, xs[lo:hi])
	}
	return out
}

func randomInts(n int) []int {
	xs := make([]int, n)
	for i := range xs {
		xs[i] = rand.Intn(2000) - 1000
	}
	return xs
}

func randomFloats(n int) []float64 {
	xs := make([]float64, n)
	for i := range xs {
		xs[i] = float64(rand.Intn(100))
	}
	return xs
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
