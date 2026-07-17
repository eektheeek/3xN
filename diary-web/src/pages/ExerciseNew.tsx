import { useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { PlanDraft } from '../types';
import { PLAN_DRAFT_KEY } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

function hasWorkoutDraft(): boolean {
  try {
    return !!sessionStorage.getItem(PLAN_DRAFT_KEY);
  } catch {
    return false;
  }
}

function appendToDraft(exercise: { id: string; name: string }) {
  try {
    const raw = sessionStorage.getItem(PLAN_DRAFT_KEY);
    const draft: PlanDraft = raw ? JSON.parse(raw) : { name: '', exercises: [] };
    if (!draft.exercises.some((e) => e.id === exercise.id)) {
      draft.exercises.push({ id: exercise.id, name: exercise.name });
      sessionStorage.setItem(PLAN_DRAFT_KEY, JSON.stringify(draft));
    }
  } catch {
    // ignore
  }
}

export function ExerciseNew(_props: RoutableProps) {
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const fromWorkoutBuilder = hasWorkoutDraft();

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
        supportsAssist: data.get('supportsAssist') === 'on',
      });
      if (fromWorkoutBuilder) {
        appendToDraft(ex);
        route('/workouts/new');
      } else {
        route(`/exercises/${ex.id}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div class="page">
      <header class="page-header">
        <a href={fromWorkoutBuilder ? '/workouts/new' : '/'} class="btn btn-ghost">
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
        <label class="field field--checkbox">
          <input name="supportsAssist" type="checkbox" />
          <span>С резинкой (assist)</span>
        </label>
        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting
            ? 'Создание…'
            : fromWorkoutBuilder
              ? 'Создать и добавить в тренировку'
              : 'Создать и задать цель'}
        </button>
      </form>
    </div>
  );
}
