interface SyncStatusBannerProps {
  offline: boolean;
  pendingCount: number;
  syncing?: boolean;
  message?: string;
  onRetry?: () => void;
}

/** Status for local-first workouts + sync worker. */
export function SyncStatusBanner({
  offline,
  pendingCount,
  syncing = false,
  message,
  onRetry,
}: SyncStatusBannerProps) {
  if (message) {
    return (
      <div class="sync-banner sync-banner--warn">
        <span>{message}</span>
        {onRetry && (
          <button type="button" class="sync-banner__retry" onClick={onRetry}>
            Повторить
          </button>
        )}
      </div>
    );
  }
  if (offline) {
    return (
      <div class="sync-banner sync-banner--offline">
        Офлайн — тренировка сохраняется на телефоне
      </div>
    );
  }
  if (syncing) {
    return <div class="sync-banner sync-banner--pending">Синхронизация…</div>;
  }
  if (pendingCount > 0) {
    return (
      <div class="sync-banner sync-banner--pending">
        Ждёт синхронизации ({pendingCount})
        {onRetry && (
          <button type="button" class="sync-banner__retry" onClick={onRetry}>
            Синхронизировать
          </button>
        )}
      </div>
    );
  }
  return null;
}
