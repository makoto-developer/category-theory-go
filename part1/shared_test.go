package part1

import "strconv"

// テストとベンチで共通に使う射。int → string → int → bool と対象が移っていく。
func itoaFn(n int) string   { return strconv.Itoa(n) }
func lengthFn(s string) int { return len(s) }
func isEvenFn(n int) bool   { return n%2 == 0 }

// 同じ射を関数値としても持つ。直接呼び出しとの差を測るため。
var (
	itoa   = itoaFn
	length = lengthFn
	isEven = isEvenFn
)

// Pipe 用の、同じ型の上を動く射。
func double(n int) int    { return n * 2 }
func increment(n int) int { return n + 1 }
func negate(n int) int    { return -n }
