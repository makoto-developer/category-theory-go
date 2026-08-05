# examples — サンドボックスに流し込むスニペット

記事に埋め込んでいる [CodeSandbox の Devbox](https://codesandbox.io/p/devbox/4msfly) は、ここのファイルを**正本**として動いています。Devbox 側でコードを持たず、起動時に GitHub から取ってくる作りです。

```jsonc
// Devbox の .codesandbox/tasks.json
{
  "setupTasks": [
    {
      "name": "最新のサンプルを取得",
      "command": "curl -sfL https://raw.githubusercontent.com/makoto-developer/category-theory-go/main/examples/laws/main.go -o main.go"
    }
  ],
  "tasks": {
    "go run": { "name": "Run", "command": "go run main.go", "runAtStart": true }
  }
}
```

こうしておくと、**このリポジトリに push するだけでサンドボックスの中身が最新になります**。ブラウザの中で書いたコードがどこにも残らない、という状態を避けられるのが狙いです。`gofmt` も `go vet` も CI も、ほかのコードと同じものが掛かります。

## 制約

| 項目 | 値 |
|---|---|
| Devbox の Go | `1.25`（`.devcontainer/devcontainer.json` で `mcr.microsoft.com/devcontainers/go:1.25-bookworm` を指定） |
| `go.mod` | 無し（`go run main.go` で直接動かす） |
| 外部依存 | 不可 |

**`part1/`〜`part4/` のコードはそのままは置けません。** `pgregory.net/rapid` に依存していて `go.mod` が要るからです。`examples/` に置くのは「標準ライブラリのみ・1ファイル」に収まるものだけにしてください。全部を触りたい人には [Codespaces](https://codespaces.new/makoto-developer/category-theory-go) を案内します。

> 当初の Devbox は Go 1.21 で、`iter.Seq` も `b.Loop()` も使えませんでした。加えて **gopls が入らず、埋め込みで `.go` ファイルを開くと「Failed to setup Go LSP」のエラーが出ていました**。イメージを 1.25 に上げて両方とも解消しています。

## 一覧

| ディレクトリ | 内容 | 埋め込み先 |
|---|---|---|
| [`laws/`](laws/) | 結合律・単位律を1万通りのランダム入力で検証し、法則が壊れる例（メモ化・並列化した平均・浮動小数点）を出す | [第1回](https://blog.makoto-developer.net/articles/2026-08-04-practical-category-theory-go-1) |
| [`functor-monoid/`](functor-monoid/) | Functor則の検証とループ融合の実測、並列集計が壊れる/直る境目、middleware に結合律はあるが可換性はないこと | [第2回](https://blog.makoto-developer.net/articles/2026-08-05-practical-category-theory-go-2) |
| [`applicative-fold/`](applicative-fold/) | Applicative と Monad でエラー件数が変わること、同じ式木に代数を差し替えるだけで意味が変わること、`iter.Seq` が継続渡しだったこと | [第3回](https://blog.makoto-developer.net/articles/2026-08-06-practical-category-theory-go-3) |
| [`workflow/`](workflow/) | 8行の Plan 定義から実行・dry-run・Mermaid 図を導く。デコレータを巻く順序が結果を決めること | [第4回](https://blog.makoto-developer.net/articles/2026-08-07-practical-category-theory-go-4) |

## 増やすとき

1. `examples/<name>/main.go` を追加する（上の制約を守る）
2. `go vet ./examples/...` と `go run ./examples/<name>` を通す
3. Devbox を新しく作り、`.codesandbox/tasks.json` の `setupTasks` で `<name>/main.go` を `curl` する
4. 記事に iframe を貼り、この表に1行足す
