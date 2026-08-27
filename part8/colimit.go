package part8

// GROUP BY は2つの操作でできている。
//
//  ① 同じキーの行を同一視する（商を取る）
//  ② 各同値類を可換モノイドへ畳み込む（SUM・COUNT など）
//
// 余等化子（coequalizer）にあたるのは ① だけ。平行2射を明示するとこうなる。
// 行の集合を R、キー写像を k: R → K として、
//
//	E = { (r₁, r₂) ∈ R×R | k(r₁) = k(r₂) },  π₁, π₂ : E ⇉ R
//
// この π₁, π₂ の余等化子が R/~ 、つまり同じキーの行を潰した商になる。
// ② はその上に乗せる別の操作で、余等化子そのものではない。

type Row struct {
	Key   string
	Value int
}

type Group struct {
	Key   string
	Sum   int
	Count int
}

// Quotient は余等化子そのもの。同じキーの行を潰して、キーの代表だけを返す。
// 畳み込みは載せない。
func Quotient(rows []Row) []string {
	seen := make(map[string]bool, len(rows))
	var out []string
	for _, r := range rows {
		if !seen[r.Key] {
			seen[r.Key] = true
			out = append(out, r.Key)
		}
	}
	return out
}

// GroupSum は商を取ったうえで、各同値類を (Sum, Count) へ畳み込む。
// 畳み込み先は可換モノイドなので、同値類の中の順序に依らない。
// キーの順は入力の初出順で安定させる。
func GroupSum(rows []Row) []Group {
	idx := make(map[string]int)
	var out []Group
	for _, r := range rows {
		if i, ok := idx[r.Key]; ok {
			out[i].Sum += r.Value
			out[i].Count++
			continue
		}
		idx[r.Key] = len(out)
		out = append(out, Group{Key: r.Key, Sum: r.Value, Count: 1})
	}
	return out
}

// FilterRows は集約の前に絞る（SQL の WHERE にあたる）。
func FilterRows(rows []Row, keep func(Row) bool) []Row {
	var out []Row
	for _, r := range rows {
		if keep(r) {
			out = append(out, r)
		}
	}
	return out
}

// FilterGroups は集約の後に絞る（SQL の HAVING にあたる）。
func FilterGroups(gs []Group, keep func(Group) bool) []Group {
	var out []Group
	for _, g := range gs {
		if keep(g) {
			out = append(out, g)
		}
	}
	return out
}

// --- 入れ替えてよい場合 ---------------------------------------------------

// GroupThenFilterByKey は「集約してからキーで絞る」。
func GroupThenFilterByKey(rows []Row, keepKey func(string) bool) []Group {
	return FilterGroups(GroupSum(rows), func(g Group) bool { return keepKey(g.Key) })
}

// FilterByKeyThenGroup は「キーで絞ってから集約する」。
// 押し下げてよい理由は「極限と余極限が交換するから」ではなく、
// 述語がキー写像 k を通って因子化する（keep = keepKey ∘ k）ため。
// そのとき述語は各同値類を丸ごと残すか丸ごと捨てるかしかしないので、
// 商を取る前に適用しても後に適用しても同じ類が残る。
func FilterByKeyThenGroup(rows []Row, keepKey func(string) bool) []Group {
	return GroupSum(FilterRows(rows, func(r Row) bool { return keepKey(r.Key) }))
}

// --- 入れ替えてはいけない場合 ---------------------------------------------

// GroupThenFilterBySum は「集約してから合計で絞る」（HAVING SUM(x) > n）。
func GroupThenFilterBySum(rows []Row, min int) []Group {
	return FilterGroups(GroupSum(rows), func(g Group) bool { return g.Sum > min })
}

// FilterBySumThenGroup は、上の述語を素朴に行へ押し下げたもの（WHERE x > n）。
// これは「同じ述語を前に出した」ものではない。SUM は行1件には定義できないので、
// 別の述語をこしらえている。だから答えが変わっても不思議はない——
// この関数が示すのは、この素朴な書き換えが不正だということだけ。
func FilterBySumThenGroup(rows []Row, min int) []Group {
	return GroupSum(FilterRows(rows, func(r Row) bool { return r.Value > min }))
}
