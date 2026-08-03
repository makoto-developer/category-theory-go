package part1

import (
	"errors"
	"testing"

	"pgregory.net/rapid"
)

// 純粋な射なら、メモ化しても元の射と区別がつかない。
func TestMemoizeIsTransparentForPureMorphism(t *testing.T) {
	memoized := Memoize(itoa)

	rapid.Check(t, func(t *rapid.T) {
		x := rapid.Int().Draw(t, "x")

		if memoized(x) != itoa(x) {
			t.Fatalf("メモ化で結果が変わった: x=%d, memo=%q, 元=%q", x, memoized(x), itoa(x))
		}
		if memoized(x) != memoized(x) {
			t.Fatalf("メモ化した射が同じ入力に違う結果を返した: x=%d", x)
		}
	})
}

// 純粋でない射をメモ化すると、2回目から結果が食い違う。
// 「純粋関数にしておけ」という経験則が、圏の言葉では何を守っているのかがここに出る。
func TestMemoizeChangesResultForImpureMorphism(t *testing.T) {
	impure := Counter()
	memoized := Memoize(impure)

	first := memoized(10)
	second := memoized(10)
	direct := impure(10)

	if first != second {
		t.Fatalf("メモ化した射が同じ値を返していない: 1回目=%d, 2回目=%d", first, second)
	}
	if direct == first {
		t.Fatalf("メモ化の有無で結果が一致してしまった（Counter が純粋になっている）: %d", direct)
	}
	t.Logf("メモ化: %d, %d / メモ化なし: %d", first, second, direct)
}

// 純粋でない射をリトライすると副作用が二重に走る。
func TestRetryRunsEffectTwiceForImpureMorphism(t *testing.T) {
	calls := 0
	flaky := func(x int) (int, error) {
		calls++
		if calls == 1 {
			return 0, errors.New("一時的な失敗")
		}
		return x, nil
	}

	got, err := Retry(flaky, 3)(7)

	if err != nil {
		t.Fatalf("リトライで成功しなかった: %v", err)
	}
	if got != 7 {
		t.Fatalf("結果が違う: got=%d, want=7", got)
	}
	if calls != 2 {
		t.Fatalf("呼び出し回数が想定と違う: got=%d, want=2", calls)
	}
}
