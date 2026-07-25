import { openTunnelUnlock } from '../tunnel/gate';
import { forceAppUpdate } from '../utils/forceAppUpdate';

/** Dev shortcuts stacked as a footer on every screen. */
export function DevLinksBar() {
  return (
    <footer class="dev-footer">
      <a href="/debug/sync">Sync debug</a>
      <button type="button" class="linkish" onClick={() => openTunnelUnlock()}>
        Подтвердить туннель
      </button>
      <button type="button" class="linkish" onClick={() => void forceAppUpdate()}>
        Обновить приложение
      </button>
    </footer>
  );
}
