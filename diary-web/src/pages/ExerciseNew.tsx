import { useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { ExerciseKind } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

export function ExerciseNew(_props: RoutableProps) {
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [kind, setKind] = useState<ExerciseKind>('reps');

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    setSubmitting(true);
    setError('');
    try {
      const ex = await api.createExercise({
        name: String(data.get('name')).trim(),
        muscleGroup: String(data.get('muscleGroup') ?? '').trim(),
        kind,
        supportsAssist: false,
      });
      route(`/exercises/${ex.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost" onClick={(e) => { e.preventDefault(); history.back(); }}>
          ← Назад
        </a>
        <h1>Новое упражнение</h1>
      </header>

      <ErrorBanner message={error} />

      <form class="card form" onSubmit={handleSubmit}>
        <label class="field">
          <span>Название</span>
          <input name="name" type="text" required placeholder="Подтягивания широким хватом" />
        </label>
        <label class="field">
          <span>Группа мышц</span>
          <input name="muscleGroup" type="text" placeholder="спина" />
        </label>

        <div class="field">
          <span class="field__label">Тип упражнения</span>
          <div class="kind-picker" role="radiogroup" aria-label="Тип упражнения">
            <button
              type="button"
              class={`kind-picker__option${kind === 'reps' ? ' kind-picker__option--active' : ''}`}
              aria-pressed={kind === 'reps'}
              onClick={() => setKind('reps')}
            >
              <strong>Повторы</strong>
              <span>сила, вес</span>
            </button>
            <button
              type="button"
              class={`kind-picker__option${kind === 'hold' ? ' kind-picker__option--active' : ''}`}
              aria-pressed={kind === 'hold'}
              onClick={() => setKind('hold')}
            >
              <strong>Удержание</strong>
              <span>статика, секунды</span>
            </button>
          </div>
          <p class="muted field__hint">
            {kind === 'hold'
              ? 'Цель и лог — в секундах на подход.'
              : 'Ассист (резинка) можно включить позже на карточке упражнения.'}
          </p>
        </div>

        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting ? 'Создание…' : 'Создать'}
        </button>
      </form>
    </div>
  );
}
