# part1 — 型と関数がつくる圏

記事: [合成できることが、すべての出発点だった](https://blog.makoto-developer.net/articles/2026-08-04-practical-category-theory-go-1)

Go の型を対象、`func(A) B` を射とみなすと圏になります。ここではその公理（結合律・単位律）を property-based test で確かめ、圏が壊れる瞬間を実際に見て、合成の値段を測ります。

## 動かす

```bash
go test -v ./part1/                        # 法則の検証
go test -bench=. -benchmem -run='^$' ./part1/   # 合成のコスト
```

[Codespaces](https://codespaces.new/makoto-developer/category-theory-go) で開けば、そのままターミナルに貼るだけで動きます。

## 何を確認すればいいか

### 1. 圏の公理が本当に成り立つ

```bash
go test -v -run 'TestComposeIsAssociative|TestIdentityIsUnit' ./part1/
```

`rapid` がランダムな入力を大量に投げても、括り方を変えた2つの合成が一致し続けることを見てください。**結合律はリファクタリングの許可証**です。「この2ステップを関数に切り出す」が安全なのは、これが成り立っているからです。

### 2. 空の合成が恒等射になる

```bash
go test -v -run TestEmptyPipeIsIdentity ./part1/
```

`Pipe()`（引数ゼロ）が素通しになります。「ミドルウェアを1個も設定しなかったら素通し」が自然に感じられる理由がこれです。

### 3. 圏が壊れる瞬間

```bash
go test -v -run 'TestMemoizeChangesResultForImpureMorphism|TestRetryRunsEffectTwiceForImpureMorphism' ./part1/
```

ログに注目してください。

```
メモ化: 11, 11 / メモ化なし: 12
```

同じ `impure(10)` が、キャッシュを挟むかどうかで違う値になります。純粋でない関数は Go の関数ではあっても**圏の射ではない**、というのがここで見えます。リトライのテストは、副作用が二重に走る様子（課金が2回走る事故の原型）です。

### 4. 合成の値段

```bash
go test -bench='Composed' -benchmem -run='^$' ./part1/
```

見どころは `BenchmarkComposedStored` と `BenchmarkComposedInline` の差です。

| | 手元での測定値 |
|---|---|
| 手書き | 1.57 ns |
| `Compose` を変数に保存して呼ぶ | 1.73 ns |
| 呼ぶたびに `Compose` し直す | **12.7 ns / 32 B** |

**合成は構築時に1回**。リクエストごとに組み直すと7倍以上のペナルティが付きます（手書きと比べれば8倍）。

なお上2行の差（0.16 ns）は測定のばらつきに埋もれます。**再現するのは3行目の桁違いのほうだけ**なので、そちらを見てください。詳しくは [BENCHMARKS.md](../BENCHMARKS.md) の但し書きを参照。

### 5. `func(A) (B, error)` は合成できない

```bash
go test -v -run 'TestComposeEIsAssociative|TestComposeEShortCircuits' ./part1/
```

`Compose` では型が繋がらないので専用の合成器 `ComposeE` を用意しています。これも圏（Kleisli 圏）になることと、エラー時に後続が呼ばれないことを確認できます。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `compose.go` | `Compose` / `Identity` / `Pipe`。これで圏の定義は全部 |
| `purity.go` | `Memoize` / `Counter` / `Retry`。圏を壊すための道具 |
| `compose_error.go` | `ComposeE` / `IdentityE`。Kleisli 合成の先取り |
| `compose_test.go` | 結合律・単位律・畳み込みの向き |
| `purity_test.go` | メモ化とリトライで壊れることの実証 |
| `compose_error_test.go` | error を返す射の結合律。`sameError` を使う理由は記事の 5.4 に |
| `bench_test.go` | 合成のコスト。段数・保存の有無・内側の重さで比較 |

## 数値について

ベンチマークの数値は実行環境で変わります。記事の数字は Apple M5 Pro / Go 1.25.1 のもので、クラウドの共有マシンでは**絶対値は一致しません**。見るべきは絶対値ではなく、**「変数に保存」と「毎回組み直す」の比**のほうです。
