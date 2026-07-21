import type { ActiveWorkoutDraft } from '../types';
import { activeWorkoutKey } from '../types';

export function readDraft(planId: string): ActiveWorkoutDraft | null {
  try {
    const raw = localStorage.getItem(activeWorkoutKey(planId));
    if (!raw) return null;
    return JSON.parse(raw) as ActiveWorkoutDraft;
  } catch {
    return null;
  }
}

export function writeDraft(planId: string, draft: ActiveWorkoutDraft) {
  localStorage.setItem(activeWorkoutKey(planId), JSON.stringify(draft));
}

export function clearDraft(planId: string) {
  localStorage.removeItem(activeWorkoutKey(planId));
}
