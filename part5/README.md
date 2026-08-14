# part5 — Cayley 表現と、コストモデルの天井

Cayley 表現（モノイド $M$ を $\mathrm{End}(M)$ へ $m \mapsto (m \cdot -)$ で埋め込むもの）が、
Go でいつ効いていつ効かないかを測ります。

第3回が測ったのは Go の**型システム**の天井でした。ここで測るのは**コストモデル**の天井です。

## 動かす

```bash
go test -v ./part5/
go test -bench=. -benchmem -run='^$' ./part5/
```

[Codespaces](https://codespaces.new/makoto-developer/category-theory-go) でも同じコマンドがそのまま動きます。

## 何を確認すればいいか

### 1. 同じ変換が、同じコードで、正反対の結果になる

`FoldNaive` と `FoldCayley` は `Monoid[M]` を受け取るジェネリック関数です。**3つのモノイドに対して
走るのは同じコードで、違うのは `Append` のコストがどちら側に付いているかだけ**です。

```bash
go test -bench='ConsMonoid|SnocMonoid' -benchmem -run='^$' ./part5/
```

| モノイド | `Append` のコスト | 素朴（左結合） | Cayley 表現 |
|---|---|---:|---:|
| cons リスト | $O(\lvert 左 \rvert)$ | 1,172,245,874 ns | **684,970 ns** |
| snoc リスト | $O(\lvert 右 \rvert)$ | **189,493 ns** | 903,319,615 ns |

**1,711倍速くなり、4,767倍遅くなります。** 変換は同じ、法則も同じ。

`allocs/op` を見ると何が起きたか分かります。cons の素朴版が **49,995,021**、snoc の Cayley 版が
**50,015,006**。**同じ $O(n^2)$ の仕事が反対側へ移っただけ**です。

### 2. 圏論の側からはこの2つを区別できない

```bash
go test -v -run 'MonoidLaws|Homomorphism' ./part5/
```

`TestConsMonoidLaws` と `TestSnocMonoidLaws` は**どちらも通ります**。結合律も単位律も同じように
成り立つ。`TestCayleyIsMonoidHomomorphism` が示すとおり埋め込みは準同型で、
`TestFoldsAgree` のとおり**畳み方を変えても答えは変わりません**。

圏論が保証しているのは「意味が変わらないこと」だけで、**速さについては何も言っていない**。
それがこのパートの主題です。

### 3. 文字列では Cayley 表現は効かない

```bash
go test -bench=StringMonoid -benchmem -run='^$' ./part5/
```

| 方式 | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| 素朴（`+` の左結合） | 31,791,395 | **428,519,972** | 10,007 |
| Cayley 表現 | 43,324,062 | **429,479,382** | 30,002 |
| `strings.Builder` | 81,266 | 383,738 | 23 |
| `strings.Builder` + `Grow` | **50,124** | 81,920 | 1 |
| `strings.Join` | 77,807 | 81,920 | 1 |

Go の文字列の `+` は $O(\lvert 左 \rvert + \lvert 右 \rvert)$ です。左右どちらに結合を寄せても
総コピー量が変わらないので、**Cayley 表現はコピーを1バイトも減らしません**（428.5MB → 429.5MB）。
クロージャ2個/要素ぶんアロケーションが3倍になり、**1.4倍遅くなるだけ**です。

Haskell の差分リストや `ShowS` が速いのは、リストの `++` が「左辺のぶんだけ払う」形だからです。
**同じ最適化を Go の文字列へ持ち込むことはできません。**

### 4. `strings.Builder` の速さは圏論由来ではない

素朴 → `Builder` が **391倍**、`Builder` → `Grow` 付きが **1.6倍**。

支配しているのは事前確保ではなく、$O(n^2) \to O(n)$ にする**可変バッファの償却成長**です。
これは Cayley 表現とは別の仕組みで、圏論の定理からは出てきません。

### 5. そして Cayley 表現は、どちらの場合も「Go で普通に書く形」に負ける

- 文字列: `Builder`+`Grow`（50,124 ns）に **864倍**負ける
- cons リスト: `PrependReverse`（559,464 ns）に **1.2倍**負ける

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `cayley.go` | `Monoid[M]`・`Cayley`・`FoldNaive`・`FoldCayley`・文字列モノイド・`strings.Builder` 版 |
| `list.go` | cons リスト（`Concat` は $O(\lvert 左 \rvert)$）と snoc リスト（`ConcatR` は $O(\lvert 右 \rvert)$） |

## Go のバージョン差

Go 1.23.12 / 1.24.13 / 1.25.13 / 1.26.5 / 1.26.6 で確認しました。**30本すべてのベンチで
`allocs/op` が一致**し、`ns/op` は 1.23 比 0.48〜2.09 倍にばらけて傾向がありません。
**この結果はランタイムの最適化の話ではなく、データ構造のコストモデルの話**なので、
バージョンには依存しません。

測定には [benchsweep](https://github.com/makoto-developer/benchsweep) を使いました
（この module は `go 1.25.1` を宣言しているので 1.23 ではビルドできません。
バージョン横断の測定は同じ実装を `go 1.23` の単独 module に置いて行っています）。
