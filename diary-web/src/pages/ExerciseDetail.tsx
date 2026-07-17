import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Exercise, SetTargetBody } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { TargetForm } from '../components/TargetForm';

interface ExerciseDetailProps extends RoutableProps {
  id?: string;
}

export function ExerciseDetail({ id }: ExerciseDetailProps) {
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [showTargetForm, setShowTargetForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api
      .getExercise(id)
      .then((ex) => {
        setExercise(ex);
        setShowTargetForm(!ex.target);
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleSetTarget = async (body: SetTargetBody) => {
    if (!id) return;
    setSubmitting(true);
    setError('');
    try {
      await api.setTarget(id, body);
      const ex = await api.getExercise(id);
      setExercise(ex);
      setShowTargetForm(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSubmitting(false);
    }
  };

  if (!id) {
    return (
      <div class="page">
        <ErrorBanner message="Не указан id упражнения" />
      </div>
    );
  }

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost" onClick={(e) => { e.preventDefault(); history.back(); }}>
          ← Назад
        </a>
        <h1>{exercise?.name ?? 'Упражнение'}</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {exercise && (
        <>
          {exercise.muscleGroup && <p class="muted">{exercise.muscleGroup}</p>}

          {exercise.target && !showTargetForm && (
            <section class="card">
              <h2 class="card__title">Текущая цель</h2>
              <p>
                {exercise.target.sets}×{exercise.target.reps} · {exercise.target.weightKg} кг
                {exercise.supportsAssist && exercise.target.assistKg > 0
                  ? ` · резинка ${exercise.target.assistKg} кг`
                  : ''}
              </p>
              <button type="button" class="btn btn-secondary" onClick={() => setShowTargetForm(true)}>
                Изменить цель
              </button>
            </section>
          )}

          {showTargetForm && (
            <section>
              <h2 class="section-title">{exercise.target ? 'Изменить цель' : 'Задать цель'}</h2>
              <TargetForm
                initial={{
                  sets: exercise.target?.sets ?? 3,
                  reps: exercise.target?.reps ?? 12,
                  weightKg: exercise.target?.weightKg ?? 0,
                  assistKg: exercise.target?.assistKg ?? 0,
                }}
                supportsAssist={exercise.supportsAssist}
                onSubmit={handleSetTarget}
                submitting={submitting}
              />
              {exercise.target && (
                <button type="button" class="btn btn-ghost" onClick={() => setShowTargetForm(false)}>
                  Отмена
                </button>
              )}
            </section>
          )}
        </>
      )}
    </div>
  );
}
