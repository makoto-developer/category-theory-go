package part10

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"
)

// Σ ⊣ Δ の随伴。Hom_D(Σ I, J) ≅ Hom_C(I, Δ J) が全単射であること。
// 射がないこの場合、これは余積の普遍性そのもの。
func TestSigmaIsLeftAdjointToDelta(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nE := rapid.IntRange(0, 8).Draw(t, "nE")
		nC := rapid.IntRange(0, 8).Draw(t, "nC")
		// 任意の (I_E → J, I_C → J) を作る
		ke := rapid.IntRange(1, 9).Draw(t, "ke")
		kc := rapid.IntRange(1, 9).Draw(t, "kc")
		fromE := func(id int) int { return id*ke + 1 }
		fromC := func(id int) int { return id*kc + 2 }

		// 組 → Σ I → J → 組 で元に戻る
		h := SigmaAdjunctLeft(fromE, fromC)
		gotE, gotC := SigmaAdjunctRight(h)
		for id := 0; id < nE; id++ {
			if gotE(id) != fromE(id) {
				t.Fatalf("Employee 側が戻らない: %d", id)
			}
		}
		for id := 0; id < nC; id++ {
			if gotC(id) != fromC(id) {
				t.Fatalf("Contractor 側が戻らない: %d", id)
			}
		}

		// Σ I → J → 組 → Σ I → J でも元に戻る
		orig := func(tg Tagged) int {
			if tg.FromEmployee {
				return tg.ID * 3
			}
			return tg.ID*5 + 7
		}
		e2, c2 := SigmaAdjunctRight(orig)
		back := SigmaAdjunctLeft(e2, c2)
		for id := 0; id < nE; id++ {
			tg := Tagged{FromEmployee: true, ID: id}
			if back(tg) != orig(tg) {
				t.Fatalf("往復しない: %v", tg)
			}
		}
		for id := 0; id < nC; id++ {
			tg := Tagged{FromEmployee: false, ID: id}
			if back(tg) != orig(tg) {
				t.Fatalf("往復しない: %v", tg)
			}
		}
	})
}

// Δ ⊣ Π の随伴。Hom_C(Δ J, I) ≅ Hom_D(J, Π I)。積の普遍性そのもの。
func TestDeltaIsLeftAdjointToPi(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 8).Draw(t, "n")
		ke := rapid.IntRange(1, 9).Draw(t, "ke")
		kc := rapid.IntRange(1, 9).Draw(t, "kc")
		toE := func(id int) int { return id*ke + 1 }
		toC := func(id int) int { return id*kc + 2 }

		h := PiAdjunctLeft(toE, toC)
		gotE, gotC := PiAdjunctRight(h)
		for id := 0; id < n; id++ {
			if gotE(id) != toE(id) || gotC(id) != toC(id) {
				t.Fatalf("成分が戻らない: %d", id)
			}
		}

		orig := func(id int) MigPair { return MigPair{E: id * 3, C: id*5 + 7} }
		e2, c2 := PiAdjunctRight(orig)
		back := PiAdjunctLeft(e2, c2)
		for id := 0; id < n; id++ {
			if back(id) != orig(id) {
				t.Fatalf("往復しない: %d", id)
			}
		}
	})
}

// 3方向の行数。Δ は変えず、Σ は足し、Π は掛ける。
func TestRowCountsOfTheThreeDirections(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nE := rapid.IntRange(0, 20).Draw(t, "nE")
		nC := rapid.IntRange(0, 20).Draw(t, "nC")
		i := TwoTables{Employees: make([]int, nE), Contractors: make([]int, nC)}

		if got := len(Sigma(i)); got != nE+nC {
			t.Fatalf("Σ の行数は和のはず: %d ≠ %d", got, nE+nC)
		}
		if got := len(Pi(i)); got != nE*nC {
			t.Fatalf("Π の行数は積のはず: %d ≠ %d", got, nE*nC)
		}
		e, c := Delta(OneTable{Persons: make([]int, nE)})
		if len(e) != nE || len(c) != nE {
			t.Fatalf("Δ は行数を変えないはず")
		}
	})
}

var (
	sinkTagged []Tagged
	sinkPairs  []MigPair
	sinkInts   []int
)

// 3方向のコスト。Δ は無料、Σ は和、Π は積。
func BenchmarkMigrationDirections(b *testing.B) {
	for _, n := range []int{1000, 4000} {
		i := TwoTables{Employees: make([]int, n), Contractors: make([]int, n)}
		j := OneTable{Persons: make([]int, n)}

		b.Run(fmt.Sprintf("n=%04d/1_delta", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				e, _ := Delta(j)
				sinkInts = e
			}
		})
		b.Run(fmt.Sprintf("n=%04d/2_sigma", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkTagged = Sigma(i)
			}
		})
		b.Run(fmt.Sprintf("n=%04d/3_pi", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkPairs = Pi(i)
			}
		})
	}
}

var sinkBad []string

// パス等式の検査コスト。行数に比例するはず。
func BenchmarkCheckEquations(b *testing.B) {
	for _, n := range []int{1000, 100000} {
		inst := makeValidInstance(n, 17)
		b.Run(fmt.Sprintf("n=%06d/1_totality", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkBad = inst.CheckTotality()
			}
		})
		b.Run(fmt.Sprintf("n=%06d/2_totality_prebuilt_index", n), func(b *testing.B) {
			idx := inst.BuildIndex()
			b.ResetTimer()
			b.ReportAllocs()
			for b.Loop() {
				sinkBad = inst.CheckTotalityWithIndex(idx)
			}
		})
		b.Run(fmt.Sprintf("n=%06d/3_equations", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkBad = inst.CheckEquations()
			}
		})
	}
}
