package part10

import (
	"math/rand/v2"
	"strings"
	"testing"
)

// Spivak の例。上司は同じ部署にいる、という制約をパス等式で書く。
// RDB では外部キーで書けない種類の制約。
func orgSchema() Schema {
	return Schema{
		Objects: []string{"Employee", "Department"},
		Arrows: map[string]Arrow{
			"worksIn": {From: "Employee", To: "Department"},
			"manager": {From: "Employee", To: "Employee"},
			"head":    {From: "Department", To: "Employee"},
		},
		Equations: []Equation{
			// 上司は同じ部署にいる
			{From: "Employee", To: "Department", Left: []string{"manager", "worksIn"}, Right: []string{"worksIn"}},
			// 部署長はその部署に所属する
			{From: "Department", To: "Department", Left: []string{"head", "worksIn"}, Right: nil},
		},
	}
}

func TestSchemaValidates(t *testing.T) {
	if err := orgSchema().Validate(); err != nil {
		t.Fatalf("スキーマが不正: %v", err)
	}
}

func TestSchemaRejectsBrokenPaths(t *testing.T) {
	s := orgSchema()
	// 繋がらないパスを入れる
	s.Equations = append(s.Equations, Equation{
		From: "Employee", To: "Employee", Left: []string{"worksIn", "worksIn"}, Right: nil,
	})
	err := s.Validate()
	if err == nil {
		t.Fatalf("繋がらないパスを弾けていない")
	}
	if !strings.Contains(err.Error(), "繋がっていない") {
		t.Fatalf("想定と違う理由で落ちた: %v", err)
	}
	t.Logf("弾いた: %v", err)
}

// 等式を満たすインスタンスを作る。部署ごとに1人の長を置き、全員の上司をその長にする。
func makeValidInstance(nEmp, nDept int) Instance {
	emp := make([]int, nEmp)
	for i := range emp {
		emp[i] = i
	}
	dept := make([]int, nDept)
	for i := range dept {
		dept[i] = i
	}
	worksIn := make(map[int]int, nEmp)
	manager := make(map[int]int, nEmp)
	head := make(map[int]int, nDept)
	for d := 0; d < nDept; d++ {
		head[d] = d // 部署 d の長は社員 d（社員 d は部署 d にいる）
	}
	for e := 0; e < nEmp; e++ {
		d := e % nDept
		worksIn[e] = d
		manager[e] = head[d] // 上司は自分の部署の長
	}
	return Instance{
		Schema: orgSchema(),
		Rows:   map[string][]int{"Employee": emp, "Department": dept},
		Maps:   map[string]map[int]int{"worksIn": worksIn, "manager": manager, "head": head},
	}
}

func TestValidInstancePassesBothChecks(t *testing.T) {
	i := makeValidInstance(200, 7)
	if bad := i.CheckTotality(); len(bad) > 0 {
		t.Fatalf("全域性が破れた: %v", bad[:min(3, len(bad))])
	}
	if bad := i.CheckEquations(); len(bad) > 0 {
		t.Fatalf("パス等式が破れた: %v", bad[:min(3, len(bad))])
	}
}

// 外部キーとしては正しいのに、パス等式を破るデータを作れる。
// ここが「RDB の制約では書けない」ことの具体例。
func TestForeignKeysCanBeValidWhileEquationsBreak(t *testing.T) {
	i := makeValidInstance(20, 3)
	// 社員1の上司を、別の部署の長に付け替える。外部キーとしては正しい。
	i.Maps["manager"][1] = i.Maps["head"][(i.Maps["worksIn"][1]+1)%3]

	if bad := i.CheckTotality(); len(bad) > 0 {
		t.Fatalf("外部キーは正しいはず: %v", bad)
	}
	bad := i.CheckEquations()
	if len(bad) == 0 {
		t.Fatalf("パス等式が破れるはず")
	}
	t.Logf("外部キーは全部正しいのに、パス等式が %d 件破れた。例: %s", len(bad), bad[0])
}

// 無作為なデータでは、外部キーとしては常に正しいのにパス等式は高い割合で破れる。
// 「ほぼ必ず」は確率的な主張なので、件数を数えて割合で述べる。
func TestRandomDataKeepsForeignKeysButBreaksEquations(t *testing.T) {
	const trials = 500
	broken := 0
	for seed := 1; seed <= trials; seed++ {
		r := rand.New(rand.NewPCG(uint64(seed), 99))
		nEmp, nDept := 3+r.IntN(18), 2+r.IntN(4)

		emp := make([]int, nEmp)
		for i := range emp {
			emp[i] = i
		}
		dept := make([]int, nDept)
		for i := range dept {
			dept[i] = i
		}
		worksIn, manager, head := map[int]int{}, map[int]int{}, map[int]int{}
		for e := range emp {
			worksIn[e] = r.IntN(nDept)
			manager[e] = r.IntN(nEmp)
		}
		for d := range dept {
			head[d] = r.IntN(nEmp)
		}
		i := Instance{Schema: orgSchema(),
			Rows: map[string][]int{"Employee": emp, "Department": dept},
			Maps: map[string]map[int]int{"worksIn": worksIn, "manager": manager, "head": head}}

		// 外部キーとしては常に正しい（行き先は必ず実在する）
		if bad := i.CheckTotality(); len(bad) > 0 {
			t.Fatalf("seed %d: 全域性は満たすはず: %v", seed, bad)
		}
		if len(i.CheckEquations()) > 0 {
			broken++
		}
	}
	t.Logf("外部キーは %d/%d 件すべて正しく、パス等式は %d/%d 件で破れた（%.1f%%）",
		trials, trials, broken, trials, float64(broken)/trials*100)
	if broken < trials*8/10 {
		t.Fatalf("この分布なら大半で破れるはず: %d/%d", broken, trials)
	}
}

// 索引を作り直さない全域性検査は、作り直す版と同じ答えを返す。
func TestTotalityWithIndexAgrees(t *testing.T) {
	i := makeValidInstance(500, 11)
	idx := i.BuildIndex()
	if a, b := len(i.CheckTotality()), len(i.CheckTotalityWithIndex(idx)); a != b {
		t.Fatalf("件数が違う: %d vs %d", a, b)
	}
	// 壊してからも一致すること
	i.Maps["worksIn"][0] = 9999
	if a, b := len(i.CheckTotality()), len(i.CheckTotalityWithIndex(i.BuildIndex())); a != b || a == 0 {
		t.Fatalf("壊した状態で一致しない: %d vs %d", a, b)
	}
}
