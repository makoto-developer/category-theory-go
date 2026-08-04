# part4 — ワークフローエンジンで実際に比べる

記事: [圏論でワークフローエンジンを組み直したら](https://blog.makoto-developer.net/articles/2026-08-07-practical-category-theory-go-4)

注文処理（在庫確保 → 決済 → 配送手配 → 通知）を、**素直な手続き型**と**射の合成**の2通りで実装しました。行数・アロケーション・レイテンシを比べ、**壊れたところ**も記録しています。

## 動かす

```bash
go test -v ./part4/
go test -bench=. -benchmem -run='^$' ./part4/
```

CodeSandbox なら **「第4回: ワークフローエンジン（素直な実装 vs 射の合成）」** のタスクです。

## 何を確認すればいいか

### 1. まず、2つの実装が同じ結果を返す

```bash
go test -v -run TestNaiveAndComposedAgree ./part4/
```

`naive.go` の71行と、`workflow.go` + `program.go` の合成版が同じ `State` を返します。ここが出発点です。

### 2. 同じ定義から4つのものが出てくる（この記事の山場）

```bash
go test -v -run 'TestDryRunTouchesNothing|TestMermaidIsGeneratedFromTheSamePlan|TestExplainListsRetryPolicy' ./part4/
```

`OrderPlan` という**8行の定義**から、実行・dry-run・Mermaid図・設定一覧の4つが導けます。ログに図がそのまま出ます。

```
flowchart TD
    start([注文処理])
    n0[在庫確保<br/>最大3回<br/>2s]
    start --> n0
    n1[決済<br/>5s]
    ...
```

**この図はコードから生成されています。** ステップを足せば図に出る。リトライ回数を変えれば図の数字が変わる。設計ドキュメントが実装とずれる問題が構造的に起きません。

dry-run のテストは「外部サービスが1回も呼ばれない」ことを assert しています。素直な実装で dry-run を足すと、本番コードに `if dryRun` が4か所混ざります。

### 3. 定義を1行変えるだけで方針が変わる

```bash
go test -v -run TestPolicyChangeDoesNotTouchBusinessLogic ./part4/
```

`plan.Nodes[0].Retries = 5` と書き換えるだけで、業務ロジックに一切触れずリトライ方針が変わります。運用中の設定変更で効くところです。

### 4. デコレータは単体でテストできる

```bash
go test -v -run 'TestWithRetryOnlyRetriesTemporaryErrors|TestTimeoutWrapsAllRetries' ./part4/
```

在庫サービスも決済サービスも出てきません。**リトライという関心事だけを、3行の偽ステップでテスト**しています。

`TestTimeoutWrapsAllRetries` のログに注目してください。

```
50ms のタイムアウトで 3 回試行された
```

1回20msの処理を最大10回リトライする設定に50msのタイムアウトを掛けると、3回で打ち切られます。**デコレータを巻く順序が結果を決めている**ことの証明です（順序が逆なら200ms走り続ける）。

### 5. 値段を測る

```bash
go test -bench=BenchmarkWorkflow -benchmem -run='^$' ./part4/
```

| 実装 | 機能 | ns/op | allocs |
|---|---|---:|---:|
| 素直な手続き型 | リトライ + 計測 | 224 | 0 |
| 射の合成 | 同上 + ステップ名 | 345 | 0 |
| 射の合成 | 上記 + ステップ単位タイムアウト | 868 | 12 |
| 射の合成（デコレータ無し） | なし | 66.5 | 0 |

最初は「4倍遅い」と見えました。でも**条件が揃っていませんでした**（合成版だけタイムアウトを持っている）。揃えると 224 → 345 ns、**1ステップあたり30ns**。しかも最後の行を見てください。**素の合成（66.5 ns）は素直な実装より速い**。素直な実装が各ステップで `Recorder`（mutex + map）を叩いているからです。

### 6. 壊れたところを自分の目で見る

```bash
go test -v -run TestPanicTraceShape ./part4/
```

同じ panic を両実装で起こして、スタックトレースを並べて表示します。書く前は「`func1` が並んで読めない」と思っていましたが、実際は違いました。

- デコレータ名（`compileNode.WithRetry.func3` など）は**ちゃんと読める**
- でも **`Node.Name`（「在庫確保」）はどこにも出ない**。データであって関数名ではないから
- フレームが 2 → 6 に増える
- `State` が16進で展開されて1行が200文字を超える

**型で進捗を保証することを諦めた**話（記事 6.1）も、`domain.go` の `State` を見ると分かります。全ステップを `State → State` にしたので、`Sequence(Reserve, Arrange, Charge)` と順序を間違えてもコンパイルが通ります。型が変わる射はリストに入らず、リストに入らなければ `Plan` が作れない——というトレードオフです。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `domain.go` | `Order` / `State` / 外部サービスのインターフェース |
| `naive.go` | 素直な実装（71行。リトライ・計測・キャンセルが各ステップに散らばる） |
| `workflow.go` | `Endo` / `Then` / `Sequence` と、`WithRetry`・`WithTiming`・`WithTimeout`・`WithLabel` |
| `program.go` | `Plan` / `Node` と4つの解釈（`Compile`・`DryRun`・`Mermaid`・`Explain`） |
| `trace_test.go` | パニック時のスタックトレース比較（記事の 6.3 の出力はここで取得） |
