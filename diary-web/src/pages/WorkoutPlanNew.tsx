import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Exercise, PlanDraft, PlanDraftExercise } from '../types';
import { PLAN_DRAFT_KEY } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

function loadDraft(): PlanDraft {
  try {
    const raw = sessionStorage.getItem(PLAN_DRAFT_KEY);
    if (raw) return JSON.parse(raw) as PlanDraft;
  } catch {
    // ignore
  }
  return { name: '', exercises: [] };
}

function saveDraft(draft: PlanDraft) {
  sessionStorage.setItem(PLAN_DRAFT_KEY, JSON.stringify(draft));
}

export function WorkoutPlanNew(_props: RoutableProps) {
  const [draft, setDraft] = useState<PlanDraft>(loadDraft);
  const [catalog, setCatalog] = useState<Exercise[]>([]);
  const [showPicker, setShowPicker] = useState(false);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    saveDraft(draft);
  }, [draft]);

  useEffect(() => {
    if (showPicker) {
      api.listExercises().then(setCatalog).catch((e: Error) => setError(e.message));
    }
  }, [showPicker]);

  const addExercise = (ex: Exercise) => {
    if (draft.exercises.some((e) => e.id === ex.id)) return;
    setDraft({
      ...draft,
      exercises: [...draft.exercises, { id: ex.id, name: ex.name }],
    });
    setShowPicker(false);
  };

  const removeExercise = (id: string) => {
    setDraft({
      ...draft,
      exercises: draft.exercises.filter((e) => e.id !== id),
    });
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const name = draft.name.trim();
    if (!name) {
      setError('Укажите название тренировки');
      return;
    }
    if (draft.exercises.length === 0) {
      setError('Добавьте хотя бы одно упражнение');
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      const plan = await api.createWorkoutPlan({
        name,
        exerciseIds: draft.exercises.map((e) => e.id),
      });
      sessionStorage.removeItem(PLAN_DRAFT_KEY);
      route(`/workouts/${plan.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSubmitting(false);
    }
  };

  const available = catalog.filter((ex) => !draft.exercises.some((d) => d.id === ex.id));

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Новая тренировка</h1>
      </header>

      <ErrorBanner message={error} />

      <form class="form" onSubmit={handleSubmit}>
        <label class="field">
          <span>Название</span>
          <input
            type="text"
            required
            placeholder="Ноги + спина"
            value={draft.name}
            onInput={(e) =>
              setDraft({ ...draft, name: (e.target as HTMLInputElement).value })
            }
          />
        </label>

        <h2 class="section-title">Упражнения в тренировке</h2>

        {draft.exercises.length === 0 && (
          <p class="muted">Добавьте упражнения из каталога или создайте новое.</p>
        )}

        {draft.exercises.map((ex: PlanDraftExercise) => (
          <div key={ex.id} class="selected-exercise">
            <span class="selected-exercise__name">{ex.name}</span>
            <button type="button" class="btn-remove" onClick={() => removeExercise(ex.id)} aria-label="Удалить">
              ×
            </button>
          </div>
        ))}

        <div class="actions">
          <button type="button" class="btn btn-secondary btn-block" onClick={() => setShowPicker(!showPicker)}>
            {showPicker ? 'Скрыть каталог' : 'Добавить из каталога'}
          </button>
          <a href="/exercises/new" class="btn btn-secondary btn-block">
            Создать новое упражнение
          </a>
        </div>

        {showPicker && (
          <div class="picker-list">
            {available.length === 0 && <p class="muted">Нет доступных упражнений.</p>}
            <ul class="list">
              {available.map((ex) => (
                <li key={ex.id}>
                  <button type="button" class="list-item" style="width:100%;text-align:left;border:none;cursor:pointer" onClick={() => addExercise(ex)}>
                    <span class="list-item__title">{ex.name}</span>
                    {ex.muscleGroup && <span class="list-item__meta">{ex.muscleGroup}</span>}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}

        <button type="submit" class="btn btn-primary btn-block" disabled={submitting}>
          {submitting ? 'Сохранение…' : 'Сохранить тренировку'}
        </button>
      </form>
    </div>
  );
}
