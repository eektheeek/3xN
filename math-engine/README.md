# Math Engine

This folder contains the training math core (Inside-Out stage 1).
It is focused on pure physiology calculations and deterministic rules.
No HTTP, database, or UI logic is allowed here.
First scope: data contracts, validation, and core metric formulas.

## CLI

```bash
cd math-engine
go run ./cmd/core-cli validate-input -f schema/core_input_v1.json
go run ./cmd/core-cli run-analysis -f schema/core_input_v1.json
```

`run-analysis` prints `analysis_result_v1` JSON to stdout (volume, per-exercise e1RM, recommendation for highest-e1RM exercise).
