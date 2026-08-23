package store

import (
	"path/filepath"
	"testing"
)

func TestRecoverRunningTaskAfterReopen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "recover.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	now := Now().Format(timeLayout)
	if _, err := st.DB().Exec(`
		INSERT INTO image_pairs (name, dims_json, spacing_json, axes, modality, created_at, updated_at)
		VALUES ('pair', '[2,2]', '[1,1]', 'RAS', 'MRI', ?, ?);
		INSERT INTO deform_fields (image_pair_id, hash, dims_json, displacements_json, created_at, updated_at)
		VALUES (1, 'hash-1', '[2,2]', '[]', ?, ?);
		INSERT INTO check_tasks (image_pair_id, field_id, params_id, status, created_at, updated_at)
		VALUES (1, 1, 1, 'running', ?, ?);`, now, now, now, now, now, now); err != nil {
		st.Close()
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err := reopened.Recover(); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := reopened.DB().QueryRow(`SELECT status FROM check_tasks WHERE id = 1`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "queued" {
		t.Fatalf("recovered task status=%q, want queued", status)
	}
}
