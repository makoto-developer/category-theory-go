# part9 — 余モナド。そしてレンズは Store 余モナドの余代数だった

第3回で Functor・Applicative・Monad を積み上げました。Monad の矢印を裏返すと **Comonad（余モナド）**になります。Monad が「値を文脈に入れる」なら、Comonad は「文脈から値を取り出す」。

> **前提を先に。** ここで余モナドと言えるのは「**空でなく、焦点が範囲内にある** `Zipper`」に限ります。`Extract` は `w.Items[w.Pos]` なので、空だと panic します。部分関数なので、counit 則をそもそも評価できません。加えて `Items` は不変として扱います（`Valid()` と `TestExtractIsUndefinedOnInvalidZipper` を参照）。

```go
Extract  : W[A] → A            // return を裏返したもの（counit）
Duplicate: W[A] → W[W[A]]      // join を裏返したもの
Extend   : (W[A] → B) → W[A] → W[B]
```

## 動かす

```bash
go test -v ./part9/
go test -bench=. -benchmem -run='^$' ./part9/
```

## 何を確認すればいいか

### 1. 移動平均は extend そのもの

```bash
go test -v -run 'MovingAverage|ComonadLaws' ./part9/
```

`Zipper`（焦点つきのリスト）は余モナドの典型例です。「いまの位置の周りを見て1つの値を返す」関数——つまり移動平均——は、そのまま `extend` に渡せる形をしています。

```go
func MovingAverage(n int) func(Zipper[float64]) float64
```

`Extend(MovingAverage(5), w)` で全位置の移動平均が一度に出ます。`TestMovingAverageAgreesWithLoop` が手書きループと同じ答えになることを確認します。

Comonad 則3つも property-based test で検査しています（`extend extract = id`、`extract ∘ extend f = f`、`extend f ∘ extend g = extend (f ∘ extend g)`）。

### 2. duplicate を実体化すると 4.99倍のメモリ

```bash
go test -bench=MovingAverage -benchmem -run='^$' ./part9/
```

法則は `extend f = fmap f ∘ duplicate` と言います。この定義をそのまま実装すると、中間に `W[W[A]]` ができます。

n=100,000 での比較です。

| 方式 | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| 法則どおり（fmap f ∘ duplicate） | 853,967 | 4,005,890 | 2 |
| extend を直接書く | 345,431 | 802,817 | 1 |
| 手書きループ | **266,042** | 802,817 | 1 |

**2.47倍遅く、メモリは 4.99倍。** part6 で見た hylomorphism の中間構造と同じ形の話です。

そして直接版と手書きループは**メモリが同じ**で、時間だけ 1.30倍。差の中身は間接呼び出し・インライン化の可否・一時 `Zipper` の構築などで、切り分けてはいません。

### 3. 予想と逆: duplicate は $O(n)$ だった

```bash
go test -bench=DuplicateAlone -benchmem -run='^$' ./part9/
```

`Duplicate` は $n$ 個の `Zipper` を作ります。各 `Zipper` は $n$ 要素のリストを指しているので、**$O(n^2)$ になると予想していました**。

| n | B/op | 要素あたり B |
|---:|---:|---:|
| 10 | 320 | 32.00 |
| 100 | 3,456 | 34.56 |
| 1,000 | 32,768 | 32.77 |
| 10,000 | 327,680 | 32.77 |
| 100,000 | 3,203,076 | 32.03 |

**n を10,000倍にしても要素あたりは平坦。$O(n)$ です。** ぶれはアロケータのサイズクラスへの切り上げによるもの。

計算量そのものはコードから読めます。`Duplicate` は長さ $n$ のスライスを1本確保して1回走査し、各要素に固定サイズの `Zipper`（スライスヘッダ24B + 位置 int 8B = 32B）を置くだけ。**`Items` を共有するので、配列は1本しかありません。** ベンチマークはその読みと整合しています。

もし `Items` まで複製していたら n=100,000 で $100{,}000^2 \times 8$ バイト——十進で80GB です。

**part7 で「Go の浅いコピーは危険だ」と書きました。同じ性質が、ここでは `duplicate` のメモリを線形に保っています。**

> 正確に言うと、余モナド則そのものは深いコピーでも成り立ちます。浅い共有が効いているのは**メモリ使用量と実用性**のほうです。そして共有されたスライスを呼び出し側が書き換えれば、値として等しい両辺の違いを観測できてしまう——`Items` を不変として扱う、という前提がここにあります。

### 4. レンズは Store 余モナドの余代数だった

```bash
go test -v -run 'Store|Coalgebra|Broken' ./part9/
```

`Store` 余モナドは「位置 $S$ と、位置から値を引く関数 $S \to A$」の組です。

```go
type Store[S, A any] struct {
	Peek func(S) A
	Pos  S
}
```

part7 のレンズ $\mathrm{Lens}[S,A]$ は、**この余モナドの余代数** $S \to \mathrm{Store}[A, S]$ に移せます。

```go
func LensToCoalgebra[S, A any](l Lens[S, A]) func(S) Store[A, S] {
	return func(s S) Store[A, S] {
		return Store[A, S]{
			Peek: func(a A) S { return l.Set(s, a) },
			Pos:  l.Get(s),
		}
	}
}
```

そして**レンズ則3つと余代数則2つは、ちょうど対応します**。

| 余代数則 | レンズ則 |
|---|---|
| counit: $\varepsilon \circ \alpha = \mathrm{id}$ | **get-set** |
| coassociativity: $W(\alpha) \circ \alpha = \delta \circ \alpha$ | **set-get** と **set-set** |

`TestLawfulLensIsAStoreCoalgebra` が、法則を満たすレンズが余代数則も満たすことを確認します。

対応が本物であることは、**破れる側**でも確かめられます。

- `TestBrokenSetSetBreaksCoassociativity` — set-set だけを破るレンズを作ると、coassociativity も破れる
- `TestBrokenSetGetBreaksCoassociativity` — set-get を破ると、同じく破れる

`TestLensCoalgebraRoundTrip` で往復も確認しています。

**part7 で「3則は非干渉を見ていない」と書きました。その3則が、余代数の2則と同じものだった**ということです。見ていないものは、裏返しても見えません。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `comonad.go` | `Zipper`・`Extract`・`Duplicate`・`Extend`・移動平均 |
| `store.go` | `Store` 余モナドと、レンズ ↔ 余代数の往復 |
| `comonad_test.go` | Comonad 則、直接版と duplicate 経由の一致、移動平均 |
| `store_test.go` | 余代数則とレンズ則の対応、破れる側の確認 |
| `bench_test.go` | 3方式の比較、duplicate 単独 |
