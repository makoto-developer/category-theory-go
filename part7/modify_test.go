package part7

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

// 中間の段。触らないフィールド（Payload）は複製しない。
// 永続データ構造と同じで、変えていない部分は共有してよい。
func m4to3() LensM[L4, L3] {
	return LensM[L4, L3]{
		Get: func(x L4) L3 { return x.Inner },
		Mod: func(x L4, f func(L3) L3) L4 { x.Inner = f(x.Inner); return x },
	}
}

func m3to2() LensM[L3, L2] {
	return LensM[L3, L2]{
		Get: func(x L3) L2 { return x.Inner },
		Mod: func(x L3, f func(L2) L2) L3 { x.Inner = f(x.Inner); return x },
	}
}

func m2to1() LensM[L2, L1] {
	return LensM[L2, L1]{
		Get: func(x L2) L1 { return x.Inner },
		Mod: func(x L2, f func(L1) L1) L2 { x.Inner = f(x.Inner); return x },
	}
}

// 葉だけが複製する。読むときも書くときも複製するので、
// f に渡した値も、f が返した値も、外と共有されない。
func m1payload() LensM[L1, []string] {
	return LensM[L1, []string]{
		Get: func(x L1) []string { return cloneSlice(x.Payload) },
		Mod: func(x L1, f func([]string) []string) L1 {
			x.Payload = cloneSlice(f(cloneSlice(x.Payload)))
			return x
		},
	}
}

func deepM(depth int) LensM[L4, []string] {
	switch depth {
	case 3:
		return ComposeM(m4to3(), ComposeM(m3to2(), ComposeM(m2to1(), m1payload())))
	default:
		panic("この検証では深さ3だけを使う")
	}
}

// Mod 版でも非干渉の4性質が成り立つことを確かめる。
// 中間で複製しないので、成り立たなければこの案は使えない。
func TestModifyLensIsIndependent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payload := rapid.SliceOfN(rapid.String(), 1, 5).Draw(t, "payload")
		incoming := rapid.SliceOfN(rapid.String(), 1, 5).Draw(t, "incoming")
		mk := func() L4 {
			return L4{Payload: slices.Clone(payload),
				Inner: L3{Payload: slices.Clone(payload),
					Inner: L2{Payload: slices.Clone(payload),
						Inner: L1{Payload: slices.Clone(payload)}}}}
		}
		l := deepM(3)

		// ① Get の戻り値を書き換えても元は変わらない
		x := mk()
		l.Get(x)[0] += "___"
		if !slices.Equal(x.Inner.Inner.Inner.Payload, payload) {
			t.Fatalf("① Get の戻り値から漏れた")
		}

		// ② Set に渡した値をあとから書き換えても格納先は変わらない
		x = mk()
		in := slices.Clone(incoming)
		u := SetM(l, x, in)
		stored := slices.Clone(u.Inner.Inner.Inner.Payload)
		in[0] += "___"
		if !slices.Equal(u.Inner.Inner.Inner.Payload, stored) {
			t.Fatalf("② Set の入力から漏れた")
		}

		// ③ Set の結果を書き換えても元は変わらない
		x = mk()
		u = SetM(l, x, slices.Clone(incoming))
		u.Inner.Inner.Inner.Payload[0] += "___"
		if !slices.Equal(x.Inner.Inner.Inner.Payload, payload) {
			t.Fatalf("③ Set の結果から漏れた")
		}

		// ④ Mod に渡した関数が入力を壊しても元は変わらない
		x = mk()
		_ = l.Mod(x, func(ts []string) []string { ts[0] += "___"; return ts })
		if !slices.Equal(x.Inner.Inner.Inner.Payload, payload) {
			t.Fatalf("④ Mod の中の書き換えが漏れた")
		}
	})
}

// Get/Set 版と Mod 版が同じ答えを返す。表現が違うだけで、意味は同じ。
func TestSetAndModifyLensesAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payload := rapid.SliceOfN(rapid.String(), 0, 5).Draw(t, "payload")
		incoming := rapid.SliceOfN(rapid.String(), 0, 5).Draw(t, "incoming")
		x := L4{Payload: slices.Clone(payload),
			Inner: L3{Payload: slices.Clone(payload),
				Inner: L2{Payload: slices.Clone(payload),
					Inner: L1{Payload: slices.Clone(payload)}}}}

		viaSet := Compose(l4to3(), Compose(l3to2(), Compose(l2to1(), l1payload()))).Set(x, incoming)
		viaMod := SetM(deepM(3), x, incoming)
		if !slices.Equal(viaSet.Inner.Inner.Inner.Payload, viaMod.Inner.Inner.Inner.Payload) {
			t.Fatalf("焦点の値が違う: %v vs %v", viaSet.Inner.Inner.Inner.Payload, viaMod.Inner.Inner.Inner.Payload)
		}
	})
}

// Mod 版の複製量が深さによらないことを見る。中間の段が複製しないので、
// 深さを増やしても葉の2回ぶんから増えないはず。
func BenchmarkModifyDepth(b *testing.B) {
	const size = 4096
	next := make([]string, size)
	p := makePayload(size)
	x1 := L1{Payload: p}
	x2 := L2{Payload: p, Inner: x1}
	x3 := L3{Payload: p, Inner: x2}
	x4 := L4{Payload: p, Inner: x3}

	b.Run("depth=0", func(b *testing.B) {
		l := m1payload()
		b.ReportAllocs()
		for b.Loop() {
			sink1 = SetM(l, x1, next)
		}
	})
	b.Run("depth=1", func(b *testing.B) {
		l := ComposeM(m2to1(), m1payload())
		b.ReportAllocs()
		for b.Loop() {
			sink2 = SetM(l, x2, next)
		}
	})
	b.Run("depth=2", func(b *testing.B) {
		l := ComposeM(m3to2(), ComposeM(m2to1(), m1payload()))
		b.ReportAllocs()
		for b.Loop() {
			sink3 = SetM(l, x3, next)
		}
	})
	b.Run("depth=3", func(b *testing.B) {
		l := deepM(3)
		b.ReportAllocs()
		for b.Loop() {
			sink4 = SetM(l, x4, next)
		}
	})
}

// 深さ3で、Get/Set 版と Mod 版の複製量を並べる。
func BenchmarkDepth3SetVsModify(b *testing.B) {
	const size = 4096
	next := make([]string, size)
	p := makePayload(size)
	x := L4{Payload: p, Inner: L3{Payload: p, Inner: L2{Payload: p, Inner: L1{Payload: p}}}}

	b.Run("1_get_set", func(b *testing.B) {
		l := Compose(l4to3(), Compose(l3to2(), Compose(l2to1(), l1payload())))
		b.ReportAllocs()
		for b.Loop() {
			sink4 = l.Set(x, next)
		}
	})
	b.Run("2_modify", func(b *testing.B) {
		l := deepM(3)
		b.ReportAllocs()
		for b.Loop() {
			sink4 = SetM(l, x, next)
		}
	})
}
