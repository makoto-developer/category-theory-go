package part4

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

// 2つの実装が同じ結果を返すことを、まず押さえる。
func TestNaiveAndComposedAgree(t *testing.T) {
	ctx := context.Background()

	naiveDeps, _, _, _, _ := newDeps()
	naive := NaiveWorkflow{Deps: naiveDeps, Recorder: NewRecorder(), Attempts: 3}
	naiveOut, naiveErr := naive.Run(ctx, State{Order: sampleOrder})

	compDeps, _, _, _, _ := newDeps()
	composed := Compile(OrderPlan(compDeps), NewRecorder())
	compOut, compErr := composed(ctx, State{Order: sampleOrder})

	if naiveErr != nil || compErr != nil {
		t.Fatalf("どちらかが失敗した: naive=%v, composed=%v", naiveErr, compErr)
	}
	if naiveOut != compOut {
		t.Fatalf("結果が違う:\n  naive   =%+v\n  composed=%+v", naiveOut, compOut)
	}
}

// 一時的な失敗はリトライで吸収される。
func TestRetryAbsorbsTemporaryFailures(t *testing.T) {
	deps, inv, _, _, _ := newDeps()
	inv.tempFailures = 2

	out, err := Compile(OrderPlan(deps), NewRecorder())(context.Background(), State{Order: sampleOrder})

	if err != nil {
		t.Fatalf("リトライで回復できていない: %v", err)
	}
	if inv.calls != 3 {
		t.Fatalf("呼び出し回数が想定と違う: got=%d, want=3", inv.calls)
	}
	if out.ReservationID != "RSV-1" {
		t.Fatalf("在庫確保の結果が入っていない: %+v", out)
	}
}

// 恒久的な失敗はリトライしない。無駄な再試行は課金や在庫を二重に動かす。
func TestPermanentFailureIsNotRetried(t *testing.T) {
	deps, _, pay, _, _ := newDeps()
	pay.permanent = ErrPaymentDenid

	_, err := Compile(OrderPlan(deps), NewRecorder())(context.Background(), State{Order: sampleOrder})

	if !errors.Is(err, ErrPaymentDenid) {
		t.Fatalf("想定したエラーが返っていない: %v", err)
	}
	if pay.calls != 1 {
		t.Fatalf("恒久エラーなのに再試行された: calls=%d", pay.calls)
	}
}

// 失敗したステップの名前がエラーに残る。
func TestErrorCarriesStepName(t *testing.T) {
	deps, _, _, ship, _ := newDeps()
	ship.permanent = errors.New("配送業者がダウン")

	_, err := Compile(OrderPlan(deps), NewRecorder())(context.Background(), State{Order: sampleOrder})

	if err == nil || !strings.Contains(err.Error(), "配送手配") {
		t.Fatalf("ステップ名が含まれていない: %v", err)
	}
}

// キャンセル済みの context では最初のステップすら実行されない。
func TestCanceledContextStopsBeforeFirstStep(t *testing.T) {
	deps, inv, _, _, _ := newDeps()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Compile(OrderPlan(deps), NewRecorder())(ctx, State{Order: sampleOrder})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("キャンセルが伝わっていない: %v", err)
	}
	if inv.calls != 0 {
		t.Fatalf("キャンセル済みなのに在庫確保が走った: calls=%d", inv.calls)
	}
}

// 空のワークフローは恒等射。State をそのまま返す。
func TestEmptyPlanIsIdentity(t *testing.T) {
	in := State{Order: sampleOrder}

	out, err := Compile(Plan{Name: "空"}, NewRecorder())(context.Background(), in)

	if err != nil || out != in {
		t.Fatalf("恒等射になっていない: out=%+v, err=%v", out, err)
	}
}

// 計測は全ステップに自動的に付く。業務ロジック側には1行も書いていない。
func TestTimingIsRecordedForEveryStep(t *testing.T) {
	deps, _, _, _, _ := newDeps()
	rec := NewRecorder()

	if _, err := Compile(OrderPlan(deps), rec)(context.Background(), State{Order: sampleOrder}); err != nil {
		t.Fatalf("実行に失敗: %v", err)
	}

	want := []string{"在庫確保", "決済", "配送手配", "通知"}
	if !slices.Equal(rec.Order, want) {
		t.Fatalf("計測されたステップが違う: got=%v, want=%v", rec.Order, want)
	}
}

// タイムアウトは各ステップに個別に効く。
func TestTimeoutAppliesPerStep(t *testing.T) {
	slow := &slowInventory{delay: 50 * time.Millisecond}
	deps, _, pay, ship, notif := newDeps()
	deps.Inventory = slow

	plan := OrderPlan(deps)
	plan.Nodes[0].Timeout = 10 * time.Millisecond
	plan.Nodes[0].Retries = 1

	_, err := Compile(plan, NewRecorder())(context.Background(), State{Order: sampleOrder})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("タイムアウトしていない: %v", err)
	}
	if pay.calls != 0 || ship.calls != 0 || notif.calls != 0 {
		t.Fatal("タイムアウト後に後続が実行された")
	}
}

type slowInventory struct{ delay time.Duration }

func (s *slowInventory) Reserve(ctx context.Context, _ Order) (string, error) {
	select {
	case <-time.After(s.delay):
		return "RSV-slow", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
