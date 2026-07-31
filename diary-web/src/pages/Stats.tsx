import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Exercise, ExerciseStats, StatsPeriod } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { VolumeChart } from '../components/VolumeChart';
import { formatVolumeLabel } from '../utils/workoutStats';

export function Stats(_props: RoutableProps) {
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [exerciseId, setExerciseId] = useState('');
  const [period, setPeriod] = useState<StatsPeriod>('30d');
  const [stats, setStats] = useState<ExerciseStats | null>(null);
  const [error, setError] = useState('');
  const [loadingList, setLoadingList] = useState(true);
  const [loadingStats, setLoadingStats] = useState(false);

  useEffect(() => {
    api
      .listExercises()
      .then((list) => {
        setExercises(list);
        if (list.length > 0) setExerciseId(list[0].id);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoadingList(false));
  }, []);

  useEffect(() => {
    if (!exerciseId) {
      setStats(null);
      return;
    }
    let cancelled = false;
    setLoadingStats(true);
    setError('');
    api
      .getExerciseStats(exerciseId, period)
      .then((data) => {
        if (!cancelled) setStats(data);
      })
      .catch((err: Error) => {
        if (!cancelled) {
          setStats(null);
          setError(err.message);
        }
      })
      .finally(() => {
        if (!cancelled) setLoadingStats(false);
      });
    return () => {
      cancelled = true;
    };
  }, [exerciseId, period]);

  const kind = stats?.kind ?? exercises.find((e) => e.id === exerciseId)?.kind ?? 'reps';
  const volumeUnit = kind === 'hold' ? 'время' : 'повторы';

  return (
    <div class="page page--with-tabs">
      <header class="page-header">
        <h1>Статистика</h1>
      </header>

      <ErrorBanner message={error} />

      {loadingList && <p class="muted">Загрузка…</p>}

      {!loadingList && exercises.length === 0 && (
        <p class="muted">Сначала добавьте упражнения в каталог.</p>
      )}

      {exercises.length > 0 && (
        <>
          <label class="field">
            <span class="field__label">Упражнение</span>
            <select
              value={exerciseId}
              onChange={(e) => setExerciseId((e.target as HTMLSelectElement).value)}
            >
              {exercises.map((ex) => (
                <option key={ex.id} value={ex.id}>
                  {ex.name}
                </option>
              ))}
            </select>
          </label>

          <div class="period-toggle" role="group" aria-label="Период">
            <button
              type="button"
              class={`period-toggle__btn${period === '30d' ? ' period-toggle__btn--active' : ''}`}
              onClick={() => setPeriod('30d')}
            >
              30 дней
            </button>
            <button
              type="button"
              class={`period-toggle__btn${period === 'all' ? ' period-toggle__btn--active' : ''}`}
              onClick={() => setPeriod('all')}
            >
              Всё
            </button>
          </div>

          {loadingStats && <p class="muted">Считаем…</p>}

          {!loadingStats && stats && (
            <>
              <div class="stats-cards">
                <div class="stats-card">
                  <span class="stats-card__value">{stats.summary.sessionCount}</span>
                  <span class="stats-card__label">тренировок</span>
                </div>
                <div class="stats-card">
                  <span class="stats-card__value">{stats.summary.avgSets}</span>
                  <span class="stats-card__label">ср. подходов</span>
                </div>
                <div class="stats-card">
                  <span class="stats-card__value">
                    {formatVolumeLabel(stats.summary.totalVolume, kind)}
                  </span>
                  <span class="stats-card__label">сумма ({volumeUnit})</span>
                </div>
              </div>

              {stats.points.length === 0 ? (
                <p class="muted">Пока нет тренировок с этим упражнением.</p>
              ) : (
                <section class="card">
                  <h2 class="section-title">Объём по тренировкам</h2>
                  <VolumeChart points={stats.points} kind={kind} />                </section>
              )}
            </>
          )}
        </>
      )}
    </div>
  );
}
