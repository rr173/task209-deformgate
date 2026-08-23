package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"task209-deformgate/internal/model"
)

// CreateImagePair 插入一个新的影像对（默认状态 registered）。
func (s *Store) CreateImagePair(p *model.ImagePair) error {
	dims, err := json.Marshal(p.Dims)
	if err != nil {
		return err
	}
	spacing, err := json.Marshal(p.Spacing)
	if err != nil {
		return err
	}
	now := Now().Format(timeLayout)
	res, err := s.db.Exec(
		`INSERT INTO image_pairs (name, dims_json, spacing_json, axes, modality, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, string(dims), string(spacing), p.Axes, p.Modality, model.PairRegistered, now, now,
	)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	p.Status = model.PairRegistered
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// GetImagePair 按 ID 读取影像对；不存在返回 model.ErrNotFound。
func (s *Store) GetImagePair(id int64) (*model.ImagePair, error) {
	var p model.ImagePair
	var dimsJSON, spacingJSON string
	err := s.db.QueryRow(
		`SELECT id, name, dims_json, spacing_json, axes, modality, status, created_at, updated_at
		 FROM image_pairs WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &dimsJSON, &spacingJSON, &p.Axes, &p.Modality, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(dimsJSON), &p.Dims); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(spacingJSON), &p.Spacing); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListImagePairs 返回全部影像对（按 ID 升序）。
func (s *Store) ListImagePairs() ([]model.ImagePair, error) {
	rows, err := s.db.Query(
		`SELECT id, name, dims_json, spacing_json, axes, modality, status, created_at, updated_at
		 FROM image_pairs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ImagePair
	for rows.Next() {
		var p model.ImagePair
		var dimsJSON, spacingJSON string
		if err := rows.Scan(&p.ID, &p.Name, &dimsJSON, &spacingJSON, &p.Axes, &p.Modality, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(dimsJSON), &p.Dims)
		_ = json.Unmarshal([]byte(spacingJSON), &p.Spacing)
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpdateImagePairMeta 更新影像对可编辑元数据（名称、坐标方向、模态）。
// 已封存的影像对不可修改。
func (s *Store) UpdateImagePairMeta(id int64, name, axes, modality string) error {
	return s.tx(func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRow(`SELECT status FROM image_pairs WHERE id = ?`, id).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrNotFound
			}
			return err
		}
		if status == model.PairArchived {
			return model.ErrForbidden
		}
		_, err := tx.Exec(
			`UPDATE image_pairs SET name = ?, axes = ?, modality = ?, updated_at = ? WHERE id = ?`,
			name, axes, modality, Now().Format(timeLayout), id,
		)
		return err
	})
}

// SetImagePairStatus 迁移影像对状态（受状态机约束，见 service 层）。
func (s *Store) SetImagePairStatus(id int64, status string) error {
	res, err := s.db.Exec(
		`UPDATE image_pairs SET status = ?, updated_at = ? WHERE id = ?`,
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
