// Package part3 は連載「Goで書く実践圏論」第3回の検証コード。
// Applicative と Monad の差、F代数による fold、iter.Seq と継続渡しの関係、
// そして高階カインド多相が無い Go でどう凌ぐかを扱う。
package part3

import "errors"

// Validated は「値、または失敗の集まり」。失敗を蓄積できるのが (T, error) との違い。
type Validated[T any] struct {
	Value  T
	Errors []error
}

// Valid は成功した検証結果。
func Valid[T any](v T) Validated[T] { return Validated[T]{Value: v} }

// Invalid は失敗した検証結果。
func Invalid[T any](errs ...error) Validated[T] { return Validated[T]{Errors: errs} }

// OK は失敗が無いかを返す。
func (v Validated[T]) OK() bool { return len(v.Errors) == 0 }

// Err は蓄積した失敗を1つの error にまとめる。errors.Join がモノイドなのでこう書ける。
func (v Validated[T]) Err() error { return errors.Join(v.Errors...) }

// MapV は Validated 上の Functor。失敗しているときは射を適用しない。
func MapV[A, B any](v Validated[A], f func(A) B) Validated[B] {
	if !v.OK() {
		return Validated[B]{Errors: v.Errors}
	}
	return Valid(f(v.Value))
}

// Combine2 は2つの検証結果を合わせる。Applicative の要。
// 片方が失敗していても、もう片方の検証は済んでいるので、両方の失敗を集められる。
func Combine2[A, B, C any](a Validated[A], b Validated[B], f func(A, B) C) Validated[C] {
	if !a.OK() || !b.OK() {
		return Validated[C]{Errors: append(append([]error{}, a.Errors...), b.Errors...)}
	}
	return Valid(f(a.Value, b.Value))
}

// Combine3 は3つ版。Go には可変長の型パラメータが無いので、数だけ書くことになる。
func Combine3[A, B, C, D any](a Validated[A], b Validated[B], c Validated[C], f func(A, B, C) D) Validated[D] {
	if !a.OK() || !b.OK() || !c.OK() {
		errs := append([]error{}, a.Errors...)
		errs = append(errs, b.Errors...)
		errs = append(errs, c.Errors...)
		return Validated[D]{Errors: errs}
	}
	return Valid(f(a.Value, b.Value, c.Value))
}

// AndThen は Monad の bind。後段が前段の値に依存できる代わりに、
// 前段が失敗した時点で後段は走らないため、失敗はひとつしか集まらない。
func AndThen[A, B any](v Validated[A], f func(A) Validated[B]) Validated[B] {
	if !v.OK() {
		return Validated[B]{Errors: v.Errors}
	}
	return f(v.Value)
}
