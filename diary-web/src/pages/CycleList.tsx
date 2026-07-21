import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { Cycle } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

export function CycleList(_props: RoutableProps) {
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [newName, setNewName] = useState('');

  const reload = () =>
    api
      .listCycles()
      .then(setCycles)
      .catch((e: Error) => setError(e.message));

  useEffect(() => {
    reload().finally(() => setLoading(false));
  }, []);

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

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Циклы</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      <p class="muted" style="font-size:0.85rem;margin-bottom:1rem">
        Отметьте циклы для главной. После прохождения шага отметка ставится сама.
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

      <ul class="list">
        {cycles.map((c) => (
          <li key={c.id} class="list-item" style="gap:0.65rem">
            <label
              class="chip-row"
              style="align-items:center;gap:0.5rem;margin:0;cursor:pointer"
              onClick={(e) => void toggleOnHome(c, e)}
            >
              <input
                type="checkbox"
                checked={c.onHome}
                disabled={togglingId === c.id}
                readOnly
              />
              <span style="font-size:0.85rem">{c.onHome ? 'На главной' : 'Скрыт с главной'}</span>
            </label>
            <a
              href={`/cycles/${c.id}`}
              style="display:flex;flex-direction:column;gap:0.35rem;text-decoration:none;color:inherit"
            >
              <span class="list-item__title">{c.name}</span>
              <span class="list-item__meta">
                {c.steps.length === 0
                  ? 'Шагов нет'
                  : c.completed
                    ? `Завершён · ${c.steps.length} шагов`
                    : `Шаг ${c.currentStep} из ${c.steps.length}`}
              </span>
              {c.steps.length > 0 && (
                <div class="chip-row">
                  {c.steps.map((s) => (
                    <span key={s.id} class="chip">
                      {s.position}. {s.workoutPlanName}
                    </span>
                  ))}
                </div>
              )}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
