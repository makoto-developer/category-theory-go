# part7 — レンズ則は Go では足りない。そして代金は表現で決まる

レンズ（Lens）は「大きい構造 S の中の小さい部分 A」への焦点を、`Get` と `Set` の組で表したものです。合成できるので、入れ子の奥まで届きます。

レンズには**3つの法則**があります。

- **get-set**: 取り出してそのまま戻せば、元と同じ（`Set(s, Get(s)) == s`）
- **set-get**: 入れたものが取り出せる（`Get(Set(s, a)) == a`）
- **set-set**: 二度入れたら後勝ち（`Set(Set(s, a), b) == Set(s, b)`）

**この3つを満たしても、Go では元の値が書き換わります。**

## 動かす

```bash
go test -v ./part7/
go test -bench=. -benchmem -run='^$' ./part7/
```

## 何を確認すればいいか

### 1. 3則を全部通るレンズが、元の値を書き換える

```bash
go test -v -run 'Aliasing|Independen|Leaks' ./part7/
```

Go の struct は代入でコピーされます。**ただし浅く**です。`[]string` はポインタ・長さ・容量の3語で、コピーされるのはその3語だけ。指している配列は共有されたままです。

だから「素直に書いた」レンズはこうなります。

```go
Get: func(p Profile) []string { return p.Tags },
Set: func(p Profile, ts []string) Profile { p.Tags = ts; return p },
```

`TestAliasingLensStillObeysAllThreeLaws` が、**これが3則を全部通る**ことを property-based test で示します（100ケース）。偶然ではありません。setter が対象フィールドをそのまま置き換えるので、3つの等式は代数的に成立します。

それでも壊れています。`TestAliasingLensBreaksAllFourIndependenceProperties` が、**4つの性質を4つとも破る**ことを示します。

| 性質 | 内容 |
|---|---|
| ① | `Get` の戻り値を書き換えても元は変わらない |
| ② | `Set` に渡した値をあとから書き換えても格納先は変わらない |
| ③ | `Set` の結果を書き換えても元は変わらない |
| ④ | `Modify` に渡した関数が入力を壊しても元は変わらない |

**この4つはレンズ則3つのどれからも導けません。** 3則が語るのは $s$、$\mathrm{set}(s,a)$、$\mathrm{get}(s)$ という**値の間の等式**だけで、「書き換え」という操作が法則の語彙にないからです。

> 3則を「任意の後続の書き換え文脈でも区別できない」という観測的等価性で述べ直せば、共有は検出できます。ただしそれは、まさにこの非干渉を等価性の定義に組み込んだ**別の仕様**であって、古典的な3つの等式ではありません。

### 2. 非干渉を守ると、Get/Set 表現では合成の深さに比例して複製が増える

```bash
go test -bench=LensDepth -benchmem -run='^$' ./part7/
```

非干渉を満たすには複製するしかありません。そして `Compose` の `Set` はこの形です。

```go
Set: func(s S, b B) S { return outer.Set(s, inner.Set(outer.Get(s), b)) }
```

`inner.Set` が複製し、それを受け取った `outer.Set` が**もう一度複製します**。外側は、渡された値が既に自分専用かを知りません。

| 深さ | 通る Set の数 | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|---:|
| 0 | 1 | 8,720 | 65,536 | 1 |
| 1 | 2 | 17,327 | 131,072 | 2 |
| 2 | 3 | 30,149 | 196,608 | 3 |
| 3 | 4 | 47,424 | 262,144 | 4 |

**1段につき 65,536 バイト・1 alloc ちょうど。** 4,096要素の `[]string` の複製1回ぶんです。

### 3. ただし、それは表現の問題だった

```bash
go test -bench='ModifyDepth|Depth3' -benchmem -run='^$' ./part7/
```

**この重複は避けられます。** レンズを `Get`/`Set` ではなく `Get`/`Mod`（焦点に関数を適用する）で表すと、中間の段が値を組み直さなくなります。

```go
type LensM[S, A any] struct {
	Get func(S) A
	Mod func(S, func(A) A) S
}
```

中間の段は関数を内側へ渡すだけなので、複製する理由がありません。複製するのは葉だけです。

| 深さ | Get/Set 表現 B/op | Mod 表現 B/op |
|---:|---:|---:|
| 0 | 65,536 | 131,120 |
| 1 | 131,072 | 131,168 |
| 2 | 196,608 | 131,216 |
| 3 | 262,144 | **131,264** |

Get/Set 表現は1段につき 65,536 バイト増え、**Mod 表現は 48 バイトしか増えません**（クロージャぶん）。深さ0では Mod 表現のほうが高い（葉が読み書き両方で複製するため）ものの、**深さ1で並び、それ以降は開きます**。

`TestSetAndModifyLensesAgree` が両者の答えが同じであることを、`TestModifyLensIsIndependent` が Mod 表現も非干渉の4性質を満たすことを確認しています。

**圏論的には同じレンズです。変わったのは Go での表し方だけで、代金はそこで決まっています。**

### 4. 参照型を通らなくても、合成そのものに代金がある

```bash
go test -bench=UpdateEmail -benchmem -run='^$' ./part7/
```

`Account.Profile.Contact.Email` を差し替えるだけの経路（焦点までの更新対象にスライスも map もない。ただし経路上の `Profile` は `Tags` と `Meta` を浅く共有します）。

| 方式 | 中央値 ns/op | allocs/op |
|---|---:|---:|
| レンズ3枚を合成 | 47.24 | 0 |
| 手書き | 3.838 | 0 |

**12.3倍。** どちらもアロケーションはゼロなので、差は合成した関数値の呼び出しと中間値の受け渡しです。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `lens.go` | `Lens[S,A]`・`Compose`・`Modify`・`Identity` |
| `modify_lens.go` | `LensM[S,A]`・`ComposeM`・`SetM`（Mod 表現） |
| `model.go` | 測定対象の入れ子と、非干渉を守る版／共有したままの版 |
| `lens_test.go` | レンズ則3つ、合成の結合律と単位律 |
| `aliasing_test.go` | 3則を通るのに非干渉の4性質を破ることの検証 |
| `depth_test.go` | 合成の深さと複製量（Get/Set 表現） |
| `modify_test.go` | Mod 表現の非干渉・同値性・深さごとの複製量 |
| `bench_test.go` | 全方式の比較 |
