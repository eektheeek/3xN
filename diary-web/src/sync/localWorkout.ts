import { newId } from '../utils/id';
import { api } from '../api/client';
import type { CreateSetBody, Exercise, WorkoutPlan } from '../types';
import { ACTIVE_CYCLE_KEY } from '../types';
import { clearDraft, writeDraft } from '../utils/activeWorkout';
import {
  countPendingOps,
  enqueueOp,
  getActiveWorkoutSession,
  getCatalogCache,
  getWorkoutSession,
  putCatalogCache,
  putWorkoutSession,
} from './store';
import { requestSync } from './syncWorker';
import type { LocalWorkoutSession } from './types';

export function defaultSetsForExercise(exercise: Exercise): CreateSetBody[] {
  const n = exercise.target?.sets ?? 3;
  const isHold = exercise.kind === 'hold';
  const reps = isHold ? 0 : (exercise.target?.reps ?? 12);
  const durationSec = isHold ? (exercise.target?.holdSec ?? 60) : 0;
  const weightKg = exercise.target?.weightKg ?? 0;
  const assistKg = isHold ? 0 : (exercise.target?.assistKg ?? 0);
  return Array.from({ length: n }, (_, i) => ({
    setNumber: i + 1,
    reps,
    durationSec,
    weightKg,
    assistKg,
  }));
}

/** Merge one plan + its exercises into the offline catalog cache. */
export async function cachePlanBundle(
  plan: WorkoutPlan,
  exercises: Exercise[],
): Promise<void> {
  const existing = await getCatalogCache();
  const plans = [...(existing?.plans ?? []).filter((p) => p.id !== plan.id), plan];
  const byId = new Map((existing?.exercises ?? []).map((e) => [e.id, e]));
  for (const ex of exercises) {
    byId.set(ex.id, ex);
  }
  await putCatalogCache({
    plans,
    exercises: [...byId.values()],
    cycles: existing?.cycles ?? [],
  });
}

export type PlanBundle = {
  plan: WorkoutPlan;
  exercises: Exercise[];
  fromCache: boolean;
};

async function readPlanBundleFromCache(planId: string): Promise<PlanBundle | null> {
  const cache = await getCatalogCache();
  const plan = cache?.plans.find((p) => p.id === planId);
  if (!plan) return null;
  const ids = new Set((plan.exercises ?? []).map((e) => e.exerciseId));
  const exercises = (cache?.exercises ?? []).filter((e) => ids.has(e.id));
  if (exercises.length < ids.size) return null;
  return { plan, exercises, fromCache: true };
}

async function refreshPlanBundleFromNetwork(planId: string): Promise<PlanBundle> {
  const plan = await api.getWorkoutPlan(planId);
  const exercises: Exercise[] = [];
  for (const slot of plan.exercises ?? []) {
    exercises.push(await api.getExercise(slot.exerciseId));
  }
  await cachePlanBundle(plan, exercises);
  return { plan, exercises, fromCache: false };
}

/**
 * Cache-first plan load. Returns IndexedDB snapshot immediately when present,
 * then refreshes from API in the background via `onFresh`.
 */
export async function loadPlanBundle(
  planId: string,
  onFresh?: (bundle: PlanBundle) => void,
): Promise<PlanBundle | null> {
  const cached = await readPlanBundleFromCache(planId);

  if (typeof navigator !== 'undefined' && !navigator.onLine) {
    return cached;
  }

  const refresh = async () => {
    const fresh = await refreshPlanBundleFromNetwork(planId);
    onFresh?.(fresh);
    return fresh;
  };

  if (cached) {
    void refresh().catch(() => {
      /* keep cache */
    });
    return cached;
  }

  try {
    return await refresh();
  } catch {
    return null;
  }
}

export async function startLocalWorkout(input: {
  plan: WorkoutPlan;
  exercises: Exercise[];
  isDeload?: boolean;
}): Promise<LocalWorkoutSession> {
  const id = newId();
  const now = new Date().toISOString();
  const cycleId = localStorage.getItem(ACTIVE_CYCLE_KEY) || undefined;
  const exerciseById = new Map(input.exercises.map((e) => [e.id, e]));

  const logs = (input.plan.exercises ?? []).map((slot, i) => {
    const exercise = exerciseById.get(slot.exerciseId);
    return {
      exerciseId: slot.exerciseId,
      position: i + 1,
      sets: exercise ? defaultSetsForExercise(exercise) : [],
    };
  });

  const session: LocalWorkoutSession = {
    id,
    planId: input.plan.id,
    planSnapshot: input.plan,
    exerciseSnapshots: input.exercises,
    cycleId,
    isDeload: input.isDeload ?? false,
    startedAt: now,
    logs,
    syncStatus: 'local',
    createdAt: now,
    updatedAt: now,
  };

  await putWorkoutSession(session);
  await enqueueOp({
    sessionId: id,
    type: 'start',
    payload: {
      planId: input.plan.id,
      performedAt: now,
      isDeload: session.isDeload,
    },
  });

  requestSync();

  // Keep legacy draft in sync until stage 7 removes it.
  writeDraft(input.plan.id, {
    sessionId: id,
    planId: input.plan.id,
    startedAt: now,
    savedExerciseIds: [],
    logs: logs.map((l) => ({ exerciseId: l.exerciseId, sets: l.sets })),
  });

  return session;
}

export async function persistLocalLogs(
  sessionId: string,
  logs: { exerciseId: string; position: number; sets: CreateSetBody[] }[],
  savedExerciseIds: string[] = [],
): Promise<LocalWorkoutSession> {
  const session = await getWorkoutSession(sessionId);
  if (!session) {
    throw new Error('Локальная тренировка не найдена');
  }
  const next = await putWorkoutSession({
    ...session,
    logs,
  });
  writeDraft(session.planId, {
    sessionId,
    planId: session.planId,
    startedAt: session.startedAt,
    savedExerciseIds,
    logs: logs.map((l) => ({ exerciseId: l.exerciseId, sets: l.sets })),
  });
  return next;
}

export async function saveLocalExerciseResult(
  sessionId: string,
  exerciseId: string,
  position: number,
  sets: CreateSetBody[],
): Promise<void> {
  const session = await getWorkoutSession(sessionId);
  if (!session) {
    throw new Error('Локальная тренировка не найдена');
  }

  const logs = session.logs.map((log) =>
    log.exerciseId === exerciseId ? { ...log, position, sets } : log,
  );
  if (!logs.some((l) => l.exerciseId === exerciseId)) {
    logs.push({ exerciseId, position, sets });
  }

  await putWorkoutSession({ ...session, logs });
  await enqueueOp({
    sessionId,
    type: 'saveExercise',
    payload: { exerciseId, position, sets },
  });
  requestSync();
}

export async function finishLocalWorkout(
  sessionId: string,
  durationSec: number,
): Promise<LocalWorkoutSession> {
  const session = await getWorkoutSession(sessionId);
  if (!session) {
    throw new Error('Локальная тренировка не найдена');
  }

  const finishedAt = new Date().toISOString();
  const cycleId =
    session.cycleId || localStorage.getItem(ACTIVE_CYCLE_KEY) || undefined;

  const next = await putWorkoutSession({
    ...session,
    finishedAt,
    durationSec,
    cycleId,
    syncStatus: 'local',
  });

  await enqueueOp({
    sessionId,
    type: 'finish',
    payload: { durationSec, cycleId },
  });

  clearDraft(session.planId);
  requestSync();
  return next;
}

export async function getResumableLocalSession(
  planId: string,
): Promise<LocalWorkoutSession | undefined> {
  const active = await getActiveWorkoutSession(planId);
  if (active && !active.finishedAt) return active;
  return undefined;
}

export async function pendingOpsLabel(sessionId?: string): Promise<string> {
  const n = await countPendingOps(sessionId);
  if (n === 0) return '';
  return n === 1 ? '1 операция ждёт синхронизации' : `${n} операций ждут синхронизации`;
}
