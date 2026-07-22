/**
 * Offline workout sync module.
 *
 * Stage 0: types + module shell.
 * Stage 1: IndexedDB (workout_sessions, sync_ops, catalog).
 * Stage 3: local-first workout UI helpers.
 * Stage 4: SyncWorker (replay pending ops over HTTP).
 */

export type {
  CatalogCache,
  LocalWorkoutExerciseLog,
  LocalWorkoutSession,
  SyncOp,
  SyncOpFinishPayload,
  SyncOpPayload,
  SyncOpSaveExercisePayload,
  SyncOpStartPayload,
  SyncOpStatus,
  SyncOpType,
  WorkoutSessionSyncStatus,
} from './types';

export { SYNC_DB_NAME, SYNC_DB_VERSION, SYNC_STORE } from './types';

export {
  closeOfflineDB,
  deleteOfflineDB,
  openOfflineDB,
  type OfflineDB,
} from './db';

export {
  countPendingOps,
  deleteOpsForSession,
  deleteWorkoutSession,
  enqueueOp,
  getActiveWorkoutSession,
  getCatalogCache,
  getWorkoutSession,
  listAllSyncOps,
  listOpsForSession,
  listPendingOps,
  listWorkoutSessions,
  markOp,
  putCatalogCache,
  putWorkoutSession,
  requeueFailedAndStuckOps,
  type EnqueueOpInput,
  type MarkOpInput,
} from './store';

export {
  cachePlanBundle,
  defaultSetsForExercise,
  finishLocalWorkout,
  getResumableLocalSession,
  loadPlanBundle,
  pendingOpsLabel,
  persistLocalLogs,
  saveLocalExerciseResult,
  startLocalWorkout,
} from './localWorkout';

export {
  getSyncWorkerState,
  onSyncChange,
  requestSync,
  runSync,
  startSyncWorker,
} from './syncWorker';

export {
  cacheCycles,
  isPlanCached,
  loadHomeCatalog,
  loadHomeCatalogCacheFirst,
  readCachedHomeCatalog,
  refreshHomeCatalogFromNetwork,
  writeHomeCatalog,
  type HomeCatalog,
} from './catalog';
