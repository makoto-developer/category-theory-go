package part1

import (
	"strconv"
	"testing"
)

var (
	benchInput = 1234567
	sinkBool   bool
	sinkInt    int
	sinkFunc   func(int) bool
)

// 何も抽象化しない素の呼び出し。コンパイラがインライン化できる形。
func BenchmarkDirectCall(b *testing.B) {
	for b.Loop() {
		sinkBool = isEvenFn(lengthFn(itoaFn(benchInput)))
	}
}

// 同じ処理を関数値（変数に入れた射）経由で呼ぶ。間接呼び出しのコストだけが乗る。
func BenchmarkFuncValueCall(b *testing.B) {
	for b.Loop() {
		sinkBool = isEven(length(itoa(benchInput)))
	}
}

// Compose で組んだ射を呼ぶ。合成はループの外で済ませてある。
func BenchmarkComposed(b *testing.B) {
	composed := Compose(Compose(itoa, length), isEven)

	for b.Loop() {
		sinkBool = composed(benchInput)
	}
}

// 同じ型の上を3段。Pipe はクロージャが入れ子になる。
func BenchmarkPipe3(b *testing.B) {
	piped := Pipe(double, increment, negate)

	for b.Loop() {
		sinkInt = piped(benchInput)
	}
}

// 比較用: Pipe と同じ処理を手で書いた場合。
func BenchmarkPipe3ByHand(b *testing.B) {
	for b.Loop() {
		sinkInt = negate(increment(double(benchInput)))
	}
}

// 合成した射を変数に保持してから呼ぶ場合。呼び出しは関数値経由になる。
func BenchmarkComposedStored(b *testing.B) {
	composed := Compose(double, increment)

	for b.Loop() {
		sinkInt = composed(benchInput)
	}
}

// 同じ合成をその場で組んで即座に呼ぶ場合。コンパイラが合成ごと畳み込めるかを見る。
func BenchmarkComposedInline(b *testing.B) {
	for b.Loop() {
		sinkInt = Compose(double, increment)(benchInput)
	}
}

// 3段を Compose の入れ子で静的に組んだ場合。Pipe（スライス経由）との差を見る。
func BenchmarkComposedStored3(b *testing.B) {
	composed := Compose(Compose(double, increment), negate)

	for b.Loop() {
		sinkInt = composed(benchInput)
	}
}

// 比較用: 合成を挟まず手で書いた場合。
func BenchmarkComposedByHand(b *testing.B) {
	for b.Loop() {
		sinkInt = increment(double(benchInput))
	}
}

// 合成の段数を増やすと1段あたり何ナノ秒増えるかを測る。
func BenchmarkPipeDepth(b *testing.B) {
	for _, depth := range []int{1, 2, 4, 8, 16} {
		fs := make([]func(int) int, depth)
		for i := range fs {
			fs[i] = increment
		}
		piped := Pipe(fs...)

		b.Run("composed/"+strconv.Itoa(depth), func(b *testing.B) {
			for b.Loop() {
				sinkInt = piped(benchInput)
			}
		})
		b.Run("loop/"+strconv.Itoa(depth), func(b *testing.B) {
			for b.Loop() {
				x := benchInput
				for range depth {
					x = increment(x)
				}
				sinkInt = x
			}
		})
	}
}

// 合成そのものの構築コスト。リクエストごとに組み直す設計をするなら効いてくる。
func BenchmarkComposeConstruction(b *testing.B) {
	for b.Loop() {
		sinkFunc = Compose(Compose(itoa, length), isEven)
	}
}
