import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Exercise, ExerciseKind, IntervalProtocol, SetTargetBody } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { TargetForm } from '../components/TargetForm';
import { tabataRounds } from '../utils/beep';
import { formatTargetLabel } from '../utils/workoutStats';

interface ExerciseDetailProps extends RoutableProps {
  id?: string;
}

export function ExerciseDetail({ id }: ExerciseDetailProps) {
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [protocols, setProtocols] = useState<IntervalProtocol[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [showTargetForm, setShowTargetForm] = useState(false);
  const [editing, setEditing] = useState(false);
  const [editKind, setEditKind] = useState<ExerciseKind>('reps');
  const [submitting, setSubmitting] = useState(false);
  const [savingMeta, setSavingMeta] = useState(false);
  const [savingProtocol, setSavingProtocol] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    Promise.all([api.getExercise(id), api.listIntervalProtocols()])
      .then(([ex, list]) => {
        setExercise(ex);
        setEditKind(ex.kind ?? 'reps');
        setProtocols(list);
        setShowTargetForm(!ex.target);
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleSetTarget = async (body: SetTargetBody) => {
    if (!id) return;
    setSubmitting(true);
    setError('');
    try {
      await api.setTarget(id, body);
      const ex = await api.getExercise(id);
      setExercise(ex);
      setShowTargetForm(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdateMeta = async (e: Event) => {
    e.preventDefault();
    if (!id || !exercise) return;
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    const kind = (String(data.get('kind')) as ExerciseKind) || 'reps';
    setSavingMeta(true);
    setError('');
    try {
      const ex = await api.updateExercise(id, {
        name: String(data.get('name')).trim(),
        muscleGroup: String(data.get('muscleGroup') ?? '').trim(),
        kind,
        supportsAssist: kind === 'reps' && data.get('supportsAssist') === 'on',
        protocolId: exercise.protocolId ?? '',
      });
      setExercise(ex);
      setEditKind(ex.kind ?? 'reps');
      setEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSavingMeta(false);
    }
  };

  const handleAttachProtocol = async (protocolId: string) => {
    if (!id || !exercise) return;
    setSavingProtocol(true);
    setError('');
    try {
      const ex = await api.updateExercise(id, {
        name: exercise.name,
        muscleGroup: exercise.muscleGroup,
        kind: exercise.kind ?? 'reps',
        supportsAssist: exercise.supportsAssist,
        protocolId,
      });
      setExercise(ex);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось привязать табату');
    } finally {
      setSavingProtocol(false);
    }
  };

  if (!id) {
    return (
      <div class="page">
        <ErrorBanner message="Не указан id упражнения" />
      </div>
    );
  }

  const sets = exercise?.target?.sets ?? 0;
  const rounds =
    exercise?.protocol && sets > 0 ? tabataRounds(sets, exercise.protocol.warmupExtra) : null;
  const kind = exercise?.kind ?? 'reps';

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost" onClick={(e) => { e.preventDefault(); history.back(); }}>
          ← Назад
        </a>
        <h1>{exercise?.name ?? 'Упражнение'}</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {exercise && (
        <>
          {!editing && (
            <section class="card">
              <h2 class="card__title">Упражнение</h2>
              <p style="margin:0 0 0.35rem">
                <strong>{exercise.name}</strong>
              </p>
              {exercise.muscleGroup && <p class="muted">{exercise.muscleGroup}</p>}
              <p class="muted" style="margin-bottom:0.75rem">
                {kind === 'hold'
                  ? 'Удержание (статика)'
                  : exercise.supportsAssist
                    ? 'Повторы · с резинкой (assist)'
                    : 'Повторы · классическое'}
              </p>
              <button
                type="button"
                class="btn btn-secondary"
                onClick={() => {
                  setEditKind(kind);
                  setEditing(true);
                }}
              >
                Редактировать
              </button>
            </section>
          )}

          {editing && (
            <form class="card form" onSubmit={handleUpdateMeta}>
              <h2 class="card__title">Редактировать</h2>
              <label class="field">
                <span>Название</span>
                <input name="name" type="text" required defaultValue={exercise.name} />
              </label>
              <label class="field">
                <span>Группа мышц</span>
                <input name="muscleGroup" type="text" defaultValue={exercise.muscleGroup} placeholder="спина" />
              </label>
              <div class="field">
                <span class="field__label">Тип упражнения</span>
                <div class="kind-picker" role="radiogroup" aria-label="Тип упражнения">
                  <button
                    type="button"
                    class={`kind-picker__option${editKind === 'reps' ? ' kind-picker__option--active' : ''}`}
                    aria-pressed={editKind === 'reps'}
                    onClick={() => setEditKind('reps')}
                  >
                    <strong>Повторы</strong>
                    <span>сила, вес</span>
                  </button>
                  <button
                    type="button"
                    class={`kind-picker__option${editKind === 'hold' ? ' kind-picker__option--active' : ''}`}
                    aria-pressed={editKind === 'hold'}
                    onClick={() => setEditKind('hold')}
                  >
                    <strong>Удержание</strong>
                    <span>статика, секунды</span>
                  </button>
                </div>
                <input type="hidden" name="kind" value={editKind} />
              </div>
              {editKind === 'reps' && (
                <label class="field field--checkbox">
                  <input name="supportsAssist" type="checkbox" defaultChecked={exercise.supportsAssist} />
                  <span>С резинкой (assist)</span>
                </label>
              )}
              <div class="actions" style="display:flex;gap:0.5rem;flex-wrap:wrap">
                <button type="submit" class="btn btn-primary" disabled={savingMeta}>
                  {savingMeta ? 'Сохранение…' : 'Сохранить'}
                </button>
                <button type="button" class="btn btn-ghost" disabled={savingMeta} onClick={() => setEditing(false)}>
                  Отмена
                </button>
              </div>
            </form>
          )}

          <section class="card">
            <h2 class="card__title">Табата</h2>
            {exercise.protocol ? (
              <>
                <p style="margin:0 0 0.35rem">
                  <strong>{exercise.protocol.name}</strong>
                </p>
                <p class="muted">
                  подг. {exercise.protocol.prepareSec}с · {exercise.protocol.workSec}с /{' '}
                  {exercise.protocol.restSec}с
                  {exercise.protocol.warmupExtra ? ' · +1 разминка' : ''}
                </p>
                {rounds != null && (
                  <p class="muted">При цели {sets}×… → {rounds} раунд(ов)</p>
                )}
                {!exercise.target && (
                  <p class="muted">Задай цель — от числа подходов зависит число раундов.</p>
                )}
              </>
            ) : (
              <p class="muted">Табата не привязана.</p>
            )}
            <label class="field" style="margin-top:0.75rem">
              <span>Выбрать из библиотеки</span>
              <select
                value={exercise.protocolId ?? ''}
                disabled={savingProtocol || protocols.length === 0}
                onChange={(e) => void handleAttachProtocol((e.target as HTMLSelectElement).value)}
              >
                <option value="">— без табаты —</option>
                {protocols.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name} ({p.prepareSec}+{p.workSec}/{p.restSec}
                    {p.warmupExtra ? ' +разм.' : ''})
                  </option>
                ))}
              </select>
            </label>
            {protocols.length === 0 && (
              <p class="muted" style="font-size:0.85rem">
                Сначала создай протокол в <a href="/protocols">библиотеке табат</a>.
              </p>
            )}
          </section>

          {exercise.target && !showTargetForm && (
            <section class="card">
              <h2 class="card__title">Текущая цель</h2>
              <p>
                {formatTargetLabel(exercise.target, {
                  kind,
                  supportsAssist: exercise.supportsAssist,
                })}
              </p>
              <button type="button" class="btn btn-secondary" onClick={() => setShowTargetForm(true)}>
                Изменить цель
              </button>
            </section>
          )}

          {showTargetForm && (
            <section>
              <h2 class="section-title">{exercise.target ? 'Изменить цель' : 'Задать цель'}</h2>
              <TargetForm
                initial={{
                  sets: exercise.target?.sets ?? 3,
                  reps: exercise.target?.reps ?? 12,
                  holdSec: exercise.target?.holdSec ?? 60,
                  weightKg: exercise.target?.weightKg ?? 0,
                  assistKg: exercise.target?.assistKg ?? 0,
                }}
                kind={kind}
                supportsAssist={exercise.supportsAssist}
                onSubmit={handleSetTarget}
                submitting={submitting}
              />
              {exercise.target && (
                <button type="button" class="btn btn-ghost" onClick={() => setShowTargetForm(false)}>
                  Отмена
                </button>
              )}
            </section>
          )}
        </>
      )}
    </div>
  );
}
