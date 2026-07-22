import 'fake-indexeddb/auto';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { newId } from '../utils/id';
import {
  countPendingOps,
  deleteOfflineDB,
  enqueueOp,
  getActiveWorkoutSession,
  getCatalogCache,
  getWorkoutSession,
  listOpsForSession,
  listPendingOps,
  markOp,
  putCatalogCache,
  putWorkoutSession,
} from './store';
import type { LocalWorkoutSession } from './types';

function sampleSession(overrides: Partial<LocalWorkoutSession> = {}): LocalWorkoutSession {
  const now = new Date().toISOString();
  return {
    id: newId(),
    planId: 'plan-1',
    planSnapshot: { id: 'plan-1', name: 'Test', createdAt: now, exercises: [] },
    exerciseSnapshots: [],
    isDeload: false,
    startedAt: now,
    logs: [],
    syncStatus: 'local',
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
}

beforeEach(async () => {
  await deleteOfflineDB();
});

afterEach(async () => {
  await deleteOfflineDB();
});

describe('workout_sessions store', () => {
  it('puts and gets a workout session', async () => {
    const session = sampleSession({ id: 'sess-1' });
    await putWorkoutSession(session);
    const loaded = await getWorkoutSession('sess-1');
    expect(loaded?.id).toBe('sess-1');
    expect(loaded?.planId).toBe('plan-1');
  });

  it('returns the active (unfinished) session for a plan', async () => {
    await putWorkoutSession(
      sampleSession({
        id: 'done',
        finishedAt: new Date().toISOString(),
        durationSec: 60,
      }),
    );
    await putWorkoutSession(sampleSession({ id: 'active' }));

    const active = await getActiveWorkoutSession('plan-1');
    expect(active?.id).toBe('active');
  });
});

describe('sync_ops queue', () => {
  it('assigns ascending seq and keeps pending order', async () => {
    const sessionId = 'sess-q';

    const start = await enqueueOp({
      sessionId,
      type: 'start',
      payload: { planId: 'plan-1', performedAt: new Date().toISOString(), isDeload: false },
    });
    const save = await enqueueOp({
      sessionId,
      type: 'saveExercise',
      payload: {
        exerciseId: 'ex-1',
        position: 1,
        sets: [{ setNumber: 1, reps: 10, durationSec: 0, weightKg: 0, assistKg: 0 }],
      },
    });
    const finish = await enqueueOp({
      sessionId,
      type: 'finish',
      payload: { durationSec: 100, cycleId: 'cycle-1' },
    });

    expect(start.seq).toBe(1);
    expect(save.seq).toBe(2);
    expect(finish.seq).toBe(3);
    expect(start.status).toBe('pending');

    const pending = await listPendingOps(sessionId);
    expect(pending.map((op) => op.seq)).toEqual([1, 2, 3]);
    expect(pending.map((op) => op.type)).toEqual(['start', 'saveExercise', 'finish']);
    expect(await countPendingOps(sessionId)).toBe(3);
  });

  it('markOp transitions pending → syncing → synced and drops from pending list', async () => {
    const sessionId = 'sess-mark';
    const op = await enqueueOp({
      sessionId,
      type: 'start',
      payload: { planId: 'plan-1', performedAt: new Date().toISOString(), isDeload: false },
    });

    const syncing = await markOp(op.id, { status: 'syncing', bumpAttempts: true });
    expect(syncing.status).toBe('syncing');
    expect(syncing.attempts).toBe(1);
    expect(await listPendingOps(sessionId)).toHaveLength(0);

    const synced = await markOp(op.id, { status: 'synced' });
    expect(synced.status).toBe('synced');
    expect(synced.syncedAt).toBeTruthy();

    const all = await listOpsForSession(sessionId);
    expect(all).toHaveLength(1);
    expect(all[0]?.status).toBe('synced');
  });

  it('keeps seq order across sessions independently', async () => {
    await enqueueOp({
      sessionId: 'a',
      type: 'start',
      payload: { planId: 'p', performedAt: new Date().toISOString(), isDeload: false },
    });
    await enqueueOp({
      sessionId: 'b',
      type: 'start',
      payload: { planId: 'p', performedAt: new Date().toISOString(), isDeload: false },
    });
    await enqueueOp({
      sessionId: 'a',
      type: 'finish',
      payload: { durationSec: 1 },
    });

    const a = await listOpsForSession('a');
    const b = await listOpsForSession('b');
    expect(a.map((op) => op.seq)).toEqual([1, 2]);
    expect(b.map((op) => op.seq)).toEqual([1]);
  });

  it('marks failed with lastError', async () => {
    const op = await enqueueOp({
      sessionId: 'sess-err',
      type: 'finish',
      payload: { durationSec: 10 },
    });
    const failed = await markOp(op.id, {
      status: 'failed',
      lastError: 'network down',
      bumpAttempts: true,
    });
    expect(failed.status).toBe('failed');
    expect(failed.lastError).toBe('network down');
    expect(failed.attempts).toBe(1);
  });
});

describe('catalog cache', () => {
  it('stores and loads catalog', async () => {
    await putCatalogCache({ plans: [], exercises: [], cycles: [] });
    const cache = await getCatalogCache();
    expect(cache?.id).toBe('catalog');
    expect(cache?.plans).toEqual([]);
  });
});
