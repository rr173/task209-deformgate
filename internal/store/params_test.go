package store

import (
	"testing"

	"task209-deformgate/internal/model"
)

// TestActivateParamsDeactivatesOthers 验证：激活新版本后，旧版本必须自动失活，
// 任意时刻只能有一个激活版本。
func TestActivateParamsDeactivatesOthers(t *testing.T) {
	st, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// 迁移已写入一个默认激活版本（id=1）。
	if err := st.CreateParams(dummyParams("v2")); err != nil {
		t.Fatal(err)
	}
	v2 := lastID(t, st)
	if err := st.CreateParams(dummyParams("v3")); err != nil {
		t.Fatal(err)
	}
	v3 := lastID(t, st)

	// 激活 v2：默认版本（id=1）应自动失活。
	if err := st.ActivateParams(v2); err != nil {
		t.Fatal(err)
	}
	assertSingleActive(t, st, v2)

	// 再激活 v3：v2 应自动失活，仅 v3 保持激活。
	if err := st.ActivateParams(v3); err != nil {
		t.Fatal(err)
	}
	assertSingleActive(t, st, v3)
}

func dummyParams(name string) *model.CheckParams {
	return &model.CheckParams{
		Name:              name,
		JacMin:            0.5,
		JacMax:            2.0,
		ICEThreshold:      2.0,
		CoverageThreshold: 0.95,
	}
}

func lastID(t *testing.T, st *Store) int64 {
	t.Helper()
	var id int64
	if err := st.db.QueryRow(`SELECT id FROM check_params ORDER BY id DESC LIMIT 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func assertSingleActive(t *testing.T, st *Store, wantID int64) {
	t.Helper()
	var activeCount int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM check_params WHERE active = 1`).Scan(&activeCount); err != nil {
		t.Fatal(err)
	}
	if activeCount != 1 {
		t.Fatalf("active count=%d, want 1 (only the newly activated version should be active)", activeCount)
	}
	var activeID int64
	if err := st.db.QueryRow(`SELECT id FROM check_params WHERE active = 1`).Scan(&activeID); err != nil {
		t.Fatal(err)
	}
	if activeID != wantID {
		t.Fatalf("active id=%d, want %d", activeID, wantID)
	}
}
