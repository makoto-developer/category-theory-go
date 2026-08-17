package part5

// 同じ「文字列の列」を表すモノイドを2つ用意する。違いは Append のコストが
// どちら側に付いているかだけで、単位元も結合律も同じように成り立つ。
// この2つは要素列の対応でモノイド同型（cayley_test.go で検査する）。
// 同型は演算と単位元を保つが、実行コストは保たない。そこを分けるのがこのパートの主題。

// --- cons リスト: Append のコストは左辺に付く ---------------------------

// List は先頭に伸びるリスト。
type List struct {
	Head string
	Tail *List
}

// Concat は左辺の背骨を作り直し、右辺はそのまま共有する。よって O(|a|)。
func Concat(a, b *List) *List {
	if a == nil {
		return b
	}
	return &List{Head: a.Head, Tail: Concat(a.Tail, b)}
}

// ConsMonoid の Append は左辺のぶんだけ払う。左結合で畳むと acc が毎回コピーされる。
var ConsMonoid = Monoid[*List]{
	Empty:  nil,
	Append: Concat,
}

func SingleCons(s string) *List { return &List{Head: s} }

func (l *List) Slice() []string {
	var out []string
	for ; l != nil; l = l.Tail {
		out = append(out, l.Head)
	}
	return out
}

// PrependReverse は Go で普通に書く形。前に積んで最後に反転する。
func PrependReverse(xs []string) *List {
	var acc *List
	for _, x := range xs {
		acc = &List{Head: x, Tail: acc}
	}
	var out *List
	for ; acc != nil; acc = acc.Tail {
		out = &List{Head: acc.Head, Tail: out}
	}
	return out
}

// --- snoc リスト: Append のコストは右辺に付く ---------------------------

// RList は末尾に伸びるリスト。cons リストの矢印を裏返しただけ。
type RList struct {
	Init *RList
	Last string
}

// ConcatR は右辺の背骨を作り直し、左辺を共有する。よって O(|b|)。
func ConcatR(a, b *RList) *RList {
	if b == nil {
		return a
	}
	return &RList{Init: ConcatR(a, b.Init), Last: b.Last}
}

// SnocMonoid は ConsMonoid と同じ法則を満たすが、コストの付き方だけが逆。
var SnocMonoid = Monoid[*RList]{
	Empty:  nil,
	Append: ConcatR,
}

func SingleSnoc(s string) *RList { return &RList{Last: s} }

func (l *RList) Slice() []string {
	if l == nil {
		return nil
	}
	return append(l.Init.Slice(), l.Last)
}
