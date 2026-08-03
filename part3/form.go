package part3

import (
	"errors"
	"strings"
)

// Registration は登録フォームの入力。
type Registration struct {
	Name  string
	Email string
	Age   int
}

var (
	ErrNameEmpty   = errors.New("名前が空です")
	ErrNameTooLong = errors.New("名前は32文字までです")
	ErrEmailNoAt   = errors.New("メールアドレスに @ が含まれていません")
	ErrEmailEmpty  = errors.New("メールアドレスが空です")
	ErrAgeNegative = errors.New("年齢が負の数です")
	ErrAgeTooLarge = errors.New("年齢が150を超えています")
)

func ValidateName(s string) Validated[string] {
	switch {
	case s == "":
		return Invalid[string](ErrNameEmpty)
	case len(s) > 32:
		return Invalid[string](ErrNameTooLong)
	}
	return Valid(s)
}

func ValidateEmail(s string) Validated[string] {
	switch {
	case s == "":
		return Invalid[string](ErrEmailEmpty)
	case !strings.Contains(s, "@"):
		return Invalid[string](ErrEmailNoAt)
	}
	return Valid(s)
}

func ValidateAge(n int) Validated[int] {
	switch {
	case n < 0:
		return Invalid[int](ErrAgeNegative)
	case n > 150:
		return Invalid[int](ErrAgeTooLarge)
	}
	return Valid(n)
}

// ValidateApplicative は3項目を独立に検証してから合わせる。
// 各検証が互いに依存しないので、失敗をすべて集められる。
func ValidateApplicative(name, email string, age int) Validated[Registration] {
	return Combine3(
		ValidateName(name),
		ValidateEmail(email),
		ValidateAge(age),
		func(n, e string, a int) Registration {
			return Registration{Name: n, Email: e, Age: a}
		},
	)
}

// ValidateMonadic は前段が成功したときだけ次に進む。
// 後段が前段の値を使えるかわりに、最初の失敗で打ち切られる。
func ValidateMonadic(name, email string, age int) Validated[Registration] {
	return AndThen(ValidateName(name), func(n string) Validated[Registration] {
		return AndThen(ValidateEmail(email), func(e string) Validated[Registration] {
			return MapV(ValidateAge(age), func(a int) Registration {
				return Registration{Name: n, Email: e, Age: a}
			})
		})
	})
}
