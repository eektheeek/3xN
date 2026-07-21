export interface ExerciseTarget {
  exerciseId: string;
  sets: number;
  reps: number;
  weightKg: number;
  assistKg: number;
}

export interface Exercise {
  id: string;
  name: string;
  muscleGroup: string;
  supportsAssist: boolean;
  protocolId?: string;
  createdAt: string;
  target?: ExerciseTarget;
  protocol?: IntervalProtocol;
}

export interface IntervalProtocol {
  id: string;
  name: string;
  prepareSec: number;
  workSec: number;
  restSec: number;
  warmupExtra: boolean;
  createdAt: string;
}

export interface WorkoutPlanExercise {
  id: string;
  workoutPlanId: string;
  exerciseId: string;
  position: number;
  exerciseName?: string;
}

export interface WorkoutPlan {
  id: string;
  name: string;
  createdAt: string;
  exercises?: WorkoutPlanExercise[];
}

export interface Set {
  id: string;
  workoutSessionExerciseId: string;
  setNumber: number;
  reps: number;
  weightKg: number;
  assistKg: number;
}

export interface WorkoutSessionExercise {
  id: string;
  workoutSessionId: string;
  exerciseId: string;
  position: number;
  exerciseName?: string;
  supportsAssist?: boolean;
  sets?: Set[];
}

export interface WorkoutSessionSummary {
  id: string;
  performedAt: string;
  durationSec: number;
  isDeload: boolean;
  createdAt: string;
  exerciseCount: number;
  exerciseNames: string[];
}

export interface WorkoutSession {
  id: string;
  performedAt: string;
  startedAt?: string;
  durationSec: number;
  isDeload: boolean;
  workoutPlanId?: string;
  createdAt: string;
  exercises?: WorkoutSessionExercise[];
}

export interface CycleStep {
  id: string;
  cycleId: string;
  position: number;
  workoutPlanId: string;
  workoutPlanName?: string;
}

export interface Cycle {
  id: string;
  name: string;
  currentStep: number;
  completed: boolean;
  onHome: boolean;
  createdAt: string;
  steps: CycleStep[];
}

export interface CreateCycleBody {
  name: string;
}

export interface ReplaceCycleStepsBody {
  workoutPlanIds: string[];
}

export interface CreateExerciseBody {
  name: string;
  muscleGroup?: string;
  supportsAssist?: boolean;
  protocolId?: string;
}

export type UpdateExerciseBody = CreateExerciseBody;

export interface CreateIntervalProtocolBody {
  name: string;
  prepareSec: number;
  workSec: number;
  restSec: number;
  warmupExtra: boolean;
}

export type UpdateIntervalProtocolBody = CreateIntervalProtocolBody;

export interface SetTargetBody {
  sets: number;
  reps: number;
  weightKg: number;
  assistKg: number;
}

export interface CreateSetBody {
  setNumber: number;
  reps: number;
  weightKg: number;
  assistKg: number;
}

export interface CreateWorkoutPlanBody {
  name: string;
  exerciseIds: string[];
}

export interface CreateWorkoutSessionBody {
  performedAt?: string;
  isDeload?: boolean;
  exercises: {
    exerciseId: string;
    sets: CreateSetBody[];
  }[];
}

export interface PlanDraftExercise {
  id: string;
  name: string;
}

export interface PlanDraft {
  name: string;
  exercises: PlanDraftExercise[];
}

export const PLAN_DRAFT_KEY = 'workoutPlanDraft';

export interface SaveSessionExerciseBody {
  position: number;
  sets: CreateSetBody[];
}

export interface StartWorkoutSessionBody {
  performedAt?: string;
  isDeload?: boolean;
  workoutPlanId?: string;
}

export interface ActiveWorkoutDraft {
  sessionId: string;
  planId: string;
  startedAt: string;
  savedExerciseIds: string[];
  logs: {
    exerciseId: string;
    sets: CreateSetBody[];
  }[];
}

export function activeWorkoutKey(planId: string): string {
  return `activeWorkout:${planId}`;
}

/** Selected training cycle on the home screen. */
export const ACTIVE_CYCLE_KEY = 'activeCycleId';
