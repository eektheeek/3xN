import type {
  CreateExerciseBody,
  CreateWorkoutPlanBody,
  CreateWorkoutSessionBody,
  Exercise,
  ExerciseTarget,
  SaveSessionExerciseBody,
  SetTargetBody,
  StartWorkoutSessionBody,
  WorkoutPlan,
  WorkoutSession,
  WorkoutSessionExercise,
  WorkoutSessionSummary,
} from '../types';

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // ignore
    }
    throw new ApiError(res.status, message);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export const api = {
  listExercises: () => request<Exercise[]>('/v1/exercises'),
  getExercise: (id: string) => request<Exercise>(`/v1/exercises/${id}`),
  createExercise: (body: CreateExerciseBody) =>
    request<Exercise>('/v1/exercises', { method: 'POST', body: JSON.stringify(body) }),
  setTarget: (id: string, body: SetTargetBody) =>
    request<ExerciseTarget>(`/v1/exercises/${id}/target`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  listWorkoutPlans: () => request<WorkoutPlan[]>('/v1/workout-plans'),
  getWorkoutPlan: (id: string) => request<WorkoutPlan>(`/v1/workout-plans/${id}`),
  createWorkoutPlan: (body: CreateWorkoutPlanBody) =>
    request<WorkoutPlan>('/v1/workout-plans', { method: 'POST', body: JSON.stringify(body) }),

  createWorkoutSession: (body: CreateWorkoutSessionBody) =>
    request<WorkoutSession>('/v1/workout-sessions', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  listWorkoutSessions: () => request<WorkoutSessionSummary[]>('/v1/workout-sessions'),
  startWorkoutSession: (body: StartWorkoutSessionBody = {}) =>
    request<WorkoutSession>('/v1/workout-sessions/start', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  getWorkoutSession: (id: string) => request<WorkoutSession>(`/v1/workout-sessions/${id}`),
  saveSessionExercise: (sessionId: string, exerciseId: string, body: SaveSessionExerciseBody) =>
    request<WorkoutSessionExercise>(
      `/v1/workout-sessions/${sessionId}/exercises/${exerciseId}`,
      { method: 'PUT', body: JSON.stringify(body) },
    ),
};

export { ApiError };
