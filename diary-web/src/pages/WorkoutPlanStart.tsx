import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { ActiveWorkoutDraft, CreateSetBody, Exercise } from '../types';
import { activeWorkoutKey } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { SetRow } from '../components/SetRow';
import { formatExerciseStats } from '../utils/workoutStats';

interface WorkoutPlanStartProps extends RoutableProps {
  id?: string;
}

type ExerciseLog = {
  exercise: Exercise;
  sets: CreateSetBody[];
};

function defaultSets(exercise: Exercise): CreateSetBody[] {
  const n = exercise.target?.sets ?? 3;
  const reps = exercise.target?.reps ?? 12;
  const weightKg = exercise.target?.weightKg ?? 0;
  const assistKg = exercise.target?.assistKg ?? 0;
  return Array.from({ length: n }, (_, i) => ({
    setNumber: i + 1,
    reps,
    weightKg,
    assistKg,
  }));
}

function setsFromSession(
  exercise: Exercise,
  savedSets: { setNumber: number; reps: number; weightKg: number; assistKg: number }[],
): CreateSetBody[] {
  if (savedSets.length === 0) return defaultSets(exercise);
  return savedSets.map((s) => ({
    setNumber: s.setNumber,
    reps: s.reps,
    weightKg: s.weightKg,
    assistKg: s.assistKg,
  }));
}

function readDraft(planId: string): ActiveWorkoutDraft | null {
  try {
    const raw = localStorage.getItem(activeWorkoutKey(planId));
    if (!raw) return null;
    return JSON.parse(raw) as ActiveWorkoutDraft;
  } catch {
    return null;
  }
}

function writeDraft(planId: string, draft: ActiveWorkoutDraft) {
  localStorage.setItem(activeWorkoutKey(planId), JSON.stringify(draft));
}

function clearDraft(planId: string) {
  localStorage.removeItem(activeWorkoutKey(planId));
}

export function WorkoutPlanStart({ id }: WorkoutPlanStartProps) {
  const [logs, setLogs] = useState<ExerciseLog[]>([]);
  const [planName, setPlanName] = useState('');
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [savedExerciseIds, setSavedExerciseIds] = useState<Set<string>>(new Set());
  const [savingIndex, setSavingIndex] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (!id) return;
    (async () => {
      try {
        const plan = await api.getWorkoutPlan(id);
        setPlanName(plan.name);

        const draft = readDraft(id);
        let activeSessionId = draft?.sessionId ?? null;
        let savedIds = new Set(draft?.savedExerciseIds ?? []);
        let sessionExercises: { exerciseId: string; sets?: CreateSetBody[] }[] = [];

        if (activeSessionId) {
          try {
            const session = await api.getWorkoutSession(activeSessionId);
            sessionExercises = (session.exercises ?? []).map((ex) => ({
              exerciseId: ex.exerciseId,
              sets: ex.sets?.map((s) => ({
                setNumber: s.setNumber,
                reps: s.reps,
                weightKg: s.weightKg,
                assistKg: s.assistKg,
              })),
            }));
            savedIds = new Set(sessionExercises.map((ex) => ex.exerciseId));
          } catch {
            activeSessionId = null;
            savedIds = new Set();
            sessionExercises = [];
          }
        }

        const items: ExerciseLog[] = [];
        for (const slot of plan.exercises ?? []) {
          const exercise = await api.getExercise(slot.exerciseId);
          const draftLog = draft?.logs.find((l) => l.exerciseId === exercise.id);
          const sessionEx = sessionExercises.find((ex) => ex.exerciseId === exercise.id);

          let sets = defaultSets(exercise);
          if (draftLog?.sets.length) {
            sets = draftLog.sets;
          } else if (sessionEx?.sets?.length) {
            sets = setsFromSession(exercise, sessionEx.sets);
          }

          items.push({ exercise, sets });
        }

        setSessionId(activeSessionId);
        setSavedExerciseIds(savedIds);
        setLogs(items);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка загрузки');
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  useEffect(() => {
    if (!id || loading || logs.length === 0) return;
    const draft: ActiveWorkoutDraft = {
      sessionId: sessionId ?? '',
      planId: id,
      savedExerciseIds: [...savedExerciseIds],
      logs: logs.map((log) => ({
        exerciseId: log.exercise.id,
        sets: log.sets,
      })),
    };
    writeDraft(id, draft);
  }, [id, loading, logs, sessionId, savedExerciseIds]);

  const updateSets = (index: number, sets: CreateSetBody[]) => {
    const copy = [...logs];
    copy[index] = { ...copy[index], sets };
    setLogs(copy);
    setSavedExerciseIds((prev) => {
      const next = new Set(prev);
      next.delete(copy[index].exercise.id);
      return next;
    });
  };

  const ensureSession = async (): Promise<string> => {
    if (sessionId) return sessionId;
    const session = await api.startWorkoutSession({
      performedAt: new Date().toISOString(),
      isDeload: false,
    });
    setSessionId(session.id);
    return session.id;
  };

  const saveExercise = async (logIndex: number) => {
    if (!id) return;
    const log = logs[logIndex];
    setSavingIndex(logIndex);
    setError('');
    try {
      const sid = await ensureSession();
      await api.saveSessionExercise(sid, log.exercise.id, {
        position: logIndex + 1,
        sets: log.sets,
      });
      setSavedExerciseIds((prev) => new Set(prev).add(log.exercise.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSavingIndex(null);
    }
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!id) return;
    setSubmitting(true);
    setError('');
    try {
      const sid = await ensureSession();
      for (let i = 0; i < logs.length; i++) {
        const log = logs[i];
        if (!savedExerciseIds.has(log.exercise.id)) {
          await api.saveSessionExercise(sid, log.exercise.id, {
            position: i + 1,
            sets: log.sets,
          });
        }
      }
      clearDraft(id);
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSubmitting(false);
    }
  };

  if (!id) {
    return (
      <div class="page">
        <ErrorBanner message="Не указан id тренировки" />
      </div>
    );
  }

  if (done) {
    return (
      <div class="page page--center">
        <h1>Готово</h1>
        <p class="muted">Тренировка «{planName}» сохранена.</p>
        <a href="/" class="btn btn-primary">
          На главную
        </a>
      </div>
    );
  }

  return (
    <div class="page">
      <header class="page-header">
        <a href={`/workouts/${id}`} class="btn btn-ghost">
          ← Назад
        </a>
        <h1>{planName || 'Тренировка'}</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {!loading && (
        <form class="form" onSubmit={handleSubmit}>
          {logs.map((log, logIndex) => {
            const isSaved = savedExerciseIds.has(log.exercise.id);
            const isSaving = savingIndex === logIndex;
            return (
              <section key={log.exercise.id} class="exercise-block">
                <div class="exercise-block__head">
                  <h3>{log.exercise.name}</h3>
                  {isSaved && <span class="badge badge--ok">Сохранено</span>}
                </div>
                {log.exercise.target && (
                  <p class="muted exercise-block__goal">
                    Цель: {log.exercise.target.sets}×{log.exercise.target.reps}
                    {log.exercise.target.assistKg > 0
                      ? ` · резинка ${log.exercise.target.assistKg} кг`
                      : ''}
                  </p>
                )}
                {!log.exercise.target && (
                  <p class="muted exercise-block__goal">
                    <a href={`/exercises/${log.exercise.id}`}>Задать цель</a>
                  </p>
                )}
                <p class="exercise-block__stats">
                  {formatExerciseStats(log.sets, !log.exercise.supportsAssist)}
                </p>
                {log.sets.map((s, setIndex) => (
                  <SetRow
                    key={s.setNumber}
                    setNumber={s.setNumber}
                    value={s}
                    supportsAssist={log.exercise.supportsAssist}
                    targetReps={log.exercise.target?.reps}
                    targetAssistKg={log.exercise.target?.assistKg}
                    onChange={(next) => {
                      const sets = [...log.sets];
                      sets[setIndex] = next;
                      updateSets(logIndex, sets);
                    }}
                  />
                ))}
                <button
                  type="button"
                  class="btn btn-secondary btn-block"
                  disabled={isSaving}
                  onClick={() => void saveExercise(logIndex)}
                >
                  {isSaving ? 'Сохранение…' : isSaved ? 'Обновить результат' : 'Сохранить результат'}
                </button>
              </section>
            );
          })}
          <button type="submit" class="btn btn-primary btn-block" disabled={submitting || logs.length === 0}>
            {submitting ? 'Сохранение…' : 'Завершить тренировку'}
          </button>
        </form>
      )}
    </div>
  );
}
