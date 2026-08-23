package store

import (
	"database/sql"
	"errors"

	"task209-deformgate/internal/model"
)

// CreateTask 插入检查任务。若同一 (影像对, 形变场) 已存在活动任务，
// 唯一索引会触发冲突，返回 model.ErrConflict。
func (s *Store) CreateTask(t *model.CheckTask) error {
	now := Now().Format(timeLayout)
	res, err := s.db.Exec(
		`INSERT INTO check_tasks (image_pair_id, field_id, params_id, status, message, created_at, updated_at)
		 VALUES (?, ?, ?, ?, '', ?, ?)`,
		t.ImagePairID, t.FieldID, t.ParamsID, model.TaskQueued, now, now,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrConflict
		}
		return err
	}
	t.ID, _ = res.LastInsertId()
	t.Status = model.TaskQueued
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

// GetTask 按 ID 读取任务。
func (s *Store) GetTask(id int64) (*model.CheckTask, error) {
	var t model.CheckTask
	err := s.db.QueryRow(
		`SELECT id, image_pair_id, field_id, params_id, status, message, created_at, updated_at
		 FROM check_tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.ImagePairID, &t.FieldID, &t.ParamsID, &t.Status, &t.Message, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListTasks 返回全部检查任务（按 ID 升序）。
func (s *Store) ListTasks() ([]model.CheckTask, error) {
	rows, err := s.db.Query(
		`SELECT id, image_pair_id, field_id, params_id, status, message, created_at, updated_at
		 FROM check_tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.CheckTask
	for rows.Next() {
		var t model.CheckTask
		if err := rows.Scan(&t.ID, &t.ImagePairID, &t.FieldID, &t.ParamsID, &t.Status, &t.Message, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// SetTaskStatus 迁移任务状态。
func (s *Store) SetTaskStatus(id int64, status string) error {
	res, err := s.db.Exec(
		`UPDATE check_tasks SET status = ?, updated_at = ? WHERE id = ?`,
		status, Now().Format(timeLayout), id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SetTaskMessage 更新任务失败/完成消息。
func (s *Store) SetTaskMessage(id int64, msg string) error {
	_, err := s.db.Exec(
		`UPDATE check_tasks SET message = ?, updated_at = ? WHERE id = ?`,
		msg, Now().Format(timeLayout), id,
	)
	return err
}

// RecoverRunningToQueued 把中断的 running 任务回退为 queued（重启恢复）。
func (s *Store) RecoverRunningToQueued() error {
	_, err := s.db.Exec(
		`UPDATE check_tasks SET status = ?, updated_at = ? WHERE status = ?`,
		model.TaskQueued, Now().Format(timeLayout), model.TaskRunning,
	)
	return err
}

// isUniqueViolation 判断是否为唯一约束冲突（SQLite 错误码 19）。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite 返回带 "constraint" / "UNIQUE" 的错误文本。
	msg := err.Error()
	return contains(msg, "constraint failed") || contains(msg, "UNIQUE constraint")
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
