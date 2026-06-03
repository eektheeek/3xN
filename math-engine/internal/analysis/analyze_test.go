package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/validate"
)

func TestRun_coreInputExample(t *testing.T) {
	path := filepath.Join("..", "..", "schema", "core_input_v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var in contracts.CoreInput
	if err := json.Unmarshal(data, &in); err != nil {
		t.Fatal(err)
	}
	if err := validate.Validate(in); err != nil {
		t.Fatal(err)
	}

	out, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	if out.Version != "analysis_result_v1" {
		t.Fatalf("version = %q", out.Version)
	}
	if out.ComputedMetrics.VolumeLoad == nil || out.ComputedMetrics.VolumeLoad.Value <= 0 {
		t.Fatal("expected session volumeLoad")
	}
	if len(out.ExerciseMetrics) != 2 {
		t.Fatalf("exercise metrics = %d, want 2", len(out.ExerciseMetrics))
	}

	var warmup, working *contracts.ExerciseMetric
	for i := range out.ExerciseMetrics {
		em := &out.ExerciseMetrics[i]
		switch em.BlockType {
		case entities.BlockTypeWarmup:
			warmup = em
		case entities.BlockTypeWorking:
			working = em
		}
	}
	if warmup == nil || working == nil {
		t.Fatal("expected warmup and working blocks in output")
	}
	if warmup.Recommendation != nil {
		t.Fatal("warmup must not have recommendation")
	}
	if warmup.E1RM != nil || warmup.VolumeLoad != nil {
		t.Fatal("warmup must not have e1rm or volume metrics")
	}
	if working.E1RM == nil || working.VolumeLoad == nil {
		t.Fatal("working block must have e1rm and volume")
	}
	if working.Recommendation == nil || working.Recommendation.Action == "" {
		t.Fatal("working block must have recommendation")
	}
	if working.ExerciseID != "back-squat" {
		t.Fatalf("working id = %q, want back-squat", working.ExerciseID)
	}
	if out.SessionContext.TrainingWeekIndex != 1 {
		t.Fatalf("week index = %d, want 1", out.SessionContext.TrainingWeekIndex)
	}
}
