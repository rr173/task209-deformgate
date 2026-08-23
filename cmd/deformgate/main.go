// 医学影像配准形变场质量门服务（task209-deformgate）
//
// 服务入口：支持 --addr 监听地址、--db 数据库路径与 --smoke-test 自检。
// --smoke-test 不启动长驻服务，而是真实创建影像对、追加形变场（含折叠场与
// 健康场）、执行雅可比/逆一致性/边界覆盖质量评估、发布结果，并关闭重开
// 数据库验证持久化与重启恢复，最后以 0 退出码结束（Docker 验证契约）。
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"task209-deformgate/internal/config"
	"task209-deformgate/internal/httpapi"
	"task209-deformgate/internal/service"
	"task209-deformgate/internal/store"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}

	if cfg.SmokeTest {
		if err := runSmoke(cfg); err != nil {
			log.Fatalf("smoke-test 失败: %v", err)
		}
		fmt.Println("smoke-test OK")
		return
	}

	if err := serve(cfg); err != nil {
		log.Fatalf("服务退出: %v", err)
	}
}

// serve 启动 HTTP 服务。
func serve(cfg config.Config) error {
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("打开数据库 %s 失败: %w", cfg.DBPath, err)
	}
	defer st.Close()

	app, err := service.NewApp(st)
	if err != nil {
		return fmt.Errorf("初始化服务失败: %w", err)
	}
	server := httpapi.New(app)

	log.Printf("task209-deformgate 监听 %s（db=%s）", cfg.Addr, cfg.DBPath)
	return http.ListenAndServe(cfg.Addr, server.Handler())
}

// runSmoke 执行端到端自检。
func runSmoke(cfg config.Config) error {
	if err := os.MkdirAll(cfg.SmokeDir, 0o755); err != nil {
		return err
	}
	dbPath := filepath.Join(cfg.SmokeDir, "smoke-deformgate.db")
	// 清理旧库，保证自检从干净 schema 开始（幂等重跑）。
	for _, f := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		_ = os.Remove(f)
	}
	res, err := service.RunSmokeTest(dbPath, cfg.SmokeDir)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	for _, s := range res.Steps {
		fmt.Println(" -", s)
	}
	return nil
}
