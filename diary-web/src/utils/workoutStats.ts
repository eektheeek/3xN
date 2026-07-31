import type { CreateSetBody, ExerciseKind, Set, SetTargetBody } from '../types';
import { formatDurationLabel } from './dates';

type SetLike = Pick<Set, 'reps' | 'weightKg' | 'durationSec'> | CreateSetBody;

export function countSets(sets: SetLike[]): number {
  return sets.length;
}

/** Total reps across all sets. */
export function calcTotalReps(sets: SetLike[]): number {
  return sets.reduce((sum, s) => sum + s.reps, 0);
}

/** Total hold seconds across all sets. */
export function calcTotalHoldSec(sets: SetLike[]): number {
  return sets.reduce((sum, s) => sum + (s.durationSec ?? 0), 0);
}

/** Volume load for classic (non-assist) exercises: Σ(reps × weightKg). */
export function calcTonnage(sets: SetLike[]): number {
  return sets.reduce((sum, s) => sum + s.reps * s.weightKg, 0);
}

export function formatTonnage(kg: number): string {
  if (Number.isInteger(kg)) return `${kg} кг`;
  return `${kg.toFixed(1)} кг`;
}

/** Target volume in chart units: sets×reps or sets×holdSec. */
export function calcTargetVolume(
  target: Pick<SetTargetBody, 'sets' | 'reps' | 'holdSec'>,
  kind: ExerciseKind = 'reps',
): number {
  if (kind === 'hold') return target.sets * target.holdSec;
  return target.sets * target.reps;
}

export function formatVolumeLabel(volume: number, kind: ExerciseKind): string {
  if (kind === 'hold') {
    return formatDurationLabel(volume) || '0с';
  }
  return String(volume);
}

export function formatExerciseStats(
  sets: SetLike[],
  opts: { kind?: ExerciseKind; withTonnage?: boolean } = {},
): string {
  const kind = opts.kind ?? 'reps';
  if (kind === 'hold') {
    const total = calcTotalHoldSec(sets);
    const parts = [`Подходов: ${countSets(sets)}`, `Время: ${formatDurationLabel(total) || '0с'}`];
    const weights = sets.map((s) => s.weightKg).filter((w) => w > 0);
    if (weights.length > 0) {
      const maxW = Math.max(...weights);
      parts.push(Number.isInteger(maxW) ? `${maxW} кг` : `${maxW.toFixed(1)} кг`);
    }
    return parts.join(' · ');
  }
  const parts = [`Подходов: ${countSets(sets)}`, `Повторов: ${calcTotalReps(sets)}`];
  if (opts.withTonnage) {
    parts.push(`Тоннаж: ${formatTonnage(calcTonnage(sets))}`);
  }
  return parts.join(' · ');
}

export function formatTargetLabel(
  target: Pick<SetTargetBody, 'sets' | 'reps' | 'holdSec' | 'weightKg' | 'assistKg'>,
  opts: { kind?: ExerciseKind; supportsAssist?: boolean },
): string {
  if (opts.kind === 'hold') {
    let label = `${target.sets}×${target.holdSec}с`;
    if (target.weightKg > 0) {
      label += Number.isInteger(target.weightKg)
        ? ` · ${target.weightKg} кг`
        : ` · ${target.weightKg.toFixed(1)} кг`;
    }
    return label;
  }
  let label = `${target.sets}×${target.reps}`;
  if (opts.supportsAssist && target.assistKg > 0) {
    label += ` · резинка ${target.assistKg} кг`;
  } else if (target.weightKg > 0) {
    label += ` · ${target.weightKg} кг`;
  }
  return label;
}
