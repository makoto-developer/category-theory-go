package part3

import (
	"slices"
	"strconv"
	"testing"
)

// 3通りの凌ぎ方は、どれも同じ答えを返す。違うのは型安全性と実行コストだけ。
func TestAllHKTWorkaroundsAgree(t *testing.T) {
	xs := []int{1, 2, 3}
	want := []string{"1", "2", "3"}

	generic := MapSliceGeneric(xs, strconv.Itoa)
	if !slices.Equal(generic, want) {
		t.Fatalf("ジェネリクス版が違う: %v", generic)
	}

	boxed := make(AnySlice, len(xs))
	for i, x := range xs {
		boxed[i] = x
	}
	erased := boxed.MapAny(func(a any) any { return strconv.Itoa(a.(int)) })
	for i, v := range erased.(AnySlice) {
		if v.(string) != want[i] {
			t.Fatalf("型消去版が違う: %v", erased)
		}
	}

	dict := MapTwiceWith(
		SliceFunctor[int, int](),
		SliceFunctor[int, string](),
		xs,
		func(n int) int { return n },
		strconv.Itoa,
	)
	if !slices.Equal(dict, want) {
		t.Fatalf("辞書渡し版が違う: %v", dict)
	}
}

// 辞書渡しなら、Functor の種類を引数で差し替えられる。
// 高階カインドがあれば1本で書けるところを、辞書を渡すことで近いことをしている。
func TestDictionaryAbstractsOverFunctor(t *testing.T) {
	double := func(n int) int { return n * 2 }

	fromSlice := MapTwiceWith(
		SliceFunctor[int, int](), SliceFunctor[int, string](),
		[]int{1, 2}, double, strconv.Itoa,
	)
	if !slices.Equal(fromSlice, []string{"2", "4"}) {
		t.Fatalf("スライスの結果が違う: %v", fromSlice)
	}

	fromOption := MapTwiceWith(
		OptionFunctor[int, int](), OptionFunctor[int, string](),
		Option[int]{Value: 3, Some: true}, double, strconv.Itoa,
	)
	if !fromOption.Some || fromOption.Value != "6" {
		t.Fatalf("Option の結果が違う: %+v", fromOption)
	}
}
