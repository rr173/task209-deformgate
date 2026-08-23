package gating

import (
	"errors"
	"testing"

	"task209-deformgate/internal/model"
)

func TestValidateVersionTransition(t *testing.T) {
	valid := [][2]string{
		{model.ResultDraft, model.ResultPublished},
		{model.ResultPublished, model.ResultSuperseded},
		{model.ResultDraft, model.ResultDraft},
	}
	for _, pair := range valid {
		if err := ValidateVersionTransition(pair[0], pair[1]); err != nil {
			t.Fatalf("%s -> %s rejected: %v", pair[0], pair[1], err)
		}
	}
	if err := ValidateVersionTransition(model.ResultDraft, model.ResultSuperseded); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("invalid transition error=%v, want ErrConflict", err)
	}
}

func TestBuildVersionChainSeparatesPublishedAndHistory(t *testing.T) {
	results := []model.QualityResult{
		{ID: 1, Status: model.ResultSuperseded},
		{ID: 2, Status: model.ResultPublished},
		{ID: 3, Status: model.ResultDraft},
	}
	chain := BuildVersionChain(42, results)
	if chain.ImagePairID != 42 {
		t.Fatalf("pair id=%d, want 42", chain.ImagePairID)
	}
	if chain.Published == nil || chain.Published.ID != 2 {
		t.Fatalf("published result=%v, want result 2", chain.Published)
	}
	if len(chain.History) != 2 || chain.History[0].ID != 1 || chain.History[1].ID != 3 {
		t.Fatalf("history=%v, want superseded 1 and draft 3", chain.History)
	}
}

func TestValidatePublishEnforcesDraftAndArchiveRules(t *testing.T) {
	draft := model.QualityResult{Status: model.ResultDraft}
	if err := ValidatePublish(draft, model.PairPending); err != nil {
		t.Fatalf("draft result should publish: %v", err)
	}
	if err := ValidatePublish(model.QualityResult{Status: model.ResultPublished}, model.PairPending); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("published result error=%v, want ErrConflict", err)
	}
	if err := ValidatePublish(draft, model.PairArchived); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("archived pair error=%v, want ErrForbidden", err)
	}
}
