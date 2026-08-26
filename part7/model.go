package part7

// 測る対象の入れ子。Tags と Meta が参照型なのがこの回の主題。
// struct を値で代入すると Tags のヘッダはコピーされるが、
// 指している配列は共有されたままになる。

type Account struct {
	ID      int64
	Profile Profile
}

type Profile struct {
	Name    string
	Contact Contact
	Tags    []string
	Meta    map[string]string
}

type Contact struct {
	Email string
	Phone string
}

// --- 非干渉を守るレンズ -------------------------------------------------
//
// 「レンズ則を守る」ではない。下のエイリアスする版も3則は満たす。
// ここで複製しているのは、Get の戻り値や Set に渡した値を書き換えても
// 元が変わらないという別の性質（非干渉）のため。

func ProfileL() Lens[Account, Profile] {
	return Lens[Account, Profile]{
		Get: func(a Account) Profile { return a.Profile },
		// Profile を差し替えるので、中の参照型もこの時点で複製しておく。
		Set: func(a Account, p Profile) Account { a.Profile = cloneProfile(p); return a },
	}
}

func ContactL() Lens[Profile, Contact] {
	return Lens[Profile, Contact]{
		Get: func(p Profile) Contact { return p.Contact },
		Set: func(p Profile, c Contact) Profile { p.Contact = c; return p },
	}
}

func EmailL() Lens[Contact, string] {
	return Lens[Contact, string]{
		Get: func(c Contact) string { return c.Email },
		Set: func(c Contact, e string) Contact { c.Email = e; return c },
	}
}

// TagsL は非干渉を守る版。get で複製を返し、set でも複製を格納する。
// どちらか片方でも共有したままにすると、呼び出し側の書き換えが元に漏れる。
func TagsL() Lens[Profile, []string] {
	return Lens[Profile, []string]{
		Get: func(p Profile) []string { return cloneSlice(p.Tags) },
		Set: func(p Profile, ts []string) Profile { p.Tags = cloneSlice(ts); return p },
	}
}

// --- 非干渉を破るレンズ（素直に書くとこうなる。3則は満たす） -------------

// TagsAliasing は「普通に書いた」版。struct はコピーされるので一見正しく見える。
// スライスのヘッダしかコピーされないため、指している配列は元と共有される。
func TagsAliasing() Lens[Profile, []string] {
	return Lens[Profile, []string]{
		Get: func(p Profile) []string { return p.Tags },
		Set: func(p Profile, ts []string) Profile { p.Tags = ts; return p },
	}
}

// ProfileAliasing も同じ。Profile を値で代入するだけ。
func ProfileAliasing() Lens[Account, Profile] {
	return Lens[Account, Profile]{
		Get: func(a Account) Profile { return a.Profile },
		Set: func(a Account, p Profile) Account { a.Profile = p; return a },
	}
}

// --- 複製 ---------------------------------------------------------------

func cloneSlice(xs []string) []string {
	if xs == nil {
		return nil
	}
	out := make([]string, len(xs))
	copy(out, xs)
	return out
}

func cloneMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneProfile(p Profile) Profile {
	p.Tags = cloneSlice(p.Tags)
	p.Meta = cloneMap(p.Meta)
	return p
}
