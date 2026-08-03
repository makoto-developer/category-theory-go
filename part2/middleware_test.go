package part2

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

// 結合律: middleware をどう括ってまとめても、実行順は変わらない。
// 「認証系だけ先に束ねておく」といったリファクタリングが安全なのはこのため。
func TestChainIsAssociative(t *testing.T) {
	var flat, grouped, nested []string

	run := func(build func(a, b, c Middleware) Middleware, order *[]string) {
		a, b, c := Tap("a", order), Tap("b", order), Tap("c", order)
		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, "handler")
		})
		build(a, b, c)(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	run(func(a, b, c Middleware) Middleware { return Chain(a, b, c) }, &flat)
	run(func(a, b, c Middleware) Middleware { return Chain(Chain(a, b), c) }, &grouped)
	run(func(a, b, c Middleware) Middleware { return Chain(a, Chain(b, c)) }, &nested)

	if !slices.Equal(flat, grouped) || !slices.Equal(flat, nested) {
		t.Fatalf("結合律が破れた:\n  平坦=%v\n  前2つを束ねた=%v\n  後2つを束ねた=%v", flat, grouped, nested)
	}
	t.Logf("実行順: %v", flat)
}

// 単位律: 何もしない middleware を挟んでも実行順は変わらない。
func TestPassthroughIsUnit(t *testing.T) {
	var withUnit, without []string

	run := func(build func(a Middleware) Middleware, order *[]string) {
		a := Tap("a", order)
		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, "handler")
		})
		build(a)(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	run(func(a Middleware) Middleware { return Chain(Passthrough, a, Passthrough) }, &withUnit)
	run(func(a Middleware) Middleware { return Chain(a) }, &without)

	if !slices.Equal(withUnit, without) {
		t.Fatalf("単位律が破れた: 素通しあり=%v, なし=%v", withUnit, without)
	}
}

// 空のチェーンは恒等射。設定なしで素通しになるのは、この構造の帰結。
func TestEmptyChainIsIdentity(t *testing.T) {
	called := false
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })

	Chain()(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("空のチェーンがハンドラに届いていない")
	}
}

// 結合律はあるが可換律は無い。括り方は変えてよいが、並び順は変えてはいけない。
func TestChainIsNotCommutative(t *testing.T) {
	var forward, reversed []string

	run := func(build func(a, b Middleware) Middleware, order *[]string) {
		a, b := Tap("a", order), Tap("b", order)
		final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, "handler")
		})
		build(a, b)(final).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	run(func(a, b Middleware) Middleware { return Chain(a, b) }, &forward)
	run(func(a, b Middleware) Middleware { return Chain(b, a) }, &reversed)

	if slices.Equal(forward, reversed) {
		t.Fatal("順序を入れ替えても同じ結果になった（可換になっている）")
	}
	t.Logf("a,b の順: %v", forward)
	t.Logf("b,a の順: %v", reversed)
}
