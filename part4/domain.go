// Package part4 は連載「Goで書く実践圏論」第4回の検証コード。
// 同じ注文処理ワークフローを「素直な手続き型」と「射の合成」の2通りで書き、
// 行数・アロケーション・レイテンシ・テストの手間を比べる。
package part4

import (
	"context"
	"errors"
)

// Order は処理対象の注文。金額は最小単位の整数で持つ（float64 は結合的でないため）。
type Order struct {
	ID     string
	UserID string
	Amount int
}

// State はワークフローの途中経過。各ステップはこれを受け取って更新した State を返す。
type State struct {
	Order         Order
	ReservationID string
	PaymentID     string
	TrackingNo    string
	Notified      bool
}

var (
	ErrOutOfStock   = errors.New("在庫が足りない")
	ErrPaymentDenid = errors.New("決済が拒否された")
	ErrTemporary    = errors.New("一時的な失敗")
)

// Inventory 以下は外部サービス。テストでは差し替える。
type Inventory interface {
	Reserve(ctx context.Context, o Order) (string, error)
}

type Payments interface {
	Charge(ctx context.Context, o Order) (string, error)
}

type Shipping interface {
	Arrange(ctx context.Context, o Order, reservationID string) (string, error)
}

type Notifier interface {
	Notify(ctx context.Context, userID, trackingNo string) error
}

// Deps はワークフローが必要とする依存一式。
type Deps struct {
	Inventory Inventory
	Payments  Payments
	Shipping  Shipping
	Notifier  Notifier
}
