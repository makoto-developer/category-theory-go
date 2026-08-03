package part4

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Node はワークフローの1ステップの「定義」。実行の仕方ではなく、何をどう実行するかを値で持つ。
// 値なので、実行する以外の解釈（図を描く、設定を並べる）も同じ定義から導ける。
type Node struct {
	Name     string
	Retries  int
	Timeout  time.Duration
	Run      Endo
	Parallel []Node // 空でなければ、これらを並列に実行して失敗を集める
}

// Plan はワークフローそのもの。実行はしない。
type Plan struct {
	Name  string
	Nodes []Node
}

// --- 解釈1: 実行可能な射にコンパイルする ---

// Compile は Plan を1本の射に変える。ここでリトライ・計測・タイムアウトが巻かれる。
func Compile(p Plan, rec *Recorder) Endo {
	steps := make([]Endo, 0, len(p.Nodes))
	for _, n := range p.Nodes {
		steps = append(steps, compileNode(n, rec))
	}
	return Sequence(steps...)
}

func compileNode(n Node, rec *Recorder) Endo {
	step := n.Run
	if len(n.Parallel) > 0 {
		step = parallelStep(n.Parallel, rec)
	}

	step = WithLabel(step, n.Name)
	if n.Retries > 1 {
		step = WithRetry(step, n.Retries)
	}
	if n.Timeout > 0 {
		step = WithTimeout(step, n.Timeout)
	}
	return WithTiming(step, n.Name, rec)
}

// parallelStep は独立したステップを同時に走らせ、失敗を全部集める。
// 第3回の Applicative と同じ形。State は変更せず、副作用だけを持つステップに使う。
func parallelStep(nodes []Node, rec *Recorder) Endo {
	return func(ctx context.Context, s State) (State, error) {
		errs := make([]error, len(nodes))
		var wg sync.WaitGroup

		for i, n := range nodes {
			wg.Add(1)
			go func(i int, n Node) {
				defer wg.Done()
				_, errs[i] = compileNode(n, rec)(ctx, s)
			}(i, n)
		}
		wg.Wait()

		return s, errors.Join(errs...)
	}
}

// --- 解釈2: 実行せずに手順を並べる（dry-run） ---

// DryRun は外部サービスを一切呼ばずに、実行される順序を返す。
func DryRun(p Plan) []string {
	var out []string
	for _, n := range p.Nodes {
		if len(n.Parallel) > 0 {
			names := make([]string, len(n.Parallel))
			for i, c := range n.Parallel {
				names[i] = c.Name
			}
			out = append(out, n.Name+"(並列: "+strings.Join(names, ", ")+")")
			continue
		}
		out = append(out, n.Name)
	}
	return out
}

// --- 解釈3: 図を描く ---

// Mermaid は同じ定義から Mermaid のフローチャートを生成する。
// ドキュメントとコードがずれないのは、両方が同じ Plan から導かれるため。
func Mermaid(p Plan) string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")
	sb.WriteString("    start([" + p.Name + "])\n")

	prev := "start"
	for i, n := range p.Nodes {
		id := fmt.Sprintf("n%d", i)
		if len(n.Parallel) > 0 {
			sb.WriteString("    subgraph " + id + "[" + n.Name + "]\n")
			for j, c := range n.Parallel {
				sb.WriteString(fmt.Sprintf("        %s_%d[%s]\n", id, j, c.Name))
			}
			sb.WriteString("    end\n")
		} else {
			sb.WriteString("    " + id + "[" + label(n) + "]\n")
		}
		sb.WriteString("    " + prev + " --> " + id + "\n")
		prev = id
	}

	sb.WriteString("    " + prev + " --> done([完了])\n")
	return sb.String()
}

func label(n Node) string {
	parts := []string{n.Name}
	if n.Retries > 1 {
		parts = append(parts, fmt.Sprintf("最大%d回", n.Retries))
	}
	if n.Timeout > 0 {
		parts = append(parts, n.Timeout.String())
	}
	return strings.Join(parts, "<br/>")
}

// --- 解釈4: 設定を一覧する ---

// Explain は運用時に「どのステップが何回リトライするのか」を答えるための一覧を返す。
func Explain(p Plan) string {
	var sb strings.Builder
	sb.WriteString(p.Name + "\n")
	for _, n := range p.Nodes {
		retries := "リトライなし"
		if n.Retries > 1 {
			retries = fmt.Sprintf("最大%d回", n.Retries)
		}
		timeout := "制限なし"
		if n.Timeout > 0 {
			timeout = n.Timeout.String()
		}
		sb.WriteString(fmt.Sprintf("  - %s: %s / %s\n", n.Name, retries, timeout))
	}
	return sb.String()
}

// OrderPlan は注文処理の定義。実行方法ではなく構造だけを書いている。
func OrderPlan(d Deps) Plan {
	return Plan{
		Name: "注文処理",
		Nodes: []Node{
			{Name: "在庫確保", Retries: 3, Timeout: 2 * time.Second, Run: Reserve(d)},
			{Name: "決済", Retries: 1, Timeout: 5 * time.Second, Run: Charge(d)},
			{Name: "配送手配", Retries: 3, Timeout: 2 * time.Second, Run: Arrange(d)},
			{Name: "通知", Retries: 3, Run: Notify(d)},
		},
	}
}
