package part4

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

// dry-run は外部サービスを一切呼ばない。同じ定義から手順だけを取り出せる。
func TestDryRunTouchesNothing(t *testing.T) {
	deps, inv, pay, ship, notif := newDeps()

	steps := DryRun(OrderPlan(deps))

	want := []string{"在庫確保", "決済", "配送手配", "通知"}
	if !slices.Equal(steps, want) {
		t.Fatalf("手順が違う: got=%v, want=%v", steps, want)
	}
	if inv.calls+pay.calls+ship.calls+notif.calls != 0 {
		t.Fatal("dry-run で外部サービスが呼ばれた")
	}
}

// 同じ定義から図が出る。コードと図がずれない。
func TestMermaidIsGeneratedFromTheSamePlan(t *testing.T) {
	deps, _, _, _, _ := newDeps()

	diagram := Mermaid(OrderPlan(deps))

	for _, want := range []string{"flowchart TD", "在庫確保", "最大3回", "決済", "配送手配", "通知", "完了"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("図に %q が含まれていない:\n%s", want, diagram)
		}
	}
	t.Logf("生成された図:\n%s", diagram)
}

// 運用時の問い「どのステップが何回リトライするのか」に、定義から答えられる。
func TestExplainListsRetryPolicy(t *testing.T) {
	deps, _, _, _, _ := newDeps()

	explained := Explain(OrderPlan(deps))

	if !strings.Contains(explained, "決済: リトライなし") {
		t.Fatalf("決済のリトライ設定が読み取れない:\n%s", explained)
	}
	if !strings.Contains(explained, "在庫確保: 最大3回 / 2s") {
		t.Fatalf("在庫確保の設定が読み取れない:\n%s", explained)
	}
	t.Logf("設定一覧:\n%s", explained)
}

// 並列ノードは独立したステップを同時に走らせ、失敗を全部集める（Applicative）。
func TestParallelNodeCollectsAllFailures(t *testing.T) {
	failA := errors.New("メール送信に失敗")
	failB := errors.New("プッシュ通知に失敗")

	plan := Plan{
		Name: "通知だけ",
		Nodes: []Node{{
			Name: "通知",
			Parallel: []Node{
				{Name: "メール", Run: failing(failA)},
				{Name: "プッシュ", Run: failing(failB)},
				{Name: "SMS", Run: Identity},
			},
		}},
	}

	_, err := Compile(plan, NewRecorder())(context.Background(), State{Order: sampleOrder})

	if !errors.Is(err, failA) || !errors.Is(err, failB) {
		t.Fatalf("両方の失敗が集まっていない: %v", err)
	}
}

func failing(err error) Endo {
	return func(_ context.Context, s State) (State, error) { return s, err }
}

// 定義を書き換えるだけでリトライ方針を変えられる。業務ロジックには触らない。
func TestPolicyChangeDoesNotTouchBusinessLogic(t *testing.T) {
	deps, inv, _, _, _ := newDeps()
	inv.tempFailures = 4

	plan := OrderPlan(deps)
	plan.Nodes[0].Retries = 5

	if _, err := Compile(plan, NewRecorder())(context.Background(), State{Order: sampleOrder}); err != nil {
		t.Fatalf("リトライ回数を増やしても回復しなかった: %v", err)
	}
	if inv.calls != 5 {
		t.Fatalf("リトライ回数が反映されていない: calls=%d, want=5", inv.calls)
	}
}
