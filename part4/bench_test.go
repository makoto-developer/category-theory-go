package part4

import (
	"context"
	"testing"
)

var (
	sinkState State
	sinkEndo  Endo
	sinkDiag  string
)

// 素直な実装と、合成した射の実行コストを比べる。外部サービスは失敗しない設定。
func BenchmarkWorkflow(b *testing.B) {
	ctx := context.Background()
	in := State{Order: sampleOrder}

	b.Run("naive", func(b *testing.B) {
		deps, _, _, _, _ := newDeps()
		w := NaiveWorkflow{Deps: deps, Recorder: NewRecorder(), Attempts: 3}
		for b.Loop() {
			sinkState, _ = w.Run(ctx, in)
		}
	})

	// naive にはステップ単位のタイムアウトが無いので、条件を揃えた版も測る。
	b.Run("composed_no_timeout", func(b *testing.B) {
		deps, _, _, _, _ := newDeps()
		plan := OrderPlan(deps)
		for i := range plan.Nodes {
			plan.Nodes[i].Timeout = 0
		}
		compiled := Compile(plan, NewRecorder())
		for b.Loop() {
			sinkState, _ = compiled(ctx, in)
		}
	})

	b.Run("composed", func(b *testing.B) {
		deps, _, _, _, _ := newDeps()
		compiled := Compile(OrderPlan(deps), NewRecorder())
		for b.Loop() {
			sinkState, _ = compiled(ctx, in)
		}
	})
}

// 計測を外した場合。時間計測そのもののコストを切り分ける。
func BenchmarkWorkflowWithoutTiming(b *testing.B) {
	ctx := context.Background()
	in := State{Order: sampleOrder}
	deps, _, _, _, _ := newDeps()

	bare := Sequence(Reserve(deps), Charge(deps), Arrange(deps), Notify(deps))

	for b.Loop() {
		sinkState, _ = bare(ctx, in)
	}
}

// Plan から射を組み立てるコスト。起動時に1回だけ払えばよい。
func BenchmarkCompile(b *testing.B) {
	deps, _, _, _, _ := newDeps()
	plan := OrderPlan(deps)
	rec := NewRecorder()

	for b.Loop() {
		sinkEndo = Compile(plan, rec)
	}
}

// 図の生成コスト。ドキュメントを毎回生成しても問題ない水準か。
func BenchmarkMermaid(b *testing.B) {
	deps, _, _, _, _ := newDeps()
	plan := OrderPlan(deps)

	for b.Loop() {
		sinkDiag = Mermaid(plan)
	}
}
