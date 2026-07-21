import type { ExerciseKind, SetTargetBody } from '../types';

interface TargetFormProps {
  initial: SetTargetBody;
  kind: ExerciseKind;
  supportsAssist: boolean;
  onSubmit: (body: SetTargetBody) => void;
  submitting?: boolean;
}

export function TargetForm({ initial, kind, supportsAssist, onSubmit, submitting }: TargetFormProps) {
  const isHold = kind === 'hold';

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    if (isHold) {
      onSubmit({
        sets: Number(data.get('sets')),
        reps: 0,
        holdSec: Number(data.get('holdSec')),
        weightKg: Number(data.get('weightKg') || 0),
        assistKg: 0,
      });
      return;
    }
    onSubmit({
      sets: Number(data.get('sets')),
      reps: Number(data.get('reps')),
      holdSec: 0,
      weightKg: supportsAssist ? 0 : Number(data.get('weightKg')),
      assistKg: supportsAssist ? Number(data.get('assistKg')) : 0,
    });
  };

  return (
    <form class="card form" onSubmit={handleSubmit}>
      <label class="field">
        <span>Подходы</span>
        <input name="sets" type="number" inputMode="numeric" min="1" required defaultValue={initial.sets} />
      </label>
      {isHold ? (
        <>
          <label class="field">
            <span>Секунды (цель на подход)</span>
            <input
              name="holdSec"
              type="number"
              inputMode="numeric"
              min="1"
              required
              defaultValue={initial.holdSec || 60}
            />
          </label>
          <label class="field">
            <span>Утяжеление (кг)</span>
            <input
              name="weightKg"
              type="number"
              inputMode="decimal"
              min="0"
              step="0.5"
              required
              defaultValue={initial.weightKg}
            />
          </label>
        </>
      ) : (
        <>
          <label class="field">
            <span>Повторы</span>
            <input name="reps" type="number" inputMode="numeric" min="1" required defaultValue={initial.reps} />
          </label>
          <label class={`field${supportsAssist ? ' field--inactive' : ''}`}>
            <span>Вес (кг)</span>
            {supportsAssist ? (
              <input
                name="weightKg"
                type="number"
                inputMode="decimal"
                value={0}
                readOnly
                tabIndex={-1}
                class="input--inactive"
              />
            ) : (
              <input
                name="weightKg"
                type="number"
                inputMode="decimal"
                min="0"
                step="0.5"
                required
                defaultValue={initial.weightKg}
              />
            )}
          </label>
          {supportsAssist && (
            <label class="field">
              <span>Резинка (кг помощи)</span>
              <input
                name="assistKg"
                type="number"
                inputMode="decimal"
                min="0"
                step="0.5"
                required
                defaultValue={initial.assistKg}
              />
            </label>
          )}
        </>
      )}
      <button type="submit" class="btn btn-primary" disabled={submitting}>
        {submitting ? 'Сохранение…' : 'Сохранить цель'}
      </button>
    </form>
  );
}
