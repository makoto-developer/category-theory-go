package part4

import (
	"context"
	"errors"
	"testing"
	"time"
)

// デコレータは業務ロジックを一切使わずに単体でテストできる。
// 素直な実装では、リトライの挙動を確かめるのにワークフロー全体を動かす必要がある。
func TestWithRetryOnlyRetriesTemporaryErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCalls int
	}{
		{"一時的な失敗は再試行する", ErrTemporary, 3},
		{"恒久的な失敗は再試行しない", ErrOutOfStock, 1},
		{"成功したら再試行しない", nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			step := Endo(func(_ context.Context, s State) (State, error) {
				calls++
				return s, tt.err
			})

			_, err := WithRetry(step, 3)(context.Background(), State{})

			if calls != tt.wantCalls {
				t.Fatalf("呼び出し回数が違う: got=%d, want=%d", calls, tt.wantCalls)
			}
			if !errors.Is(err, tt.err) {
				t.Fatalf("エラーが違う: got=%v, want=%v", err, tt.err)
			}
		})
	}
}

// 失敗しても計測は記録される。失敗したステップこそ所要時間が知りたい。
func TestWithTimingRecordsEvenOnFailure(t *testing.T) {
	rec := NewRecorder()
	failing := Endo(func(_ context.Context, s State) (State, error) {
		time.Sleep(time.Millisecond)
		return s, ErrOutOfStock
	})

	if _, err := WithTiming(failing, "失敗するステップ", rec)(context.Background(), State{}); err == nil {
		t.Fatal("エラーが返っていない")
	}

	if d, ok := rec.Timings["失敗するステップ"]; !ok || d < time.Millisecond {
		t.Fatalf("失敗時に計測されていない: %v", rec.Timings)
	}
}

// タイムアウトはリトライ全体に効く。デコレータを巻く順序が結果を決めている。
func TestTimeoutWrapsAllRetries(t *testing.T) {
	calls := 0
	slow := Endo(func(ctx context.Context, s State) (State, error) {
		calls++
		select {
		case <-time.After(20 * time.Millisecond):
			return s, ErrTemporary
		case <-ctx.Done():
			return s, ctx.Err()
		}
	})

	// compileNode と同じ順序: リトライの外側にタイムアウトを巻く
	step := WithTimeout(WithRetry(slow, 10), 50*time.Millisecond)

	_, err := step(context.Background(), State{})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("タイムアウトしていない: %v", err)
	}
	if calls >= 10 {
		t.Fatalf("タイムアウトがリトライを止めていない: calls=%d", calls)
	}
	t.Logf("50ms のタイムアウトで %d 回試行された", calls)
}
