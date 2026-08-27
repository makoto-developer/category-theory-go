package part10

import "fmt"

// Instance は「関手の候補」を表すデータ構造。それ単独では関手ではない。
//
// 商圏 Schema から Set への関手を定めるには、次が全部要る。
//
//	Schema.Validate()  … 表示自身が整合している
//	CheckTotality()    … 生成射が全域で、行き先が終域の行集合に入る
//	CheckEquations()   … 提示されたパス等式を保存する
//
// 参照整合性に対応するのは2番目だけ。関手性はもっと広い。
type Instance struct {
	Schema Schema
	// Rows[object] = その対象の行 ID
	Rows map[string][]int
	// Maps[arrow][fromID] = toID
	Maps map[string]map[int]int
}

// BuildIndex は対象ごとの行集合を作る。CheckTotalityWithIndex に渡す。
func (i Instance) BuildIndex() map[string]map[int]bool {
	idx := make(map[string]map[int]bool, len(i.Rows))
	for obj, rows := range i.Rows {
		m := make(map[int]bool, len(rows))
		for _, id := range rows {
			m[id] = true
		}
		idx[obj] = m
	}
	return idx
}

// CheckTotalityWithIndex は索引を作り直さない版。
// CheckTotality と CheckEquations を比べるとき、索引構築の有無が混ざらないようにする。
func (i Instance) CheckTotalityWithIndex(idx map[string]map[int]bool) []string {
	var bad []string
	for name, a := range i.Schema.Arrows {
		m := i.Maps[name]
		target := idx[a.To]
		for _, id := range i.Rows[a.From] {
			to, ok := m[id]
			if !ok {
				bad = append(bad, fmt.Sprintf("射 %s: %s の行 %d に行き先が無い", name, a.From, id))
				continue
			}
			if !target[to] {
				bad = append(bad, fmt.Sprintf("射 %s: %s の行 %d の行き先 %d が %s に無い", name, a.From, id, to, a.To))
			}
		}
	}
	return bad
}

// CheckTotality は射が全域写像になっているかを見る。
// RDB でいう「外部キーが NULL でなく、参照先が実在する」に対応する。
// 呼ぶたびに終域の索引を作り直すので、索引を使い回せる場面では
// CheckTotalityWithIndex のほうが速い。
func (i Instance) CheckTotality() []string {
	var bad []string
	for name, a := range i.Schema.Arrows {
		m := i.Maps[name]
		target := make(map[int]bool, len(i.Rows[a.To]))
		for _, id := range i.Rows[a.To] {
			target[id] = true
		}
		for _, id := range i.Rows[a.From] {
			to, ok := m[id]
			if !ok {
				bad = append(bad, fmt.Sprintf("射 %s: %s の行 %d に行き先が無い", name, a.From, id))
				continue
			}
			if !target[to] {
				bad = append(bad, fmt.Sprintf("射 %s: %s の行 %d の行き先 %d が %s に無い", name, a.From, id, to, a.To))
			}
		}
	}
	return bad
}

// follow はパスを1つの行 ID に適用する。
func (i Instance) follow(path []string, id int) (int, bool) {
	cur := id
	for _, name := range path {
		next, ok := i.Maps[name][cur]
		if !ok {
			return 0, false
		}
		cur = next
	}
	return cur, true
}

// CheckEquations はパス等式が全行で成り立つかを見る。
// 「上司は同じ部署にいる」のような制約は、単独の外部キー制約では書けない。
//
// これは CheckTotality の代わりにはならない。等式に現れない射の欠落は拾わないし、
// follow は途中の値が各対象の Rows に入っているかを確認しない。両方が要る。
func (i Instance) CheckEquations() []string {
	var bad []string
	for _, e := range i.Schema.Equations {
		for _, id := range i.Rows[e.From] {
			l, okL := i.follow(e.Left, id)
			r, okR := i.follow(e.Right, id)
			if !okL || !okR {
				bad = append(bad, fmt.Sprintf("等式 %q: 行 %d でパスが辿れない", e, id))
				continue
			}
			if l != r {
				bad = append(bad, fmt.Sprintf("等式 %q: 行 %d で %d ≠ %d", e, id, l, r))
			}
		}
	}
	return bad
}
