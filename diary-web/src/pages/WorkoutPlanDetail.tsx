import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import type { Exercise, WorkoutPlan } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { SyncStatusBanner } from '../components/SyncStatusBanner';
import {
  getResumableLocalSession,
  loadPlanBundle,
  startLocalWorkout,
} from '../sync/localWorkout';
import { useSyncBanner } from '../sync/useSyncBanner';

interface WorkoutPlanDetailProps extends RoutableProps {
  id?: string;
}

export function WorkoutPlanDetail({ id }: WorkoutPlanDetailProps) {
  const [plan, setPlan] = useState<WorkoutPlan | null>(null);
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [starting, setStarting] = useState(false);
  const [canResume, setCanResume] = useState(false);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const { offline, pendingCount, syncing, syncError, retry } = useSyncBanner(activeSessionId);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    (async () => {
      const applyBundle = async (bundle: {
        plan: WorkoutPlan;
        exercises: Exercise[];
        fromCache: boolean;
      }) => {
        if (cancelled) return;
        setPlan(bundle.plan);
        setExercises(bundle.exercises);
        setError('');
        setLoading(false);
        const active = await getResumableLocalSession(id);
        if (!cancelled) {
          setCanResume(Boolean(active));
          setActiveSessionId(active?.id ?? null);
        }
      };

      try {
        const bundle = await loadPlanBundle(id, (fresh) => {
          void applyBundle(fresh);
        });
        if (cancelled) return;
        if (!bundle) {
          setError(
            offline
              ? 'План недоступен офлайн. Открой его хотя бы раз с интернетом.'
              : 'Не удалось загрузить тренировку',
          );
          setLoading(false);
          return;
        }
        await applyBundle(bundle);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : 'Ошибка загрузки');
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id, offline]);

  const handleStart = async () => {
    if (!id || !plan) return;
    setStarting(true);
    setError('');
    try {
      const session = await startLocalWorkout({ plan, exercises });
      setActiveSessionId(session.id);
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

      <SyncStatusBanner
        offline={offline}
        pendingCount={pendingCount}
        syncing={syncing}
        message={syncError || undefined}
        onRetry={retry}
      />
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
