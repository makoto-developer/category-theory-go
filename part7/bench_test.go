package part7

import (
	"fmt"
	"strconv"
	"testing"
)

var sinkAcc Account

func makeAccount(nTags, nMeta int) Account {
	tags := make([]string, nTags)
	for i := range tags {
		tags[i] = "tag" + strconv.Itoa(i)
	}
	meta := make(map[string]string, nMeta)
	for i := range nMeta {
		meta["k"+strconv.Itoa(i)] = "v" + strconv.Itoa(i)
	}
	return Account{ID: 1, Profile: Profile{
		Name:    "n",
		Contact: Contact{Email: "a@example.com", Phone: "000"},
		Tags:    tags,
		Meta:    meta,
	}}
}

// 参照型を通らない経路。レンズ合成そのものの代金を見る。
func BenchmarkUpdateEmail(b *testing.B) {
	acc := makeAccount(8, 8)
	// レンズはループの外で組む。中で組むと合成の費用が毎回混ざる。
	email := Compose(Compose(ProfileAliasing(), ContactL()), EmailL())

	b.Run("1_lens", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkAcc = email.Set(acc, "b@example.com")
		}
	})
	b.Run("2_hand", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			a := acc
			a.Profile.Contact.Email = "b@example.com"
			sinkAcc = a
		}
	})
}

// 参照型を通る経路。法則を守るための複製がいくらかかるかを見る。
func BenchmarkUpdateTags(b *testing.B) {
	for _, n := range []int{0, 8, 128, 4096} {
		acc := makeAccount(n, 8)
		next := make([]string, n)
		copy(next, acc.Profile.Tags)

		lawful := Compose(ProfileL(), TagsL())
		aliasing := Compose(ProfileAliasing(), TagsAliasing())

		b.Run(fmt.Sprintf("tags=%04d/1_lawful_lens", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkAcc = lawful.Set(acc, next)
			}
		})
		b.Run(fmt.Sprintf("tags=%04d/2_aliasing_lens", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkAcc = aliasing.Set(acc, next)
			}
		})
		// 手書きで、法則を守る形。レンズと同じ仕事をする。
		b.Run(fmt.Sprintf("tags=%04d/3_hand_copying", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a := acc
				p := a.Profile
				ts := make([]string, len(next))
				copy(ts, next)
				p.Tags = ts
				m := make(map[string]string, len(p.Meta))
				for k, v := range p.Meta {
					m[k] = v
				}
				p.Meta = m
				a.Profile = p
				sinkAcc = a
			}
		})
		// 手書きで、素直に書いた形。共有したまま。
		b.Run(fmt.Sprintf("tags=%04d/4_hand_aliasing", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				a := acc
				a.Profile.Tags = next
				sinkAcc = a
			}
		})
	}
}
