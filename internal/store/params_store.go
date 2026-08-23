package store

import (
	"database/sql"
	"errors"

	"task209-deformgate/internal/model"
)

// CreateParams 插入新的检查参数版本（默认未激活）。
func (s *Store) CreateParams(p *model.CheckParams) error {
	now := Now().Format(timeLayout)
	res, err := s.db.Exec(
		`INSERT INTO check_params (name, jac_min, jac_max, ice_threshold, coverage_threshold, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 0, ?, ?)`,
		p.Name, p.JacMin, p.JacMax, p.ICEThreshold, p.CoverageThreshold, now, now,
	)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	p.Active = false
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// GetParams 按 ID 读取参数版本。
func (s *Store) GetParams(id int64) (*model.CheckParams, error) {
	var p model.CheckParams
	var active int
	err := s.db.QueryRow(
		`SELECT id, name, jac_min, jac_max, ice_threshold, coverage_threshold, active, created_at, updated_at
		 FROM check_params WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.JacMin, &p.JacMax, &p.ICEThreshold, &p.CoverageThreshold, &active, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	return &p, nil
}

// ListParams 返回全部参数版本（按 ID 升序）。
func (s *Store) ListParams() ([]model.CheckParams, error) {
	rows, err := s.db.Query(
		`SELECT id, name, jac_min, jac_max, ice_threshold, coverage_threshold, active, created_at, updated_at
		 FROM check_params ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.CheckParams
	for rows.Next() {
		var p model.CheckParams
		var active int
		if err := rows.Scan(&p.ID, &p.Name, &p.JacMin, &p.JacMax, &p.ICEThreshold, &p.CoverageThreshold, &active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Active = active == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetActiveParams 返回当前激活的参数版本；无激活版本时回退到最新版本。
func (s *Store) GetActiveParams() (*model.CheckParams, error) {
	var p model.CheckParams
	var active int
	err := s.db.QueryRow(
		`SELECT id, name, jac_min, jac_max, ice_threshold, coverage_threshold, active, created_at, updated_at
		 FROM check_params ORDER BY active DESC, id DESC LIMIT 1`,
	).Scan(&p.ID, &p.Name, &p.JacMin, &p.JacMax, &p.ICEThreshold, &p.CoverageThreshold, &active, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	return &p, nil
}

// ActivateParams 激活指定参数版本：事务内先把所有版本置为未激活，再激活目标。
func (s *Store) ActivateParams(id int64) error {
	return s.tx(func(tx *sql.Tx) error {
		var exists int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM check_params WHERE id = ?`, id).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return model.ErrNotFound
		}
		if _, err := tx.Exec(`UPDATE check_params SET active = 0 WHERE id = ?`, id); err != nil {
			return err
		}
		_, err := tx.Exec(`UPDATE check_params SET active = 1, updated_at = ? WHERE id = ?`,
			Now().Format(timeLayout), id)
		return err
	})
}
