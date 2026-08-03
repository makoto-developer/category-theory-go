package part4

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// NaiveWorkflow は素直な手続き型の実装。各ステップにリトライと計測をその場で書く。
type NaiveWorkflow struct {
	Deps     Deps
	Recorder *Recorder
	Attempts int
}

// Run は在庫確保 → 決済 → 配送手配 → 通知を順に実行する。
// リトライ・計測・キャンセル確認が各ステップに散らばるのが、この書き方の特徴。
func (w NaiveWorkflow) Run(ctx context.Context, s State) (State, error) {
	if err := ctx.Err(); err != nil {
		return s, err
	}

	start := time.Now()
	var reservationID string
	var err error
	for i := range w.Attempts {
		reservationID, err = w.Deps.Inventory.Reserve(ctx, s.Order)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrTemporary) || i == w.Attempts-1 {
			w.Recorder.Record("reserve", time.Since(start))
			return s, fmt.Errorf("在庫確保: %w", err)
		}
	}
	w.Recorder.Record("reserve", time.Since(start))
	s.ReservationID = reservationID

	if err := ctx.Err(); err != nil {
		return s, err
	}

	start = time.Now()
	var paymentID string
	for i := range w.Attempts {
		paymentID, err = w.Deps.Payments.Charge(ctx, s.Order)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrTemporary) || i == w.Attempts-1 {
			w.Recorder.Record("charge", time.Since(start))
			return s, fmt.Errorf("決済: %w", err)
		}
	}
	w.Recorder.Record("charge", time.Since(start))
	s.PaymentID = paymentID

	if err := ctx.Err(); err != nil {
		return s, err
	}

	start = time.Now()
	var trackingNo string
	for i := range w.Attempts {
		trackingNo, err = w.Deps.Shipping.Arrange(ctx, s.Order, s.ReservationID)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrTemporary) || i == w.Attempts-1 {
			w.Recorder.Record("arrange", time.Since(start))
			return s, fmt.Errorf("配送手配: %w", err)
		}
	}
	w.Recorder.Record("arrange", time.Since(start))
	s.TrackingNo = trackingNo

	if err := ctx.Err(); err != nil {
		return s, err
	}

	start = time.Now()
	for i := range w.Attempts {
		err = w.Deps.Notifier.Notify(ctx, s.Order.UserID, s.TrackingNo)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrTemporary) || i == w.Attempts-1 {
			w.Recorder.Record("notify", time.Since(start))
			return s, fmt.Errorf("通知: %w", err)
		}
	}
	w.Recorder.Record("notify", time.Since(start))
	s.Notified = true

	return s, nil
}
