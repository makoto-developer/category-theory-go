package part8

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"pgregory.net/rapid"
)

func tripleKeys(ts []Triple) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		out = append(out, fmt.Sprintf("%d|%d,%d|%d", t.U.ID, t.A.Dept, t.A.Project, t.P.Project))
	}
	slices.Sort(out)
	return out
}

// 引き戻しの補題を実務の言葉にすると「結合の順序を変えても同じ答えになる」。
// これを property-based test で確かめる。ここが成り立たなければ、
// 以下のベンチマークは単に別の計算を比べているだけになる。
func TestJoinOrdersAgree(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		us := rapid.SliceOfN(rapid.Custom(func(t *rapid.T) User {
			return User{ID: rapid.IntRange(0, 50).Draw(t, "id"), Dept: rapid.IntRange(1, 4).Draw(t, "dept")}
		}), 0, 12).Draw(t, "us")
		as := rapid.SliceOfN(rapid.Custom(func(t *rapid.T) Assignment {
			return Assignment{Dept: rapid.IntRange(1, 4).Draw(t, "d"), Project: rapid.IntRange(1, 6).Draw(t, "p")}
		}), 0, 12).Draw(t, "as")
		ps := rapid.SliceOfN(rapid.Custom(func(t *rapid.T) ActiveProject {
			return ActiveProject{Project: rapid.IntRange(1, 6).Draw(t, "p")}
		}), 0, 6).Draw(t, "ps")

		left := tripleKeys(JoinLeftFirst(us, as, ps))
		right := tripleKeys(JoinRightFirst(us, as, ps))
		if !slices.Equal(left, right) {
			t.Fatalf("結合順で答えが変わった:\n左優先 %v\n右優先 %v", left, right)
		}
	})
}

// 3-way 結合のデータ。
// dept は 10種類しかないので User ⋈ Assignment は膨らむ。
// active な project は5件しかないので Assignment ⋈ ActiveProject は縮む。
func makeJoinData(nUsers, nAssign, nProjects, nActive int) ([]User, []Assignment, []ActiveProject) {
	r := rand.New(rand.NewPCG(1, 2))
	us := make([]User, nUsers)
	for i := range us {
		us[i] = User{ID: i, Dept: 1 + r.IntN(10)}
	}
	as := make([]Assignment, nAssign)
	for i := range as {
		as[i] = Assignment{Dept: 1 + r.IntN(10), Project: 1 + r.IntN(nProjects)}
	}
	ps := make([]ActiveProject, nActive)
	for i := range ps {
		ps[i] = ActiveProject{Project: 1 + i}
	}
	return us, as, ps
}

var sinkTriples []Triple

// 同じ答えを返す2つの順序で、実行時間と確保量がどれだけ違うか。
func BenchmarkJoinOrder(b *testing.B) {
	us, as, ps := makeJoinData(4000, 4000, 2000, 5)
	lf, rf := IntermediateSizes(us, as, ps)
	b.Logf("中間データの要素数: 左優先 %d / 右優先 %d", lf, rf)

	b.Run("1_left_first", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkTriples = JoinLeftFirst(us, as, ps)
		}
	})
	b.Run("2_right_first", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkTriples = JoinRightFirst(us, as, ps)
		}
	})
}

var sinkPairs []Pair[int, int]

// 引き戻しの作り方3種。できる対象は同じで、作り方だけが違う。
func BenchmarkPullbackConstruction(b *testing.B) {
	for _, n := range []int{100, 1000} {
		r := rand.New(rand.NewPCG(3, 4))
		as := make([]int, n)
		bs := make([]int, n)
		for i := range as {
			as[i] = r.IntN(n)
			bs[i] = r.IntN(n)
		}
		f := func(a int) int { return a % (n / 10) }
		g := func(x int) int { return x % (n / 10) }

		b.Run(fmt.Sprintf("n=%04d/1_nested", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkPairs = PullbackNested(as, bs, f, g)
			}
		})
		b.Run(fmt.Sprintf("n=%04d/2_hash", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkPairs = PullbackHash(as, bs, f, g)
			}
		})
		b.Run(fmt.Sprintf("n=%04d/3_product_then_equalizer", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkPairs = PullbackViaProductAndEqualizer(as, bs, f, g)
			}
		})
	}
}
