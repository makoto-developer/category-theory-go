# part3 — 積み上げた先の天井

記事: [Goに高階カインドはない](https://blog.makoto-developer.net/articles/2026-08-06-practical-category-theory-go-3)

Functor から Applicative、Monad へ。式木を fold ひとつで畳む。`iter.Seq` が継続渡しだったこと。そして Go の型システムの天井を測ります。

## 動かす

```bash
go test -v ./part3/
go test -bench=. -benchmem -run='^$' ./part3/
```

[Codespaces](https://codespaces.new/makoto-developer/category-theory-go) でも同じコマンドがそのまま動きます。

## 何を確認すればいいか

### 1. Applicative と Monad で、ユーザーが見る画面が変わる

```bash
go test -v -run TestApplicativeCollectsAllErrors ./part3/
```

同じ「3項目すべてが不正」な入力に対して、ログがこうなります。

```
Applicative: 3件 -> 名前が空です
    メールアドレスに @ が含まれていません
    年齢が負の数です
Monad      : 1件 -> 名前が空です
```

Monad 版だとユーザーは3往復します。**「フォームのエラーは全部まとめて返せ」という UX の常識は、圏論の言葉では「その検証は Applicative で書け」**でした。書けるかどうかの判定基準は「各項目が互いに独立か」だけです。

```bash
go test -v -run TestApplicativeNeverCollectsFewerErrors ./part3/
```

成否の判定は一致し、**失敗の件数だけが違う**ことを property-based test で確かめています。

### 2. F代数 —— 解釈を増やしても再帰は増えない

```bash
go test -v -run 'TestFoldInterpretsSameTreeDifferently|TestSimplifyAlgebra' ./part3/
```

同じ式木に代数を差し替えるだけで、評価・整形・変数収集・深さ・節点数・簡約が全部書けます。**再帰は `Fold` の中の1か所にしかありません**。

構成子を足すと `Algebra` にフィールドが増え、**既存の代数がすべてコンパイルエラーになります**。`switch` の網羅性が検査されない問題が、ここでは構造体のフィールドとして解決しています。

```bash
go test -v -run TestSimplifyPreservesMeaning ./part3/
```

ランダムな式木で「簡約しても評価結果が変わらない」ことを検証します。**最適化パスの正しさをこの形でテストできる**のが実務的な利点です。

### 3. 代数の積 —— 木を1回たどって2つの答え

```bash
go test -v -run TestProductAlgebraMatchesTwoFolds ./part3/
go test -bench=TwoFoldsVsProduct -benchmem -run='^$' ./part3/
```

「評価しながら節点数も数える」が、`EvalAlgebra` にも `CountAlgebra` にも手を入れずにできます（48.7μs → 38.8μs）。

### 4. `iter.Seq` は継続渡しだった

```bash
go test -v -run 'TestContRoundTrip|TestContFunctorLaws|TestSeqStopsOnBreak' ./part3/
```

`FromCont` の実装を見てください。**継続に恒等射を渡すと値が出てきます**。

ただし逆向きは戻りません。`func(k func(int) int) int { return 42 }` のように継続を捨てる関数も `Cont[int, int]` なので、**`Cont[A, A]` と `A` は同型ではありません**。米田の補題が要求しているのは「任意の行き先 $R$ で自然に振る舞う」ことで、`Cont[A, A]` と書いた時点で $R$ が固定され、その条件が落ちています。Go の型ではこの自然性を書けません。詳しくは記事の第3回 3.2 を参照してください。

### 5. `iter.Pull` は120倍遅い（今回いちばん実用的な数字）

```bash
go test -bench=SeqStyles -benchmem -run='^$' ./part3/
```

| 要素数 | スライス | push (`iter.Seq`) | pull (`iter.Pull`) |
|---:|---:|---:|---:|
| 100,000 | 27.4 μs | 127.8 μs | **3,340 μs** |

要素あたりでスライス 0.27 ns、push 1.28 ns、pull **33 ns**。`iter.Pull` は内部でコルーチンを作り、1要素ごとに制御を往復させるためです。**pull でなければ書けない処理（マージ・先読み）にだけ使ってください。**

理論上 push と pull は相互変換できます。それでも100倍以上違う。**同型であることと、同じコストであることは別**です。

### 6. 高階カインドが無いことの値段

```bash
go test -v -run 'TestAllHKTWorkaroundsAgree|TestDictionaryAbstractsOverFunctor' ./part3/
go test -bench=HKTWorkarounds -benchmem -run='^$' ./part3/
```

`Functor[F[_]]` は書けません。回避策3つのコストがこれです。

| 方式 | ns/op | allocs/op |
|---|---:|---:|
| 手書きループ | 12,600 | 901 |
| ジェネリクス | 12,500 | 901 |
| **型消去（`any`）** | **22,800** | **1,902** |
| 辞書渡し | 11,800 | 901 |

**ジェネリクスと辞書渡しはゼロコスト。型消去だけが1.8倍遅く、アロケーション2倍。** 高階カインドの代わりに `any` を選ぶ理由は、性能の面ではありません。

### 7. Functional Option パターンの正体

```bash
go test -v -run 'TestOptionsAreMonoid|TestCurryOptionMatchesDirectSetter|TestOptionsAreNotCommutative' ./part3/
```

`WithHost(h)` は `func(Config, string) Config` をカリー化したものです。カリー化の結果 `ServerOption` は自己射になり、その全体がモノイドになります。だから `ApplyOptions` でまとめられます。

middleware・Functional Option・比較関数が、すべて**同じ「自己射のモノイド」**だと分かるのがこの節の狙いです。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `validation.go` / `form.go` | `Validated[T]`・`Combine2/3`（Applicative）・`AndThen`（Monad） |
| `fold.go` | 式木・`Algebra[T]`・`Fold`・各種代数・`ProductAlgebra` |
| `cps.go` | `Cont[R,A]`・`MapCont`・push/pull の合計 |
| `hkt.go` | 型消去・ジェネリクス・辞書渡しの3方式 |
| `adjunction.go` | `ServerOption`・`ApplyOptions`・`CurryOption` |
