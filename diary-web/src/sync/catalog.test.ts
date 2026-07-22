import 'fake-indexeddb/auto';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { deleteOfflineDB, getCatalogCache } from './store';
import {
  isPlanCached,
  loadHomeCatalog,
  loadHomeCatalogCacheFirst,
  writeHomeCatalog,
} from './catalog';

const listWorkoutPlans = vi.fn();
const listExercises = vi.fn();
const listCycles = vi.fn();

vi.mock('../api/client', () => ({
  api: {
    listWorkoutPlans: (...args: unknown[]) => listWorkoutPlans(...args),
    listExercises: (...args: unknown[]) => listExercises(...args),
    listCycles: (...args: unknown[]) => listCycles(...args),
  },
}));

beforeEach(async () => {
  await deleteOfflineDB();
  listWorkoutPlans.mockReset();
  listExercises.mockReset();
  listCycles.mockReset();
  vi.stubGlobal('navigator', { onLine: true });
});

afterEach(async () => {
  await deleteOfflineDB();
  vi.unstubAllGlobals();
});

describe('loadHomeCatalog', () => {
  it('caches plans/exercises/cycles from API', async () => {
    listWorkoutPlans.mockResolvedValue([
      {
        id: 'plan-1',
        name: 'Test',
        createdAt: '2026-07-22T00:00:00Z',
        exercises: [{ id: 'slot-1', workoutPlanId: 'plan-1', exerciseId: 'ex-1', position: 1 }],
      },
    ]);
    listExercises.mockResolvedValue([
      {
        id: 'ex-1',
        name: 'Pull-up',
        muscleGroup: 'back',
        kind: 'reps',
        supportsAssist: true,
        createdAt: '2026-07-22T00:00:00Z',
      },
    ]);
    listCycles.mockResolvedValue([]);

    const data = await loadHomeCatalog();
    expect(data.fromCache).toBe(false);
    expect(data.plans).toHaveLength(1);

    const cache = await getCatalogCache();
    expect(cache?.plans[0]?.id).toBe('plan-1');
    expect(await isPlanCached('plan-1')).toBe(true);
  });

  it('falls back to IndexedDB when API fails', async () => {
    listWorkoutPlans.mockResolvedValue([
      { id: 'plan-1', name: 'Test', createdAt: 't', exercises: [] },
    ]);
    listExercises.mockResolvedValue([]);
    listCycles.mockResolvedValue([]);
    await loadHomeCatalog();

    listWorkoutPlans.mockRejectedValue(new Error('offline'));
    const data = await loadHomeCatalog();
    expect(data.fromCache).toBe(true);
    expect(data.plans[0]?.id).toBe('plan-1');
  });
});

describe('loadHomeCatalogCacheFirst', () => {
  it('returns cache immediately and refreshes in background', async () => {
    await writeHomeCatalog({
      plans: [{ id: 'plan-1', name: 'Cached', createdAt: 't', exercises: [] }],
      exercises: [],
      cycles: [],
    });

    let resolvePlans!: (v: unknown) => void;
    listWorkoutPlans.mockReturnValue(
      new Promise((resolve) => {
        resolvePlans = resolve;
      }),
    );
    listExercises.mockResolvedValue([]);
    listCycles.mockResolvedValue([]);

    const freshPromise = new Promise<{ fromCache: boolean; name?: string }>((resolve) => {
      void loadHomeCatalogCacheFirst((fresh) => {
        resolve({ fromCache: fresh.fromCache, name: fresh.plans[0]?.name });
      }).then((first) => {
        expect(first?.fromCache).toBe(true);
        expect(first?.plans[0]?.name).toBe('Cached');
        resolvePlans([{ id: 'plan-1', name: 'Fresh', createdAt: 't', exercises: [] }]);
      });
    });

    const fresh = await freshPromise;
    expect(fresh.fromCache).toBe(false);
    expect(fresh.name).toBe('Fresh');
  });
});
