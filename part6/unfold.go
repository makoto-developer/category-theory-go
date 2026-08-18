// Package part6 は連載「Goで書く実践圏論」発展編2本目の検証コード。
//
// 第3回で F代数と fold（catamorphism）を見た。その矢印を全部裏返すと
// F余代数と unfold（anamorphism）になる。fold が「畳んで壊す」なら
// unfold は「種から生やす」。
//
// このパッケージが測るのは、両者を繋いだ hylomorphism（unfold してから fold）で
// 中間構造が消えるかどうか。理論上は消える（deforestation）。Go では消えるのか。
package part6

import "iter"

// Coalgebra は、A を固定したリスト関手 F(X) = 1 + A×X の余代数 S → F(S) を Go で表したもの。
// bool が直和のタグで、false が 1（終わり）、true が A×S（1要素と次の種）にあたる。
// Go に直和型が無いので、false のときも A と S を返せてしまう点は忠実でない。
// Algebra が T を組み立てるのに対して、こちらは S を展開する。
type Coalgebra[S, A any] func(S) (A, S, bool)

// Unfold は種から有限の列を生やす。停止しない余代数を渡すと戻らない
// （無限に生やしたいときは UnfoldSeq のほう）。
func Unfold[S, A any](co Coalgebra[S, A], seed S) []A {
	var out []A
	for {
		a, next, ok := co(seed)
		if !ok {
			return out
		}
		out = append(out, a)
		seed = next
	}
}

// UnfoldSeq は同じ展開を iter.Seq として返す。中間のスライスを作らない。
func UnfoldSeq[S, A any](co Coalgebra[S, A], seed S) iter.Seq[A] {
	return func(yield func(A) bool) {
		for {
			a, next, ok := co(seed)
			if !ok || !yield(a) {
				return
			}
			seed = next
		}
	}
}

// --- 余代数の例 ---------------------------------------------------------

// CountTo は 1..n を生やす余代数。
func CountTo(n int) Coalgebra[int, int] {
	return func(i int) (int, int, bool) {
		if i > n {
			return 0, 0, false
		}
		return i, i + 1, true
	}
}

// Page はカーソル方式のページング。次カーソルが空なら終わり、という実務でよくある形。
type Page struct {
	Items      []string
	NextCursor string
}

// Cursor はページングの種。「次があるか」を種そのものが持つ。
// 進行状態をクロージャに隠さず種に入れておくと、同じ種から何度でも展開し直せる。
// （Go の型は純粋性を保証しないし、fetch 自身は通信するので純粋ではない。
// ここで揃えているのは「種が同じなら同じところから再開できる」ことまで。）
type Cursor struct {
	Value string
	Done  bool
}

// Paginate はカーソルを種として、ページを生やす余代数。
// 終了条件が種に入っているので、呼び出し側にループも打ち切り判定も要らない。
func Paginate(fetch func(cursor string) Page) Coalgebra[Cursor, Page] {
	return func(c Cursor) (Page, Cursor, bool) {
		if c.Done {
			return Page{}, c, false
		}
		p := fetch(c.Value)
		return p, Cursor{Value: p.NextCursor, Done: p.NextCursor == ""}, true
	}
}
