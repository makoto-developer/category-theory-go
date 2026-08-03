package part3

import "time"

// ServerConfig は Functional Option パターンの題材。
type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
}

// ServerOption は設定を書き換える射。Config → Config の自己射になっている。
type ServerOption func(ServerConfig) ServerConfig

func WithHost(h string) ServerOption {
	return func(c ServerConfig) ServerConfig { c.Host = h; return c }
}

func WithPort(p int) ServerOption {
	return func(c ServerConfig) ServerConfig { c.Port = p; return c }
}

func WithTimeout(d time.Duration) ServerOption {
	return func(c ServerConfig) ServerConfig { c.Timeout = d; return c }
}

// NewServer は既定値に ServerOption を順に適用する。
// ServerOption は Config → Config の自己射なので、その全体は合成に関してモノイドをなす。
// 引数ゼロなら既定値がそのまま出るのは、恒等射が単位元として働くから。
func NewServer(opts ...ServerOption) ServerConfig {
	c := ServerConfig{Host: "localhost", Port: 8080, Timeout: 30 * time.Second}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

// ApplyOptions は ServerOption を1本にまとめる。Chain（middleware）と同じ構造。
func ApplyOptions(opts ...ServerOption) ServerOption {
	return func(c ServerConfig) ServerConfig {
		for _, opt := range opts {
			c = opt(c)
		}
		return c
	}
}

// SetHost は「設定と値の組」を受け取る形。ServerOption とはカリー化で行き来できる。
func SetHost(c ServerConfig, h string) ServerConfig { c.Host = h; return c }

// CurryOption は SetHost のような2引数の設定関数を ServerOption に変える。
// func(Config, A) Config ≅ func(A) func(Config) Config という同型が
// Functional Option パターンの正体である（積と冪の随伴）。
func CurryOption[A any](f func(ServerConfig, A) ServerConfig) func(A) ServerOption {
	return func(a A) ServerOption {
		return func(c ServerConfig) ServerConfig { return f(c, a) }
	}
}
