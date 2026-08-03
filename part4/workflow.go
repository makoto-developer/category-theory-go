package part4

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Endo は State から State への射。すべてのステップがこの形をしているので、
// 全体が自己射のモノイドになり、リストに並べて畳み込める。
type Endo func(context.Context, State) (State, error)

// Recorder は各ステップの所要時間を集める。並行に書かれるのでロックする。
type Recorder struct {
	mu      sync.Mutex
	Timings map[string]time.Duration
	Order   []string
}

func NewRecorder() *Recorder {
	return &Recorder{Timings: make(map[string]time.Duration)}
}

func (r *Recorder) Record(name string, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, seen := r.Timings[name]; !seen {
		r.Order = append(r.Order, name)
	}
	r.Timings[name] += d
}

// Identity は何もしない射。空のワークフローがこれになる。
func Identity(_ context.Context, s State) (State, error) { return s, nil }

// Then は射を合成する。context のキャンセルは合成器の側で一度だけ見る。
func Then(f, g Endo) Endo {
	return func(ctx context.Context, s State) (State, error) {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		next, err := f(ctx, s)
		if err != nil {
			return next, err
		}
		return g(ctx, next)
	}
}

// Sequence は射の列を1本にまとめる。空なら Identity になる。
func Sequence(steps ...Endo) Endo {
	out := Endo(Identity)
	for _, step := range steps {
		out = Then(out, step)
	}
	return out
}

// --- 射を包む変換（デコレータ） ---
// どれも Endo → Endo なので、対象を変えずに射だけを取り替えている。

// WithRetry は一時的な失敗のときだけ呼び直す。
func WithRetry(s Endo, attempts int) Endo {
	return func(ctx context.Context, st State) (State, error) {
		var err error
		var next State
		for range attempts {
			next, err = s(ctx, st)
			if err == nil || !errors.Is(err, ErrTemporary) {
				return next, err
			}
		}
		return next, err
	}
}

// WithTiming は所要時間を記録する。
func WithTiming(s Endo, name string, rec *Recorder) Endo {
	return func(ctx context.Context, st State) (State, error) {
		start := time.Now()
		next, err := s(ctx, st)
		rec.Record(name, time.Since(start))
		return next, err
	}
}

// WithTimeout は制限時間を付ける。
func WithTimeout(s Endo, d time.Duration) Endo {
	return func(ctx context.Context, st State) (State, error) {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return s(ctx, st)
	}
}

// WithLabel は失敗したときにどのステップかを添える。
func WithLabel(s Endo, name string) Endo {
	return func(ctx context.Context, st State) (State, error) {
		next, err := s(ctx, st)
		if err != nil {
			return next, fmt.Errorf("%s: %w", name, err)
		}
		return next, nil
	}
}

// --- 業務のステップ本体 ---
// 横断的な関心事（リトライ・計測・ラベル）が一切入っていないことに注目。

func Reserve(d Deps) Endo {
	return func(ctx context.Context, s State) (State, error) {
		id, err := d.Inventory.Reserve(ctx, s.Order)
		if err != nil {
			return s, err
		}
		s.ReservationID = id
		return s, nil
	}
}

func Charge(d Deps) Endo {
	return func(ctx context.Context, s State) (State, error) {
		id, err := d.Payments.Charge(ctx, s.Order)
		if err != nil {
			return s, err
		}
		s.PaymentID = id
		return s, nil
	}
}

func Arrange(d Deps) Endo {
	return func(ctx context.Context, s State) (State, error) {
		no, err := d.Shipping.Arrange(ctx, s.Order, s.ReservationID)
		if err != nil {
			return s, err
		}
		s.TrackingNo = no
		return s, nil
	}
}

func Notify(d Deps) Endo {
	return func(ctx context.Context, s State) (State, error) {
		if err := d.Notifier.Notify(ctx, s.Order.UserID, s.TrackingNo); err != nil {
			return s, err
		}
		s.Notified = true
		return s, nil
	}
}
