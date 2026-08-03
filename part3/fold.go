package part3

import (
	"fmt"
	"slices"
)

// Expr は算術式の木。sealed interface で構成子を固定する。
type Expr interface{ isExpr() }

type Num struct{ V float64 }
type Var struct{ Name string }
type Add struct{ L, R Expr }
type Mul struct{ L, R Expr }

func (Num) isExpr() {}
func (Var) isExpr() {}
func (Add) isExpr() {}
func (Mul) isExpr() {}

// Algebra は「各構成子をどう畳むか」の指定。圏論でいう F代数にあたる。
// 木の形（再帰の仕方）はここに現れず、Fold の側だけが知っている。
type Algebra[T any] struct {
	Num func(float64) T
	Var func(string) T
	Add func(T, T) T
	Mul func(T, T) T
}

// Fold は木を1回たどって T にする（catamorphism）。
// 再帰の書き方はここ1か所にしかないので、解釈を増やしても再帰は増えない。
func Fold[T any](e Expr, alg Algebra[T]) T {
	switch n := e.(type) {
	case Num:
		return alg.Num(n.V)
	case Var:
		return alg.Var(n.Name)
	case Add:
		return alg.Add(Fold(n.L, alg), Fold(n.R, alg))
	case Mul:
		return alg.Mul(Fold(n.L, alg), Fold(n.R, alg))
	}
	panic(fmt.Sprintf("未知の構成子: %T", e))
}

// EvalAlgebra は式を評価する。
func EvalAlgebra(env map[string]float64) Algebra[float64] {
	return Algebra[float64]{
		Num: func(v float64) float64 { return v },
		Var: func(name string) float64 { return env[name] },
		Add: func(l, r float64) float64 { return l + r },
		Mul: func(l, r float64) float64 { return l * r },
	}
}

// PrintAlgebra は式を文字列にする。
var PrintAlgebra = Algebra[string]{
	Num: func(v float64) string { return fmt.Sprintf("%g", v) },
	Var: func(name string) string { return name },
	Add: func(l, r string) string { return "(" + l + " + " + r + ")" },
	Mul: func(l, r string) string { return "(" + l + " * " + r + ")" },
}

// VarsAlgebra は出現する変数名を集める。
var VarsAlgebra = Algebra[[]string]{
	Num: func(float64) []string { return nil },
	Var: func(name string) []string { return []string{name} },
	Add: mergeVars,
	Mul: mergeVars,
}

func mergeVars(l, r []string) []string {
	out := append(append([]string{}, l...), r...)
	slices.Sort(out)
	return slices.Compact(out)
}

// DepthAlgebra は木の深さを測る。
var DepthAlgebra = Algebra[int]{
	Num: func(float64) int { return 1 },
	Var: func(string) int { return 1 },
	Add: func(l, r int) int { return 1 + max(l, r) },
	Mul: func(l, r int) int { return 1 + max(l, r) },
}

// CountAlgebra は節点数を数える。
var CountAlgebra = Algebra[int]{
	Num: func(float64) int { return 1 },
	Var: func(string) int { return 1 },
	Add: func(l, r int) int { return 1 + l + r },
	Mul: func(l, r int) int { return 1 + l + r },
}

// SimplifyAlgebra は畳み込みながら式を簡約する。
// 結果が Expr なので、fold の行き先は数値や文字列でなくてもよいと分かる。
var SimplifyAlgebra = Algebra[Expr]{
	Num: func(v float64) Expr { return Num{V: v} },
	Var: func(name string) Expr { return Var{Name: name} },
	Add: func(l, r Expr) Expr {
		ln, lok := l.(Num)
		rn, rok := r.(Num)
		switch {
		case lok && rok:
			return Num{V: ln.V + rn.V}
		case lok && ln.V == 0:
			return r
		case rok && rn.V == 0:
			return l
		}
		return Add{L: l, R: r}
	},
	Mul: func(l, r Expr) Expr {
		ln, lok := l.(Num)
		rn, rok := r.(Num)
		switch {
		case lok && rok:
			return Num{V: ln.V * rn.V}
		case (lok && ln.V == 0) || (rok && rn.V == 0):
			return Num{V: 0}
		case lok && ln.V == 1:
			return r
		case rok && rn.V == 1:
			return l
		}
		return Mul{L: l, R: r}
	},
}

// Pair は2つの結果を同時に運ぶ。
type Pair[A, B any] struct {
	First  A
	Second B
}

// ProductAlgebra は2つの代数を1つにまとめる。木を1回たどるだけで両方の答えが出る。
// 代数そのものが合成できる、というのがF代数を使ういちばんの利点。
func ProductAlgebra[A, B any](a Algebra[A], b Algebra[B]) Algebra[Pair[A, B]] {
	return Algebra[Pair[A, B]]{
		Num: func(v float64) Pair[A, B] { return Pair[A, B]{a.Num(v), b.Num(v)} },
		Var: func(name string) Pair[A, B] { return Pair[A, B]{a.Var(name), b.Var(name)} },
		Add: func(l, r Pair[A, B]) Pair[A, B] {
			return Pair[A, B]{a.Add(l.First, r.First), b.Add(l.Second, r.Second)}
		},
		Mul: func(l, r Pair[A, B]) Pair[A, B] {
			return Pair[A, B]{a.Mul(l.First, r.First), b.Mul(l.Second, r.Second)}
		},
	}
}

// EvalDirect は代数を経由せず手書きで評価する。Fold との速度比較用。
func EvalDirect(e Expr, env map[string]float64) float64 {
	switch n := e.(type) {
	case Num:
		return n.V
	case Var:
		return env[n.Name]
	case Add:
		return EvalDirect(n.L, env) + EvalDirect(n.R, env)
	case Mul:
		return EvalDirect(n.L, env) * EvalDirect(n.R, env)
	}
	panic(fmt.Sprintf("未知の構成子: %T", e))
}
