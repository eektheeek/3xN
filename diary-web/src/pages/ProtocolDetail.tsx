import { useEffect, useState } from 'preact/hooks';
import { route } from 'preact-router';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { IntervalProtocol } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

interface ProtocolDetailProps extends RoutableProps {
  id?: string;
}

export function ProtocolDetail({ id }: ProtocolDetailProps) {
  const [protocol, setProtocol] = useState<IntervalProtocol | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (!id) return;
    api
      .getIntervalProtocol(id)
      .then(setProtocol)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleSave = async (e: Event) => {
    e.preventDefault();
    if (!id) return;
    const form = e.target as HTMLFormElement;
    const data = new FormData(form);
    setSaving(true);
    setError('');
    try {
      const p = await api.updateIntervalProtocol(id, {
        name: String(data.get('name')).trim(),
        workSec: Number(data.get('workSec')),
        restSec: Number(data.get('restSec')),
        warmupExtra: data.get('warmupExtra') === 'on',
      });
      setProtocol(p);
      setEditing(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось сохранить');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!id || !confirm('Удалить эту табату? С упражнений она снимется.')) return;
    setDeleting(true);
    setError('');
    try {
      await api.deleteIntervalProtocol(id);
      route('/protocols');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось удалить');
      setDeleting(false);
    }
  };

  if (!id) {
    return (
      <div class="page">
        <ErrorBanner message="Не указан id" />
      </div>
    );
  }

  return (
    <div class="page">
      <header class="page-header">
        <a href="/protocols" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>{protocol?.name ?? 'Табата'}</h1>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {protocol && !editing && (
        <section class="card">
          <h2 class="card__title">{protocol.name}</h2>
          <p>
            Работа {protocol.workSec}с · Отдых {protocol.restSec}с
            {protocol.warmupExtra ? ' · +1 разминка' : ''}
          </p>
          <p class="muted">Раунды = подходы из цели упражнения{protocol.warmupExtra ? ' + 1' : ''}.</p>
          <div class="actions" style="display:flex;gap:0.5rem;flex-wrap:wrap">
            <button type="button" class="btn btn-secondary" onClick={() => setEditing(true)}>
              Редактировать
            </button>
            <button type="button" class="btn btn-ghost" disabled={deleting} onClick={() => void handleDelete()}>
              {deleting ? 'Удаление…' : 'Удалить'}
            </button>
          </div>
        </section>
      )}

      {protocol && editing && (
        <form class="card form" onSubmit={handleSave}>
          <label class="field">
            <span>Название</span>
            <input name="name" type="text" required defaultValue={protocol.name} />
          </label>
          <label class="field">
            <span>Работа, сек</span>
            <input name="workSec" type="number" min="1" required defaultValue={protocol.workSec} />
          </label>
          <label class="field">
            <span>Отдых, сек</span>
            <input name="restSec" type="number" min="0" required defaultValue={protocol.restSec} />
          </label>
          <label class="field field--checkbox">
            <input name="warmupExtra" type="checkbox" defaultChecked={protocol.warmupExtra} />
            <span>+1 разминочный раунд</span>
          </label>
          <div class="actions" style="display:flex;gap:0.5rem;flex-wrap:wrap">
            <button type="submit" class="btn btn-primary" disabled={saving}>
              {saving ? 'Сохранение…' : 'Сохранить'}
            </button>
            <button type="button" class="btn btn-ghost" onClick={() => setEditing(false)}>
              Отмена
            </button>
          </div>
        </form>
      )}
    </div>
  );
}
