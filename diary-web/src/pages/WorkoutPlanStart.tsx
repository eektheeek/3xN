import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { ActiveWorkoutDraft, CreateSetBody, Exercise } from '../types';
import { ACTIVE_CYCLE_KEY } from '../types';
import { clearDraft, readDraft, writeDraft } from '../utils/activeWorkout';
import { ErrorBanner } from '../components/ErrorBanner';
import { SetRow } from '../components/SetRow';
import { TabataTimer } from '../components/TabataTimer';
import { formatDuration } from '../utils/dates';
import { formatExerciseStats } from '../utils/workoutStats';
import { preloadTabataAudio, unlockAudio } from '../utils/beep';

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

function elapsedSec(startedAt: string): number {
  const t = Date.parse(startedAt);
  if (Number.isNaN(t)) return 0;
  return Math.max(0, Math.floor((Date.now() - t) / 1000));
}

export function WorkoutPlanStart({ id }: WorkoutPlanStartProps) {
  const [logs, setLogs] = useState<ExerciseLog[]>([]);
  const [planName, setPlanName] = useState('');
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [startedAt, setStartedAt] = useState<string | null>(null);
  const [elapsed, setElapsed] = useState(0);
  const [savedExerciseIds, setSavedExerciseIds] = useState<Set<string>>(new Set());
  const [savingIndex, setSavingIndex] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);
  const [finalDurationSec, setFinalDurationSec] = useState(0);
  const [tabataIndex, setTabataIndex] = useState<number | null>(null);

  const started = Boolean(sessionId && startedAt);

  useEffect(() => {
    preloadTabataAudio();
  }, []);

  useEffect(() => {
    if (!id) return;
    (async () => {
      try {
        const plan = await api.getWorkoutPlan(id);
        setPlanName(plan.name);

        const draft = readDraft(id);
        let activeSessionId = draft?.sessionId || null;
        let activeStartedAt = draft?.startedAt || null;
        let savedIds = new Set(draft?.savedExerciseIds ?? []);
        let sessionExercises: { exerciseId: string; sets?: CreateSetBody[] }[] = [];

        if (activeSessionId) {
          try {
            const session = await api.getWorkoutSession(activeSessionId);
            if (session.durationSec > 0) {
              clearDraft(id);
              activeSessionId = null;
              activeStartedAt = null;
              savedIds = new Set();
            } else {
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
              activeStartedAt = session.startedAt || activeStartedAt;
            }
          } catch {
            activeSessionId = null;
            activeStartedAt = null;
            savedIds = new Set();
            sessionExercises = [];
          }
        }

        if (!activeSessionId || !activeStartedAt) {
          route(`/workouts/${id}`, true);
          return;
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
        setStartedAt(activeStartedAt);
        setSavedExerciseIds(savedIds);
        setLogs(items);
        setElapsed(elapsedSec(activeStartedAt));
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка загрузки');
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  useEffect(() => {
    if (!id || loading || !started || !sessionId || !startedAt) return;
    const draft: ActiveWorkoutDraft = {
      sessionId,
      planId: id,
      startedAt,
      savedExerciseIds: [...savedExerciseIds],
      logs: logs.map((log) => ({
        exerciseId: log.exercise.id,
        sets: log.sets,
      })),
    };
    writeDraft(id, draft);
  }, [id, loading, logs, sessionId, startedAt, savedExerciseIds, started]);

  useEffect(() => {
    if (!startedAt || done) return;
    setElapsed(elapsedSec(startedAt));
    const timer = window.setInterval(() => {
      setElapsed(elapsedSec(startedAt));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [startedAt, done]);

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

  const saveExercise = async (logIndex: number) => {
    if (!sessionId) return;
    const log = logs[logIndex];
    setSavingIndex(logIndex);
    setError('');
    try {
      await api.saveSessionExercise(sessionId, log.exercise.id, {
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
    if (!id || !sessionId || !startedAt) return;
    setSubmitting(true);
    setError('');
    try {
      for (let i = 0; i < logs.length; i++) {
        const log = logs[i];
        if (!savedExerciseIds.has(log.exercise.id)) {
          await api.saveSessionExercise(sessionId, log.exercise.id, {
            position: i + 1,
            sets: log.sets,
          });
        }
      }
      const durationSec = elapsedSec(startedAt);
      const cycleId = localStorage.getItem(ACTIVE_CYCLE_KEY) || undefined;
      await api.finishWorkoutSession(sessionId, {
        durationSec,
        cycleId,
      });

      clearDraft(id);
      setFinalDurationSec(durationSec);
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
        {finalDurationSec > 0 && (
          <p class="muted">Время: {formatDuration(finalDurationSec)}</p>
        )}
        <a href="/" class="btn btn-primary">
          На главную
        </a>
        <a href="/diary" class="btn btn-secondary">
          В дневник
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

      {!loading && started && (
        <>
          <div class="workout-timer" aria-live="polite">
            <span class="workout-timer__label">Время</span>
            <span class="workout-timer__value">{formatDuration(elapsed)}</span>
          </div>

          <form class="form" onSubmit={handleSubmit}>
            {logs.map((log, logIndex) => {
              const isSaved = savedExerciseIds.has(log.exercise.id);
              const isSaving = savingIndex === logIndex;
              const protocol = log.exercise.protocol;
              const canTabata = Boolean(protocol && (log.exercise.target?.sets ?? log.sets.length) > 0);
              const tabataOpen = tabataIndex === logIndex && protocol;
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
                  {canTabata && protocol && !tabataOpen && (
                    <button
                      type="button"
                      class="btn btn-secondary btn-block"
                      style="margin-bottom:0.65rem"
                      onClick={() => {
                        unlockAudio();
                        setTabataIndex(logIndex);
                      }}
                    >
                      Старт табаты
                      {protocol.prepareSec > 0 ? ` · подг. ${protocol.prepareSec}с` : ''}
                      {protocol.warmupExtra ? ' · +разминка' : ''} · {protocol.workSec}/{protocol.restSec}с
                    </button>
                  )}
                  {tabataOpen && (
                    <TabataTimer
                      protocol={protocol}
                      sets={log.exercise.target?.sets ?? log.sets.length}
                      onClose={() => setTabataIndex(null)}
                    />
                  )}
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
        </>
      )}
    </div>
  );
}
