# 実測結果

記事に載せている数字の正本。すべてこのリポジトリで実際に走らせたもの。

**測定環境**

```
goos: darwin
goarch: arm64
pkg: github.com/makoto-developer/category-theory-go
cpu: Apple M5 Pro (18 threads)
go version go1.25.1 darwin/arm64
```

再現コマンド:

```bash
go test -bench=. -benchmem -count=6 -run='^$' ./part1/
```

数値は `-count=6` の代表値（外れ値を除いた中央付近）。

---

## part1 — 合成のコスト

### 合成の組み方による差（int → int の軽い射）

| 書き方 | ns/op | B/op | allocs/op | 手書きとの差 |
|---|---:|---:|---:|---:|
| 手書き `increment(double(x))` | 1.57 | 0 | 0 | — |
| `Compose` 2段・変数に保存して呼ぶ | 1.73 | 0 | 0 | +0.16 ns |
| `Compose` 2段・呼ぶたびにその場で合成 | 12.7 | 32 | 1 | **+11.1 ns** |
| 手書き3段 `negate(increment(double(x)))` | 1.54 | 0 | 0 | — |
| `Compose` 3段・変数に保存（静的な入れ子） | 2.50 | 0 | 0 | +0.96 ns |
| `Pipe` 3段（スライス経由で組む） | 3.90 | 0 | 0 | +2.36 ns |
| 合成そのものの構築コスト（`Compose` 2回） | 22.0 | 64 | 2 | — |

読み取れること:

- **合成を変数に保存して使い回す限り、2段の合成コストは 0.16 ns。実質ゼロ**。コンパイラが関数値の呼び出しを潰している
- **呼ぶたびに合成し直すと 10倍近く遅くなる**（1.73 → 12.7 ns）。クロージャがヒープに逃げ、32 B のアロケーションが毎回発生する。合成はハンドラ構築時に済ませ、リクエストごとに組み直さないこと
- 同じ3段でも、`Compose` の静的な入れ子（2.50 ns）よりスライス経由の `Pipe`（3.90 ns）が遅い。スライスの中身はコンパイル時に分からないため、各段が本物の間接呼び出しになる

### 合成の段数を増やしたときの傾き

| 段数 | `Pipe` で合成 (ns/op) | ループで直接呼ぶ (ns/op) |
|---:|---:|---:|
| 1 | 2.02 | 1.62 |
| 2 | 2.95 | 1.58 |
| 4 | 4.82 | 2.72 |
| 8 | 8.78 | 4.85 |
| 16 | 18.4 | 8.25 |

- 合成 1段あたり **約 1.09 ns**（(18.4 − 2.02) / 15）
- ループで直接呼ぶ場合は 1段あたり **約 0.44 ns**
- 差の **約 0.65 ns/段** が、クロージャを1枚挟むことの代金

### 内側の処理が重い場合

| 書き方 | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| 手書き `isEven(length(strconv.Itoa(x)))` | 11.9 | 8 | 1 |
| 同じ処理を関数値経由で呼ぶ | 12.2 | 8 | 1 |
| `Compose` 3段・変数に保存 | 13.4 | 8 | 1 |

`strconv.Itoa` が 1 回アロケートする（8 B）ため、そちらが支配的になる。合成のコストは全体の約 10%（+1.5 ns）に薄まる。

**判断基準**: 射1本の中身が 10 ns 以上（文字列変換1回、mutex 1回、map アクセス数回）なら合成のコストは誤差。1 ns 級の算術だけを繋ぐホットループでは、合成は 1.5〜2.5 倍の差になる。

### インライン化の判定

```bash
go test -gcflags='-m' -run='^$' -bench='^$' ./part1/ 2>&1 | grep compose.go
```

```
part1/compose.go:7:6: can inline Compose[go.shape.int,go.shape.string,go.shape.int]
part1/compose.go:8:9: can inline Compose[go.shape.int,go.shape.string,go.shape.int].func1
part1/compose.go:12:6: can inline Identity[go.shape.int]
part1/compose.go:16:6: can inline Pipe[go.shape.int]
```

`Compose` も、それが返すクロージャも「インライン化可能」と判定されている。ジェネリック関数は `go.shape.*`（同じメモリ表現を持つ型をまとめた形）単位でインスタンス化されるため、`Compose[int,string,int]` と `Compose[int,int,bool]` は別々に現れる。
