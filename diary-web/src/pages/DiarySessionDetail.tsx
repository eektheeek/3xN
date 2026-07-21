import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Exercise, WorkoutSession } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { formatSessionDateTime, formatDurationLabel } from '../utils/dates';
import { formatExerciseStats, formatTargetLabel } from '../utils/workoutStats';

interface DiarySessionDetailProps extends RoutableProps {
  id?: string;
}

export function DiarySessionDetail({ id }: DiarySessionDetailProps) {
  const [session, setSession] = useState<WorkoutSession | null>(null);
  const [targets, setTargets] = useState<Map<string, Exercise>>(new Map());
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    (async () => {
      try {
        const [loaded, catalog] = await Promise.all([
          api.getWorkoutSession(id),
          api.listExercises(),
        ]);
        setSession(loaded);
        setTargets(new Map(catalog.map((ex) => [ex.id, ex])));
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Ошибка загрузки');
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  if (!id) {
    return (
      <div class="page">
        <ErrorBanner message="Не указан id записи" />
      </div>
    );
  }

  return (
    <div class="page">
      <header class="page-header">
        <a href="/diary" class="btn btn-ghost">
          ← Дневник
        </a>
        <h1>Тренировка</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {session && (
        <>
          <p class="diary-detail__date">{formatSessionDateTime(session.performedAt)}</p>
          {session.cycleName && (
            <p class="diary-detail__cycle">
              {session.cycleId ? (
                <a href={`/cycles/${session.cycleId}`}>
                  Цикл: {session.cycleName}
                  {session.cycleStep ? ` · шаг ${session.cycleStep}` : ''}
                </a>
              ) : (
                <>
                  Цикл: {session.cycleName}
                  {session.cycleStep ? ` · шаг ${session.cycleStep}` : ''}
                </>
              )}
            </p>
          )}
          {session.durationSec > 0 && (
            <p class="diary-detail__duration">Длительность: {formatDurationLabel(session.durationSec)}</p>
          )}
          {session.isDeload && <span class="badge badge--muted">Разгрузка</span>}

          {(session.exercises ?? []).map((block) => {
            const exercise = targets.get(block.exerciseId);
            const target = exercise?.target;
            const name = block.exerciseName || exercise?.name || 'Упражнение';
            const sets = block.sets ?? [];
            const kind = block.kind ?? exercise?.kind ?? 'reps';
            const supportsAssist = block.supportsAssist ?? exercise?.supportsAssist ?? false;
            const isHold = kind === 'hold';

            return (
              <section key={block.id} class="exercise-block">
                <div class="exercise-block__head">
                  <h3>{name}</h3>
                </div>
                {target && (
                  <p class="muted exercise-block__goal">
                    Цель: {formatTargetLabel(target, { kind, supportsAssist })}
                  </p>
                )}
                <p class="exercise-block__stats">
                  {formatExerciseStats(sets, {
                    kind,
                    withTonnage: !isHold && !supportsAssist,
                  })}
                </p>
                {sets.map((set) => {
                  const goalMet = isHold
                    ? target != null && set.durationSec >= target.holdSec
                    : target != null && set.reps >= target.reps;
                  return (
                    <div
                      key={set.id}
                      class={`set-row set-row--readonly${goalMet ? ' set-row--done' : ''}`}
                    >
                      <div class="set-row__head">
                        <span class="set-row__label">Подход {set.setNumber}</span>
                        {target && (
                          <span class={`set-row__plan${goalMet ? ' set-row__plan--done' : ''}`}>
                            план {isHold ? `${target.holdSec}с` : target.reps}
                          </span>
                        )}
                      </div>
                      <div class="set-row__fact">
                        <span class="set-row__fact-label">Факт</span>
                        <strong class={goalMet ? 'input--goal-met-text' : undefined}>
                          {isHold ? `${set.durationSec}с` : set.reps}
                        </strong>
                        {isHold && set.weightKg > 0 && (
                          <span class="muted"> · {set.weightKg} кг</span>
                        )}
                        {!isHold && !supportsAssist && (
                          <span class="muted"> · {set.weightKg} кг</span>
                        )}
                        {!isHold && supportsAssist && (
                          <span class="muted"> · рез. {set.assistKg} кг</span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </section>
            );
          })}
        </>
      )}
    </div>
  );
}
