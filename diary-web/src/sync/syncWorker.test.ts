import 'fake-indexeddb/auto';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { deleteOfflineDB, enqueueOp, getWorkoutSession, listOpsForSession, putWorkoutSession } from './store';
import { runSync } from './syncWorker';
import type { LocalWorkoutSession } from './types';
import { newId } from '../utils/id';

const startWorkoutSession = vi.fn();
const saveSessionExercise = vi.fn();
const finishWorkoutSession = vi.fn();

vi.mock('../api/client', () => ({
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
  api: {
    startWorkoutSession: (...args: unknown[]) => startWorkoutSession(...args),
    saveSessionExercise: (...args: unknown[]) => saveSessionExercise(...args),
    finishWorkoutSession: (...args: unknown[]) => finishWorkoutSession(...args),
  },
}));

function sampleSession(id: string): LocalWorkoutSession {
  const now = new Date().toISOString();
  return {
    id,
    planId: 'plan-1',
    planSnapshot: { id: 'plan-1', name: 'Test', createdAt: now, exercises: [] },
    exerciseSnapshots: [],
    isDeload: false,
    startedAt: now,
    logs: [],
    syncStatus: 'local',
    createdAt: now,
    updatedAt: now,
  };
}

beforeEach(async () => {
  await deleteOfflineDB();
  startWorkoutSession.mockReset().mockResolvedValue({ id: 'x' });
  saveSessionExercise.mockReset().mockResolvedValue({});
  finishWorkoutSession.mockReset().mockResolvedValue({});
  vi.stubGlobal('navigator', { onLine: true });
});

afterEach(async () => {
  await deleteOfflineDB();
  vi.unstubAllGlobals();
});

describe('runSync', () => {
  it('replays start → save → finish in order', async () => {
    const sessionId = newId();
    await putWorkoutSession(sampleSession(sessionId));
    await enqueueOp({
      sessionId,
      type: 'start',
      payload: { planId: 'plan-1', performedAt: '2026-07-22T12:00:00Z', isDeload: false },
    });
    await enqueueOp({
      sessionId,
      type: 'saveExercise',
      payload: {
        exerciseId: 'ex-1',
        position: 1,
        sets: [{ setNumber: 1, reps: 10, durationSec: 0, weightKg: 0, assistKg: 0 }],
      },
    });
    await enqueueOp({
      sessionId,
      type: 'finish',
      payload: { durationSec: 120, cycleId: 'cycle-1' },
    });

    await runSync();

    expect(startWorkoutSession).toHaveBeenCalledTimes(1);
    expect(startWorkoutSession.mock.calls[0]?.[0]).toMatchObject({
      id: sessionId,
      workoutPlanId: 'plan-1',
    });
    expect(saveSessionExercise).toHaveBeenCalledTimes(1);
    expect(finishWorkoutSession).toHaveBeenCalledTimes(1);

    const ops = await listOpsForSession(sessionId);
    expect(ops.every((op) => op.status === 'synced')).toBe(true);

    const session = await getWorkoutSession(sessionId);
    expect(session?.syncStatus).toBe('synced');
  });
});
