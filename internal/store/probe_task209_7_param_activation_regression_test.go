package store

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug07ParamActivationRegression(t *testing.T) {
	st, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	p, err := st.GetActiveParams()
	if err != nil {
		t.Fatal(err)
	}
	next := &model.CheckParams{Name: "next", JacMin: 0.4, JacMax: 2.2, ICEThreshold: 1.5, CoverageThreshold: 0.9}
	if err := st.CreateParams(next); err != nil {
		t.Fatal(err)
	}
	if err := st.ActivateParams(next.ID); err != nil {
		t.Fatal(err)
	}
	var active int
	if err := st.DB().QueryRow(`SELECT COUNT(*) FROM check_params WHERE active = 1`).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != 1 || p.ID == next.ID {
		t.Fatalf("active parameter count=%d, old=%d new=%d", active, p.ID, next.ID)
	}
}
