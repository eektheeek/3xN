import { useEffect, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { ErrorBanner } from '../components/ErrorBanner';
import {
  listAllSyncOps,
  listWorkoutSessions,
  requeueFailedAndStuckOps,
} from '../sync/store';
import type { LocalWorkoutSession, SyncOp } from '../sync/types';
import {
  getSyncWorkerState,
  onSyncChange,
  runSync,
} from '../sync/syncWorker';
import { openTunnelUnlock } from '../tunnel/gate';
import { formatSessionDateTime } from '../utils/dates';

function shortId(id: string): string {
  return id.length <= 12 ? id : `${id.slice(0, 8)}…`;
}

function opSummary(op: SyncOp): string {
  if (op.type === 'saveExercise' && 'exerciseId' in op.payload) {
    return `ex ${shortId(op.payload.exerciseId)} · sets ${op.payload.sets.length}`;
  }
  if (op.type === 'finish' && 'durationSec' in op.payload) {
    return `${op.payload.durationSec}s` + (op.payload.cycleId ? ` · cycle` : '');
  }
  if (op.type === 'start' && 'planId' in op.payload) {
    return `plan ${shortId(op.payload.planId)}`;
  }
  return '';
}

function opDatesLine(op: SyncOp): string {
  const parts = [`создана ${formatSessionDateTime(op.createdAt)}`];
  if (op.type === 'start' && 'performedAt' in op.payload) {
    parts.push(`тренировка ${formatSessionDateTime(op.payload.performedAt)}`);
  }
  if (op.syncedAt) {
    parts.push(`синк ${formatSessionDateTime(op.syncedAt)}`);
  }
  return parts.join(' · ');
}

function sessionSummary(s: LocalWorkoutSession): string {
  const parts = [
    s.finishedAt ? 'finished' : 'active',
    `logs ${s.logs.length}`,
    s.syncStatus,
  ];
  if (s.durationSec != null) parts.push(`${s.durationSec}s`);
  if (s.lastSyncError) parts.push(s.lastSyncError);
  return parts.join(' · ');
}

export function SyncDebug(_props: RoutableProps) {
  const [ops, setOps] = useState<SyncOp[]>([]);
  const [sessions, setSessions] = useState<LocalWorkoutSession[]>([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [workerState, setWorkerState] = useState(getSyncWorkerState());
  const [online, setOnline] = useState(navigator.onLine);

  const reload = async () => {
    setError('');
    try {
      const [nextOps, nextSessions] = await Promise.all([
        listAllSyncOps(),
        listWorkoutSessions(),
      ]);
      setOps(nextOps);
      setSessions(nextSessions);
      setWorkerState(getSyncWorkerState());
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось прочитать IndexedDB');
    }
  };

  useEffect(() => {
    void reload();
    const off = onSyncChange(() => {
      void reload();
    });
    const onNet = () => setOnline(navigator.onLine);
    window.addEventListener('online', onNet);
    window.addEventListener('offline', onNet);
    return () => {
      off();
      window.removeEventListener('online', onNet);
      window.removeEventListener('offline', onNet);
    };
  }, []);

  const handleSync = async () => {
    setBusy(true);
    setError('');
    try {
      await runSync();
      await reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Sync failed');
    } finally {
      setBusy(false);
    }
  };

  const handleRequeue = async () => {
    setBusy(true);
    setError('');
    try {
      const n = await requeueFailedAndStuckOps();
      await runSync();
      await reload();
      if (n === 0) setError('Нечего возвращать в очередь (нет failed/syncing)');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Requeue failed');
    } finally {
      setBusy(false);
    }
  };

  const pending = ops.filter((o) => o.status === 'pending').length;
  const failed = ops.filter((o) => o.status === 'failed').length;
  const synced = ops.filter((o) => o.status === 'synced').length;

  return (
    <div class="page">
      <header class="page-header">
        <a href="/" class="btn btn-ghost">
          ← Назад
        </a>
        <h1>Sync debug</h1>
      </header>

      <p class="muted" style="margin-top:0">
        Только просмотр IndexedDB. Данные не удаляются.
      </p>

      <ErrorBanner message={error} />

      <section class="card" style="margin-bottom:1rem">
        <p style="margin:0 0 0.35rem">
          Сеть: <strong>{online ? 'online' : 'offline'}</strong>
          {' · '}
          worker: <strong>{workerState.state}</strong>
          {workerState.lastError ? ` · ${workerState.lastError}` : ''}
        </p>
        <p class="muted" style="margin:0 0 0.75rem">
          ops: {ops.length} (pending {pending}, failed {failed}, synced {synced}) · sessions:{' '}
          {sessions.length}
        </p>
        <div class="actions" style="gap:0.5rem">
          <button type="button" class="btn btn-secondary" disabled={busy} onClick={() => void reload()}>
            Обновить
          </button>
          <button type="button" class="btn btn-primary" disabled={busy || !online} onClick={() => void handleSync()}>
            {busy ? '…' : 'Синхронизировать'}
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            disabled={busy}
            onClick={() => void handleRequeue()}
          >
            Failed → pending + sync
          </button>
          <button type="button" class="btn btn-secondary" onClick={() => openTunnelUnlock()}>
            Разблокировать туннель
          </button>
        </div>
      </section>

      <section style="margin-bottom:1.25rem">
        <h2 class="section-title">Локальные тренировки</h2>
        {sessions.length === 0 ? (
          <p class="muted">Пусто</p>
        ) : (
          <ul class="list">
            {sessions.map((s) => (
              <li key={s.id} class="list-item" style="cursor:default">
                <span class="list-item__title">{s.planSnapshot.name || s.planId}</span>
                <span class="list-item__meta">{shortId(s.id)}</span>
                <span class="list-item__meta">{sessionSummary(s)}</span>
                <span class="list-item__meta">
                  exercises:{' '}
                  {s.logs.map((l) => `${shortId(l.exerciseId)}(${l.sets.length})`).join(', ') || '—'}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section>
        <h2 class="section-title">Очередь sync_ops</h2>
        {ops.length === 0 ? (
          <p class="muted">Очередь пуста</p>
        ) : (
          <ul class="list">
            {ops.map((op) => (
              <li key={op.id} class="list-item" style="cursor:default">
                <span class="list-item__title">
                  #{op.seq} {op.type}{' '}
                  <span class={`badge ${op.status === 'synced' ? 'badge--ok' : op.status === 'failed' ? 'badge--warn' : 'badge--muted'}`}>
                    {op.status}
                  </span>
                </span>
                <span class="list-item__meta">{opDatesLine(op)}</span>
                <span class="list-item__meta">session {shortId(op.sessionId)}</span>
                <span class="list-item__meta">{opSummary(op)}</span>
                <span class="list-item__meta">
                  attempts {op.attempts}
                  {op.lastError ? ` · ${op.lastError}` : ''}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
