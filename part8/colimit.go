package part8

// 集約（GROUP BY）は余極限の側にある。同じキーを持つ行を「同一視する」操作、
// つまり余等化子（coequalizer）にあたる。
//
// 極限どうし（JOIN と JOIN）は入れ替えてよい。引き戻しの補題がそれを保証する。
// では極限と余極限——filter と GROUP BY——は入れ替えてよいのか。
// ここは保証が無いので、条件を見て判断することになる。

type Row struct {
	Key   string
	Value int
}

type Group struct {
	Key   string
	Sum   int
	Count int
}

// GroupSum は同じ Key の行を1つに潰す。キーの順は入力の初出順で安定させる。
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
// 述語がキーだけを見ているので、上と同じ答えになる。押し下げてよい。
func FilterByKeyThenGroup(rows []Row, keepKey func(string) bool) []Group {
	return GroupSum(FilterRows(rows, func(r Row) bool { return keepKey(r.Key) }))
}

// --- 入れ替えてはいけない場合 ---------------------------------------------

// GroupThenFilterBySum は「集約してから合計で絞る」（HAVING SUM(x) > n）。
func GroupThenFilterBySum(rows []Row, min int) []Group {
	return FilterGroups(GroupSum(rows), func(g Group) bool { return g.Sum > min })
}

// FilterBySumThenGroup は、上の述語を素朴に行へ押し下げたもの（WHERE x > n）。
// 述語が集約の結果に依存しているので、これは同じ答えにならない。
func FilterBySumThenGroup(rows []Row, min int) []Group {
	return GroupSum(FilterRows(rows, func(r Row) bool { return r.Value > min }))
}
