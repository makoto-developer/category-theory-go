package part3

import (
	"errors"
	"testing"

	"pgregory.net/rapid"
)

// 3項目すべてが不正な入力。Applicative なら3件、Monad なら1件しか集まらない。
func TestApplicativeCollectsAllErrors(t *testing.T) {
	app := ValidateApplicative("", "no-at-mark", -1)
	mon := ValidateMonadic("", "no-at-mark", -1)

	if len(app.Errors) != 3 {
		t.Fatalf("Applicative が全部の失敗を集めていない: %d件 %v", len(app.Errors), app.Errors)
	}
	if len(mon.Errors) != 1 {
		t.Fatalf("Monad が途中で止まっていない: %d件 %v", len(mon.Errors), mon.Errors)
	}
	t.Logf("Applicative: %d件 -> %v", len(app.Errors), app.Err())
	t.Logf("Monad      : %d件 -> %v", len(mon.Errors), mon.Err())
}

// 成功する入力では両者は完全に一致する。差が出るのは失敗時だけ。
func TestApplicativeAndMonadicAgreeOnSuccess(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.StringMatching(`[a-z]{1,20}`).Draw(t, "name")
		local := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "local")
		age := rapid.IntRange(0, 150).Draw(t, "age")

		app := ValidateApplicative(name, local+"@example.com", age)
		mon := ValidateMonadic(name, local+"@example.com", age)

		if !app.OK() || !mon.OK() || app.Value != mon.Value {
			t.Fatalf("成功時に結果が食い違った: app=%+v, mon=%+v", app, mon)
		}
	})
}

// 失敗の件数は、Applicative が常に Monad 以上になる。
func TestApplicativeNeverCollectsFewerErrors(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.SampledFrom([]string{"", "ok"}).Draw(t, "name")
		email := rapid.SampledFrom([]string{"", "bad", "a@b.com"}).Draw(t, "email")
		age := rapid.SampledFrom([]int{-1, 200, 30}).Draw(t, "age")

		app := ValidateApplicative(name, email, age)
		mon := ValidateMonadic(name, email, age)

		if len(app.Errors) < len(mon.Errors) {
			t.Fatalf("Applicative のほうが失敗を取りこぼした: app=%d, mon=%d", len(app.Errors), len(mon.Errors))
		}
		if app.OK() != mon.OK() {
			t.Fatalf("成否の判定が食い違った: app=%v, mon=%v", app.OK(), mon.OK())
		}
	})
}

// Validated の Functor 則。
func TestValidatedFunctorLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.Int().Draw(t, "n")
		v := Valid(n)

		if got := MapV(v, func(x int) int { return x }); got.Value != n {
			t.Fatalf("恒等の保存が破れた: got=%d, want=%d", got.Value, n)
		}

		f := func(x int) int { return x * 2 }
		g := func(x int) int { return x + 1 }

		twice := MapV(MapV(v, f), g)
		fused := MapV(v, func(x int) int { return g(f(x)) })

		if twice.Value != fused.Value {
			t.Fatalf("合成の保存が破れた: 2回=%d, 合成=%d", twice.Value, fused.Value)
		}
	})
}

// 蓄積した失敗は errors.Join でまとめられ、errors.Is で個別に判定できる。
func TestErrJoinsAllErrors(t *testing.T) {
	v := ValidateApplicative("", "bad", -1)

	err := v.Err()
	for _, want := range []error{ErrNameEmpty, ErrEmailNoAt, ErrAgeNegative} {
		if !errors.Is(err, want) {
			t.Fatalf("%v が含まれていない: %v", want, err)
		}
	}
}
