package part7

import (
	"fmt"
	"testing"
)

// 入れ子の深さを変えられる型。各段が同じスライスを抱える。
type Nest struct {
	Payload []string
	Inner   *Nest
}

// nestL は1段ぶんのレンズ。法則を守るために、格納するスライスを複製する。
func nestL() Lens[Nest, *Nest] {
	return Lens[Nest, *Nest]{
		Get: func(n Nest) *Nest { return n.Inner },
		Set: func(n Nest, in *Nest) Nest {
			n.Payload = cloneSlice(n.Payload) // 自分の参照型を、通るたびに複製する
			n.Inner = in
			return n
		},
	}
}

func payloadL() Lens[*Nest, []string] {
	return Lens[*Nest, []string]{
		Get: func(n *Nest) []string { return cloneSlice(n.Payload) },
		Set: func(n *Nest, ts []string) *Nest {
			c := *n
			c.Payload = cloneSlice(ts)
			return &c
		},
	}
}

func makeNest(depth, size int) Nest {
	payload := make([]string, size)
	for i := range payload {
		payload[i] = "x"
	}
	root := Nest{Payload: payload}
	cur := &root
	for range depth {
		cur.Inner = &Nest{Payload: payload}
		cur = cur.Inner
	}
	return root
}

// depth 段のレンズを合成して、いちばん奥の Payload に焦点を合わせる。
func deepLens(depth int) Lens[Nest, []string] {
	l := Compose(nestL(), payloadL())
	for range depth - 1 {
		l = Compose(nestL(), Compose(Lens[*Nest, Nest]{
			Get: func(n *Nest) Nest { return *n },
			Set: func(_ *Nest, n Nest) *Nest { return &n },
		}, l))
	}
	return l
}

var sinkNest Nest

// 合成の深さを増やすと、法則を守るための複製が何倍に重なるかを測る。
func BenchmarkLensDepth(b *testing.B) {
	const size = 4096
	next := make([]string, size)
	for _, depth := range []int{1, 2, 3, 4} {
		n := makeNest(depth, size)
		l := deepLens(depth)
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkNest = l.Set(n, next)
			}
		})
	}
}
