package part6

import (
	"slices"
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

// 3通りの hylomorphism は同じ答えを返す。中間構造の有無は意味を変えない。
func TestHylosAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 200).Draw(t, "n")
		add := func(acc, x int) int { return acc + x }

		want := n * (n + 1) / 2
		for name, got := range map[string]int{
			"HyloVia":   HyloVia(CountTo(n), add, 0, 1),
			"HyloFused": HyloFused(CountTo(n), add, 0, 1),
			"HyloSeq":   HyloSeq(CountTo(n), add, 0, 1),
		} {
			if got != want {
				t.Fatalf("%s = %d, want %d", name, got, want)
			}
		}
	})
}

// Unfold と UnfoldSeq が同じ列を生む。
func TestUnfoldAndSeqAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 100).Draw(t, "n")
		co := CountTo(n)

		var fromSeq []int
		for v := range UnfoldSeq(co, 1) {
			fromSeq = append(fromSeq, v)
		}
		if got := Unfold(co, 1); !slices.Equal(got, fromSeq) {
			t.Fatalf("Unfold=%v, UnfoldSeq=%v", got, fromSeq)
		}
	})
}

// 途中で break しても、余代数はそこから先を評価しない（遅延であること）。
func TestUnfoldSeqStopsOnBreak(t *testing.T) {
	calls := 0
	co := Coalgebra[int, int](func(i int) (int, int, bool) {
		calls++
		return i, i + 1, true // 無限に続く
	})

	var got []int
	for v := range UnfoldSeq(co, 1) {
		got = append(got, v)
		if len(got) == 3 {
			break
		}
	}
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("got=%v", got)
	}
	if calls != 3 {
		t.Fatalf("break したのに余代数が %d 回呼ばれた（3回のはず）", calls)
	}
}

// 木を作っても作らなくてもソート結果は同じ。
func TestMergeSortsAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		xs := rapid.SliceOfN(rapid.Int(), 0, 200).Draw(t, "xs")
		want := slices.Clone(xs)
		slices.Sort(want)

		if got := MergeSortHylo(slices.Clone(xs)); !slices.Equal(got, want) {
			t.Fatalf("MergeSortHylo = %v, want %v", got, want)
		}
		if got := MergeSortFused(slices.Clone(xs)); !slices.Equal(got, want) {
			t.Fatalf("MergeSortFused = %v, want %v", got, want)
		}
	})
}

// ページングは、呼び出し側にループを書かずに全ページを取れる。
func TestPaginateCollectsAllPages(t *testing.T) {
	pages := map[string]Page{
		"":   {Items: []string{"a", "b"}, NextCursor: "c1"},
		"c1": {Items: []string{"c"}, NextCursor: "c2"},
		"c2": {Items: []string{"d", "e"}, NextCursor: ""},
	}
	fetches := 0
	fetch := func(cursor string) Page { fetches++; return pages[cursor] }

	var got []string
	for p := range UnfoldSeq(Paginate(fetch), Cursor{}) {
		got = append(got, p.Items...)
	}

	if want := []string{"a", "b", "c", "d", "e"}; !slices.Equal(got, want) {
		t.Fatalf("got=%v, want=%v", got, want)
	}
	if fetches != 3 {
		t.Fatalf("fetch が %d 回（3回のはず）", fetches)
	}
}

// 進行状態が種に入っているので、同じ種から何度でも展開し直せる。
// 状態をクロージャに隠すと、Unfold を2回走らせたときに2回目が空になる。
// （これは「余代数が純粋である」ことの検査ではない。この例で再開できることの検査。）
func TestPaginateCanRestartFromSameSeed(t *testing.T) {
	pages := map[string]Page{
		"":   {Items: []string{"a"}, NextCursor: "c1"},
		"c1": {Items: []string{"b"}, NextCursor: ""},
	}
	co := Paginate(func(cursor string) Page { return pages[cursor] })

	first := Unfold(co, Cursor{})
	second := Unfold(co, Cursor{})
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("同じ種で2回展開したのにページ数が違う: %d, %d", len(first), len(second))
	}
	for i := range first {
		if !slices.Equal(first[i].Items, second[i].Items) || first[i].NextCursor != second[i].NextCursor {
			t.Fatalf("%d ページ目の中身が違う: %+v vs %+v", i, first[i], second[i])
		}
	}
}

// 余代数を差し替えるだけで生やすものが変わる（fold で代数を差し替えるのと同じ）。
func TestSwappingCoalgebraChangesWhatGrows(t *testing.T) {
	labels := Coalgebra[int, string](func(i int) (string, int, bool) {
		if i > 3 {
			return "", 0, false
		}
		return "#" + strconv.Itoa(i), i + 1, true
	})

	if got, want := Unfold(labels, 1), []string{"#1", "#2", "#3"}; !slices.Equal(got, want) {
		t.Fatalf("got=%v, want=%v", got, want)
	}
}
