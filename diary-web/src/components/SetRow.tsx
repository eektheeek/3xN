import type { CreateSetBody, ExerciseKind } from '../types';

interface SetRowProps {
  setNumber: number;
  value: CreateSetBody;
  kind: ExerciseKind;
  supportsAssist: boolean;
  targetReps?: number;
  targetHoldSec?: number;
  targetAssistKg?: number;
  onChange: (value: CreateSetBody) => void;
}

export function SetRow({
  setNumber,
  value,
  kind,
  supportsAssist,
  targetReps,
  targetHoldSec,
  targetAssistKg,
  onChange,
}: SetRowProps) {
  const isHold = kind === 'hold';

  const update = (field: keyof CreateSetBody, raw: string) => {
    const next = { ...value, [field]: Number(raw) };
    if (supportsAssist && !isHold) next.weightKg = 0;
    if (isHold) {
      next.reps = 0;
      next.assistKg = 0;
    }
    onChange(next);
  };

  const repsMet = !isHold && targetReps != null && value.reps >= targetReps;
  const holdMet = isHold && targetHoldSec != null && value.durationSec >= targetHoldSec;
  const goalMet = isHold ? holdMet : repsMet;
  const planLabel = isHold ? targetHoldSec : targetReps;

  return (
    <div class={`set-row${goalMet ? ' set-row--done' : ''}`}>
      <div class="set-row__head">
        <span class="set-row__label">Подход {setNumber}</span>
        {planLabel != null && (
          <span class={`set-row__plan${goalMet ? ' set-row__plan--done' : ''}`}>
            план {isHold ? `${planLabel}с` : planLabel}
          </span>
        )}
      </div>
      {isHold ? (
        <>
          <label class="field field--inline">
            <span>сек</span>
            <input
              type="number"
              inputMode="numeric"
              min="0"
              value={value.durationSec}
              class={goalMet ? 'input--goal-met' : undefined}
              onInput={(e) => update('durationSec', (e.target as HTMLInputElement).value)}
            />
          </label>
          <label class="field field--inline">
            <span>кг</span>
            <input
              type="number"
              inputMode="decimal"
              min="0"
              step="0.5"
              value={value.weightKg}
              onInput={(e) => update('weightKg', (e.target as HTMLInputElement).value)}
            />
          </label>
        </>
      ) : (
        <>
          <label class="field field--inline">
            <span>Факт</span>
            <input
              type="number"
              inputMode="numeric"
              min="0"
              value={value.reps}
              class={goalMet ? 'input--goal-met' : undefined}
              onInput={(e) => update('reps', (e.target as HTMLInputElement).value)}
            />
          </label>
          <label class={`field field--inline${supportsAssist ? ' field--inactive' : ''}`}>
            <span>кг</span>
            <input
              type="number"
              inputMode="decimal"
              min="0"
              step="0.5"
              value={supportsAssist ? 0 : value.weightKg}
              readOnly={supportsAssist}
              tabIndex={supportsAssist ? -1 : 0}
              class={supportsAssist ? 'input--inactive' : undefined}
              onInput={(e) => {
                if (!supportsAssist) update('weightKg', (e.target as HTMLInputElement).value);
              }}
            />
          </label>
          {supportsAssist && (
            <label class="field field--inline">
              <span>
                рез.
                {targetAssistKg != null && targetAssistKg > 0 && (
                  <span class="field-plan-hint"> ({targetAssistKg})</span>
                )}
              </span>
              <input
                type="number"
                inputMode="decimal"
                min="0"
                step="0.5"
                value={value.assistKg}
                onInput={(e) => update('assistKg', (e.target as HTMLInputElement).value)}
              />
            </label>
          )}
        </>
      )}
    </div>
  );
}
