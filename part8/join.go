package part8

// 引き戻しの補題（pullback lemma）を実務の言葉に直すと、結合の順序を
// 入れ替えても同じ答えになる、という保証になる。保証があるのは「同じ答え」までで、
// どの順で計算するかは圏論の外にある。ここはその幅を測るための題材。

type User struct {
	ID   int
	Dept int
}

type Assignment struct {
	Dept    int
	Project int
}

type ActiveProject struct {
	Project int
}

type Triple struct {
	U User
	A Assignment
	P ActiveProject
}

// JoinLeftFirst は (User ⋈dept Assignment) ⋈project ActiveProject の順で計算する。
func JoinLeftFirst(us []User, as []Assignment, ps []ActiveProject) []Triple {
	ua := PullbackHash(us, as,
		func(u User) int { return u.Dept },
		func(a Assignment) int { return a.Dept })

	uap := PullbackHash(ua, ps,
		func(p Pair[User, Assignment]) int { return p.R.Project },
		func(p ActiveProject) int { return p.Project })

	out := make([]Triple, 0, len(uap))
	for _, t := range uap {
		out = append(out, Triple{U: t.L.L, A: t.L.R, P: t.R})
	}
	return out
}

// JoinRightFirst は User ⋈dept (Assignment ⋈project ActiveProject) の順で計算する。
// 答えは上と同じ。作られる中間データの大きさだけが違う。
func JoinRightFirst(us []User, as []Assignment, ps []ActiveProject) []Triple {
	ap := PullbackHash(as, ps,
		func(a Assignment) int { return a.Project },
		func(p ActiveProject) int { return p.Project })

	uap := PullbackHash(us, ap,
		func(u User) int { return u.Dept },
		func(p Pair[Assignment, ActiveProject]) int { return p.L.Dept })

	out := make([]Triple, 0, len(uap))
	for _, t := range uap {
		out = append(out, Triple{U: t.L, A: t.R.L, P: t.R.R})
	}
	return out
}

// IntermediateSizes は、それぞれの順で作られる中間データの要素数を返す。
// 時間差の原因が中間データの大きさにあることを、数で示すため。
func IntermediateSizes(us []User, as []Assignment, ps []ActiveProject) (leftFirst, rightFirst int) {
	ua := PullbackHash(us, as,
		func(u User) int { return u.Dept },
		func(a Assignment) int { return a.Dept })
	ap := PullbackHash(as, ps,
		func(a Assignment) int { return a.Project },
		func(p ActiveProject) int { return p.Project })
	return len(ua), len(ap)
}
