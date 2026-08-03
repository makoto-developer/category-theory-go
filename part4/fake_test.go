package part4

import "context"

// テスト用の外部サービス。tempFailures 回だけ一時的に失敗し、その後成功する。
// permanent が非 nil なら常にそれを返す。
type fakeService struct {
	calls        int
	tempFailures int
	permanent    error
}

func (f *fakeService) next(value string) (string, error) {
	f.calls++
	if f.permanent != nil {
		return "", f.permanent
	}
	if f.calls <= f.tempFailures {
		return "", ErrTemporary
	}
	return value, nil
}

type fakeInventory struct{ fakeService }
type fakePayments struct{ fakeService }
type fakeShipping struct{ fakeService }
type fakeNotifier struct{ fakeService }

func (f *fakeInventory) Reserve(_ context.Context, _ Order) (string, error) {
	return f.next("RSV-1")
}

func (f *fakePayments) Charge(_ context.Context, _ Order) (string, error) {
	return f.next("PAY-1")
}

func (f *fakeShipping) Arrange(_ context.Context, _ Order, _ string) (string, error) {
	return f.next("TRK-1")
}

func (f *fakeNotifier) Notify(_ context.Context, _, _ string) error {
	_, err := f.next("")
	return err
}

// newDeps は4つのフェイクを束ねて返す。個別に設定を変えたいときは戻り値をいじる。
func newDeps() (Deps, *fakeInventory, *fakePayments, *fakeShipping, *fakeNotifier) {
	inv, pay, ship, notif := &fakeInventory{}, &fakePayments{}, &fakeShipping{}, &fakeNotifier{}
	return Deps{Inventory: inv, Payments: pay, Shipping: ship, Notifier: notif}, inv, pay, ship, notif
}

var sampleOrder = Order{ID: "ORD-1", UserID: "U-1", Amount: 1200}
