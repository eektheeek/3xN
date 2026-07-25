import { useEffect, useState } from 'preact/hooks';
import { looksLikeNgrokInterstitial } from '../api/ngrokDetect';
import {
  clearTunnelBlocked,
  isTunnelBlocked,
  onTunnelBlockedChange,
  openTunnelUnlock,
  reportTunnelBlocked,
} from '../tunnel/gate';

async function probeTunnel(): Promise<void> {
  try {
    const res = await fetch('/healthz', {
      headers: { 'ngrok-skip-browser-warning': 'true' },
      cache: 'no-store',
    });
    const text = await res.text();
    if (looksLikeNgrokInterstitial(text, res.headers.get('content-type'))) {
      reportTunnelBlocked();
      return;
    }
    if (!res.ok) return;
    JSON.parse(text);
    clearTunnelBlocked();
  } catch {
    // Network down — not necessarily ngrok interstitial.
  }
}

/** Shown when API returns free-ngrok Visit Site HTML instead of JSON. */
export function TunnelGateBanner() {
  const [blocked, setBlocked] = useState(isTunnelBlocked());

  useEffect(() => onTunnelBlockedChange(() => setBlocked(isTunnelBlocked())), []);

  useEffect(() => {
    void probeTunnel();
    const onVis = () => {
      if (document.visibilityState === 'visible') void probeTunnel();
    };
    document.addEventListener('visibilitychange', onVis);
    return () => document.removeEventListener('visibilitychange', onVis);
  }, []);

  if (!blocked) return null;

  return (
    <div class="tunnel-banner" role="alert">
      <div class="tunnel-banner__text">
        Нужно подтвердить доступ к туннелю (Visit Site). Без этого API недоступен.
      </div>
      <div class="tunnel-banner__actions">
        <button type="button" class="btn btn-primary" onClick={() => openTunnelUnlock()}>
          Открыть подтверждение
        </button>
        <button type="button" class="btn btn-ghost" onClick={() => clearTunnelBlocked()}>
          Скрыть
        </button>
      </div>
    </div>
  );
}
