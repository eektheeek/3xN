import { render } from 'preact';
import { registerSW } from 'virtual:pwa-register';
import { App } from './app.tsx';
import { startSyncWorker } from './sync/syncWorker';

registerSW({
  immediate: true,
  onRegisteredSW(_swUrl, registration) {
    // Pull a fresh SW when the tab/PWA becomes visible again.
    const ping = () => {
      void registration?.update();
    };
    ping();
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') ping();
    });
    setInterval(ping, 60_000);
  },
});
startSyncWorker();
render(<App />, document.getElementById('app')!);
