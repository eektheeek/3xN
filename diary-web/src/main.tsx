import { render } from 'preact';
import { App } from './app.tsx';
import { startSyncWorker } from './sync/syncWorker';

startSyncWorker();
render(<App />, document.getElementById('app')!);
