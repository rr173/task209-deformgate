package service

import (
	"testing"

	"task209-deformgate/internal/field"
	"task209-deformgate/internal/model"
	"task209-deformgate/internal/store"
)

func TestTask209Bug09CheckCompletionRegression(t *testing.T) {
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
	f, err := app.AttachField(pair.ID, model.Dims{2, 2}, field.SynthesizeIdentity(model.Dims{2, 2}))
	if err != nil {
		t.Fatal(err)
	}
	task, err := app.CreateCheck(pair.ID, f.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.RunCheck(task.ID); err != nil {
		t.Fatal(err)
	}
	got, err := app.GetCheck(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.TaskCompleted {
		panic("completed check task did not persist completed status")
	}
}
