import { closeOfflineDB, deleteOfflineDB, openOfflineDB } from './db';
import { newId } from '../utils/id';
import type {
  CatalogCache,
  LocalWorkoutSession,
  SyncOp,
  SyncOpPayload,
  SyncOpStatus,
  SyncOpType,
} from './types';
import { SYNC_STORE } from './types';

function nowIso(): string {
  return new Date().toISOString();
}

// --- workout_sessions ---

export async function putWorkoutSession(
  session: LocalWorkoutSession,
): Promise<LocalWorkoutSession> {
  const db = await openOfflineDB();
  const next: LocalWorkoutSession = {
    ...session,
    updatedAt: nowIso(),
  };
  await db.put(SYNC_STORE.workoutSessions, next);
  return next;
}

export async function getWorkoutSession(
  id: string,
): Promise<LocalWorkoutSession | undefined> {
  const db = await openOfflineDB();
  return db.get(SYNC_STORE.workoutSessions, id);
}

/** Active = not finished yet (no finishedAt). Prefer most recently updated. */
export async function getActiveWorkoutSession(
  planId?: string,
): Promise<LocalWorkoutSession | undefined> {
  const db = await openOfflineDB();
  const all = await db.getAll(SYNC_STORE.workoutSessions);
  const active = all
    .filter((s) => !s.finishedAt && (!planId || s.planId === planId))
    .sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
  return active[0];
}

export async function listWorkoutSessions(): Promise<LocalWorkoutSession[]> {
  const db = await openOfflineDB();
  const all = await db.getAll(SYNC_STORE.workoutSessions);
  return all.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
}

export async function deleteWorkoutSession(id: string): Promise<void> {
  const db = await openOfflineDB();
  await db.delete(SYNC_STORE.workoutSessions, id);
}

// --- sync_ops ---

export interface EnqueueOpInput {
  sessionId: string;
  type: SyncOpType;
  payload: SyncOpPayload;
}

/** Append the next op for a session (auto-increment seq, status=pending). */
export async function enqueueOp(input: EnqueueOpInput): Promise<SyncOp> {
  const db = await openOfflineDB();
  const existing = await db.getAllFromIndex(
    SYNC_STORE.syncOps,
    'by-session-seq',
    IDBKeyRange.bound([input.sessionId, 0], [input.sessionId, Number.MAX_SAFE_INTEGER]),
  );
  const maxSeq = existing.reduce((max, op) => Math.max(max, op.seq), 0);

  const op: SyncOp = {
    id: newId(),
    sessionId: input.sessionId,
    seq: maxSeq + 1,
    type: input.type,
    status: 'pending',
    payload: input.payload,
    attempts: 0,
    createdAt: nowIso(),
  };

  await db.put(SYNC_STORE.syncOps, op);
  return op;
}

/** Pending ops for a session, ascending by seq. */
export async function listPendingOps(sessionId: string): Promise<SyncOp[]> {
  const db = await openOfflineDB();
  const ops = await db.getAllFromIndex(SYNC_STORE.syncOps, 'by-session-status', [
    sessionId,
    'pending',
  ]);
  return ops.sort((a, b) => a.seq - b.seq);
}

/** All ops for a session (any status), ascending by seq. */
export async function listOpsForSession(sessionId: string): Promise<SyncOp[]> {
  const db = await openOfflineDB();
  const ops = await db.getAllFromIndex(
    SYNC_STORE.syncOps,
    'by-session-seq',
    IDBKeyRange.bound([sessionId, 0], [sessionId, Number.MAX_SAFE_INTEGER]),
  );
  return ops.sort((a, b) => a.seq - b.seq);
}

export interface MarkOpInput {
  status: SyncOpStatus;
  lastError?: string;
  bumpAttempts?: boolean;
}

export async function markOp(opId: string, input: MarkOpInput): Promise<SyncOp> {
  const db = await openOfflineDB();
  const existing = await db.get(SYNC_STORE.syncOps, opId);
  if (!existing) {
    throw new Error(`sync op not found: ${opId}`);
  }

  const next: SyncOp = {
    ...existing,
    status: input.status,
    attempts: input.bumpAttempts ? existing.attempts + 1 : existing.attempts,
    lastError: input.lastError,
    syncedAt: input.status === 'synced' ? nowIso() : existing.syncedAt,
  };

  if (input.status !== 'failed' && input.lastError === undefined) {
    delete next.lastError;
  }

  await db.put(SYNC_STORE.syncOps, next);
  return next;
}

export async function countPendingOps(sessionId?: string): Promise<number> {
  const db = await openOfflineDB();
  if (!sessionId) {
    return (await db.getAllFromIndex(SYNC_STORE.syncOps, 'by-status', 'pending')).length;
  }
  return (
    await db.getAllFromIndex(SYNC_STORE.syncOps, 'by-session-status', [sessionId, 'pending'])
  ).length;
}

export async function deleteOpsForSession(sessionId: string): Promise<void> {
  const db = await openOfflineDB();
  const ops = await listOpsForSession(sessionId);
  const tx = db.transaction(SYNC_STORE.syncOps, 'readwrite');
  await Promise.all(ops.map((op) => tx.store.delete(op.id)));
  await tx.done;
}

// --- catalog ---

export async function putCatalogCache(
  data: Omit<CatalogCache, 'id' | 'updatedAt'> & { updatedAt?: string },
): Promise<CatalogCache> {
  const db = await openOfflineDB();
  const cache: CatalogCache = {
    id: 'catalog',
    plans: data.plans,
    exercises: data.exercises,
    cycles: data.cycles,
    updatedAt: data.updatedAt ?? nowIso(),
  };
  await db.put(SYNC_STORE.catalog, cache);
  return cache;
}

export async function getCatalogCache(): Promise<CatalogCache | undefined> {
  const db = await openOfflineDB();
  return db.get(SYNC_STORE.catalog, 'catalog');
}

export { closeOfflineDB, deleteOfflineDB, openOfflineDB };
