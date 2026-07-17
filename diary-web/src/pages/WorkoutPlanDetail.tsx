import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { WorkoutPlan } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

interface WorkoutPlanDetailProps extends RoutableProps {
  id?: string;
}

export function WorkoutPlanDetail({ id }: WorkoutPlanDetailProps) {
  const [plan, setPlan] = useState<WorkoutPlan | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    api
      .getWorkoutPlan(id)
      .then(setPlan)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (!id) {
    return (
      <div class="page">
        <ErrorBanner message="Не указан id тренировки" />
      </div>
    );
  }

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>{plan?.name ?? 'Тренировка'}</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {plan && (
        <>
          <section class="card">
            <h2 class="card__title">Состав</h2>
            <ol style="margin:0;padding-left:1.25rem">
              {plan.exercises?.map((ex) => (
                <li key={ex.id} style="margin-bottom:0.35rem">
                  {ex.exerciseName}
                </li>
              ))}
            </ol>
          </section>

          <div class="actions">
            <a href={`/workouts/${id}/start`} class="btn btn-primary btn-block">
              Открыть тренировку
            </a>
          </div>
        </>
      )}
    </div>
  );
}
