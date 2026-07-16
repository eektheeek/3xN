# Smart Trainer — decision rules (v1)

`lastWorkingWeightKg` = weight kg of the **last set** in the exercise.

| exerciseType | exerciseOutcome | Session context | Next weight (smart trainer) |
|--------------|-----------------|-----------------|----------------------------|
| warmup | (omit) | any | No suggestion |
| working | plan not completed | any | hold at last |
| working | plan completed, not ready to progress | recovery ok | hold at last |
| working | plan completed, ready to progress | deload week | deload intensity, not +step |
| working | plan completed, ready to progress | fatigue/readiness below profile thresholds | hold or decrease, not +step |
| working | plan completed, ready to progress | otherwise | last + `increaseStepKg` |
