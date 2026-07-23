import { render } from 'preact';
import { registerSW } from 'virtual:pwa-register';
import { App } from './app.tsx';
import { startSyncWorker } from './sync/syncWorker';

registerSW({ immediate: true });
startSyncWorker();
render(<App />, document.getElementById('app')!);
