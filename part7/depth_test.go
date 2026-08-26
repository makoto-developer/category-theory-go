package part7

import "testing"

// 深さを固定した値の入れ子。ポインタを使わないので、段ごとの複製だけが観測される。
// 以前はポインタと値を行き来するアダプタを噛ませていて、そのアダプタ自身の
// ヒープ確保（32B）が段ごとの差分に混ざっていた。足場を測ってしまっていた。

type L1 struct{ Payload []string }
type L2 struct {
	Payload []string
	Inner   L1
}
type L3 struct {
	Payload []string
	Inner   L2
}
type L4 struct {
	Payload []string
	Inner   L3
}

// 各段のレンズは、通るときに自分の Payload を複製する。
// これが「各段が自分で身を守る」の形。

func l4to3() Lens[L4, L3] {
	return Lens[L4, L3]{
		Get: func(x L4) L3 { return x.Inner },
		Set: func(x L4, in L3) L4 { x.Payload = cloneSlice(x.Payload); x.Inner = in; return x },
	}
}

func l3to2() Lens[L3, L2] {
	return Lens[L3, L2]{
		Get: func(x L3) L2 { return x.Inner },
		Set: func(x L3, in L2) L3 { x.Payload = cloneSlice(x.Payload); x.Inner = in; return x },
	}
}

func l2to1() Lens[L2, L1] {
	return Lens[L2, L1]{
		Get: func(x L2) L1 { return x.Inner },
		Set: func(x L2, in L1) L2 { x.Payload = cloneSlice(x.Payload); x.Inner = in; return x },
	}
}

func l1payload() Lens[L1, []string] {
	return Lens[L1, []string]{
		Get: func(x L1) []string { return cloneSlice(x.Payload) },
		Set: func(x L1, ts []string) L1 { x.Payload = cloneSlice(ts); return x },
	}
}

func makePayload(size int) []string {
	p := make([]string, size)
	for i := range p {
		p[i] = "x"
	}
	return p
}

var (
	sink1 L1
	sink2 L2
	sink3 L3
	sink4 L4
)

// 合成の深さを増やすと、複製が何回重なるかを測る。
// 深さ n の合成では、通る Set が n+1 個あるので複製は n+1 回になる。
func BenchmarkLensDepth(b *testing.B) {
	const size = 4096
	next := make([]string, size)
	p := makePayload(size)

	x1 := L1{Payload: p}
	x2 := L2{Payload: p, Inner: x1}
	x3 := L3{Payload: p, Inner: x2}
	x4 := L4{Payload: p, Inner: x3}

	// 合成なし。Set は1つだけ通る。
	b.Run("depth=0/set=1", func(b *testing.B) {
		l := l1payload()
		b.ReportAllocs()
		for b.Loop() {
			sink1 = l.Set(x1, next)
		}
	})
	b.Run("depth=1/set=2", func(b *testing.B) {
		l := Compose(l2to1(), l1payload())
		b.ReportAllocs()
		for b.Loop() {
			sink2 = l.Set(x2, next)
		}
	})
	b.Run("depth=2/set=3", func(b *testing.B) {
		l := Compose(l3to2(), Compose(l2to1(), l1payload()))
		b.ReportAllocs()
		for b.Loop() {
			sink3 = l.Set(x3, next)
		}
	})
	b.Run("depth=3/set=4", func(b *testing.B) {
		l := Compose(l4to3(), Compose(l3to2(), Compose(l2to1(), l1payload())))
		b.ReportAllocs()
		for b.Loop() {
			sink4 = l.Set(x4, next)
		}
	})
}
