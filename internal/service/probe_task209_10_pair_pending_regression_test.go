package service

import (
	"testing"

	"task209-deformgate/internal/field"
	"task209-deformgate/internal/model"
	"task209-deformgate/internal/store"
)

func TestTask209Bug10PairPendingRegression(t *testing.T) {
	st, err := store.Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	app, err := NewApp(st)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := app.CreateImagePair("pair", model.Dims{2, 2}, []float64{1, 1}, "RAS", "MRI")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.AttachField(pair.ID, model.Dims{2, 2}, field.SynthesizeIdentity(model.Dims{2, 2})); err != nil {
		t.Fatal(err)
	}
	got, err := app.GetImagePair(pair.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.PairPending {
		t.Fatalf("pair status=%q, want %q", got.Status, model.PairPending)
	}
}
