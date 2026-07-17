import type { SetTargetBody } from '../types';

interface TargetFormProps {
  initial: SetTargetBody;
  supportsAssist: boolean;
  onSubmit: (body: SetTargetBody) => void;
  submitting?: boolean;
}

export function TargetForm({ initial, supportsAssist, onSubmit, submitting }: TargetFormProps) {
  const handleSubmit = (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    onSubmit({
      sets: Number(data.get('sets')),
      reps: Number(data.get('reps')),
      weightKg: supportsAssist ? 0 : Number(data.get('weightKg')),
      assistKg: supportsAssist ? Number(data.get('assistKg')) : 0,
    });
  };

  return (
    <form class="card form" onSubmit={handleSubmit}>
      <label class="field">
        <span>Подходы</span>
        <input name="sets" type="number" min="1" required defaultValue={initial.sets} />
      </label>
      <label class="field">
        <span>Повторы</span>
        <input name="reps" type="number" min="1" required defaultValue={initial.reps} />
      </label>
      <label class={`field${supportsAssist ? ' field--inactive' : ''}`}>
        <span>Вес (кг)</span>
        {supportsAssist ? (
          <input
            name="weightKg"
            type="number"
            value={0}
            readOnly
            tabIndex={-1}
            class="input--inactive"
          />
        ) : (
          <input name="weightKg" type="number" min="0" step="0.5" required defaultValue={initial.weightKg} />
        )}
      </label>
      {supportsAssist && (
        <label class="field">
          <span>Резинка (кг помощи)</span>
          <input name="assistKg" type="number" min="0" step="0.5" required defaultValue={initial.assistKg} />
        </label>
      )}
      <button type="submit" class="btn btn-primary" disabled={submitting}>
        {submitting ? 'Сохранение…' : 'Сохранить цель'}
      </button>
    </form>
  );
}
