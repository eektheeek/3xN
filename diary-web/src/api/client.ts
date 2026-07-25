import type {
  CreateCycleBody,
  CreateExerciseBody,
  CreateIntervalProtocolBody,
  CreateWorkoutPlanBody,
  CreateWorkoutSessionBody,
  Cycle,
  Exercise,
  ExerciseTarget,
  IntervalProtocol,
  ReplaceCycleStepsBody,
  SaveSessionExerciseBody,
  SetTargetBody,
  StartWorkoutSessionBody,
  UpdateExerciseBody,
  UpdateIntervalProtocolBody,
  WorkoutPlan,
  WorkoutSession,
  WorkoutSessionExercise,
  WorkoutSessionSummary,
} from '../types';
import { looksLikeNgrokInterstitial } from './ngrokDetect';
import { clearTunnelBlocked, reportTunnelBlocked } from '../tunnel/gate';

// Empty = same-origin (Vite proxies /v1 → diary-api). Set VITE_API_URL for a remote API host.
const API_URL = import.meta.env.VITE_API_URL ?? '';

class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

/** Free ngrok returned the Visit Site HTML page instead of API JSON. */
class NgrokInterstitialError extends ApiError {
  readonly kind = 'ngrok-interstitial' as const;

  constructor() {
    super(0, 'Нужно подтвердить доступ к туннелю (Visit Site).');
  }
}

function throwIfNgrokInterstitial(body: string, contentType: string | null): void {
  if (!looksLikeNgrokInterstitial(body, contentType)) return;
  reportTunnelBlocked();
  throw new NgrokInterstitialError();
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeoutId = window.setTimeout(() => controller.abort(), 15000);
  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...init,
      signal: controller.signal,
      headers: {
        'Content-Type': 'application/json',
        // Bypasses free-ngrok warning HTML on XHR/fetch (not on full page loads).
        'ngrok-skip-browser-warning': 'true',
        ...init?.headers,
      },
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new ApiError(0, 'Сервер не отвечает (таймаут). Проверь, что API запущен.');
    }
    throw new ApiError(0, 'Нет связи с API. Проверь Wi‑Fi и что diary-api запущен.');
  } finally {
    window.clearTimeout(timeoutId);
  }

  if (res.status === 204) {
    clearTunnelBlocked();
    return undefined as T;
  }

  const contentType = res.headers.get('content-type');
  const raw = await res.text();
  throwIfNgrokInterstitial(raw, contentType);

  if (!res.ok) {
    let message = res.statusText || `HTTP ${res.status}`;
    try {
      const body = JSON.parse(raw) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // ignore
    }
    throw new ApiError(res.status, message);
  }

  try {
    const data = JSON.parse(raw) as T;
    clearTunnelBlocked();
    return data;
  } catch {
    throwIfNgrokInterstitial(raw, contentType || 'text/html');
    throw new ApiError(0, 'Ответ API не JSON. Проверь туннель и proxy.');
  }
}

export const api = {
  listExercises: () => request<Exercise[]>('/v1/exercises'),
  getExercise: (id: string) => request<Exercise>(`/v1/exercises/${id}`),
  createExercise: (body: CreateExerciseBody) =>
    request<Exercise>('/v1/exercises', { method: 'POST', body: JSON.stringify(body) }),
  updateExercise: (id: string, body: UpdateExerciseBody) =>
    request<Exercise>(`/v1/exercises/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  setTarget: (id: string, body: SetTargetBody) =>
    request<ExerciseTarget>(`/v1/exercises/${id}/target`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  listIntervalProtocols: () => request<IntervalProtocol[]>('/v1/interval-protocols'),
  getIntervalProtocol: (id: string) => request<IntervalProtocol>(`/v1/interval-protocols/${id}`),
  createIntervalProtocol: (body: CreateIntervalProtocolBody) =>
    request<IntervalProtocol>('/v1/interval-protocols', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  updateIntervalProtocol: (id: string, body: UpdateIntervalProtocolBody) =>
    request<IntervalProtocol>(`/v1/interval-protocols/${id}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  deleteIntervalProtocol: (id: string) =>
    request<void>(`/v1/interval-protocols/${id}`, { method: 'DELETE' }),

  listWorkoutPlans: () => request<WorkoutPlan[]>('/v1/workout-plans'),
  getWorkoutPlan: (id: string) => request<WorkoutPlan>(`/v1/workout-plans/${id}`),
  createWorkoutPlan: (body: CreateWorkoutPlanBody) =>
    request<WorkoutPlan>('/v1/workout-plans', { method: 'POST', body: JSON.stringify(body) }),

  listCycles: () => request<Cycle[]>('/v1/cycles'),
  getCycle: (id: string) => request<Cycle>(`/v1/cycles/${id}`),
  createCycle: (body: CreateCycleBody) =>
    request<Cycle>('/v1/cycles', { method: 'POST', body: JSON.stringify(body) }),
  replaceCycleSteps: (id: string, body: ReplaceCycleStepsBody) =>
    request<Cycle>(`/v1/cycles/${id}/steps`, { method: 'PUT', body: JSON.stringify(body) }),
  advanceCycle: (id: string, sessionId?: string) =>
    request<Cycle>(`/v1/cycles/${id}/advance`, {
      method: 'POST',
      body: JSON.stringify(sessionId ? { sessionId } : {}),
    }),
  restartCycle: (id: string) =>
    request<Cycle>(`/v1/cycles/${id}/restart`, { method: 'POST', body: '{}' }),
  repeatCycle: (id: string) =>
    request<Cycle>(`/v1/cycles/${id}/repeat`, { method: 'POST', body: '{}' }),
  setCycleOnHome: (id: string, onHome: boolean) =>
    request<Cycle>(`/v1/cycles/${id}/on-home`, {
      method: 'PUT',
      body: JSON.stringify({ onHome }),
    }),

  createWorkoutSession: (body: CreateWorkoutSessionBody) =>
    request<WorkoutSession>('/v1/workout-sessions', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  listWorkoutSessions: () => request<WorkoutSessionSummary[]>('/v1/workout-sessions'),
  startWorkoutSession: (body: StartWorkoutSessionBody) =>
    request<WorkoutSession>('/v1/workout-sessions/start', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  getWorkoutSession: (id: string) => request<WorkoutSession>(`/v1/workout-sessions/${id}`),
  finishWorkoutSession: (id: string, body: { durationSec: number; cycleId?: string }) =>
    request<WorkoutSession>(`/v1/workout-sessions/${id}/finish`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  saveSessionExercise: (sessionId: string, exerciseId: string, body: SaveSessionExerciseBody) =>
    request<WorkoutSessionExercise>(
      `/v1/workout-sessions/${sessionId}/exercises/${exerciseId}`,
      { method: 'PUT', body: JSON.stringify(body) },
    ),
};

export { ApiError, NgrokInterstitialError };
