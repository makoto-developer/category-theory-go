package part4

import (
	"context"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// captureTrace は f の中で起きたパニックのスタックトレースを返す。
func captureTrace(f func()) (trace string) {
	defer func() {
		if r := recover(); r != nil {
			trace = string(debug.Stack())
		}
	}()
	f()
	return ""
}

// panicking は必ずパニックする射。
func panicking() Endo {
	return func(_ context.Context, s State) (State, error) {
		panic("在庫サービスが nil を返した")
	}
}

type panickingInventory struct{}

func (panickingInventory) Reserve(_ context.Context, _ Order) (string, error) {
	panic("在庫サービスが nil を返した")
}

// 同じパニックが、2つの実装でスタックトレースにどう出るかを比べる。
// 記事に載せた出力はこのテストで実際に取得したもの。
func TestPanicTraceShape(t *testing.T) {
	naiveTrace := captureTrace(func() {
		w := NaiveWorkflow{
			Deps:     Deps{Inventory: panickingInventory{}},
			Recorder: NewRecorder(),
			Attempts: 3,
		}
		_, _ = w.Run(context.Background(), State{Order: sampleOrder})
	})

	plan := Plan{
		Name:  "落ちるワークフロー",
		Nodes: []Node{{Name: "在庫確保", Retries: 3, Timeout: time.Second, Run: panicking()}},
	}
	composedTrace := captureTrace(func() {
		_, _ = Compile(plan, NewRecorder())(context.Background(), State{Order: sampleOrder})
	})

	if naiveTrace == "" || composedTrace == "" {
		t.Fatal("パニックが捕まえられなかった")
	}

	t.Logf("素直な実装:\n%s", part4Frames(naiveTrace))
	t.Logf("合成版:\n%s", part4Frames(composedTrace))

	// 合成版のほうがフレームが深くなる。これが読みにくさの正体。
	if part4FrameCount(composedTrace) <= part4FrameCount(naiveTrace) {
		t.Fatalf("合成版のフレームが増えていない: naive=%d, composed=%d",
			part4FrameCount(naiveTrace), part4FrameCount(composedTrace))
	}
}

// part4Frames はスタックトレースから part4 のフレームだけを抜き出す。
func part4Frames(trace string) string {
	var out []string
	lines := strings.Split(trace, "\n")
	for i, line := range lines {
		if strings.Contains(line, "category-theory-go/part4.") && i+1 < len(lines) {
			out = append(out, strings.TrimSpace(line), "    "+strings.TrimSpace(lines[i+1]))
		}
	}
	return strings.Join(out, "\n")
}

func part4FrameCount(trace string) int {
	return strings.Count(trace, "category-theory-go/part4.")
}
