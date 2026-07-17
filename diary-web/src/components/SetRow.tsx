import type { CreateSetBody } from '../types';

interface SetRowProps {
  setNumber: number;
  value: CreateSetBody;
  supportsAssist: boolean;
  targetReps?: number;
  targetAssistKg?: number;
  onChange: (value: CreateSetBody) => void;
}

export function SetRow({
  setNumber,
  value,
  supportsAssist,
  targetReps,
  targetAssistKg,
  onChange,
}: SetRowProps) {
  const update = (field: keyof CreateSetBody, raw: string) => {
    const next = { ...value, [field]: Number(raw) };
    if (supportsAssist) next.weightKg = 0;
    onChange(next);
  };

  const repsMet = targetReps != null && value.reps >= targetReps;

  return (
    <div class="set-row">
      <div class="set-row__head">
        <span class="set-row__label">Подход {setNumber}</span>
        {targetReps != null && (
          <span class={`set-row__plan${repsMet ? ' set-row__plan--done' : ''}`}>
            план {targetReps}
          </span>
        )}
      </div>
      <label class="field field--inline">
        <span>Факт</span>
        <input
          type="number"
          min="0"
          value={value.reps}
          class={repsMet ? 'input--goal-met' : undefined}
          onInput={(e) => update('reps', (e.target as HTMLInputElement).value)}
        />
      </label>
      <label class={`field field--inline${supportsAssist ? ' field--inactive' : ''}`}>
        <span>кг</span>
        <input
          type="number"
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
            min="0"
            step="0.5"
            value={value.assistKg}
            onInput={(e) => update('assistKg', (e.target as HTMLInputElement).value)}
          />
        </label>
      )}
    </div>
  );
}
