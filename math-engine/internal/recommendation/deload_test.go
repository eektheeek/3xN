package recommendation

import (
	"testing"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/entities"
)

func policy3plus1() entities.DeloadPolicy {
	return entities.DeloadPolicy{
		Strategy:         "fixed_3_plus_1",
		LoadWeeks:        3,
		DeloadWeeks:      1,
		IntensityDropPct: 0.10,
		VolumeDropPct:    0.40,
		MinRIR:           3,
	}
}

func TestIsDeloadWeek_fixed3plus1(t *testing.T) {
	p := policy3plus1()
	for _, tc := range []struct {
		week int
		want bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, true},
		{5, false},
		{7, false},
		{8, true},
	} {
		if got := IsDeloadWeek(p, tc.week); got != tc.want {
			t.Errorf("week %d: got %v, want %v", tc.week, got, tc.want)
		}
	}
}

func TestIsDeloadWeek_fixed2plus1(t *testing.T) {
	p := entities.DeloadPolicy{Strategy: "fixed_2_plus_1", LoadWeeks: 2, DeloadWeeks: 1}
	cases := map[int]bool{1: false, 2: false, 3: true, 4: false, 6: true}
	for week, want := range cases {
		if got := IsDeloadWeek(p, week); got != want {
			t.Errorf("week %d: got %v, want %v", week, got, want)
		}
	}
}

func TestIsDeloadWeek_custom(t *testing.T) {
	p := entities.DeloadPolicy{
		Strategy:    "custom",
		LoadWeeks:   4,
		DeloadWeeks: 2,
	}
	if IsDeloadWeek(p, 4) {
		t.Fatal("week 4 should be load")
	}
	if !IsDeloadWeek(p, 5) {
		t.Fatal("week 5 should be deload")
	}
	if !IsDeloadWeek(p, 6) {
		t.Fatal("week 6 should be deload")
	}
	if IsDeloadWeek(p, 7) {
		t.Fatal("week 7 should restart cycle (load)")
	}
}

func TestIsDeloadWeek_invalidWeek(t *testing.T) {
	if IsDeloadWeek(policy3plus1(), 0) {
		t.Fatal("week 0 should not be deload")
	}
}
