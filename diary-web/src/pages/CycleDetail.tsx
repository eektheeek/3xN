import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Cycle, WorkoutPlan } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

interface CycleDetailProps extends RoutableProps {
  id?: string;
}

type DraftStep = {
  workoutPlanId: string;
  workoutPlanName: string;
};

export function CycleDetail({ id }: CycleDetailProps) {
  const [cycle, setCycle] = useState<Cycle | null>(null);
  const [plans, setPlans] = useState<WorkoutPlan[]>([]);
  const [draft, setDraft] = useState<DraftStep[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showPicker, setShowPicker] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [savedFlash, setSavedFlash] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    Promise.all([api.getCycle(id), api.listWorkoutPlans()])
      .then(([c, p]) => {
        setCycle(c);
        setPlans(p);
        setDraft(
          c.steps.map((s) => ({
            workoutPlanId: s.workoutPlanId,
            workoutPlanName: s.workoutPlanName ?? s.workoutPlanId,
          })),
        );
        setDirty(false);
        setSavedFlash(false);
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  const move = (index: number, delta: number) => {
    const next = index + delta;
    if (next < 0 || next >= draft.length) return;
    const copy = [...draft];
    const tmp = copy[index]!;
    copy[index] = copy[next]!;
    copy[next] = tmp;
    setDraft(copy);
    setDirty(true);
    setSavedFlash(false);
  };

  const removeAt = (index: number) => {
    setDraft(draft.filter((_, i) => i !== index));
    setDirty(true);
    setSavedFlash(false);
  };

  const addPlan = (plan: WorkoutPlan) => {
    setDraft([
      ...draft,
      { workoutPlanId: plan.id, workoutPlanName: plan.name },
    ]);
    setShowPicker(false);
    setDirty(true);
    setSavedFlash(false);
  };

  const handleSave = async () => {
    if (!id || !dirty || saving) return;
    setSaving(true);
    setError('');
    try {
      const updated = await api.replaceCycleSteps(id, {
        workoutPlanIds: draft.map((s) => s.workoutPlanId),
      });
      setCycle(updated);
      setDraft(
        updated.steps.map((s) => ({
          workoutPlanId: s.workoutPlanId,
          workoutPlanName: s.workoutPlanName ?? s.workoutPlanId,
        })),
      );
      setDirty(false);
      setSavedFlash(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSaving(false);
    }
  };
  return (
    <div class="page">
      <header class="page-header">
        <a href="/cycles" class="btn btn-ghost">
          ← Циклы
        </a>
        <h1>{cycle?.name ?? 'Цикл'}</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {!loading && cycle && (
        <>
          <p class="muted" style="font-size:0.85rem;margin-bottom:1rem">
            {draft.length === 0
              ? 'Добавьте тренировки из библиотеки — порядок = шаги цикла.'
              : `Сейчас курсор: шаг ${cycle.currentStep} из ${Math.max(draft.length, cycle.steps.length) || 1}.`}
            {dirty ? ' Есть несохранённые изменения.' : ''}
          </p>

          <ul class="list">
            {draft.map((step, index) => (
              <li key={`${step.workoutPlanId}-${index}`} class="list-item">
                <span class="list-item__title">
                  Шаг {index + 1}. {step.workoutPlanName}
                </span>
                <div class="chip-row" style="margin-top:0.5rem">
                  <button
                    type="button"
                    class="btn btn-secondary"
                    disabled={index === 0}
                    onClick={() => move(index, -1)}
                  >
                    ↑
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary"
                    disabled={index === draft.length - 1}
                    onClick={() => move(index, 1)}
                  >
                    ↓
                  </button>
                  <button type="button" class="btn btn-ghost" onClick={() => removeAt(index)}>
                    Убрать
                  </button>
                </div>
              </li>
            ))}
          </ul>

          {draft.length === 0 && (
            <p class="muted">Пока шагов нет.</p>
          )}

          {!showPicker ? (
            <button
              type="button"
              class="btn btn-secondary btn-block"
              style="margin-top:1rem"
              onClick={() => setShowPicker(true)}
            >
              + Добавить тренировку
            </button>
          ) : (
            <section style="margin-top:1rem">
              <div class="page-header" style="margin-bottom:0.5rem">
                <h2 class="section-title" style="margin:0;flex:1">
                  Выберите тренировку
                </h2>
                <button type="button" class="btn btn-ghost" onClick={() => setShowPicker(false)}>
                  Отмена
                </button>
              </div>
              {plans.length === 0 ? (
                <p class="muted">
                  Нет доступных шаблонов.{' '}
                  <a href="/workouts/new">Соберите тренировку</a> в библиотеке.
                </p>
              ) : (
                <ul class="list">
                  {plans.map((plan) => (
                    <li key={plan.id}>
                      <button
                        type="button"
                        class="list-item"
                        style="width:100%;text-align:left;cursor:pointer;border:none;background:inherit"
                        onClick={() => addPlan(plan)}
                      >
                        <span class="list-item__title">{plan.name}</span>
                        <span class="list-item__meta">
                          {plan.exercises?.length ?? 0} упражн.
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          )}

          <button
            type="button"
            class={`btn btn-block${dirty || saving ? ' btn-primary' : ' btn-secondary'}`}
            style="margin-top:1rem"
            disabled={!dirty || saving}
            onClick={() => void handleSave()}
          >
            {saving ? 'Сохранение…' : dirty ? 'Сохранить стек' : savedFlash ? 'Сохранено' : 'Сохранить стек'}
          </button>
        </>
      )}
    </div>
  );
}
