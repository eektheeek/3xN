package validate

import (
	"time"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
	"github.com/go-playground/validator/v10"
)

var engine = validator.New()

func init() {
	engine.RegisterStructValidation(validateSessionWindow, contracts.RawTrainingLog{})
}

// Validate checks CoreInput using struct tags (go-playground/validator).
func Validate(input contracts.CoreInput) error {
	return engine.Struct(input)
}

func validateSessionWindow(sl validator.StructLevel) {
	log, ok := sl.Current().Interface().(contracts.RawTrainingLog)
	if !ok {
		return
	}

	started, err1 := time.Parse(time.RFC3339, log.StartedAt)
	completed, err2 := time.Parse(time.RFC3339, log.CompletedAt)
	if err1 != nil || err2 != nil {
		return
	}
	if !completed.After(started) {
		sl.ReportError(log.CompletedAt, "completedAt", "CompletedAt", "gt_session_start", "")
	}
}
