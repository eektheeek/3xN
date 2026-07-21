import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Cycle, Exercise, WorkoutPlan } from '../types';
import { ACTIVE_CYCLE_KEY } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

function formatTarget(ex: Exercise): string | null {
  if (!ex.target) return null;
  const parts = [`${ex.target.sets}×${ex.target.reps}`];
  if (ex.target.weightKg > 0) parts.push(`${ex.target.weightKg} кг`);
  if (ex.supportsAssist && ex.target.assistKg > 0) parts.push(`рез. ${ex.target.assistKg} кг`);
  return parts.join(' · ');
}

function writeActiveCycleId(id: string) {
  localStorage.setItem(ACTIVE_CYCLE_KEY, id);
}

function currentStepOf(cycle: Cycle) {
  if (cycle.completed || cycle.steps.length === 0) return null;
  return cycle.steps.find((s) => s.position === cycle.currentStep) ?? null;
}

export function WorkoutPlanList(_props: RoutableProps) {
  const [plans, setPlans] = useState<WorkoutPlan[]>([]);
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [restartingId, setRestartingId] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.listWorkoutPlans(), api.listExercises(), api.listCycles()])
      .then(([p, e, c]) => {
        setPlans(p);
        setExercises(e);
        setCycles(c);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const handleRestart = async (cycleId: string) => {
    setError('');
    setRestartingId(cycleId);
    try {
      const updated = await api.restartCycle(cycleId);
      setCycles((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось перезапустить цикл');
    } finally {
      setRestartingId(null);
    }
  };

  const startFromCycle = (cycleId: string, planId: string) => {
    writeActiveCycleId(cycleId);
    route(`/workouts/${planId}`);
  };

  return (
    <div class="page">
      <header class="page-header">
        <h1>Тренировки</h1>
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
          <a href="/workouts/new" class="btn btn-primary btn-icon" aria-label="Создать тренировку">
            +
          </a>
        </div>
      </header>

      <ErrorBanner message={error} />

      {loading && <p class="muted">Загрузка…</p>}

      {!loading && (
        <section style="margin-bottom:1.5rem">
          <h2 class="section-title">Циклы</h2>
          {cycles.filter((c) => c.onHome).length === 0 ? (
            <p class="muted">
              На главной пока пусто. В{' '}
              <a href="/cycles">Циклах</a> отметьте нужные или пройдите шаг — отметка появится сама.
            </p>
          ) : (
            <ul class="list">
              {cycles
                .filter((c) => c.onHome)
                .map((cycle) => {
                const step = currentStepOf(cycle);
                return (
                  <li
                    key={cycle.id}
                    class={`card${cycle.completed ? ' card--cycle-done' : ''}`}
                    style="padding:1rem;list-style:none"
                  >
                    <h3 style="margin:0 0 0.5rem;font-size:1.1rem">{cycle.name}</h3>

                    {cycle.steps.length === 0 ? (
                      <p class="muted" style="margin:0">
                        Нет шагов.{' '}
                        <a href={`/cycles/${cycle.id}`}>Добавьте тренировки</a>.
                      </p>
                    ) : cycle.completed ? (
                      <>
                        <p class="muted" style="margin:0 0 0.75rem">
                          Завершён · {cycle.steps.length} шагов
                        </p>
                        <button
                          type="button"
                          class="btn btn-secondary btn-block"
                          disabled={restartingId === cycle.id}
                          onClick={() => void handleRestart(cycle.id)}
                        >
                          {restartingId === cycle.id ? '…' : 'Сначала'}
                        </button>
                      </>
                    ) : step ? (
                      <>
                        <p class="muted" style="margin:0 0 0.35rem">
                          Шаг {cycle.currentStep} из {cycle.steps.length}
                        </p>
                        <p style="margin:0 0 0.75rem;font-size:1.15rem;font-weight:600">
                          {step.workoutPlanName}
                        </p>
                        <button
                          type="button"
                          class="btn btn-primary btn-block"
                          onClick={() => startFromCycle(cycle.id, step.workoutPlanId)}
                        >
                          Начать
                        </button>
                      </>
                    ) : null}
                  </li>
                );
              })}
            </ul>
          )}
        </section>
      )}

      <section>
        <h2 class="section-title">Мои тренировки</h2>
        {!loading && plans.length === 0 && (
          <p class="muted">Пока нет сохранённых тренировок. Нажмите +, чтобы собрать первую.</p>
        )}
        <ul class="list">
          {plans.map((plan) => (
            <li key={plan.id}>
              <a href={`/workouts/${plan.id}`} class="list-item">
                <span class="list-item__title">{plan.name}</span>
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
        <div class="page-header" style="margin-bottom:0.5rem">
          <h2 class="section-title" style="margin:0;flex:1">Каталог упражнений</h2>
          <a href="/exercises/new" class="btn btn-secondary btn-icon" aria-label="Новое упражнение">
            +
          </a>
        </div>
        <p class="muted" style="font-size:0.85rem">
          Цель (подходы × повторы) задаётся на карточке упражнения — при записи тренировки видно «план» и факт.
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
                  {ex.muscleGroup && <span class="list-item__meta">{ex.muscleGroup}</span>}
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
    </div>
  );
}
