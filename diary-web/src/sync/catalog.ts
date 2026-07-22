import { api } from '../api/client';
import type { Cycle, Exercise, WorkoutPlan } from '../types';
import { getCatalogCache, putCatalogCache } from './store';

export type HomeCatalog = {
  plans: WorkoutPlan[];
  exercises: Exercise[];
  cycles: Cycle[];
  fromCache: boolean;
};

/** Replace full catalog snapshot used for offline start. */
export async function writeHomeCatalog(input: {
  plans: WorkoutPlan[];
  exercises: Exercise[];
  cycles: Cycle[];
}): Promise<void> {
  await putCatalogCache({
    plans: input.plans,
    exercises: input.exercises,
    cycles: input.cycles,
  });
}

export async function readCachedHomeCatalog(): Promise<HomeCatalog | null> {
  const cache = await getCatalogCache();
  if (!cache) return null;
  return {
    plans: cache.plans,
    exercises: cache.exercises,
    cycles: cache.cycles,
    fromCache: true,
  };
}

/** Fetch from API and update IndexedDB. */
export async function refreshHomeCatalogFromNetwork(): Promise<HomeCatalog> {
  const [plans, exercises, cycles] = await Promise.all([
    api.listWorkoutPlans(),
    api.listExercises(),
    api.listCycles(),
  ]);
  await writeHomeCatalog({ plans, exercises, cycles });
  return { plans, exercises, cycles, fromCache: false };
}

/**
 * Cache-first: return IndexedDB immediately when present, refresh API in background.
 * `onFresh` is called when the network snapshot arrives.
 */
export async function loadHomeCatalogCacheFirst(
  onFresh?: (data: HomeCatalog) => void,
): Promise<HomeCatalog | null> {
  const cached = await readCachedHomeCatalog();

  const refresh = async () => {
    const fresh = await refreshHomeCatalogFromNetwork();
    onFresh?.(fresh);
    return fresh;
  };

  if (typeof navigator !== 'undefined' && !navigator.onLine) {
    return cached;
  }

  if (cached) {
    void refresh().catch(() => {
      /* keep showing cache */
    });
    return cached;
  }

  try {
    return await refresh();
  } catch {
    return null;
  }
}

/** @deprecated Prefer loadHomeCatalogCacheFirst — kept for simple call sites/tests. */
export async function loadHomeCatalog(): Promise<HomeCatalog> {
  const cached = await readCachedHomeCatalog();
  try {
    return await refreshHomeCatalogFromNetwork();
  } catch (err) {
    if (cached) return cached;
    throw err instanceof Error ? err : new Error('Не удалось загрузить каталог');
  }
}

/** Merge cycles into existing catalog (e.g. after CycleList reload). */
export async function cacheCycles(cycles: Cycle[]): Promise<void> {
  const existing = await getCatalogCache();
  await putCatalogCache({
    plans: existing?.plans ?? [],
    exercises: existing?.exercises ?? [],
    cycles,
  });
}

export async function isPlanCached(planId: string): Promise<boolean> {
  const cache = await getCatalogCache();
  if (!cache) return false;
  const plan = cache.plans.find((p) => p.id === planId);
  if (!plan) return false;
  const ids = new Set((plan.exercises ?? []).map((e) => e.exerciseId));
  if (ids.size === 0) return true;
  const have = new Set(cache.exercises.map((e) => e.id));
  for (const id of ids) {
    if (!have.has(id)) return false;
  }
  return true;
}
