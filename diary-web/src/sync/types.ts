import type { CreateSetBody, Cycle, Exercise, WorkoutPlan } from '../types';

/** Sync status of a local workout session as a whole. */
export type WorkoutSessionSyncStatus = 'local' | 'syncing' | 'synced' | 'error';

/** Status of one outbound operation in the sync queue. */
export type SyncOpStatus = 'pending' | 'syncing' | 'synced' | 'failed';

export type SyncOpType = 'start' | 'saveExercise' | 'finish';

/** One exercise block inside a local workout session. */
export interface LocalWorkoutExerciseLog {
  exerciseId: string;
  position: number;
  sets: CreateSetBody[];
}

/**
 * Local copy of a gym visit (same entity as server workout_sessions).
 * `id` is always generated on the client (see utils/id.newId).
 */
export interface LocalWorkoutSession {
  id: string;
  planId: string;
  /** Frozen plan at start time so offline UI does not need the network. */
  planSnapshot: WorkoutPlan;
  /** Exercises with targets/protocols needed to render the start screen offline. */
  exerciseSnapshots: Exercise[];
  cycleId?: string;
  isDeload: boolean;
  startedAt: string;
  finishedAt?: string;
  durationSec?: number;
  logs: LocalWorkoutExerciseLog[];
  syncStatus: WorkoutSessionSyncStatus;
  lastSyncError?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SyncOpStartPayload {
  planId: string;
  performedAt: string;
  isDeload: boolean;
}

export interface SyncOpSaveExercisePayload {
  exerciseId: string;
  position: number;
  sets: CreateSetBody[];
}

export interface SyncOpFinishPayload {
  durationSec: number;
  cycleId?: string;
}

export type SyncOpPayload =
  | SyncOpStartPayload
  | SyncOpSaveExercisePayload
  | SyncOpFinishPayload;

/**
 * One ordered outbound API operation for a workout session.
 * Process pending ops for a session strictly by ascending `seq`.
 */
export interface SyncOp {
  id: string;
  sessionId: string;
  seq: number;
  type: SyncOpType;
  status: SyncOpStatus;
  payload: SyncOpPayload;
  attempts: number;
  lastError?: string;
  createdAt: string;
  syncedAt?: string;
}

/** Cached reference data for offline start. */
export interface CatalogCache {
  id: 'catalog';
  plans: WorkoutPlan[];
  exercises: Exercise[];
  cycles: Cycle[];
  updatedAt: string;
}

/** IndexedDB schema version and store names (implemented in stage 1). */
export const SYNC_DB_NAME = 'diary-offline';
export const SYNC_DB_VERSION = 1;

export const SYNC_STORE = {
  workoutSessions: 'workout_sessions',
  syncOps: 'sync_ops',
  catalog: 'catalog',
} as const;
