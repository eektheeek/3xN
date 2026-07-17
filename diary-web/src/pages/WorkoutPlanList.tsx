import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Exercise, WorkoutPlan } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

function formatTarget(ex: Exercise): string | null {
  if (!ex.target) return null;
  const parts = [`${ex.target.sets}×${ex.target.reps}`];
  if (ex.target.weightKg > 0) parts.push(`${ex.target.weightKg} кг`);
  if (ex.supportsAssist && ex.target.assistKg > 0) parts.push(`рез. ${ex.target.assistKg} кг`);
  return parts.join(' · ');
}

export function WorkoutPlanList(_props: RoutableProps) {
  const [plans, setPlans] = useState<WorkoutPlan[]>([]);
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([api.listWorkoutPlans(), api.listExercises()])
      .then(([p, e]) => {
        setPlans(p);
        setExercises(e);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div class="page">
      <header class="page-header">
        <h1>Тренировки</h1>
        <div class="page-header__actions">
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
