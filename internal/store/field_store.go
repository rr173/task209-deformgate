package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"task209-deformgate/internal/model"
)

// CreateField 插入形变场（位移向量整体 JSON 存储，hash 唯一幂等）。
func (s *Store) CreateField(f *model.DeformField) error {
	dims, err := json.Marshal(f.Dims)
	if err != nil {
		return err
	}
	disp, err := json.Marshal(f.Displacements)
	if err != nil {
		return err
	}
	now := Now().Format(timeLayout)
	res, err := s.db.Exec(
		`INSERT INTO deform_fields (image_pair_id, hash, dims_json, displacements_json, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		f.ImagePairID, f.Hash, string(dims), string(disp), model.FieldUploading, now, now,
	)
	if err != nil {
		return err
	}
	f.ID, _ = res.LastInsertId()
	f.Status = model.FieldUploading
	f.CreatedAt = now
	f.UpdatedAt = now
	return nil
}

// GetField 按 ID 读取形变场。
func (s *Store) GetField(id int64) (*model.DeformField, error) {
	var f model.DeformField
	var dimsJSON, dispJSON string
	err := s.db.QueryRow(
		`SELECT id, image_pair_id, hash, dims_json, displacements_json, status, created_at, updated_at
		 FROM deform_fields WHERE id = ?`, id,
	).Scan(&f.ID, &f.ImagePairID, &f.Hash, &dimsJSON, &dispJSON, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(dimsJSON), &f.Dims)
	if err := json.Unmarshal([]byte(dispJSON), &f.Displacements); err != nil {
		return nil, err
	}
	return &f, nil
}

// FieldByHash 按内容哈希查找形变场（幂等去重）。
func (s *Store) FieldByHash(hash string) (*model.DeformField, error) {
	var f model.DeformField
	var dimsJSON, dispJSON string
	err := s.db.QueryRow(
		`SELECT id, image_pair_id, hash, dims_json, displacements_json, status, created_at, updated_at
		 FROM deform_fields WHERE hash = ?`, hash,
	).Scan(&f.ID, &f.ImagePairID, &f.Hash, &dimsJSON, &dispJSON, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(dimsJSON), &f.Dims)
	_ = json.Unmarshal([]byte(dispJSON), &f.Displacements)
	return &f, nil
}

// ListFieldsByPair 返回某影像对下全部形变场。
func (s *Store) ListFieldsByPair(pairID int64) ([]model.DeformField, error) {
	rows, err := s.db.Query(
		`SELECT id, image_pair_id, hash, dims_json, displacements_json, status, created_at, updated_at
		 FROM deform_fields WHERE image_pair_id = ? ORDER BY id`, pairID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.DeformField
	for rows.Next() {
		var f model.DeformField
		var dimsJSON, dispJSON string
		if err := rows.Scan(&f.ID, &f.ImagePairID, &f.Hash, &dimsJSON, &dispJSON, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(dimsJSON), &f.Dims)
		_ = json.Unmarshal([]byte(dispJSON), &f.Displacements)
		out = append(out, f)
	}
	return out, rows.Err()
}

// SetFieldStatus 迁移形变场状态。
func (s *Store) SetFieldStatus(id int64, status string) error {
	res, err := s.db.Exec(
		`UPDATE deform_fields SET status = ?, updated_at = ? WHERE id = ?`,
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

// CreateSample 插入一个采样点（(field_id, pos) 唯一幂等）。
func (s *Store) CreateSample(sp *model.SamplePoint) error {
	res, err := s.db.Exec(
		`INSERT INTO sample_points
		 (field_id, pos_x, pos_y, pos_z, forward_x, forward_y, forward_z, inverse_x, inverse_y, inverse_z, error, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sp.FieldID, sp.Pos.X, sp.Pos.Y, sp.Pos.Z,
		sp.Forward.X, sp.Forward.Y, sp.Forward.Z,
		sp.Inverse.X, sp.Inverse.Y, sp.Inverse.Z,
		sp.Error, sp.Status,
	)
	if err != nil {
		return err
	}
	sp.ID, _ = res.LastInsertId()
	return nil
}

// ListSamplesByField 返回某形变场下全部采样点。
func (s *Store) ListSamplesByField(fieldID int64) ([]model.SamplePoint, error) {
	rows, err := s.db.Query(
		`SELECT id, field_id, pos_x, pos_y, pos_z, forward_x, forward_y, forward_z,
		        inverse_x, inverse_y, inverse_z, error, status
		 FROM sample_points WHERE field_id = ? ORDER BY id`, fieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SamplePoint
	for rows.Next() {
		var sp model.SamplePoint
		if err := rows.Scan(&sp.ID, &sp.FieldID,
			&sp.Pos.X, &sp.Pos.Y, &sp.Pos.Z,
			&sp.Forward.X, &sp.Forward.Y, &sp.Forward.Z,
			&sp.Inverse.X, &sp.Inverse.Y, &sp.Inverse.Z,
			&sp.Error, &sp.Status); err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

// SetSampleChecked 标记采样点已校验并写入误差。
func (s *Store) SetSampleChecked(id int64, errVal float64) error {
	res, err := s.db.Exec(
		`UPDATE sample_points SET error = ?, status = ? WHERE id = ?`,
		errVal, model.SampleChecked, id,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}
