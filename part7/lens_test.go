package part7

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

// レンズ則は3つ。get-set（取って戻せば元）、set-get（入れたものが取れる）、
// set-set（二度入れたら後勝ち）。3つとも満たすものだけがレンズを名乗れる。
func checkLensLaws[S, A any](t *rapid.T, l Lens[S, A], s S, a, b A, eq func(S, S) bool, eqA func(A, A) bool) {
	if got := l.Set(s, l.Get(s)); !eq(got, s) {
		t.Fatalf("get-set 則が破れた")
	}
	if got := l.Get(l.Set(s, a)); !eqA(got, a) {
		t.Fatalf("set-get 則が破れた")
	}
	if !eq(l.Set(l.Set(s, a), b), l.Set(s, b)) {
		t.Fatalf("set-set 則が破れた")
	}
}

func genAccount() *rapid.Generator[Account] {
	return rapid.Custom(func(t *rapid.T) Account {
		return Account{
			ID: rapid.Int64().Draw(t, "id"),
			Profile: Profile{
				Name:    rapid.String().Draw(t, "name"),
				Contact: Contact{Email: rapid.String().Draw(t, "email"), Phone: rapid.String().Draw(t, "phone")},
				Tags:    rapid.SliceOfN(rapid.String(), 0, 5).Draw(t, "tags"),
				Meta:    rapid.MapOfN(rapid.String(), rapid.String(), 0, 3).Draw(t, "meta"),
			},
		}
	})
}

func eqProfile(a, b Profile) bool {
	if a.Name != b.Name || a.Contact != b.Contact || len(a.Meta) != len(b.Meta) {
		return false
	}
	if !slices.Equal(a.Tags, b.Tags) {
		return false
	}
	for k, v := range a.Meta {
		if b.Meta[k] != v {
			return false
		}
	}
	return true
}

func eqAccount(a, b Account) bool { return a.ID == b.ID && eqProfile(a.Profile, b.Profile) }

// 法則を守る版は3則を満たす。
func TestLawfulLensesObeyTheLaws(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		acc := genAccount().Draw(t, "acc")
		e1 := rapid.String().Draw(t, "e1")
		e2 := rapid.String().Draw(t, "e2")

		email := Compose(Compose(ProfileL(), ContactL()), EmailL())
		checkLensLaws(t, email, acc, e1, e2, eqAccount, func(a, b string) bool { return a == b })

		t1 := rapid.SliceOfN(rapid.String(), 0, 4).Draw(t, "t1")
		t2 := rapid.SliceOfN(rapid.String(), 0, 4).Draw(t, "t2")
		tags := Compose(ProfileL(), TagsL())
		checkLensLaws(t, tags, acc, t1, t2, eqAccount, slices.Equal)
	})
}

// 合成は結合的で、Identity が単位射になる（圏の公理）。
func TestLensCompositionIsAssociativeAndUnital(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		acc := genAccount().Draw(t, "acc")
		e := rapid.String().Draw(t, "e")

		left := Compose(Compose(ProfileL(), ContactL()), EmailL())
		right := Compose(ProfileL(), Compose(ContactL(), EmailL()))
		if left.Get(acc) != right.Get(acc) || !eqAccount(left.Set(acc, e), right.Set(acc, e)) {
			t.Fatalf("合成が結合的でない")
		}

		withID := Compose(Identity[Account](), left)
		if withID.Get(acc) != left.Get(acc) || !eqAccount(withID.Set(acc, e), left.Set(acc, e)) {
			t.Fatalf("Identity が単位射になっていない")
		}
	})
}
