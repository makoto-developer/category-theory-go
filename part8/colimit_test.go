package part8

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

func groupKeys(gs []Group) []string {
	out := make([]string, 0, len(gs))
	for _, g := range gs {
		out = append(out, fmt.Sprintf("%s:%d:%d", g.Key, g.Sum, g.Count))
	}
	slices.Sort(out)
	return out
}

func genRows() *rapid.Generator[[]Row] {
	return rapid.SliceOfN(rapid.Custom(func(t *rapid.T) Row {
		return Row{
			Key:   rapid.SampledFrom([]string{"a", "b", "c", "d"}).Draw(t, "key"),
			Value: rapid.IntRange(-5, 20).Draw(t, "v"),
		}
	}), 0, 30)
}

// 述語がキーだけを見ているなら、集約の前後どちらで絞っても同じ答えになる。
// 押し下げてよい。
func TestFilterOnKeyCommutesWithGrouping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rows := genRows().Draw(t, "rows")
		keep := func(k string) bool { return k == "a" || k == "c" }

		after := groupKeys(GroupThenFilterByKey(rows, keep))
		before := groupKeys(FilterByKeyThenGroup(rows, keep))
		if !slices.Equal(after, before) {
			t.Fatalf("キーで絞る場合は一致するはず:\n後 %v\n前 %v", after, before)
		}
	})
}

// 述語が集約の結果を見ているなら、押し下げると答えが変わる。
// 「変わることがある」ではなく、具体的に変わる例を出す。
func TestFilterOnAggregateDoesNotCommute(t *testing.T) {
	rows := []Row{
		{"a", 4}, {"a", 4}, {"a", 4}, // 合計 12
		{"b", 20}, // 合計 20
	}
	const min = 10

	after := GroupThenFilterBySum(rows, min)  // HAVING SUM(v) > 10
	before := FilterBySumThenGroup(rows, min) // WHERE v > 10

	if slices.Equal(groupKeys(after), groupKeys(before)) {
		t.Fatalf("差が出ないなら、この例は主張の裏づけにならない")
	}
	t.Logf("HAVING: %v", groupKeys(after))
	t.Logf("WHERE : %v", groupKeys(before))
}

// 押し下げが安全かどうかを、無作為なデータで数える。
// キーだけを見る述語は常に一致し、集約を見る述語は頻繁に食い違う。
func TestPushdownSafetyCounts(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	sameKey, sameAgg, n := 0, 0, 500
	for i := 0; i < n; i++ {
		rows := make([]Row, 1+r.IntN(20))
		for j := range rows {
			rows[j] = Row{Key: strconv.Itoa(r.IntN(4)), Value: r.IntN(15)}
		}
		keep := func(k string) bool { return k == "1" || k == "2" }
		if slices.Equal(groupKeys(GroupThenFilterByKey(rows, keep)), groupKeys(FilterByKeyThenGroup(rows, keep))) {
			sameKey++
		}
		if slices.Equal(groupKeys(GroupThenFilterBySum(rows, 10)), groupKeys(FilterBySumThenGroup(rows, 10))) {
			sameAgg++
		}
	}
	t.Logf("キーだけを見る述語: %d/%d 一致", sameKey, n)
	t.Logf("集約を見る述語　　: %d/%d 一致", sameAgg, n)
	if sameKey != n {
		t.Fatalf("キーで絞る場合は常に一致するはずが %d/%d", sameKey, n)
	}
	if sameAgg == n {
		t.Fatalf("集約で絞る場合に一度も食い違わないなら、例が弱い")
	}
}

var sinkGroups []Group

// 押し下げてよい場合、押し下げるとどれだけ得か。
func BenchmarkFilterPushdown(b *testing.B) {
	r := rand.New(rand.NewPCG(9, 10))
	rows := make([]Row, 200_000)
	for i := range rows {
		rows[i] = Row{Key: strconv.Itoa(r.IntN(1000)), Value: r.IntN(100)}
	}
	// 1000キーのうち10キーだけ残す
	keep := func(k string) bool { n, _ := strconv.Atoi(k); return n < 10 }

	b.Run("1_group_then_filter", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGroups = GroupThenFilterByKey(rows, keep)
		}
	})
	b.Run("2_filter_then_group", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGroups = FilterByKeyThenGroup(rows, keep)
		}
	})
}

// 余等化子の普遍性。h: R → X が同じキーの行を区別しない（h∘π₁ = h∘π₂）なら、
// h は商を経由してただ一通りに分解する。
// 「GROUP BY は余等化子」と言うなら、これが成り立っていないといけない。
func TestQuotientUniversalProperty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rows := genRows().Draw(t, "rows")
		if len(rows) == 0 {
			return
		}
		// h は「キーだけを見る」関数。これは h∘π₁ = h∘π₂ を満たす。
		suffix := rapid.String().Draw(t, "suffix")
		h := func(r Row) string { return r.Key + suffix }

		// 平行2射の条件: 同じキーの2行に h を当てたら同じ値になる
		for _, r1 := range rows {
			for _, r2 := range rows {
				if r1.Key == r2.Key && h(r1) != h(r2) {
					t.Fatalf("h が平行2射を等化していない")
				}
			}
		}

		// 商を経由する分解 hBar が、ただ一通りに決まる
		q := Quotient(rows)
		hBar := make(map[string]string, len(q))
		for _, k := range q {
			hBar[k] = k + suffix
		}
		for _, r := range rows {
			if got, ok := hBar[r.Key]; !ok || got != h(r) {
				t.Fatalf("h = hBar ∘ quotient になっていない: %q vs %q", got, h(r))
			}
		}
		if len(hBar) != len(q) {
			t.Fatalf("分解が一意でない")
		}
	})
}

// h が同じキーの行を区別するなら、商を経由できない。
// 余等化子の普遍性が「条件つき」であることの裏。
func TestQuotientRejectsKeyBlindFunctions(t *testing.T) {
	rows := []Row{{"a", 1}, {"a", 2}}
	h := func(r Row) string { return r.Key + strconv.Itoa(r.Value) } // 値まで見てしまう
	if h(rows[0]) == h(rows[1]) {
		t.Fatalf("この h は2行を区別するはず")
	}
	// 商は "a" の1点しかないので、h の2つの値を分けて表せない
	if len(Quotient(rows)) != 1 {
		t.Fatalf("商が1点でなければこの例は成り立たない")
	}
	t.Logf("h は同じキーの2行に別の値 (%s, %s) を与えるので、1点の商を経由できない", h(rows[0]), h(rows[1]))
}

// 「述語が集約に触れていたら押し下げ不可」は言い過ぎ。
// COUNT(*) > 0 は、存在する群では常に真なので、恒真述語に置き換えて押し下げられる。
// 判定基準は「集約に触れているか」ではなく「行単位の述語へ持ち上げられるか」。
func TestSomeAggregatePredicatesAreStillSafe(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		rows := genRows().Draw(t, "rows")

		// HAVING COUNT(*) > 0
		after := FilterGroups(GroupSum(rows), func(g Group) bool { return g.Count > 0 })
		// 恒真述語として押し下げる
		before := GroupSum(FilterRows(rows, func(Row) bool { return true }))

		if !slices.Equal(groupKeys(after), groupKeys(before)) {
			t.Fatalf("COUNT(*) > 0 は押し下げられるはず:\n後 %v\n前 %v", groupKeys(after), groupKeys(before))
		}
	})
}
