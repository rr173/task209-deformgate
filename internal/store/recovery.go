package store

// Stats 汇总服务级统计信息，供 /api/stats 与自检报告使用。
type Stats struct {
	ImagePairs int `json:"image_pairs"`
	Fields     int `json:"fields"`
	Samples    int `json:"samples"`
	Tasks      int `json:"tasks"`
	Results    int `json:"results"`
	Params     int `json:"params"`
	ActiveTask int `json:"active_tasks"`
}

// Stats 统计各实体数量与活动任务数。
func (s *Store) Stats() (Stats, error) {
	var st Stats
	counters := []struct {
		table string
		dst   *int
	}{
		{"image_pairs", &st.ImagePairs},
		{"deform_fields", &st.Fields},
		{"sample_points", &st.Samples},
		{"check_tasks", &st.Tasks},
		{"quality_results", &st.Results},
		{"check_params", &st.Params},
	}
	for _, c := range counters {
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM ` + c.table).Scan(c.dst); err != nil {
			return st, err
		}
	}
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM check_tasks WHERE status IN ('queued','running')`).Scan(&st.ActiveTask); err != nil {
		return st, err
	}
	return st, nil
}

// Recover 执行重启恢复：把中断的 running 任务回退为 queued，
// 使重启后可继续从已持久化的指标/结果续跑。
func (s *Store) Recover() error {
	return s.RecoverRunningToQueued()
}
