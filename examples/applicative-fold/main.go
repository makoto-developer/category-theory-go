// 連載「Goで書く実践圏論」第3回のスニペット。
//
// Applicative と Monad で画面が変わること、F代数で再帰が1か所に閉じること、
// iter.Seq が継続渡しだったこと。この3つを実際に走らせて確かめます。
package main

import (
	"fmt"
	"iter"
	"strings"
)

// 案内は先頭に出す。記事に埋め込むと末尾しか見えないので、末尾は検証結果のために空けておく。
func main() {
	fmt.Println("連載「Goで書く実践圏論」第3回")
	fmt.Println("  記事: https://blog.makoto-developer.net/articles/2026-08-06-practical-category-theory-go-3")
	fmt.Println("  全4回のコード（テスト・ベンチ込み）: https://github.com/makoto-developer/category-theory-go")
	fmt.Println("──────────────────────────────────────────")

	showApplicativeVsMonad()
	showOneFoldManyMeanings()
	showSeqIsContinuationPassing()
}

// ---- Applicative と Monad ----------------------------------------------

// Validated は「成功した値」か「失敗の一覧」を持つ。
type Validated[T any] struct {
	Value  T
	Errors []string
}

func Valid[T any](v T) Validated[T]              { return Validated[T]{Value: v} }
func Invalid[T any](msgs ...string) Validated[T] { return Validated[T]{Errors: msgs} }

func (v Validated[T]) OK() bool { return len(v.Errors) == 0 }

// Combine3 は Applicative。3つを「同時に」見るので、失敗はすべて集まる。
func Combine3[A, B, C, R any](a Validated[A], b Validated[B], c Validated[C], f func(A, B, C) R) Validated[R] {
	if errs := append(append(append([]string{}, a.Errors...), b.Errors...), c.Errors...); len(errs) > 0 {
		return Validated[R]{Errors: errs}
	}
	return Valid(f(a.Value, b.Value, c.Value))
}

// AndThen は Monad。前の結果に次が依存できるかわりに、最初の失敗で止まる。
func AndThen[A, B any](a Validated[A], f func(A) Validated[B]) Validated[B] {
	if !a.OK() {
		return Validated[B]{Errors: a.Errors}
	}
	return f(a.Value)
}

type user struct {
	name  string
	email string
	age   int
}

func validName(s string) Validated[string] {
	if s == "" {
		return Invalid[string]("名前が空です")
	}
	return Valid(s)
}

func validEmail(s string) Validated[string] {
	if !strings.Contains(s, "@") {
		return Invalid[string]("メールアドレスに @ が含まれていません")
	}
	return Valid(s)
}

func validAge(n int) Validated[int] {
	if n < 0 {
		return Invalid[int]("年齢が負の数です")
	}
	return Valid(n)
}

// 同じ「3項目すべてが不正」な入力を、Applicative と Monad の両方に通す。
// 成否の判定は一致し、返る失敗の件数だけが違う。
func showApplicativeVsMonad() {
	name, email, age := "", "not-an-email", -1

	applicative := Combine3(validName(name), validEmail(email), validAge(age),
		func(n, e string, a int) user { return user{n, e, a} })

	monadic := AndThen(validName(name), func(n string) Validated[user] {
		return AndThen(validEmail(email), func(e string) Validated[user] {
			return AndThen(validAge(age), func(a int) Validated[user] {
				return Valid(user{n, e, a})
			})
		})
	})

	fmt.Printf("\n[Applicative] %d件 -> %s\n", len(applicative.Errors), strings.Join(applicative.Errors, " / "))
	fmt.Printf("[Monad]       %d件 -> %s\n", len(monadic.Errors), strings.Join(monadic.Errors, " / "))
	fmt.Println("              Monad で書くとユーザーは3往復する。「エラーは全部まとめて返せ」は")
	fmt.Println("              圏論の言葉では「その検証は Applicative で書け」だった")
}

// ---- F代数 -------------------------------------------------------------

// Expr は式木。構成子は3つだけ。
type Expr struct {
	Kind string // "num" | "var" | "add" | "mul"
	Num  int
	Name string
	L, R *Expr
}

// Algebra は「各構成子をどう畳むか」を集めたもの。ここを差し替えると意味が変わる。
type Algebra[T any] struct {
	Num func(int) T
	Var func(string) T
	Add func(T, T) T
	Mul func(T, T) T
}

// Fold は木を1回たどる。再帰はこの関数の中の1か所にしかない。
func Fold[T any](e *Expr, alg Algebra[T]) T {
	switch e.Kind {
	case "num":
		return alg.Num(e.Num)
	case "var":
		return alg.Var(e.Name)
	case "add":
		return alg.Add(Fold(e.L, alg), Fold(e.R, alg))
	default:
		return alg.Mul(Fold(e.L, alg), Fold(e.R, alg))
	}
}

// 同じ木に代数を差し替えるだけで、評価も整形も変数の収集もできる。
func showOneFoldManyMeanings() {
	// (2 + x) * 3
	tree := mul(add(num(2), variable("x")), num(3))

	env := map[string]int{"x": 5}
	eval := Algebra[int]{
		Num: func(n int) int { return n },
		Var: func(s string) int { return env[s] },
		Add: func(l, r int) int { return l + r },
		Mul: func(l, r int) int { return l * r },
	}
	pretty := Algebra[string]{
		Num: func(n int) string { return fmt.Sprint(n) },
		Var: func(s string) string { return s },
		Add: func(l, r string) string { return "(" + l + " + " + r + ")" },
		Mul: func(l, r string) string { return "(" + l + " * " + r + ")" },
	}
	vars := Algebra[[]string]{
		Num: func(int) []string { return nil },
		Var: func(s string) []string { return []string{s} },
		Add: func(l, r []string) []string { return append(l, r...) },
		Mul: func(l, r []string) []string { return append(l, r...) },
	}

	fmt.Printf("\n[F代数] 整形 = %s\n", Fold(tree, pretty))
	fmt.Printf("        評価 = %d （x=5）\n", Fold(tree, eval))
	fmt.Printf("        変数 = %v\n", Fold(tree, vars))
	fmt.Println("        再帰は Fold の中の1か所だけ。構成子を足すと既存の代数が全部コンパイルエラーになる")
	fmt.Println("        （switch の網羅性が検査されない問題が、構造体のフィールドとして解決している）")
}

func num(n int) *Expr         { return &Expr{Kind: "num", Num: n} }
func variable(s string) *Expr { return &Expr{Kind: "var", Name: s} }
func add(l, r *Expr) *Expr    { return &Expr{Kind: "add", L: l, R: r} }
func mul(l, r *Expr) *Expr    { return &Expr{Kind: "mul", L: l, R: r} }

// ---- iter.Seq と継続渡し -----------------------------------------------

// Cont は継続渡し表現。「値そのもの」ではなく「値を受け取る関数を受け取るもの」。
type Cont[A any] func(func(A))

// ToCont は iter.Seq を継続渡しに移す。……が、見ての通り中身は同じ。
// range-over-func は最初から継続渡しだった。
func ToCont[A any](seq iter.Seq[A]) Cont[A] {
	return func(yield func(A)) {
		for v := range seq {
			yield(v)
		}
	}
}

// 継続に恒等射を渡すと値が出てくる。逆向きは戻らないので米田の補題の同型ではない（記事 3.2）。
func Collect[A any](c Cont[A]) []A {
	var out []A
	c(func(a A) { out = append(out, a) })
	return out
}

func showSeqIsContinuationPassing() {
	seq := func(yield func(int) bool) {
		for i := 1; i <= 5; i++ {
			if !yield(i * i) {
				return
			}
		}
	}

	fmt.Printf("\n[iter.Seq] 継続に恒等射を渡すと値が出てくる: %v\n", Collect(ToCont(iter.Seq[int](seq))))
	fmt.Println("           range-over-func は最初から継続渡し表現だった。ToCont は素通しになる")
	fmt.Println("           手元のベンチでは iter.Pull がこれの120倍遅い。同型であることと同じ値段は別")
}
