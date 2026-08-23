package store

// migrate 建表并做轻量迁移。当前版本为单轮全量建表，并写入一个默认激活的
// 检查参数版本，保证任何检查任务都有可引用的参数快照。
// 未来如需演进 schema，在此按版本号叠加增量迁移。
func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS image_pairs (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT NOT NULL,
	dims_json     TEXT NOT NULL,
	spacing_json  TEXT NOT NULL,
	axes          TEXT NOT NULL,
	modality      TEXT NOT NULL DEFAULT '',
	status        TEXT NOT NULL DEFAULT 'registered',
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS deform_fields (
	id                 INTEGER PRIMARY KEY AUTOINCREMENT,
	image_pair_id      INTEGER NOT NULL REFERENCES image_pairs(id),
	hash               TEXT NOT NULL UNIQUE,
	dims_json          TEXT NOT NULL,
	displacements_json TEXT NOT NULL,
	status             TEXT NOT NULL DEFAULT 'uploading',
	created_at         TEXT NOT NULL,
	updated_at         TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_fields_pair ON deform_fields(image_pair_id);

CREATE TABLE IF NOT EXISTS sample_points (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	field_id   INTEGER NOT NULL REFERENCES deform_fields(id) ON DELETE CASCADE,
	pos_x      REAL NOT NULL,
	pos_y      REAL NOT NULL,
	pos_z      REAL NOT NULL,
	forward_x  REAL NOT NULL,
	forward_y  REAL NOT NULL,
	forward_z  REAL NOT NULL,
	inverse_x  REAL NOT NULL,
	inverse_y  REAL NOT NULL,
	inverse_z  REAL NOT NULL,
	error      REAL NOT NULL DEFAULT 0,
	status     TEXT NOT NULL DEFAULT 'open',
	UNIQUE (field_id, pos_x, pos_y, pos_z)
);

CREATE TABLE IF NOT EXISTS check_params (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	name              TEXT NOT NULL,
	jac_min           REAL NOT NULL,
	jac_max           REAL NOT NULL,
	ice_threshold     REAL NOT NULL,
	coverage_threshold REAL NOT NULL,
	active            INTEGER NOT NULL DEFAULT 0,
	created_at        TEXT NOT NULL,
	updated_at        TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS check_tasks (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	image_pair_id INTEGER NOT NULL REFERENCES image_pairs(id),
	field_id     INTEGER NOT NULL REFERENCES deform_fields(id),
	params_id    INTEGER NOT NULL REFERENCES check_params(id),
	status       TEXT NOT NULL DEFAULT 'queued',
	message      TEXT NOT NULL DEFAULT '',
	created_at   TEXT NOT NULL,
	updated_at   TEXT NOT NULL,
	UNIQUE (image_pair_id, field_id, status)
);
-- 同一 (影像对, 形变场) 只允许一个活动任务，防止并发重复检查。
CREATE UNIQUE INDEX IF NOT EXISTS idx_task_active
	ON check_tasks(image_pair_id, field_id)
	WHERE status IN ('queued', 'running');

CREATE TABLE IF NOT EXISTS quality_results (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id       INTEGER NOT NULL REFERENCES check_tasks(id),
	image_pair_id INTEGER NOT NULL REFERENCES image_pairs(id),
	field_id      INTEGER NOT NULL REFERENCES deform_fields(id),
	params_id     INTEGER NOT NULL REFERENCES check_params(id),
	verdict       TEXT NOT NULL,
	jacobian_json TEXT NOT NULL,
	inverse_json  TEXT NOT NULL,
	boundary_json TEXT NOT NULL,
	status        TEXT NOT NULL DEFAULT 'draft',
	supersedes    INTEGER NOT NULL DEFAULT 0,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);
-- 同一影像对最多一个发布态结果。
CREATE UNIQUE INDEX IF NOT EXISTS idx_result_published
	ON quality_results(image_pair_id) WHERE status = 'published';

CREATE TABLE IF NOT EXISTS fold_regions (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	result_id   INTEGER NOT NULL REFERENCES quality_results(id) ON DELETE CASCADE,
	voxel_index INTEGER NOT NULL,
	pos_x       REAL NOT NULL,
	pos_y       REAL NOT NULL,
	pos_z       REAL NOT NULL,
	det         REAL NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_folds_result ON fold_regions(result_id);
`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	// 写入默认激活参数（仅当没有任何参数版本时）。
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM check_params`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		now := Now().Format(timeLayout)
		if _, err := s.db.Exec(
			`INSERT INTO check_params (name, jac_min, jac_max, ice_threshold, coverage_threshold, active, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
			"default", 0.5, 2.0, 2.0, 0.95, now, now,
		); err != nil {
			return err
		}
	}
	return nil
}

// timeLayout 是数据库时间戳的存储格式。
const timeLayout = "2006-01-02T15:04:05.000Z"
