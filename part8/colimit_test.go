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
