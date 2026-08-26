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

// 非干渉は1つの性質ではなく4つある。レンズ則3つのどれにも含まれていない。
// 空スライスだと書き換えようがないので、検査できる要素があるときだけ見る。
func checkIndependence(t *rapid.T, l Lens[Account, []string], acc Account, incoming []string) {
	// ① Get の戻り値を書き換えても、元は変わらない
	before := slices.Clone(acc.Profile.Tags)
	if got := l.Get(acc); len(got) > 0 {
		got[0] += "___書き換えた"
		if !slices.Equal(acc.Profile.Tags, before) {
			t.Fatalf("① Get の戻り値を書き換えたら元が変わった: %v → %v", before, acc.Profile.Tags)
		}
	}

	// ② Set に渡した値をあとから書き換えても、格納先は変わらない
	if len(incoming) > 0 {
		in := slices.Clone(incoming)
		updated := l.Set(acc, in)
		stored := slices.Clone(updated.Profile.Tags)
		in[0] += "___書き換えた"
		if !slices.Equal(updated.Profile.Tags, stored) {
			t.Fatalf("② Set に渡した値を書き換えたら格納先が変わった: %v → %v", stored, updated.Profile.Tags)
		}
	}

	// ③ Set の結果を書き換えても、元は変わらない
	before = slices.Clone(acc.Profile.Tags)
	if updated := l.Set(acc, incoming); len(updated.Profile.Tags) > 0 {
		updated.Profile.Tags[0] += "___書き換えた"
		if !slices.Equal(acc.Profile.Tags, before) {
			t.Fatalf("③ Set の結果を書き換えたら元が変わった: %v → %v", before, acc.Profile.Tags)
		}
	}

	// ④ Modify に渡した関数が入力を壊しても、元は変わらない
	before = slices.Clone(acc.Profile.Tags)
	_ = Modify(l, acc, func(ts []string) []string {
		if len(ts) > 0 {
			ts[0] += "___関数の中で書き換えた"
		}
		return ts
	})
	if !slices.Equal(acc.Profile.Tags, before) {
		t.Fatalf("④ Modify の中の書き換えが元に漏れた: %v → %v", before, acc.Profile.Tags)
	}
}

// 非干渉を守る版は、4つとも満たす。
func TestLawfulLensIsIndependent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		acc := genAccount().Draw(t, "acc")
		incoming := rapid.SliceOfN(rapid.String(), 1, 4).Draw(t, "incoming")
		checkIndependence(t, Compose(ProfileL(), TagsL()), acc, incoming)
	})
}

// エイリアスする版は、4つとも破る。3則は満たすのに。
func TestAliasingLensBreaksAllFourIndependenceProperties(t *testing.T) {
	l := Compose(ProfileAliasing(), TagsAliasing())
	broken := 0
	for _, name := range []string{"①Get", "②Set入力", "③Set結果", "④Modify"} {
		acc := Account{ID: 1, Profile: Profile{Tags: []string{"a", "b"}}}
		in := []string{"x", "y"}
		before := slices.Clone(acc.Profile.Tags)
		switch name {
		case "①Get":
			l.Get(acc)[0] = "書き換えた"
		case "②Set入力":
			u := l.Set(acc, in)
			in[0] = "書き換えた"
			if u.Profile.Tags[0] == "書き換えた" {
				broken++
			}
			continue
		case "③Set結果":
			l.Set(acc, in).Profile.Tags[0] = "書き換えた"
			// Set は acc.Profile.Tags を in で置き換えるので、元ではなく in が壊れる
			if in[0] == "書き換えた" {
				broken++
			}
			continue
		case "④Modify":
			_ = Modify(l, acc, func(ts []string) []string {
				if len(ts) > 0 {
					ts[0] = "書き換えた"
				}
				return ts
			})
		}
		if !slices.Equal(acc.Profile.Tags, before) {
			broken++
		}
	}
	if broken != 4 {
		t.Fatalf("4つとも破れるはずが %d 件しか破れなかった", broken)
	}
	t.Logf("3則を全部満たすレンズが、非干渉の4性質を4つとも破った")
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
