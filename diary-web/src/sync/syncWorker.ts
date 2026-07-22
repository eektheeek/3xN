import { api, ApiError } from '../api/client';
import {
  getWorkoutSession,
  listOpsForSession,
  listPendingOps,
  listWorkoutSessions,
  markOp,
  putWorkoutSession,
} from './store';
import type {
  SyncOp,
  SyncOpFinishPayload,
  SyncOpSaveExercisePayload,
  SyncOpStartPayload,
} from './types';

const MAX_ATTEMPTS = 5;

export type SyncWorkerState = 'idle' | 'syncing' | 'error';

let running = false;
let listenersAttached = false;
let lastError = '';
let state: SyncWorkerState = 'idle';

const bus = new EventTarget();

export function getSyncWorkerState(): { state: SyncWorkerState; lastError: string } {
  return { state, lastError };
}

export function onSyncChange(listener: () => void): () => void {
  const handler = () => listener();
  bus.addEventListener('change', handler);
  return () => bus.removeEventListener('change', handler);
}

function emitChange() {
  bus.dispatchEvent(new Event('change'));
}

function isOnline(): boolean {
  return typeof navigator === 'undefined' ? true : navigator.onLine;
}

async function resetStuckSyncingOps(): Promise<void> {
  const sessions = await listWorkoutSessions();
  for (const session of sessions) {
    const ops = await listOpsForSession(session.id);
    for (const op of ops) {
      if (op.status === 'syncing') {
        await markOp(op.id, { status: 'pending' });
      }
    }
  }
}

async function applyOp(op: SyncOp): Promise<void> {
  switch (op.type) {
    case 'start': {
      const payload = op.payload as SyncOpStartPayload;
      await api.startWorkoutSession({
        id: op.sessionId,
        performedAt: payload.performedAt,
        isDeload: payload.isDeload,
        workoutPlanId: payload.planId,
      });
      return;
    }
    case 'saveExercise': {
      const payload = op.payload as SyncOpSaveExercisePayload;
      await api.saveSessionExercise(op.sessionId, payload.exerciseId, {
        position: payload.position,
        sets: payload.sets,
      });
      return;
    }
    case 'finish': {
      const payload = op.payload as SyncOpFinishPayload;
      await api.finishWorkoutSession(op.sessionId, {
        durationSec: payload.durationSec,
        cycleId: payload.cycleId,
      });
      return;
    }
    default: {
      const _exhaustive: never = op.type;
      throw new Error(`unknown sync op type: ${_exhaustive}`);
    }
  }
}

function isFatalClientError(err: unknown): boolean {
  return err instanceof ApiError && err.status >= 400 && err.status < 500 && err.status !== 408;
}

async function refreshSessionSyncStatus(sessionId: string): Promise<void> {
  const session = await getWorkoutSession(sessionId);
  if (!session) return;

  const ops = await listOpsForSession(sessionId);
  const hasFailed = ops.some((op) => op.status === 'failed');
  const hasPending = ops.some((op) => op.status === 'pending' || op.status === 'syncing');

  let syncStatus = session.syncStatus;
  let lastSyncError = session.lastSyncError;

  if (hasFailed) {
    syncStatus = 'error';
    lastSyncError = ops.find((op) => op.status === 'failed')?.lastError ?? 'sync failed';
  } else if (hasPending) {
    syncStatus = session.finishedAt ? 'local' : 'syncing';
    lastSyncError = undefined;
  } else if (ops.length > 0 && ops.every((op) => op.status === 'synced')) {
    syncStatus = 'synced';
    lastSyncError = undefined;
  }

  await putWorkoutSession({
    ...session,
    syncStatus,
    lastSyncError,
  });
}

async function processSession(sessionId: string): Promise<boolean> {
  const pending = await listPendingOps(sessionId);
  for (const op of pending) {
    await markOp(op.id, { status: 'syncing', bumpAttempts: true });
    try {
      await applyOp(op);
      await markOp(op.id, { status: 'synced' });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'sync failed';
      const attempts = op.attempts + 1;
      if (isFatalClientError(err) || attempts >= MAX_ATTEMPTS) {
        await markOp(op.id, { status: 'failed', lastError: message });
        lastError = message;
        state = 'error';
        await refreshSessionSyncStatus(sessionId);
        emitChange();
        return false;
      }
      await markOp(op.id, { status: 'pending', lastError: message });
      lastError = message;
      state = 'error';
      await refreshSessionSyncStatus(sessionId);
      emitChange();
      return false;
    }
  }
  await refreshSessionSyncStatus(sessionId);
  return true;
}

/**
 * Replay pending sync_ops over HTTP, in seq order per session.
 * Safe to call often; concurrent calls coalesce.
 */
export async function runSync(): Promise<void> {
  if (running) return;
  if (!isOnline()) return;

  running = true;
  state = 'syncing';
  lastError = '';
  emitChange();

  try {
    await resetStuckSyncingOps();
    const sessions = await listWorkoutSessions();
    const sessionIds = [
      ...new Set(
        (
          await Promise.all(sessions.map(async (s) => ({ id: s.id, n: (await listPendingOps(s.id)).length })))
        )
          .filter((x) => x.n > 0)
          .map((x) => x.id),
      ),
    ];

    for (const sessionId of sessionIds) {
      const ok = await processSession(sessionId);
      if (!ok) break;
    }

    if (state === 'syncing') {
      state = 'idle';
      lastError = '';
    }
  } catch (err) {
    lastError = err instanceof Error ? err.message : 'sync failed';
    state = 'error';
  } finally {
    running = false;
    emitChange();
  }
}

/** Fire-and-forget sync when online. */
export function requestSync(): void {
  if (!isOnline()) return;
  void runSync();
}

/** Attach online / visibility listeners once (call from main). */
export function startSyncWorker(): void {
  if (listenersAttached || typeof window === 'undefined') return;
  listenersAttached = true;

  window.addEventListener('online', () => requestSync());
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') requestSync();
  });

  requestSync();
}
