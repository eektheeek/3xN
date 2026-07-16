package smarttrainer

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
		switch em.ExerciseType {
		case entities.ExerciseTypeWarmup:
			warmup = em
		case entities.ExerciseTypeWorking:
			working = em
		}
	}
	if warmup == nil || working == nil {
		t.Fatal("expected warmup and working in output")
	}
	if warmup.SmartTrainer != nil {
		t.Fatal("warmup must not have smartTrainer")
	}
	if working.SmartTrainer == nil {
		t.Fatal("working must have smartTrainer")
	}
	st := working.SmartTrainer
	if st.SuggestedNextWeightKg != 102.5 || st.Action != "increase" {
		t.Fatalf("smartTrainer = %.1f %q, want 102.5 increase", st.SuggestedNextWeightKg, st.Action)
	}
	if st.LastWorkingWeightKg != 100 {
		t.Fatalf("last = %.1f, want 100", st.LastWorkingWeightKg)
	}
	if st.Message == "" {
		t.Fatal("expected message for UI")
	}
}
