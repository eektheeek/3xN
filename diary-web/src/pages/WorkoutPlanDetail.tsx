import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { WorkoutPlan } from '../types';
import { clearDraft, readDraft, writeDraft } from '../utils/activeWorkout';
import { ErrorBanner } from '../components/ErrorBanner';

interface WorkoutPlanDetailProps extends RoutableProps {
  id?: string;
}

export function WorkoutPlanDetail({ id }: WorkoutPlanDetailProps) {
  const [plan, setPlan] = useState<WorkoutPlan | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [starting, setStarting] = useState(false);
  const [canResume, setCanResume] = useState(false);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    (async () => {
      try {
        const p = await api.getWorkoutPlan(id);
        if (cancelled) return;
        setPlan(p);

        const draft = readDraft(id);
        if (!draft?.sessionId) return;
        try {
          const session = await api.getWorkoutSession(draft.sessionId);
          if (cancelled) return;
          if (session.durationSec > 0) {
            clearDraft(id);
          } else {
            setCanResume(true);
          }
        } catch {
          clearDraft(id);
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Ошибка загрузки');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const handleStart = async () => {
    if (!id) return;
    setStarting(true);
    setError('');
    try {
      const session = await api.startWorkoutSession({
        performedAt: new Date().toISOString(),
        isDeload: false,
        workoutPlanId: id,
      });
      const startedAt = session.startedAt || new Date().toISOString();
      writeDraft(id, {
        sessionId: session.id,
        planId: id,
        startedAt,
        savedExerciseIds: [],
        logs: [],
      });
      route(`/workouts/${id}/start`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось начать');
    } finally {
      setStarting(false);
    }
  };

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
            {canResume ? (
              <a href={`/workouts/${id}/start`} class="btn btn-primary btn-block">
                Продолжить тренировку
              </a>
            ) : (
              <button
                type="button"
                class="btn btn-primary btn-block"
                disabled={starting || (plan.exercises?.length ?? 0) === 0}
                onClick={() => void handleStart()}
              >
                {starting ? 'Старт…' : 'Начать тренировку'}
              </button>
            )}
          </div>
        </>
      )}
    </div>
  );
}
