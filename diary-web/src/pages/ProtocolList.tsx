import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { IntervalProtocol } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';

function protocolMeta(p: IntervalProtocol): string {
  const parts = [`${p.workSec}с / ${p.restSec}с`];
  if (p.warmupExtra) parts.push('+1 разминка');
  return parts.join(' · ');
}

export function ProtocolList(_props: RoutableProps) {
  const [items, setItems] = useState<IntervalProtocol[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .listIntervalProtocols()
      .then(setItems)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Табаты</h1>
        <a href="/protocols/new" class="btn btn-primary btn-icon" aria-label="Новая табата">
          +
        </a>
      </header>

      <ErrorBanner message={error} />
      {loading && <p class="muted">Загрузка…</p>}

      {!loading && items.length === 0 && (
        <p class="muted">Пока нет сохранённых табат. Создай протокол и повесь его на упражнение.</p>
      )}

      <ul class="list">
        {items.map((p) => (
          <li key={p.id}>
            <a href={`/protocols/${p.id}`} class="list-item">
              <span class="list-item__title">{p.name}</span>
              <span class="list-item__meta">{protocolMeta(p)}</span>
              <span class="list-item__meta">Раунды = подходы цели{p.warmupExtra ? ' + 1' : ''}</span>
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
