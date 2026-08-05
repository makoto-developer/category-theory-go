// 連載「Goで書く実践圏論」第4回のスニペット。
//
// 注文処理を「射のリスト」として定義し、同じ定義から
// 実行・dry-run・Mermaid図を導きます。連載の到達点をここに縮めました。
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// 案内は先頭に出す。記事に埋め込むと末尾しか見えないので、末尾は検証結果のために空けておく。
func main() {
	fmt.Println("連載「Goで書く実践圏論」第4回")
	fmt.Println("  記事: https://blog.makoto-developer.net/articles/2026-08-07-practical-category-theory-go-4")
	fmt.Println("  全4回のコード（テスト・ベンチ込み）: https://github.com/makoto-developer/category-theory-go")
	fmt.Println("──────────────────────────────────────────")

	plan := orderPlan()

	showRun(plan)
	showDryRunTouchesNothing(plan)
	showDecoratorOrderDecidesTheResult()
	showMermaid(plan)
}

// ---- 定義 --------------------------------------------------------------

// State は工程を通り抜けていく値。全ステップが State → State なので合成できる。
type State struct {
	OrderID string
	Trail   []string
}

// Step は自己射。合成すると Step になる、が全部。
type Step func(State) (State, error)

// Node は「ステップ1つぶんの設定」。これはデータであって関数ではない。
type Node struct {
	Name    string
	Step    Step
	Retries int
	Timeout time.Duration
}

// Plan は Node のリスト。これがワークフローの定義そのもの。
type Plan struct {
	Title string
	Nodes []Node
}

// externalCalls は「外部サービスを何回叩いたか」。dry-run が本当に何も触らないことの証拠に使う。
var externalCalls int

func callExternal(name string) Step {
	return func(s State) (State, error) {
		externalCalls++
		s.Trail = append(s.Trail, name)
		return s, nil
	}
}

// これがワークフローの全定義。8行しかない。
func orderPlan() Plan {
	return Plan{
		Title: "注文処理",
		Nodes: []Node{
			{Name: "在庫確保", Step: callExternal("在庫確保"), Retries: 3, Timeout: 2 * time.Second},
			{Name: "決済", Step: callExternal("決済"), Timeout: 5 * time.Second},
			{Name: "配送手配", Step: callExternal("配送手配"), Retries: 2},
			{Name: "通知", Step: callExternal("通知")},
		},
	}
}

// ---- 解釈① 実行 -------------------------------------------------------

// Compile は Plan を1本の射に畳む。ここだけが本番で走る。
func (p Plan) Compile() Step {
	steps := make([]Step, 0, len(p.Nodes))
	for _, n := range p.Nodes {
		steps = append(steps, n.Step)
	}
	return Sequence(steps...)
}

// Sequence は射のリストを1本に合成する。恒等射から畳んでいくので、空なら素通しになる。
func Sequence(steps ...Step) Step {
	out := Step(func(s State) (State, error) { return s, nil })
	for _, step := range steps {
		out = Then(out, step)
	}
	return out
}

// Then は Kleisli 合成。error が出たら後続を呼ばない。
func Then(f, g Step) Step {
	return func(s State) (State, error) {
		s, err := f(s)
		if err != nil {
			return s, err
		}
		return g(s)
	}
}

func showRun(p Plan) {
	externalCalls = 0
	out, err := p.Compile()(State{OrderID: "A-1001"})

	fmt.Printf("\n[実行] %v （外部サービス %d 回）\n", out.Trail, externalCalls)
	if err != nil {
		fmt.Printf("       エラー: %v\n", err)
	}
}

// ---- 解釈② dry-run ----------------------------------------------------

// DryRun は同じ Plan から「何が起きるはずか」だけを出す。Step は1度も呼ばない。
func (p Plan) DryRun() []string {
	out := make([]string, 0, len(p.Nodes))
	for i, n := range p.Nodes {
		out = append(out, fmt.Sprintf("%d. %s", i+1, n.Name))
	}
	return out
}

func showDryRunTouchesNothing(p Plan) {
	externalCalls = 0
	steps := p.DryRun()

	fmt.Printf("\n[dry-run] %s\n", strings.Join(steps, " → "))
	fmt.Printf("          外部サービスの呼び出し: %d 回\n", externalCalls)
	fmt.Println("          素直な実装に dry-run を足すと、本番コードに if dryRun が4か所混ざる")
}

// ---- デコレータ（射を包む変換）-----------------------------------------

// WithRetry は失敗したら最大 n 回まで試し直す射を返す。
func WithRetry(step Step, n int) Step {
	return func(s State) (State, error) {
		var err error
		for i := 0; i < n; i++ {
			var out State
			out, err = step(s)
			if err == nil {
				return out, nil
			}
		}
		return s, err
	}
}

// WithBudget は「合計でこの回数まで」という上限を掛ける。タイムアウトの代役。
func WithBudget(step Step, budget *int) Step {
	return func(s State) (State, error) {
		if *budget <= 0 {
			return s, errors.New("打ち切り")
		}
		*budget--
		return step(s)
	}
}

// リトライとタイムアウトは、巻く順序で結果が変わる。
// 外側に上限を置けば「上限まで」で止まり、内側に置けばリトライの回数ぶん走る。
func showDecoratorOrderDecidesTheResult() {
	attempts := 0
	always := Step(func(s State) (State, error) {
		attempts++
		return s, errors.New("一時的な失敗")
	})

	budget := 3
	attempts = 0
	WithRetry(WithBudget(always, &budget), 10)(State{})
	inner := attempts

	budget = 3
	attempts = 0
	WithBudget(WithRetry(always, 10), &budget)(State{})
	outer := attempts

	fmt.Printf("\n[デコレータの順序] 上限を内側に巻くと %d 回で打ち切り / 外側に巻くと %d 回走った\n", inner, outer)
	fmt.Println("                   同じ部品でも巻く順序が結果を決める。合成は結合的だが可換ではない")
}

// ---- 解釈③ Mermaid 図 -------------------------------------------------

// Mermaid は同じ Plan から図を書き出す。ステップを足せば図に出るので、
// 設計ドキュメントが実装とずれる問題が構造的に起きなくなる。
func (p Plan) Mermaid() string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	fmt.Fprintf(&b, "    start([%s])\n", p.Title)

	prev := "start"
	for i, n := range p.Nodes {
		id := fmt.Sprintf("n%d", i)
		label := n.Name
		if n.Retries > 0 {
			label += fmt.Sprintf("<br/>最大%d回", n.Retries)
		}
		if n.Timeout > 0 {
			label += "<br/>" + n.Timeout.String()
		}
		fmt.Fprintf(&b, "    %s[%s]\n", id, label)
		fmt.Fprintf(&b, "    %s --> %s\n", prev, id)
		prev = id
	}
	return b.String()
}

func showMermaid(p Plan) {
	fmt.Println("\n[Mermaid] この図は上の8行の定義から生成しています。ステップを足せば図に出ます")
	fmt.Print(p.Mermaid())
}
