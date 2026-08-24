package store

import (
	"errors"
	"path/filepath"
	"testing"

	"task209-deformgate/internal/model"
)

// insertResultRow 写入一条 quality_results 记录，返回其 id。
// 用最小脚手架直接构造版本链场景，聚焦 PublishResult 的替代语义。
// 为满足外键约束，同时插入对应的 deform_fields / check_tasks 父行。
func insertResultRow(t *testing.T, st *Store, pairID, taskID int64, status string) int64 {
	t.Helper()
	now := Now().Format(timeLayout)
	fieldID := taskID // 每条结果配套一个独立 field 行
	if _, err := st.DB().Exec(
		`INSERT INTO deform_fields (image_pair_id, hash, dims_json, displacements_json, status, created_at, updated_at)
		 VALUES (?, ?, '[]', '[]', 'uploading', ?, ?)`,
		pairID, "field-hash-"+itoa(int(fieldID)), now, now,
	); err != nil {
		t.Fatalf("insert field row: %v", err)
	}
	if _, err := st.DB().Exec(
		`INSERT INTO check_tasks (image_pair_id, field_id, params_id, status, message, created_at, updated_at)
		 VALUES (?, ?, 1, 'queued', '', ?, ?)`,
		pairID, fieldID, now, now,
	); err != nil {
		t.Fatalf("insert task row: %v", err)
	}
	res, err := st.DB().Exec(
		`INSERT INTO quality_results
		 (task_id, image_pair_id, field_id, params_id, verdict,
		  jacobian_json, inverse_json, boundary_json, status, supersedes, created_at, updated_at)
		 VALUES (?, ?, ?, 1, 'pass', '{}', '{}', '{}', ?, 0, ?, ?)`,
		taskID, pairID, fieldID, status, now, now, now,
	)
	if err != nil {
		t.Fatalf("insert result row: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}

func TestPublishResultSupersedesPreviousPublished(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "publish.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := Now().Format(timeLayout)
	if _, err := st.DB().Exec(
		`INSERT INTO image_pairs (name, dims_json, spacing_json, axes, modality, status, created_at, updated_at)
		 VALUES ('pair', '[8,8,4]', '[1,1,1]', 'RAS', 'MRI', ?, ?, ?);`,
		model.PairPending, now, now,
	); err != nil {
		t.Fatal(err)
	}
	const pairID int64 = 1

	// 第一条结果：先发布，成为当前 published 版本。
	first := insertResultRow(t, st, pairID, 1, model.ResultDraft)
	if err := st.PublishResult(first); err != nil {
		t.Fatalf("publish first: %v", err)
	}
	if r, _ := st.GetResult(first); r.Status != model.ResultPublished {
		t.Fatalf("first result status=%s, want published", r.Status)
	}

	// 第二条结果：发布后必须替代第一条，自身成为唯一 published 版本，
	// supersedes 指向第一条，第一条降为 superseded。
	second := insertResultRow(t, st, pairID, 2, model.ResultDraft)
	if err := st.PublishResult(second); err != nil {
		t.Fatalf("publish second (supersede): %v", err)
	}

	firstAfter, err := st.GetResult(first)
	if err != nil {
		t.Fatal(err)
	}
	if firstAfter.Status != model.ResultSuperseded {
		t.Fatalf("superseded first result status=%s, want superseded", firstAfter.Status)
	}
	secondAfter, err := st.GetResult(second)
	if err != nil {
		t.Fatal(err)
	}
	if secondAfter.Status != model.ResultPublished {
		t.Fatalf("new published result status=%s, want published", secondAfter.Status)
	}
	if secondAfter.Supersedes != first {
		t.Fatalf("new result supersedes=%d, want %d (the previous published)", secondAfter.Supersedes, first)
	}

	// 唯一性：该影像对下 published 结果必须恰好一条。
	var n int
	if err := st.DB().QueryRow(
		`SELECT COUNT(*) FROM quality_results WHERE image_pair_id = ? AND status = ?`,
		pairID, model.ResultPublished,
	).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("published count=%d, want 1 (single authoritative published version)", n)
	}

	// 重复发布已发布结果应拒绝（不可变）。
	if err := st.PublishResult(second); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("republish published result error=%v, want ErrConflict", err)
	}
}
