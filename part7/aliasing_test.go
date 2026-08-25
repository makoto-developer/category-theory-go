package part7

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

// 素直に書いたレンズ（スライスを共有する版）が、3則を全部通ってしまうことを示す。
// 通るなら、レンズ則は Go のエイリアスを検出できないということになる。
func TestAliasingLensStillObeysAllThreeLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		acc := genAccount().Draw(t, "acc")
		t1 := rapid.SliceOfN(rapid.String(), 0, 4).Draw(t, "t1")
		t2 := rapid.SliceOfN(rapid.String(), 0, 4).Draw(t, "t2")

		aliasing := Compose(ProfileAliasing(), TagsAliasing())
		checkLensLaws(t, aliasing, acc, t1, t2, eqAccount, slices.Equal)
	})
}

// 4つ目の性質: Get で取り出したものを書き換えても、元の S は変わらない。
// レンズ則3つには含まれていないが、「非破壊である」と言うために要る。
func checkIndependence(t *rapid.T, l Lens[Account, []string], acc Account) {
	before := slices.Clone(acc.Profile.Tags)
	got := l.Get(acc)
	if len(got) == 0 {
		return
	}
	got[0] = got[0] + "___書き換えた"
	if !slices.Equal(acc.Profile.Tags, before) {
		t.Fatalf("Get の戻り値を書き換えたら、元の Account が変わった: %v → %v", before, acc.Profile.Tags)
	}
}

// 法則を守る版は独立性も満たす。
func TestLawfulLensIsIndependent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		acc := genAccount().Draw(t, "acc")
		checkIndependence(t, Compose(ProfileL(), TagsL()), acc)
	})
}

// エイリアスする版は独立性を破る。3則は通るのに、こちらは通らない。
func TestAliasingLensBreaksIndependence(t *testing.T) {
	acc := Account{ID: 1, Profile: Profile{Tags: []string{"a", "b"}}}
	before := slices.Clone(acc.Profile.Tags)

	aliasing := Compose(ProfileAliasing(), TagsAliasing())
	got := aliasing.Get(acc)
	got[0] = "書き換えた"

	if slices.Equal(acc.Profile.Tags, before) {
		t.Fatalf("エイリアスする版なのに元が守られてしまった（テストの前提が壊れている）")
	}
	t.Logf("3則を通ったレンズで、Get の戻り値を書き換えたら元が変わった: %v → %v", before, acc.Profile.Tags)
}

// Set 側でも同じことが起きる。渡したスライスを後から書き換えると、格納先が変わる。
func TestAliasingLensLeaksThroughSet(t *testing.T) {
	acc := Account{ID: 1, Profile: Profile{Tags: []string{"a"}}}
	incoming := []string{"x", "y"}

	aliasing := Compose(ProfileAliasing(), TagsAliasing())
	updated := aliasing.Set(acc, incoming)
	incoming[0] = "呼び出し側で書き換えた"

	if updated.Profile.Tags[0] != "呼び出し側で書き換えた" {
		t.Fatalf("Set で共有されていない（テストの前提が壊れている）")
	}

	lawful := Compose(ProfileL(), TagsL())
	incoming2 := []string{"x", "y"}
	updated2 := lawful.Set(acc, incoming2)
	incoming2[0] = "呼び出し側で書き換えた"
	if updated2.Profile.Tags[0] != "x" {
		t.Fatalf("法則を守る版なのに Set から漏れた: %v", updated2.Profile.Tags)
	}
}

// Modify も同じ罠を踏む。f がスライスを in-place で触ると元に漏れる。
func TestModifyLeaksWhenTheFunctionMutates(t *testing.T) {
	acc := Account{ID: 1, Profile: Profile{Tags: []string{"a", "b"}}}
	aliasing := Compose(ProfileAliasing(), TagsAliasing())

	// append は容量が足りていれば元の配列を書き換える。
	_ = Modify(aliasing, acc, func(ts []string) []string {
		ts = append(ts[:1], "差し込んだ")
		return ts
	})

	if acc.Profile.Tags[1] != "差し込んだ" {
		t.Skipf("この容量では in-place にならなかった: %v", acc.Profile.Tags)
	}
	t.Logf("Modify に渡した関数の append が、元の Account を書き換えた: %v", acc.Profile.Tags)
}
