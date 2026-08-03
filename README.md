# category-theory-go

連載「**Goで書く実践圏論**」の検証コード。記事に載せたコードと実測値は、すべてこのリポジトリで実際に動かしたものである。

## 再現手順

```bash
go vet ./...
go test ./...                      # 圏の法則・Functor則・Monoid則の検証（property-based test を含む）
go test -bench=. -benchmem ./...   # 実測（結果は BENCHMARKS.md）
```

Go 1.25 以降が必要（`iter.Seq` と range-over-func を使う）。外部依存は property-based test 用の [rapid](https://pgregory.net/rapid) のみ。

## 構成

| ディレクトリ | 記事 | 内容 |
|---|---|---|
| `part1/` | 第1回 | 合成 `Compose`、圏の公理（結合律・単位律）の検証、法則が壊れる例、抽象化コストの実測 |
| `part2/` | 第2回 | Functor / Monoid / Kleisli 合成 / middleware（自然変換）/ 積と余積 |
| `part3/` | 第3回 | Applicative バリデーション、F代数と fold、`iter.Seq` と CPS、高階カインドの回避策3種 |
| `part4/` | 第4回 | ワークフローエンジン（素直な実装 vs 圏論的実装）、interpreter による実行・dry-run・図生成 |

## 記事

- 第1回　合成できることが、すべての出発点だった —— 型と関数がつくる圏
- 第2回　middleware も errors.Join も圏論だった —— Go標準ライブラリに潜む構造
- 第3回　Goに高階カインドはない —— 型システムの天井を圏論で測る
- 第4回　圏論でワークフローエンジンを組み直したら —— 得したところ、壊れたところ

## ライセンス

MIT
