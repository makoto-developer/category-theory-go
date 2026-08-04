# part2 — 標準ライブラリに潜む構造

記事: [middleware も errors.Join も圏論だった](https://blog.makoto-developer.net/articles/2026-08-05-practical-category-theory-go-2)

`errors.Join`、`strings.Builder`、`slices.SortFunc`、`iter.Seq`、`net/http` の middleware。すでに使っているものを構造の目で読み直します。今回のテーマは **法則は最適化の許可証** です。

## 動かす

```bash
go test -v ./part2/
go test -bench=. -benchmem -run='^$' ./part2/
```

CodeSandbox なら **「第2回: Functor則・Monoid則・middleware の結合律」** のタスクを押してください。

## 何を確認すればいいか

### 1. Functor 則 = ループを1本にまとめてよい、という許可証

```bash
go test -v -run 'FunctorLaws' ./part2/
go test -bench=MapTwice -benchmem -run='^$' ./part2/
```

`MapSlice(MapSlice(xs, f), g)` と `MapSlice(xs, Compose(f, g))` が等しいことをテストが保証します。等しいなら速いほうを選べます。

| 要素数 | 2回に分ける | 融合 |
|---:|---:|---:|
| 100 | 570 ns / 2,688 B | 321 ns / 928 B |
| 1,000,000 | 12.4 ms / 31.7 MB | 11.9 ms / **15.7 MB** |

**メモリが常に半分**になるところが本命です（中間スライスが消える）。

### 2. `iter.Seq` は遅延なので融合が自動で起きる

```bash
go test -v -run TestSeqFunctorIsLazy ./part2/
```

5要素のうち1要素しか射が適用されないことを確認できます。同じ Functor 則を満たしながら評価戦略だけが違う、という例です。

### 3. 標準ライブラリのモノイドたち

```bash
go test -v -run 'IsMonoid|IsSameMonoid' ./part2/
```

`errors.Join`・文字列連結・`time.Duration` の最大値が、どれも結合律と単位律を満たすことを確認します。**`errors.Join` の単位元が `nil`** なので、失敗ゼロ件のときに特別扱いが要らなくなります。

### 4. 多段ソートの比較関数もモノイド（実務で一番効く）

```bash
go test -v -run 'TestCompareByIsMonoid|TestSortIsUnaffectedByGrouping' ./part2/
```

「部署順、同じなら年齢順、同じなら名前順」の比較関数はモノイドです。だから **「標準の並び順」を部品として定義し、あとから条件を足せます**。

```go
var defaultOrder = CompareBy(byDept, byAge)
searchOrder := CompareBy(byScore, defaultOrder)   // 中身を知らなくても安全に合成できる
```

### 5. 法則が無いと壊れる —— 平均と浮動小数点

```bash
go test -v -run 'TestMeanBreaksUnderParallelSplit|TestFloatAdditionIsNotAssociative' ./part2/
```

ログを見てください。

```
逐次=5.5 / 3分割=6.166666666666667 （差=0.667）
(a+b)+c = 1 / a+(b+c) = 0
```

`[1..10]` の平均を3分割すると **12% ずれます**。平均は結合的でないからです。合計と個数の組（`SumCount`）に分ければ直ることも、同じテストファイルで確認できます。

### 6. 結合律は並列化の許可証。ただし損益分岐点がある

```bash
go test -bench=BenchmarkFold -benchmem -run='^$' ./part2/
```

| 要素数 | 逐次 | 並列(18分割) |
|---:|---:|---:|
| 1,000 | 622 ns | 5,050 ns（**8倍遅い**） |
| 30,000 | 17,800 ns | 15,100 ns |
| 10,000,000 | 6.3 ms | 1.2 ms（5倍速） |

損益分岐点は **2〜3万要素**。「モノイドだから並列化できる」は正しくても、「並列化すべき」ではありません。

### 7. 同じモノイドでも実装で55倍違う

```bash
go test -bench=StringMonoid -benchmem -run='^$' ./part2/
```

`a + b` の畳み込みと `strings.Join` は同じモノイドの別実装です。答えは同じで、速度が55倍違います。**法則は「置き換えてよい」と言うだけで「同じ速さだ」とは言いません。**

### 8. middleware は括り直せるが、順序は変えられない

```bash
go test -v -run 'TestChainIsAssociative|TestChainIsNotCommutative|TestEmptyChainIsIdentity' ./part2/
```

実行順のログが出ます。

```
実行順: [a b c handler c:after b:after a:after]
```

3通りの括り方すべてで同じ順序になる（結合律）一方、並び順を変えると変わる（非可換）。**「認証系だけ束ねておく」が安全な理由**がこれです。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `functor.go` | `MapSlice` / `MapPtr` / `MapSeq` / `MapErr` |
| `monoid.go` | `Monoid[T]`・標準ライブラリのモノイド・`FoldParallel`・`CompareBy`・平均の壊れ方 |
| `kleisli.go` | `Kleisli[A,B]` / `Then` / `Pure` / `Step` / `ChainStep` |
| `middleware.go` | `Middleware` / `Chain` / `Tap` / `Passthrough` |
| `product.go` | 積（`Pair`/`Fanout`）と余積（`Either`/`Fanin`）、`Curry`/`Uncurry` |
