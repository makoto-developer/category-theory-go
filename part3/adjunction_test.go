package part3

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// ServerOption の全体はモノイド。まとめ方を変えても結果は同じ（結合律）。
func TestOptionsAreMonoid(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		host := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "host")
		port := rapid.IntRange(1, 65535).Draw(t, "port")

		flat := NewServer(WithHost(host), WithPort(port), WithTimeout(time.Second))
		grouped := NewServer(ApplyOptions(WithHost(host), WithPort(port)), WithTimeout(time.Second))
		nested := NewServer(WithHost(host), ApplyOptions(WithPort(port), WithTimeout(time.Second)))

		if flat != grouped || flat != nested {
			t.Fatalf("結合律が破れた: 平坦=%+v, 前2つ=%+v, 後2つ=%+v", flat, grouped, nested)
		}
	})
}

// ServerOption を1つも渡さなければ既定値になる（単位元）。
func TestNoOptionsGivesDefaults(t *testing.T) {
	want := ServerConfig{Host: "localhost", Port: 8080, Timeout: 30 * time.Second}

	if got := NewServer(); got != want {
		t.Fatalf("既定値が違う: got=%+v, want=%+v", got, want)
	}
}

// 同じ項目を2回設定すると後勝ち。可換ではないので順序には意味がある。
func TestOptionsAreNotCommutative(t *testing.T) {
	forward := NewServer(WithPort(1), WithPort(2))
	reversed := NewServer(WithPort(2), WithPort(1))

	if forward == reversed {
		t.Fatal("順序を入れ替えても同じになった（可換になっている）")
	}
	if forward.Port != 2 || reversed.Port != 1 {
		t.Fatalf("後勝ちになっていない: forward=%d, reversed=%d", forward.Port, reversed.Port)
	}
}

// カリー化すると、2引数の設定関数がそのまま ServerOption になる。
func TestCurryOptionMatchesDirectSetter(t *testing.T) {
	withHost := CurryOption(SetHost)

	rapid.Check(t, func(t *rapid.T) {
		host := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "host")

		curried := NewServer(withHost(host))
		direct := SetHost(NewServer(), host)

		if curried != direct {
			t.Fatalf("カリー化した設定と直接の設定が食い違う: %+v vs %+v", curried, direct)
		}
	})
}
