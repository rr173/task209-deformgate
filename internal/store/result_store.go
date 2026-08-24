package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"task209-deformgate/internal/model"
)

// CreateResult 插入质量结果（指标三块 JSON 存储，状态默认 draft）。
func (s *Store) CreateResult(r *model.QualityResult) error {
	jac, err := json.Marshal(r.Jacobian)
	if err != nil {
		return err
	}
	inv, err := json.Marshal(r.Inverse)
	if err != nil {
		return err
	}
	bnd, err := json.Marshal(r.Boundary)
	if err != nil {
		return err
	}
	now := Now().Format(timeLayout)
	res, err := s.db.Exec(
		`INSERT INTO quality_results
		 (task_id, image_pair_id, field_id, params_id, verdict, jacobian_json, inverse_json, boundary_json, status, supersedes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		r.TaskID, r.ImagePairID, r.FieldID, r.ParamsID, r.Verdict,
		string(jac), string(inv), string(bnd), model.ResultDraft, now, now,
	)
	if err != nil {
		return err
	}
	r.ID, _ = res.LastInsertId()
	r.Status = model.ResultDraft
	r.CreatedAt = now
	r.UpdatedAt = now
	return nil
}

// SaveFolds 批量写入折叠/异常体素证据。
func (s *Store) SaveFolds(resultID int64, folds []model.FoldRegion) error {
	if len(folds) == 0 {
		return nil
	}
	return s.tx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			`INSERT INTO fold_regions (result_id, voxel_index, pos_x, pos_y, pos_z, det)
			 VALUES (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, f := range folds {
			if _, err := stmt.Exec(resultID, f.Index, f.Pos.X, f.Pos.Y, f.Pos.Z, f.Det); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetResult 按 ID 读取质量结果。
func (s *Store) GetResult(id int64) (*model.QualityResult, error) {
	var r model.QualityResult
	var jacJSON, invJSON, bndJSON string
	err := s.db.QueryRow(
		`SELECT id, task_id, image_pair_id, field_id, params_id, verdict,
		        jacobian_json, inverse_json, boundary_json, status, supersedes, created_at, updated_at
		 FROM quality_results WHERE id = ?`, id,
	).Scan(&r.ID, &r.TaskID, &r.ImagePairID, &r.FieldID, &r.ParamsID, &r.Verdict,
		&jacJSON, &invJSON, &bndJSON, &r.Status, &r.Supersedes, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(jacJSON), &r.Jacobian)
	_ = json.Unmarshal([]byte(invJSON), &r.Inverse)
	_ = json.Unmarshal([]byte(bndJSON), &r.Boundary)
	return &r, nil
}

// GetResultByTask 按检查任务读取其产出的质量结果。
func (s *Store) GetResultByTask(taskID int64) (*model.QualityResult, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM quality_results WHERE task_id = ? LIMIT 1`, taskID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.GetResult(id)
}

// ListResultsByPair 返回某影像对下全部质量结果。
func (s *Store) ListResultsByPair(pairID int64) ([]model.QualityResult, error) {
	rows, err := s.db.Query(
		`SELECT id, task_id, image_pair_id, field_id, params_id, verdict,
		        jacobian_json, inverse_json, boundary_json, status, supersedes, created_at, updated_at
		 FROM quality_results WHERE image_pair_id = ? ORDER BY id`, pairID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

// ListResults 返回全部质量结果。
func (s *Store) ListResults() ([]model.QualityResult, error) {
	rows, err := s.db.Query(
		`SELECT id, task_id, image_pair_id, field_id, params_id, verdict,
		        jacobian_json, inverse_json, boundary_json, status, supersedes, created_at, updated_at
		 FROM quality_results ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

func scanResults(rows *sql.Rows) ([]model.QualityResult, error) {
	var out []model.QualityResult
	for rows.Next() {
		var r model.QualityResult
		var jacJSON, invJSON, bndJSON string
		if err := rows.Scan(&r.ID, &r.TaskID, &r.ImagePairID, &r.FieldID, &r.ParamsID, &r.Verdict,
			&jacJSON, &invJSON, &bndJSON, &r.Status, &r.Supersedes, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(jacJSON), &r.Jacobian)
		_ = json.Unmarshal([]byte(invJSON), &r.Inverse)
		_ = json.Unmarshal([]byte(bndJSON), &r.Boundary)
		out = append(out, r)
	}
	return out, rows.Err()
}

// PublishResult 发布质量结果：事务内把该影像对既有 published 结果显式替代，
// 新结果绑定 supersedes 指向被替代结果。
func (s *Store) PublishResult(id int64) error {
	return s.tx(func(tx *sql.Tx) error {
		var pairID int64
		var status string
		err := tx.QueryRow(`SELECT image_pair_id, status FROM quality_results WHERE id = ?`, id).
			Scan(&pairID, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrNotFound
		}
		if err != nil {
			return err
		}
		if status != model.ResultDraft {
			return model.ErrConflict
		}
		// 找当前 pair 的 published 结果（即被替代的旧版本）。
		// 注意：必须排除正在发布的结果自身（它此刻仍是 draft，但即便未来
		// 允许重复发布，也不应让 supersedes 指向自己）。
		var prevID int64
		err = tx.QueryRow(
			`SELECT id FROM quality_results
			 WHERE image_pair_id = ? AND status = ? AND id <> ? LIMIT 1`,
			pairID, model.ResultPublished, id,
		).Scan(&prevID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if prevID != 0 {
			if _, err := tx.Exec(
				`UPDATE quality_results SET status = ?, updated_at = ? WHERE id = ?`,
				model.ResultSuperseded, Now().Format(timeLayout), prevID,
			); err != nil {
				return err
			}
		}
		_, err = tx.Exec(
			`UPDATE quality_results SET status = ?, supersedes = ?, updated_at = ? WHERE id = ?`,
			model.ResultPublished, prevID, Now().Format(timeLayout), id,
		)
		return err
	})
}

// ListFolds 返回某结果下的折叠证据。
func (s *Store) ListFolds(resultID int64) ([]model.FoldRegion, error) {
	rows, err := s.db.Query(
		`SELECT voxel_index, pos_x, pos_y, pos_z, det FROM fold_regions WHERE result_id = ? ORDER BY voxel_index`, resultID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.FoldRegion
	for rows.Next() {
		var f model.FoldRegion
		if err := rows.Scan(&f.Index, &f.Pos.X, &f.Pos.Y, &f.Pos.Z, &f.Det); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
