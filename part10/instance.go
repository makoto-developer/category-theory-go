package part10

import "fmt"

// Instance はスキーマから Set への関手。
// 各対象に行 ID の集合を、各射に行 ID どうしの写像を割り当てる。
type Instance struct {
	Schema Schema
	// Rows[object] = その対象の行 ID
	Rows map[string][]int
	// Maps[arrow][fromID] = toID
	Maps map[string]map[int]int
}

// CheckTotality は射が全域写像になっているかを見る。
// RDB でいう「外部キーが NULL でなく、参照先が実在する」に対応する。
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
// 「上司は同じ部署にいる」のような制約は、RDB では外部キーでは書けない。
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
