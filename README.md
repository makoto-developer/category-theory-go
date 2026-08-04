# category-theory-go

連載「**Goで書く実践圏論**」の検証コードです。記事に載せたコードと実測値は、すべてこのリポジトリで実際に動かしたものです。

## ブラウザですぐ試す

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/makoto-developer/category-theory-go)

→ https://codespaces.new/makoto-developer/category-theory-go

`.devcontainer/devcontainer.json` で `golang:1.25` を指定してあるので、開くだけで環境が整い、`go test ./...` が自動で1回走ります。あとはターミナルで下の「まず何を見るか」のコマンドを打つだけです。ローカルに Go を入れなくても、圏の法則が本当に成り立つこと・壊れることを確かめられます。

> CodeSandbox でも開けるようにしていましたが、**Repositories 機能は 2026年7月1日でサポート終了**になりました（CodeSandbox 自身が Codespaces への移行を案内しています）。そのため Codespaces に切り替えています。

**ベンチマークの数値だけは注意してください。** クラウドの共有マシン上なので、記事に載せた数字（Apple M5 Pro / Go 1.25.1）とは一致しません。見るべきは絶対値ではなく比のほうです。正確な実測は手元でどうぞ。

## 手元で動かす

```bash
go vet ./...
go test ./...                            # 圏の法則・Functor則・Monoid則の検証（property-based test を含む）
go test -bench=. -benchmem -run='^$' ./...   # 実測（結果は BENCHMARKS.md）
```

Go 1.25 以降が必要です（`iter.Seq` と range-over-func を使うため）。外部依存は property-based test 用の [rapid](https://pgregory.net/rapid) だけです。

## 構成

各ディレクトリに README があります。**何を確認すればいいか**はそちらに書いてあります。

| ディレクトリ | 記事 | 内容 |
|---|---|---|
| [`part1/`](part1/) | 第1回 | 合成 `Compose`、圏の公理（結合律・単位律）の検証、法則が壊れる例、抽象化コストの実測 |
| [`part2/`](part2/) | 第2回 | Functor / Monoid / Kleisli 合成 / middleware（自然変換）/ 積と余積 |
| [`part3/`](part3/) | 第3回 | Applicative バリデーション、F代数と fold、`iter.Seq` と CPS、高階カインドの回避策3種 |
| [`part4/`](part4/) | 第4回 | ワークフローエンジン（素直な実装 vs 圏論的実装）、interpreter による実行・dry-run・図生成 |

## まず何を見るか

時間がないときの3つです。

```bash
# 1. 圏の法則が本当に成り立つ（そして具体的な射でしか証明していない、という限界も）
go test -v -run 'TestComposeIsAssociative|TestIdentityIsUnit' ./part1/

# 2. 法則が無いと壊れる。平均を並列化すると 5.5 が 6.17 になる
go test -v -run 'TestMeanBreaksUnderParallelSplit|TestFloatAdditionIsNotAssociative' ./part2/

# 3. 同じ定義から Mermaid 図が出てくる（第4回の山場）
go test -v -run TestMermaidIsGeneratedFromTheSamePlan ./part4/
```

## 記事

1. [合成できることが、すべての出発点だった —— 型と関数がつくる圏](https://blog.makoto-developer.net/articles/2026-08-04-practical-category-theory-go-1)
2. [middleware も errors.Join も圏論だった —— Go標準ライブラリに潜む構造](https://blog.makoto-developer.net/articles/2026-08-05-practical-category-theory-go-2)
3. [Goに高階カインドはない —— 型システムの天井を圏論で測る](https://blog.makoto-developer.net/articles/2026-08-06-practical-category-theory-go-3)
4. [圏論でワークフローエンジンを組み直したら —— 得したところ、壊れたところ](https://blog.makoto-developer.net/articles/2026-08-07-practical-category-theory-go-4)

実測した数字は [BENCHMARKS.md](BENCHMARKS.md) に測定条件つきでまとめてあります。

## ライセンス

MIT
