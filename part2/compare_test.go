package part2

import (
	"cmp"
	"slices"
	"testing"

	"pgregory.net/rapid"
)

type employee struct {
	Dept string
	Age  int
	Name string
}

func byDept(a, b employee) int { return cmp.Compare(a.Dept, b.Dept) }
func byAge(a, b employee) int  { return cmp.Compare(a.Age, b.Age) }
func byName(a, b employee) int { return cmp.Compare(a.Name, b.Name) }

var employeeGen = rapid.Custom(func(t *rapid.T) employee {
	return employee{
		Dept: rapid.SampledFrom([]string{"sales", "dev"}).Draw(t, "dept"),
		Age:  rapid.IntRange(20, 60).Draw(t, "age"),
		Name: rapid.SampledFrom([]string{"a", "b", "c"}).Draw(t, "name"),
	}
})

// 比較関数のモノイド則。どう括ってまとめても、同じ順序になる。
// 「部署でまとめた比較」を先に用意してから年齢を足しても、結果は変わらない。
func TestCompareByIsMonoid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := employeeGen.Draw(t, "a")
		b := employeeGen.Draw(t, "b")

		flat := CompareBy(byDept, byAge, byName)
		grouped := CompareBy(CompareBy(byDept, byAge), byName)
		nested := CompareBy(byDept, CompareBy(byAge, byName))

		if flat(a, b) != grouped(a, b) || flat(a, b) != nested(a, b) {
			t.Fatalf("結合律が破れた: a=%+v, b=%+v", a, b)
		}
	})
}

// 単位元は「常に 0 を返す比較」。挟んでも順序は変わらない。
func TestCompareByUnit(t *testing.T) {
	always0 := func(a, b employee) int { return 0 }

	rapid.Check(t, func(t *rapid.T) {
		a := employeeGen.Draw(t, "a")
		b := employeeGen.Draw(t, "b")

		withUnit := CompareBy(always0, byDept, always0)
		without := CompareBy(byDept)

		if withUnit(a, b) != without(a, b) {
			t.Fatalf("単位律が破れた: a=%+v, b=%+v", a, b)
		}
	})
}

// 実際にソートしても、括り方によらず同じ並びになる。
func TestSortIsUnaffectedByGrouping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOfN(employeeGen, 0, 20).Draw(t, "xs")

		flat := slices.Clone(xs)
		grouped := slices.Clone(xs)

		slices.SortStableFunc(flat, CompareBy(byDept, byAge, byName))
		slices.SortStableFunc(grouped, CompareBy(CompareBy(byDept, byAge), byName))

		if !slices.Equal(flat, grouped) {
			t.Fatalf("括り方でソート結果が変わった:\n  平坦=%+v\n  束ねた=%+v", flat, grouped)
		}
	})
}
