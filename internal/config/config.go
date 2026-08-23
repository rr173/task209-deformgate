// Package config 提供服务配置的读取与默认值。
package config

import "os"

// Default 常量：服务默认行为。
const (
	DefaultAddr = ":8080"
	DefaultDB   = "./task209-deformgate.db"
)

// Config 是服务的运行时配置。
type Config struct {
	Addr      string // HTTP 监听地址
	DBPath    string // SQLite 数据库路径
	SmokeTest bool   // 仅执行自检后退出
	SmokeDir  string // 自检工作目录
}

// Load 从命令行参数构造配置。
// 支持 --addr、--db、--smoke-test、--smoke-dir 四个标志。
func Load(args []string) (Config, error) {
	cfg := Config{
		Addr:   DefaultAddr,
		DBPath: DefaultDB,
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			if i+1 < len(args) {
				i++
				cfg.Addr = args[i]
			}
		case "--db":
			if i+1 < len(args) {
				i++
				cfg.DBPath = args[i]
			}
		case "--smoke-dir":
			if i+1 < len(args) {
				i++
				cfg.SmokeDir = args[i]
			}
		case "--smoke-test":
			cfg.SmokeTest = true
		}
	}
	if cfg.SmokeDir == "" {
		cfg.SmokeDir = os.TempDir()
	}
	return cfg, nil
}
