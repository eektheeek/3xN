import { useEffect, useState } from 'preact/hooks';
import { countPendingOps } from '../sync/store';
import {
  getSyncWorkerState,
  onSyncChange,
  requestSync,
  runSync,
} from '../sync/syncWorker';

/** Shared sync banner state for workout screens. */
export function useSyncBanner(sessionId?: string | null) {
  const [offline, setOffline] = useState(!navigator.onLine);
  const [pendingCount, setPendingCount] = useState(0);
  const [syncing, setSyncing] = useState(false);
  const [syncError, setSyncError] = useState('');

  const refresh = async () => {
    const { state, lastError } = getSyncWorkerState();
    setSyncing(state === 'syncing');
    setSyncError(state === 'error' ? lastError : '');
    setPendingCount(await countPendingOps(sessionId || undefined));
  };

  useEffect(() => {
    const syncOnline = () => {
      setOffline(!navigator.onLine);
      if (navigator.onLine) requestSync();
    };
    window.addEventListener('online', syncOnline);
    window.addEventListener('offline', syncOnline);
    return () => {
      window.removeEventListener('online', syncOnline);
      window.removeEventListener('offline', syncOnline);
    };
  }, []);

  useEffect(() => {
    void refresh();
    return onSyncChange(() => {
      void refresh();
    });
  }, [sessionId]);

  const retry = () => {
    void runSync();
  };

  return { offline, pendingCount, syncing, syncError, retry, refresh };
}
