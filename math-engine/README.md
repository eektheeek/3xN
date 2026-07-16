# Math Engine

Training math core (Inside-Out). Smart trainer for next working weights; no e1RM in v1.

## CLI

```bash
cd math-engine
go run ./cmd/core-cli validate-input -f schema/core_input_v1.json
go run ./cmd/core-cli run-analysis -f schema/core_input_v1.json
```

`run-analysis` calls `smarttrainer.Run` and prints `analysis_result_v1` with per-exercise `smartTrainer` (working only) and session volume.
