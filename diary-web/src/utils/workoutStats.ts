import type { CreateSetBody, Set } from '../types';

type SetLike = Pick<Set, 'reps' | 'weightKg'> | CreateSetBody;

export function countSets(sets: SetLike[]): number {
  return sets.length;
}

/** Total reps across all sets. */
export function calcTotalReps(sets: SetLike[]): number {
  return sets.reduce((sum, s) => sum + s.reps, 0);
}

/** Volume load for classic (non-assist) exercises: Σ(reps × weightKg). */
export function calcTonnage(sets: SetLike[]): number {
  return sets.reduce((sum, s) => sum + s.reps * s.weightKg, 0);
}

export function formatTonnage(kg: number): string {
  if (Number.isInteger(kg)) return `${kg} кг`;
  return `${kg.toFixed(1)} кг`;
}

export function formatExerciseStats(sets: SetLike[], withTonnage: boolean): string {
  const parts = [`Подходов: ${countSets(sets)}`, `Повторов: ${calcTotalReps(sets)}`];
  if (withTonnage) {
    parts.push(`Тоннаж: ${formatTonnage(calcTonnage(sets))}`);
  }
  return parts.join(' · ');
}
