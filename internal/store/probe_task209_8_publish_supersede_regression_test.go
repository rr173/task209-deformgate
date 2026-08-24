package store

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug08PublishSupersedeRegression(t *testing.T) {
	st, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	now := Now().Format(timeLayout)
	if _, err := st.DB().Exec(`
		INSERT INTO image_pairs (name, dims_json, spacing_json, axes, created_at, updated_at) VALUES ('p', '[2,2]', '[1,1]', 'RAS', ?, ?);
		INSERT INTO deform_fields (image_pair_id, hash, dims_json, displacements_json, created_at, updated_at) VALUES (1, 'h1', '[2,2]', '[]', ?, ?);
		INSERT INTO deform_fields (image_pair_id, hash, dims_json, displacements_json, created_at, updated_at) VALUES (1, 'h2', '[2,2]', '[]', ?, ?);
		INSERT INTO check_tasks (image_pair_id, field_id, params_id, status, created_at, updated_at) VALUES (1, 1, 1, 'completed', ?, ?);
		INSERT INTO check_tasks (image_pair_id, field_id, params_id, status, created_at, updated_at) VALUES (1, 2, 1, 'completed', ?, ?);`, now, now, now, now, now, now, now, now, now, now); err != nil {
		t.Fatal(err)
	}
	r1 := &model.QualityResult{TaskID: 1, ImagePairID: 1, FieldID: 1, ParamsID: 1, Verdict: model.VerdictPass}
	r2 := &model.QualityResult{TaskID: 2, ImagePairID: 1, FieldID: 2, ParamsID: 1, Verdict: model.VerdictPass}
	if err := st.CreateResult(r1); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateResult(r2); err != nil {
		t.Fatal(err)
	}
	if err := st.PublishResult(r1.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.PublishResult(r2.ID); err != nil {
		panic("second published result could not replace the previous result")
	}
	old, err := st.GetResult(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	newer, err := st.GetResult(r2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if old.Status != model.ResultSuperseded || newer.Status != model.ResultPublished || newer.Supersedes != old.ID {
		panic("published result replacement invariant violated")
	}
}
