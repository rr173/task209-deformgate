package service

import (
	"errors"
	"testing"

	"task209-deformgate/internal/model"
	"task209-deformgate/internal/store"
)

// newTestApp 打开一个内存数据库并构造已迁移、含默认激活参数的 App。
func newTestApp(t *testing.T) *App {
	t.Helper()
	st, err := store.Open("")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	app, err := NewApp(st)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	return app
}

// TestAttachFieldAdvancesRegisteredPairToPending 校验首次追加合法形变场后，
// 影像对状态由 registered 推进到 pending_validation（回归：原实现停在 registered）。
func TestAttachFieldAdvancesRegisteredPairToPending(t *testing.T) {
	app := newTestApp(t)

	dims := model.Dims{2, 2}
	pair, err := app.CreateImagePair("pair", dims, []float64{1, 1}, "RAS", "MRI")
	if err != nil {
		t.Fatalf("create pair: %v", err)
	}
	if pair.Status != model.PairRegistered {
		t.Fatalf("initial status=%q, want %q", pair.Status, model.PairRegistered)
	}

	disp := make([]model.Vec3, dims.Volume())
	if _, err := app.AttachField(pair.ID, dims, disp); err != nil {
		t.Fatalf("attach field: %v", err)
	}

	after, err := app.GetImagePair(pair.ID)
	if err != nil {
		t.Fatalf("get pair: %v", err)
	}
	if after.Status != model.PairPending {
		t.Fatalf("status after attach=%q, want %q", after.Status, model.PairPending)
	}
}

// TestAttachFieldRejectedOnArchivedPair 校验封存影像对仍禁止追加形变场，
// 修复不得放宽该规则。
func TestAttachFieldRejectedOnArchivedPair(t *testing.T) {
	app := newTestApp(t)

	dims := model.Dims{2, 2}
	pair, err := app.CreateImagePair("pair", dims, []float64{1, 1}, "RAS", "MRI")
	if err != nil {
		t.Fatalf("create pair: %v", err)
	}
	if err := app.ArchiveImagePair(pair.ID); err != nil {
		t.Fatalf("archive pair: %v", err)
	}

	disp := make([]model.Vec3, dims.Volume())
	if _, err := app.AttachField(pair.ID, dims, disp); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("attach to archived pair: err=%v, want ErrForbidden", err)
	}
}
