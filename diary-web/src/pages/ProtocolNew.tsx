import { useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import { ErrorBanner } from '../components/ErrorBanner';

export function ProtocolNew(_props: RoutableProps) {
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    setSubmitting(true);
    setError('');
    try {
      const p = await api.createIntervalProtocol({
        name: String(data.get('name')).trim(),
        prepareSec: Number(data.get('prepareSec')),
        workSec: Number(data.get('workSec')),
        restSec: Number(data.get('restSec')),
        warmupExtra: data.get('warmupExtra') === 'on',
      });
      route(`/protocols/${p.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось создать');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div class="page">
      <header class="page-header">
        <a href="/protocols" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Новая табата</h1>
      </header>

      <ErrorBanner message={error} />

      <form class="card form" onSubmit={handleSubmit}>
        <label class="field">
          <span>Название</span>
          <input name="name" type="text" required placeholder="40/20 с разминкой" />
        </label>
        <label class="field">
          <span>Подготовка, сек</span>
          <input name="prepareSec" type="number" min="0" required defaultValue={5} />
        </label>
        <label class="field">
          <span>Работа, сек</span>
          <input name="workSec" type="number" min="1" required defaultValue={40} />
        </label>
        <label class="field">
          <span>Отдых, сек</span>
          <input name="restSec" type="number" min="0" required defaultValue={20} />
        </label>
        <label class="field field--checkbox">
          <input name="warmupExtra" type="checkbox" />
          <span>+1 разминочный раунд (раунды = подходы + 1)</span>
        </label>
        <p class="muted" style="font-size:0.85rem">
          Число раундов берётся из цели упражнения (подходы). Здесь задаёшь только work/rest и разминку.
        </p>
        <button type="submit" class="btn btn-primary" disabled={submitting}>
          {submitting ? 'Создание…' : 'Создать'}
        </button>
      </form>
    </div>
  );
}
