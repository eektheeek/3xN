import { describe, expect, it } from 'vitest';
import { calcTargetVolume, formatVolumeLabel } from './workoutStats';

describe('calcTargetVolume', () => {
  it('multiplies sets × reps for reps kind', () => {
    expect(calcTargetVolume({ sets: 3, reps: 12, holdSec: 0 }, 'reps')).toBe(36);
  });

  it('multiplies sets × holdSec for hold kind', () => {
    expect(calcTargetVolume({ sets: 3, reps: 0, holdSec: 60 }, 'hold')).toBe(180);
  });
});

describe('formatVolumeLabel', () => {
  it('returns number string for reps', () => {
    expect(formatVolumeLabel(36, 'reps')).toBe('36');
  });

  it('formats hold seconds', () => {
    expect(formatVolumeLabel(125, 'hold')).toBe('2м 5с');
  });
});
