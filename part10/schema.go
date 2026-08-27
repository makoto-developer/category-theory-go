// Package part10 は連載「Goで書く実践圏論」発展編6本目の検証コード。
//
// スキーマを圏、データを関手として扱う（Spivak の functorial data migration）。
// 対象がテーブル、射が外部キー、そして射の間に「パス等式」を置ける。
// パス等式は普通の RDB が書けない種類の制約で、そこがこの見方の実利になる。
package part10

import (
	"fmt"
	"slices"
	"strings"
)

// Schema は圏そのものではなく、**圏の表示**（presentation）。
// 対象と生成射からなる有向グラフと、パスの間の等式を持つ。
//
// ここから圏を作るには、グラフ上の自由圏（=すべてのパス）を取り、
// Equations が生成する最小の合同関係で割る。合成射も恒等射も、
// この構成で初めて出てくる。Schema 値そのものには合成演算が無い。
//
// なお、インスタンスの検査では合同閉包を作る必要はない。
// 生成関係をすべて保存する写像は、合成によって派生する等式も自動的に保存するため。
type Schema struct {
	Objects []string
	// Arrows[name] = (from, to)
	Arrows map[string]Arrow
	// Equations は「2つのパスが等しい」という制約。恒等射は空パスで表す。
	Equations []Equation
}

type Arrow struct {
	From string
	To   string
}

// Equation は同じ対象から同じ対象への2本のパスが等しい、という制約。
type Equation struct {
	From  string
	To    string
	Left  []string // 射の名前を左から順に適用する
	Right []string
}

func (e Equation) String() string {
	l, r := strings.Join(e.Left, " ; "), strings.Join(e.Right, " ; ")
	if l == "" {
		l = "id"
	}
	if r == "" {
		r = "id"
	}
	return fmt.Sprintf("%s: %s = %s", e.From, l, r)
}

// Validate はスキーマ自身の整合を見る。パスが繋がっているか、始点と終点が合うか。
func (s Schema) Validate() error {
	for name, a := range s.Arrows {
		if !slices.Contains(s.Objects, a.From) || !slices.Contains(s.Objects, a.To) {
			return fmt.Errorf("射 %q の端点がスキーマに無い", name)
		}
	}
	for _, e := range s.Equations {
		for _, side := range [][]string{e.Left, e.Right} {
			at := e.From
			for _, name := range side {
				a, ok := s.Arrows[name]
				if !ok {
					return fmt.Errorf("等式 %q に未知の射 %q", e, name)
				}
				if a.From != at {
					return fmt.Errorf("等式 %q のパスが繋がっていない（%s の手前が %s）", e, name, at)
				}
				at = a.To
			}
			if at != e.To {
				return fmt.Errorf("等式 %q の終点が %s になっている", e, at)
			}
		}
	}
	return nil
}
