import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import type { CreateSetBody, Exercise } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { SyncStatusBanner } from '../components/SyncStatusBanner';
import { SetRow } from '../components/SetRow';
import { TabataTimer } from '../components/TabataTimer';
import { formatDuration } from '../utils/dates';
import { formatExerciseStats, formatTargetLabel } from '../utils/workoutStats';
import { preloadTabataAudio, unlockAudio } from '../utils/beep';
import {
  defaultSetsForExercise,
  finishLocalWorkout,
  getResumableLocalSession,
  persistLocalLogs,
  saveLocalExerciseResult,
} from '../sync/localWorkout';
import { useSyncBanner } from '../sync/useSyncBanner';
import { requestSync } from '../sync/syncWorker';

interface WorkoutPlanStartProps extends RoutableProps {
  id?: string;
}

type ExerciseLog = {
  exercise: Exercise;
  sets: CreateSetBody[];
};

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
  const { offline, pendingCount, syncing, syncError, retry, refresh } = useSyncBanner(sessionId);

  const started = Boolean(sessionId && startedAt);

  useEffect(() => {
    preloadTabataAudio();
  }, []);

  useEffect(() => {
    if (!id) return;
    (async () => {
      try {
        const session = await getResumableLocalSession(id);
        if (!session) {
          route(`/workouts/${id}`, true);
          return;
        }

        const exerciseById = new Map(session.exerciseSnapshots.map((e) => [e.id, e]));
        const items: ExerciseLog[] = [];
        for (const slot of session.planSnapshot.exercises ?? []) {
          const exercise = exerciseById.get(slot.exerciseId);
          if (!exercise) continue;
          const log = session.logs.find((l) => l.exerciseId === exercise.id);
          items.push({
            exercise,
            sets: log?.sets.length ? log.sets : defaultSetsForExercise(exercise),
          });
        }

        setPlanName(session.planSnapshot.name);
        setSessionId(session.id);
        setStartedAt(session.startedAt);
        setLogs(items);
        setElapsed(elapsedSec(session.startedAt));
        requestSync();
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка загрузки');
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  useEffect(() => {
    if (!sessionId || loading || !started) return;
    void persistLocalLogs(
      sessionId,
      logs.map((log, i) => ({
        exerciseId: log.exercise.id,
        position: i + 1,
        sets: log.sets,
      })),
      [...savedExerciseIds],
    ).catch(() => {
      // IndexedDB write failures surface on next explicit save/finish.
    });
  }, [logs, sessionId, loading, started, savedExerciseIds]);

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
      await saveLocalExerciseResult(sessionId, log.exercise.id, logIndex + 1, log.sets);
      setSavedExerciseIds((prev) => new Set(prev).add(log.exercise.id));
      await refresh();
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
          await saveLocalExerciseResult(sessionId, log.exercise.id, i + 1, log.sets);
        }
      }
      const durationSec = elapsedSec(startedAt);
      await finishLocalWorkout(sessionId, durationSec);
      setFinalDurationSec(durationSec);
      setDone(true);
      await refresh();
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
        <p class="muted">
          {pendingCount === 0 && !syncing && !syncError
            ? `Тренировка «${planName}» сохранена.`
            : `Тренировка «${planName}» сохранена на телефоне.`}
        </p>
        {finalDurationSec > 0 && (
          <p class="muted">Время: {formatDuration(finalDurationSec)}</p>
        )}
        <SyncStatusBanner
          offline={offline}
          pendingCount={pendingCount}
          syncing={syncing}
          message={syncError || undefined}
          onRetry={retry}
        />
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

      <SyncStatusBanner
        offline={offline}
        pendingCount={pendingCount}
        syncing={syncing}
        message={syncError || undefined}
        onRetry={retry}
      />
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
                      Цель:{' '}
                      {formatTargetLabel(log.exercise.target, {
                        kind: log.exercise.kind ?? 'reps',
                        supportsAssist: log.exercise.supportsAssist,
                      })}
                    </p>
                  )}
                  {!log.exercise.target && (
                    <p class="muted exercise-block__goal">
                      <a href={`/exercises/${log.exercise.id}`}>Задать цель</a>
                    </p>
                  )}
                  <p class="exercise-block__stats">
                    {formatExerciseStats(log.sets, {
                      kind: log.exercise.kind ?? 'reps',
                      withTonnage:
                        (log.exercise.kind ?? 'reps') !== 'hold' && !log.exercise.supportsAssist,
                    })}
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
                      kind={log.exercise.kind ?? 'reps'}
                      supportsAssist={log.exercise.supportsAssist}
                      targetReps={log.exercise.target?.reps}
                      targetHoldSec={log.exercise.target?.holdSec}
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
