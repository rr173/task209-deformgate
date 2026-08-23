// Package service 是编排层：串联影像对 → 形变场 → 检查 → 质量结果 →
// 发布 的完整业务闭环，并承载状态机流转、并发约束与错误映射。
package service

import (
	"task209-deformgate/internal/store"
)

// App 是服务编排入口，持有持久化仓储引用。
type App struct {
	st *store.Store
}

// NewApp 构造应用实例，并执行重启恢复（把中断的 running 任务回退为 queued）。
func NewApp(st *store.Store) (*App, error) {
	if err := st.Recover(); err != nil {
		return nil, err
	}
	return &App{st: st}, nil
}

// Store 暴露底层仓储（仅供同包与 httpapi 元信息端点使用）。
func (a *App) Store() *store.Store { return a.st }
