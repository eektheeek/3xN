import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import type { Cycle, Exercise, WorkoutPlan } from '../types';
import { ACTIVE_CYCLE_KEY } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { SyncStatusBanner } from '../components/SyncStatusBanner';
import { CycleTimeline } from '../components/CycleTimeline';
import { loadHomeCatalogCacheFirst, refreshHomeCatalogFromNetwork } from '../sync/catalog';
import { requestSync } from '../sync/syncWorker';

function formatTarget(ex: Exercise): string | null {
  if (!ex.target) return null;
  if (ex.kind === 'hold') {
    const parts = [`${ex.target.sets}×${ex.target.holdSec}с`];
    if (ex.target.weightKg > 0) parts.push(`${ex.target.weightKg} кг`);
    return parts.join(' · ');
  }
  const parts = [`${ex.target.sets}×${ex.target.reps}`];
  if (ex.target.weightKg > 0) parts.push(`${ex.target.weightKg} кг`);
  if (ex.supportsAssist && ex.target.assistKg > 0) parts.push(`рез. ${ex.target.assistKg} кг`);
  return parts.join(' · ');
}

function writeActiveCycleId(id: string) {
  localStorage.setItem(ACTIVE_CYCLE_KEY, id);
}

export function WorkoutPlanList(_props: RoutableProps) {
  const [plans, setPlans] = useState<WorkoutPlan[]>([]);
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [fromCache, setFromCache] = useState(false);
  const [offline, setOffline] = useState(!navigator.onLine);

  useEffect(() => {
    const onNet = () => {
      setOffline(!navigator.onLine);
      if (navigator.onLine) {
        requestSync();
        void refreshHomeCatalogFromNetwork()
          .then((data) => {
            setPlans(data.plans);
            setExercises(data.exercises);
            setCycles(data.cycles);
            setFromCache(false);
            setError('');
          })
          .catch(() => {
            /* keep cached UI */
          });
      }
    };
    window.addEventListener('online', onNet);
    window.addEventListener('offline', onNet);
    return () => {
      window.removeEventListener('online', onNet);
      window.removeEventListener('offline', onNet);
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const apply = (data: {
        plans: WorkoutPlan[];
        exercises: Exercise[];
        cycles: Cycle[];
        fromCache: boolean;
      }) => {
        if (cancelled) return;
        setPlans(data.plans);
        setExercises(data.exercises);
        setCycles(data.cycles);
        setFromCache(data.fromCache);
        setError('');
        setLoading(false);
      };

      const data = await loadHomeCatalogCacheFirst((fresh) => apply(fresh));
      if (cancelled) return;
      if (data) {
        apply(data);
        return;
      }
      setError('Не удалось загрузить каталог');
      setLoading(false);
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const homeCycles = cycles.filter((c) => c.onHome && !c.completed);

  const startFromCycle = (cycleId: string, planId: string) => {
    writeActiveCycleId(cycleId);
    route(`/workouts/${planId}`);
  };

  return (
    <div class="page">
      <header class="page-header">
        <img src="/logo-3xn.jpg" alt="3xN" class="brand-logo" width="48" height="48" />
        <div class="page-header__actions">
          <a href="/cycles" class="btn btn-secondary">
            Циклы
          </a>
          <a href="/diary" class="btn btn-secondary">
            Дневник
          </a>
          <a href="/protocols" class="btn btn-secondary">
            Табаты
          </a>
        </div>
      </header>

      <SyncStatusBanner
        offline={offline}
        pendingCount={0}
        message={
          fromCache && !offline
            ? 'Показан сохранённый каталог (нет связи с API)'
            : undefined
        }
      />
      <ErrorBanner message={error} />

      {loading && <p class="muted">Загрузка…</p>}

      {!loading && (
        <section style="margin-bottom:1.5rem">
          <h2 class="section-title">Циклы</h2>
          {homeCycles.length === 0 ? (
            <p class="muted">
              На главной пока пусто. В{' '}
              <a href="/cycles">Циклах</a> отметьте нужные или пройдите шаг — отметка появится сама.
            </p>
          ) : (
            <ul class="list">
              {homeCycles.map((cycle) => (
                <li key={cycle.id} class="card" style="padding:1rem;list-style:none">
                  <h3 style="margin:0 0 0.35rem;font-size:1.1rem">{cycle.name}</h3>

                  {cycle.steps.length === 0 ? (
                    <p class="muted" style="margin:0">
                      Нет шагов.{' '}
                      <a href={`/cycles/${cycle.id}`}>Добавьте тренировки</a>.
                    </p>
                  ) : (
                    <CycleTimeline
                      cycle={cycle}
                      onStartCurrent={(step) => startFromCycle(cycle.id, step.workoutPlanId)}
                    />
                  )}
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      <section>
        <div class="section-heading">
          <h2 class="section-title">Мои тренировки</h2>
          <a href="/workouts/new" class="btn btn-secondary btn-icon" aria-label="Создать тренировку">
            +
          </a>
        </div>
        {!loading && plans.length === 0 && (
          <p class="muted">Пока нет сохранённых тренировок. Нажмите +, чтобы собрать первую.</p>
        )}
        <ul class="list">
          {plans.map((plan) => (
            <li key={plan.id}>
              <a href={`/workouts/${plan.id}`} class="list-item">
                <span class="list-item__title">
                  {plan.name}
                  {(offline || fromCache) && (
                    <span class="badge badge--muted" style="margin-left:0.4rem">
                      офлайн
                    </span>
                  )}
                </span>
                <span class="list-item__meta">{plan.exercises?.length ?? 0} упражн.</span>
                {plan.exercises && plan.exercises.length > 0 && (
                  <div class="chip-row">
                    {plan.exercises.map((ex) => (
                      <span key={ex.id} class="chip">
                        {ex.exerciseName}
                      </span>
                    ))}
                  </div>
                )}
              </a>
            </li>
          ))}
        </ul>
      </section>

      <section style="margin-top:1.5rem">
        <div class="section-heading">
          <h2 class="section-title">Каталог упражнений</h2>
          <a href="/exercises/new" class="btn btn-secondary btn-icon" aria-label="Новое упражнение">
            +
          </a>
        </div>
        <p class="muted" style="font-size:0.85rem">
          Цель задаётся на карточке упражнения: для повторов — подходы × повторы, для удержания — подходы × секунды.
        </p>
        {!loading && exercises.length === 0 && (
          <p class="muted">Каталог пуст. Создайте упражнение для использования в тренировках.</p>
        )}
        <ul class="list">
          {exercises.map((ex) => {
            const target = formatTarget(ex);
            return (
              <li key={ex.id}>
                <a href={`/exercises/${ex.id}`} class="list-item">
                  <span class="list-item__title">{ex.name}</span>
                  <span class="list-item__meta">
                    {ex.kind === 'hold' ? 'Удержание' : ex.supportsAssist ? 'Повторы · assist' : 'Повторы'}
                    {ex.muscleGroup ? ` · ${ex.muscleGroup}` : ''}
                  </span>
                  {target ? (
                    <span class="list-item__meta">Цель: {target}</span>
                  ) : (
                    <span class="list-item__meta list-item__meta--warn">Цель не задана</span>
                  )}
                </a>
              </li>
            );
          })}
        </ul>
      </section>

      <p class="debug-link-row">
        <a href="/debug/sync">Sync debug</a>
      </p>
    </div>
  );
}
