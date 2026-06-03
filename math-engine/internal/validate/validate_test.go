package validate

import (
	"strings"
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
)

func validInput() contracts.CoreInput {
	return contracts.CoreInput{
		Version: "core_input_v1",
		RawTrainingLog: contracts.RawTrainingLog{
			SessionID:          "s1",
			UserID:             "u1",
			StartedAt:          "2026-05-28T17:00:00Z",
			CompletedAt:        "2026-05-28T18:00:00Z",
			SessionDurationSec: 3600,
			Exercises: []contracts.Exercise{{
				ExerciseID:  "ex1",
				Name:        "Squat",
				MuscleGroup: "legs",
				BlockType:   "working",
				Sets:        []contracts.SetEntry{{SetNumber: 1, WeightKg: 100, Reps: 5}},
			}},
		},
		UserMetrics: contracts.UserMetrics{
			Age: 40, BodyWeightKg: 82, RecoverySensitivity: 0.7,
			FatigueThreshold: 0.96, ReadinessThreshold: 0.75,
		},
		TrainingConstraints: contracts.TrainingConstraints{SessionsPerWeek: 3},
		ProgressionPolicy: contracts.ProgressionPolicy{
			Strategy: "percent_e1rm", TargetPercent: 0.78,
			IncreaseStepKg: 2.5, DecreaseStepKg: 2.5,
		},
		DeloadPolicy: contracts.DeloadPolicy{
			Strategy: "fixed_3_plus_1", LoadWeeks: 3, DeloadWeeks: 1,
			IntensityDropPct: 0.10, VolumeDropPct: 0.40, MinRIR: 3,
		},
		TrainingWeekIndex: 1,
		SessionModifiers: contracts.SessionModifiers{
			FatigueModifier: 0.97, ReadinessModifier: 0.99,
		},
	}
}

func TestValidate_valid(t *testing.T) {
	if err := Validate(validInput()); err != nil {
		t.Fatalf("expected valid input, got: %v", err)
	}
}

func TestValidate_invalid(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*contracts.CoreInput)
	}{
		{"wrong version", func(in *contracts.CoreInput) { in.Version = "v0" }},
		{"missing sessionId", func(in *contracts.CoreInput) { in.RawTrainingLog.SessionID = "" }},
		{"no exercises", func(in *contracts.CoreInput) { in.RawTrainingLog.Exercises = nil }},
		{"completed before start", func(in *contracts.CoreInput) {
			in.RawTrainingLog.StartedAt = "2026-05-28T18:00:00Z"
			in.RawTrainingLog.CompletedAt = "2026-05-28T17:00:00Z"
		}},
		{"bad recovery", func(in *contracts.CoreInput) { in.UserMetrics.RecoverySensitivity = 2 }},
		{"unknown deload preset", func(in *contracts.CoreInput) { in.DeloadPolicy.Strategy = "weekly_auto" }},
		{"load weeks out of range", func(in *contracts.CoreInput) { in.DeloadPolicy.LoadWeeks = 0 }},
		{"missing training week", func(in *contracts.CoreInput) { in.TrainingWeekIndex = 0 }},
		{"fatigue out of range", func(in *contracts.CoreInput) { in.SessionModifiers.FatigueModifier = 1.5 }},
		{"invalid blockType", func(in *contracts.CoreInput) { in.RawTrainingLog.Exercises[0].BlockType = "main" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput()
			tc.mutate(&in)
			if err := Validate(in); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidate_userDeloadPresets(t *testing.T) {
	presets := []contracts.DeloadPolicy{
		{Strategy: "fixed_3_plus_1", LoadWeeks: 3, DeloadWeeks: 1, IntensityDropPct: 0.10, VolumeDropPct: 0.40, MinRIR: 3},
		{Strategy: "fixed_2_plus_1", LoadWeeks: 2, DeloadWeeks: 1, IntensityDropPct: 0.12, VolumeDropPct: 0.35, MinRIR: 3},
		{Strategy: "custom", LoadWeeks: 4, DeloadWeeks: 2, IntensityDropPct: 0.08, VolumeDropPct: 0.30, MinRIR: 4},
	}

	for i, deload := range presets {
		in := validInput()
		in.DeloadPolicy = deload
		if err := Validate(in); err != nil {
			t.Fatalf("preset %d (%s): expected valid, got %v", i, deload.Strategy, err)
		}
	}
}

func TestValidate_intervalProtocol(t *testing.T) {
	in := validInput()
	in.RawTrainingLog.Exercises[0].IntervalProtocol = &contracts.IntervalProtocol{
		ProtocolID: "p1", WorkSec: 0, RestSec: 60, DisplayName: "tabata",
	}
	err := Validate(in)
	if err == nil || !strings.Contains(err.Error(), "WorkSec") {
		t.Fatalf("expected WorkSec validation error, got: %v", err)
	}
}
