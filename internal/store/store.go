// Package store 提供基于 SQLite（modernc.org/sqlite，纯 Go 驱动）的持久化层：
// 建表迁移、CRUD、事务与重启恢复。
//
// 所有写路径走单写者串行化（database/sql 连接池上限 1 + 写事务），
// 幂等约束用唯一索引保证（形变场内容哈希、(影像对,形变场) 活动任务唯一）。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store 封装 SQLite 连接与领域仓储。
type Store struct {
	db *sql.DB
}

// Open 打开（或创建）SQLite 数据库并执行迁移。
// path 为空时使用内存数据库（仅测试用）。
func Open(path string) (*Store, error) {
	if path == "" {
		path = ":memory:"
	} else if filepath.Dir(path) != "." {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
	}
	dsn := path
	if path != ":memory:" {
		dsn = fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // 单写者：避免 SQLite 锁竞争
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// DB 暴露底层连接（仅供事务性 API 内部使用）。
func (s *Store) DB() *sql.DB { return s.db }

// Now 返回统一的 UTC 时间戳字符串。
func Now() time.Time { return time.Now().UTC() }

// tx 开启写事务，并在 fn 出错时回滚。
func (s *Store) tx(fn func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
