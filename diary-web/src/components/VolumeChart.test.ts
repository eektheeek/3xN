import { describe, expect, it } from 'vitest';
import { chartYRange } from '../components/VolumeChart';

describe('chartYRange', () => {
  it('pads around a tight cluster so lines are not flat', () => {
    const { yMin, yMax } = chartYRange([24, 25, 26, 27]);
    expect(yMin).toBeLessThan(24);
    expect(yMax).toBeGreaterThan(27);
    expect(yMax - yMin).toBeLessThan(50);
  });

  it('handles identical values', () => {
    const { yMin, yMax } = chartYRange([30, 30, 30]);
    expect(yMin).toBeLessThan(30);
    expect(yMax).toBeGreaterThan(30);
  });
});
