package part10

// スキーマ間の写像 F: C → D があると、データの移行が3方向に決まる。
//
//	Σ_F ⊣ Δ_F ⊣ Π_F
//
// Δ は D 上のデータを C 上へ引き戻す（対象の名前を付け替えるだけ）。
// Σ は Δ の左随伴で、余極限の側——和。
// Π は Δ の右随伴で、極限の側——積。
//
// ここでは移行の形をいちばん単純な場合で見る。
// C は対象2つ（Employee, Contractor）で射なし、D は対象1つ（Person）で射なし、
// F は両方を Person へ送る。このとき随伴の中身は、余積と積の普遍性そのものになる。

// TwoTables は C 側のデータ。射が無いので行 ID の集合が2つあるだけ。
type TwoTables struct {
	Employees   []int
	Contractors []int
}

// OneTable は D 側のデータ。
type OneTable struct {
	Persons []int
}

// Tagged は和を取ったあとの行。どちらから来たかを覚えている（余積のタグ）。
type Tagged struct {
	FromEmployee bool
	ID           int
}

// MigPair は積を取ったあとの行。
type MigPair struct{ E, C int }

// Delta は D 上のデータを C 上へ引き戻す。Person をそのまま両方に配る。
// スライスを共有するので、行のコピーは起きない。
func Delta(j OneTable) (employees, contractors []int) {
	return j.Persons, j.Persons
}

// Sigma は Δ の左随伴。C の2つの表を D の1つの表へ「足す」。
// 余積なので行数は len(E) + len(C)。
func Sigma(i TwoTables) []Tagged {
	out := make([]Tagged, 0, len(i.Employees)+len(i.Contractors))
	for _, id := range i.Employees {
		out = append(out, Tagged{FromEmployee: true, ID: id})
	}
	for _, id := range i.Contractors {
		out = append(out, Tagged{FromEmployee: false, ID: id})
	}
	return out
}

// Pi は Δ の右随伴。C の2つの表を D の1つの表へ「掛ける」。
// 積なので行数は len(E) × len(C)。
func Pi(i TwoTables) []MigPair {
	out := make([]MigPair, 0, len(i.Employees)*len(i.Contractors))
	for _, e := range i.Employees {
		for _, c := range i.Contractors {
			out = append(out, MigPair{E: e, C: c})
		}
	}
	return out
}

// --- 随伴を普遍性として確かめるための道具 ---------------------------------
//
// 射がないので、この場合のインスタンス準同型はただの写像になる。
//
//	Hom_D(Σ I, J) ≅ Hom_C(I, Δ J)   ← 余積の普遍性
//	Hom_C(Δ J, I) ≅ Hom_D(J, Π I)   ← 積の普遍性

// SigmaAdjunctLeft は (I_E → J, I_C → J) の組を Σ I → J に移す。
func SigmaAdjunctLeft(fromE, fromC func(int) int) func(Tagged) int {
	return func(t Tagged) int {
		if t.FromEmployee {
			return fromE(t.ID)
		}
		return fromC(t.ID)
	}
}

// SigmaAdjunctRight は逆向き。Σ I → J を (I_E → J, I_C → J) に戻す。
func SigmaAdjunctRight(h func(Tagged) int) (fromE, fromC func(int) int) {
	return func(id int) int { return h(Tagged{FromEmployee: true, ID: id}) },
		func(id int) int { return h(Tagged{FromEmployee: false, ID: id}) }
}

// PiAdjunctLeft は (J → I_E, J → I_C) の組を J → Π I に移す。
func PiAdjunctLeft(toE, toC func(int) int) func(int) MigPair {
	return func(id int) MigPair { return MigPair{E: toE(id), C: toC(id)} }
}

// PiAdjunctRight は逆向き。
func PiAdjunctRight(h func(int) MigPair) (toE, toC func(int) int) {
	return func(id int) int { return h(id).E }, func(id int) int { return h(id).C }
}
