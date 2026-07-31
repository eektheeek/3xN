import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Cycle } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { CycleTimeline } from '../components/CycleTimeline';
import { cacheCycles } from '../sync/catalog';

function CycleCard({
  cycle,
  togglingId,
  repeatingId,
  onToggleHome,
  onRepeat,
}: {
  cycle: Cycle;
  togglingId: string | null;
  repeatingId: string | null;
  onToggleHome: (cycle: Cycle, e: Event) => void;
  onRepeat: (cycle: Cycle) => void;
}) {
  return (
    <li
      key={cycle.id}
      class={`list-item${cycle.completed ? ' card--cycle-done' : ''}`}
      style="gap:0.65rem"
    >
      {!cycle.completed && (
        <label
          class="chip-row"
          style="align-items:center;gap:0.5rem;margin:0;cursor:pointer"
          onClick={(e) => onToggleHome(cycle, e)}
        >
          <input
            type="checkbox"
            checked={cycle.onHome}
            disabled={togglingId === cycle.id}
            readOnly
          />
          <span style="font-size:0.85rem">
            {cycle.onHome ? 'На главной' : 'Скрыт с главной'}
          </span>
        </label>
      )}
      <div style="display:flex;flex-direction:column;gap:0.35rem">
        <a href={`/cycles/${cycle.id}`} style="text-decoration:none;color:inherit">
          <span class="list-item__title">{cycle.name}</span>
        </a>
        {cycle.steps.length === 0 ? (
          <span class="list-item__meta">Шагов нет</span>
        ) : (
          <CycleTimeline cycle={cycle} />
        )}
        {cycle.completed && cycle.steps.length > 0 && (
          <button
            type="button"
            class="btn btn-primary btn-block"
            style="margin-top:0.35rem"
            disabled={repeatingId === cycle.id}
            onClick={() => onRepeat(cycle)}
          >
            {repeatingId === cycle.id ? '…' : 'Повторить'}
          </button>
        )}
      </div>
    </li>
  );
}

export function CycleList(_props: RoutableProps) {
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [repeatingId, setRepeatingId] = useState<string | null>(null);
  const [newName, setNewName] = useState('');

  const reload = () =>
    api
      .listCycles()
      .then(async (list) => {
        setCycles(list);
        await cacheCycles(list);
      })
      .catch((e: Error) => setError(e.message));

  useEffect(() => {
    reload().finally(() => setLoading(false));
  }, []);

  const active = cycles.filter((c) => !c.completed);
  const done = cycles.filter((c) => c.completed);

  const handleCreate = async (e: Event) => {
    e.preventDefault();
    const name = newName.trim();
    if (!name) {
      setError('Укажите название цикла');
      return;
    }
    setCreating(true);
    setError('');
    try {
      const cycle = await api.createCycle({ name });
      setNewName('');
      route(`/cycles/${cycle.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать цикл');
    } finally {
      setCreating(false);
    }
  };

  const toggleOnHome = async (cycle: Cycle, e: Event) => {
    e.preventDefault();
    e.stopPropagation();
    setTogglingId(cycle.id);
    setError('');
    try {
      const updated = await api.setCycleOnHome(cycle.id, !cycle.onHome);
      setCycles((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось обновить');
    } finally {
      setTogglingId(null);
    }
  };

  const handleRepeat = async (cycle: Cycle) => {
    setRepeatingId(cycle.id);
    setError('');
    try {
      const created = await api.repeatCycle(cycle.id);
      setCycles((prev) => [created, ...prev]);
      route(`/cycles/${created.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось повторить цикл');
    } finally {
      setRepeatingId(null);
    }
  };

  return (
    <div class="page page--with-tabs">
      <header class="page-header">
        <a href="/more" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Циклы</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      <p class="muted" style="font-size:0.85rem;margin-bottom:1rem">
        Активные — на главную. Завершённые сохраняются в истории; «Повторить» создаёт новый прогон.
      </p>

      <form class="form" onSubmit={handleCreate} style="margin-bottom:1.5rem">
        <label class="field">
          <span>Новый цикл</span>
          <input
            type="text"
            value={newName}
            onInput={(e) => setNewName((e.target as HTMLInputElement).value)}
            placeholder="Дом"
            disabled={creating}
          />
        </label>
        <button type="submit" class="btn btn-primary btn-block" disabled={creating}>
          {creating ? 'Создание…' : 'Создать цикл'}
        </button>
      </form>

      {!loading && cycles.length === 0 && (
        <p class="muted">Пока нет циклов. Создайте первый и наберите шаги из «Моих тренировок».</p>
      )}

      {active.length > 0 && (
        <section style="margin-bottom:1.5rem">
          <h2 class="section-title">Активные</h2>
          <ul class="list">
            {active.map((c) => (
              <CycleCard
                key={c.id}
                cycle={c}
                togglingId={togglingId}
                repeatingId={repeatingId}
                onToggleHome={(cycle, e) => void toggleOnHome(cycle, e)}
                onRepeat={(cycle) => void handleRepeat(cycle)}
              />
            ))}
          </ul>
        </section>
      )}

      {done.length > 0 && (
        <section>
          <h2 class="section-title">Завершённые</h2>
          <p class="muted" style="font-size:0.85rem;margin-top:-0.35rem;margin-bottom:0.75rem">
            {done.length}{' '}
            {done.length === 1 ? 'прогон' : done.length < 5 ? 'прогона' : 'прогонов'} в истории
          </p>
          <ul class="list">
            {done.map((c) => (
              <CycleCard
                key={c.id}
                cycle={c}
                togglingId={togglingId}
                repeatingId={repeatingId}
                onToggleHome={(cycle, e) => void toggleOnHome(cycle, e)}
                onRepeat={(cycle) => void handleRepeat(cycle)}
              />
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
