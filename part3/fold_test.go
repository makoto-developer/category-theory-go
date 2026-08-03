package part3

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

// (x + 2) * (y + 3)
var sample = Mul{
	L: Add{L: Var{Name: "x"}, R: Num{V: 2}},
	R: Add{L: Var{Name: "y"}, R: Num{V: 3}},
}

var env = map[string]float64{"x": 1, "y": 2}

// 同じ木を、代数を差し替えるだけで違うものに変換できる。再帰は Fold の中の1か所だけ。
func TestFoldInterpretsSameTreeDifferently(t *testing.T) {
	if got := Fold(sample, EvalAlgebra(env)); got != 15 {
		t.Fatalf("評価が違う: got=%v, want=15", got)
	}
	if got := Fold(sample, PrintAlgebra); got != "((x + 2) * (y + 3))" {
		t.Fatalf("整形が違う: got=%q", got)
	}
	if got := Fold(sample, VarsAlgebra); !slices.Equal(got, []string{"x", "y"}) {
		t.Fatalf("変数収集が違う: got=%v", got)
	}
	if got := Fold(sample, DepthAlgebra); got != 3 {
		t.Fatalf("深さが違う: got=%d, want=3", got)
	}
	if got := Fold(sample, CountAlgebra); got != 7 {
		t.Fatalf("節点数が違う: got=%d, want=7", got)
	}
}

// 簡約の結果も Expr なので、そのまま次の fold に渡せる。
func TestSimplifyAlgebra(t *testing.T) {
	// (x * 1) + (0 * y) + 3 は x + 3 になる
	e := Add{
		L: Add{L: Mul{L: Var{Name: "x"}, R: Num{V: 1}}, R: Mul{L: Num{V: 0}, R: Var{Name: "y"}}},
		R: Num{V: 3},
	}

	simplified := Fold(e, SimplifyAlgebra)

	if got := Fold(simplified, PrintAlgebra); got != "(x + 3)" {
		t.Fatalf("簡約が効いていない: got=%q", got)
	}
	if got := Fold(simplified, CountAlgebra); got != 3 {
		t.Fatalf("節点数が減っていない: got=%d, want=3", got)
	}
}

// 簡約しても評価結果は変わらない（意味を保つ変換になっている）。
func TestSimplifyPreservesMeaning(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		e := genExpr(t, 4)

		before := Fold(e, EvalAlgebra(env))
		after := Fold(Fold(e, SimplifyAlgebra), EvalAlgebra(env))

		if before != after {
			t.Fatalf("簡約で意味が変わった: before=%v, after=%v", before, after)
		}
	})
}

// 代数の積: 木を1回たどるだけで2つの答えが同時に出る。
func TestProductAlgebraMatchesTwoFolds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		e := genExpr(t, 4)

		both := Fold(e, ProductAlgebra(EvalAlgebra(env), CountAlgebra))
		wantEval := Fold(e, EvalAlgebra(env))
		wantCount := Fold(e, CountAlgebra)

		if both.First != wantEval || both.Second != wantCount {
			t.Fatalf("積代数の結果が食い違う: got=%+v, want=(%v,%d)", both, wantEval, wantCount)
		}
	})
}

// 代数を経由しても、手書きの再帰と同じ答えになる。
func TestFoldMatchesDirectEval(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		e := genExpr(t, 5)

		if got, want := Fold(e, EvalAlgebra(env)), EvalDirect(e, env); got != want {
			t.Fatalf("Fold と手書きで答えが違う: got=%v, want=%v", got, want)
		}
	})
}

// genExpr は深さ制限つきでランダムな式木を作る。
func genExpr(t *rapid.T, depth int) Expr {
	if depth <= 1 {
		if rapid.Bool().Draw(t, "isVar") {
			return Var{Name: rapid.SampledFrom([]string{"x", "y"}).Draw(t, "name")}
		}
		return Num{V: float64(rapid.IntRange(0, 5).Draw(t, "num"))}
	}
	switch rapid.IntRange(0, 3).Draw(t, "node") {
	case 0:
		return Num{V: float64(rapid.IntRange(0, 5).Draw(t, "num"))}
	case 1:
		return Var{Name: rapid.SampledFrom([]string{"x", "y"}).Draw(t, "name")}
	case 2:
		return Add{L: genExpr(t, depth-1), R: genExpr(t, depth-1)}
	default:
		return Mul{L: genExpr(t, depth-1), R: genExpr(t, depth-1)}
	}
}
