# part10 — スキーマは圏の表示で、データは関手

**対象がテーブル、射が外部キー、そして射の間に等式を置ける。** これがスキーマを圏として見るということです（Spivak の functorial data migration）。

> **正確には、`Schema` は圏そのものではなく圏の表示（presentation）です。** 持っているのは生成射だけで、合成も恒等射もありません。圏になるのは、この有向グラフ上の自由圏を `Equations` が生成する最小の合同関係で割ったものです。

```go:schema.go
type Schema struct {
	Objects []string
	// Arrows[name] = (from, to)
	Arrows map[string]Arrow
	// Equations は「2つのパスが等しい」という制約。恒等射は空パスで表す。
	Equations []Equation
}
```

インスタンス（データ）は、この商圏から集合の圏への**関手**です。各対象に行の集合を、各射に行どうしの写像を割り当てます。

> `Instance` 値そのものも関手ではありません。**関手を定めるには `Validate()` と `CheckTotality()` と `CheckEquations()` が全部通る必要があります。** 参照整合性に対応するのは2番目だけです。

```go:instance.go
type Instance struct {
	Schema Schema
	// Rows[object] = その対象の行 ID
	Rows map[string][]int
	// Maps[arrow][fromID] = toID
	Maps map[string]map[int]int
}
```

## 動かす

```bash
go test -v ./part10/
go test -bench=. -benchmem -run='^$' ./part10/
```

## 何を確認すればいいか

### 1. パス等式は、RDB の外部キーでは書けない制約

```bash
go test -v -run 'Schema|Instance|ForeignKeys' ./part10/
```

「**上司は同じ部署にいる**」という制約を考えます。圏の言葉ではこう書けます。

$$\mathrm{manager} \,;\, \mathrm{worksIn} = \mathrm{worksIn}$$

社員から上司を辿ってその部署を見るのと、社員から直接部署を見るのが等しい。**射の合成が2通りあって、それが一致する**という等式です。

```go:schema_test.go
		Equations: []Equation{
			// 上司は同じ部署にいる
			{From: "Employee", To: "Department", Left: []string{"manager", "worksIn"}, Right: []string{"worksIn"}},
			// 部署長はその部署に所属する
			{From: "Department", To: "Department", Left: []string{"head", "worksIn"}, Right: nil},
		},
```

**外部キー制約ではこれを書けません。** `manager` も `worksIn` もそれぞれ正しい行を指していればよく、両者の関係は表現できないからです。

`TestForeignKeysCanBeValidWhileEquationsBreak` がそれを示します。社員1の上司を別の部署の長に付け替えると——**外部キーとしては全部正しいまま、パス等式だけが破れます**。

### 2. パス等式の検査は、外部キーの検査より安い

```bash
go test -bench=CheckEquations -benchmem -run='^$' ./part10/
```

| n | 検査 | ns/op | B/op | allocs/op |
|---:|---|---:|---:|---:|
| 100,000 | 全域性（毎回索引を作る） | 14,657,049 | 4,729,707 | 517 |
| 100,000 | 全域性（索引を使い回す） | 9,666,065 | **0** | **0** |
| 100,000 | パス等式 | **6,732,160** | **0** | **0** |

**パス等式の検査は追加確保がゼロ。** これは実装から言えます——既にある `map` を辿って比べるだけなので。

速さのほうは条件つきです。索引を作り直す版と比べれば 2.18倍ですが、**索引を使い回す版と比べると 1.44倍**まで縮みます。しかも2つは**同じ量の仕事をしていません**（全域性は3本の射について「引く」と「終域に入るか見る」、等式は2本について「引く」だけ）。

言えるのは「高級な制約だから高くつく、とは限らない」までです。

### 3. 移行の3方向 —— Δ は無料、Σ は和、Π は積

```bash
go test -bench=MigrationDirections -benchmem -run='^$' ./part10/
```

スキーマ間の写像 $F: C \to D$ があると、データの移行が**3方向**に決まります。

$$\Sigma_F \dashv \Delta_F \dashv \Pi_F$$

- $\Delta_F$ は $D$ 上のデータを $C$ 上へ引き戻す
- $\Sigma_F$ はその**左**随伴。余極限の側——**和**
- $\Pi_F$ はその**右**随伴。極限の側——**積**

Employee 表と Contractor 表を Person 表へ移す場合で測ります（`n` は片側の行数）。

| n | 方向 | ns/op | B/op | allocs/op | 出力の行数 |
|---:|---|---:|---:|---:|---:|
| 4,000 | Δ（引き戻し） | 2 | **0** | **0** | 各表 n（2表で 2n） |
| 4,000 | Σ（和） | 7,114 | 131,072 | 1 | 2n |
| 4,000 | Π（積） | 8,890,071 | 256,000,003 | 1 | n² |

$\Sigma$ の確保量は n を4倍にすると**ちょうど 4.0倍**、$\Pi$ は **15.993倍**（ほぼ16倍）です。

> **Δ の 2ns は「引き戻しが無料」という意味ではありません。** この `Delta` はスライスヘッダを2つ返すだけで、行の検査もコピーも表の構築もしていません。インライン化もされうる。言えるのは「この表現では追加確保がゼロ」までです。

**この離散スキーマでは、随伴の側と確保量の形が対応しています。** ただし一般の $\Sigma_F$ と $\Pi_F$ は Kan 拡張で、計算量は関わるコンマ圏・射・等式・表現によって変わります。**「左随伴だから常に線形」ではありません。**

### 4. 随伴を普遍性として確かめる

```bash
go test -v -run Adjoint ./part10/
```

射のないこの場合、インスタンス準同型はただの写像になるので、随伴は**余積と積の普遍性そのもの**に落ちます。

$$\mathrm{Hom}_D(\Sigma I, J) \cong \mathrm{Hom}_C(I, \Delta J), \qquad \mathrm{Hom}_C(\Delta J, I) \cong \mathrm{Hom}_D(J, \Pi I)$$

`TestSigmaIsLeftAdjointToDelta` と `TestDeltaIsLeftAdjointToPi` が、この全単射を両方向で確かめます。

## ファイルの地図

| ファイル | 中身 |
|---|---|
| `schema.go` | 圏としてのスキーマ（対象・射・パス等式）と、スキーマ自身の検証 |
| `instance.go` | 関手としてのインスタンス、全域性とパス等式の検査 |
| `migrate.go` | Σ ⊣ Δ ⊣ Π と、随伴の全単射 |
| `schema_test.go` | 外部キーが正しくてもパス等式が破れる例 |
| `migrate_test.go` | 随伴の検証、3方向の行数、コスト測定 |
