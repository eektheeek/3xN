/** Play a short beep via Web Audio API (no asset file). */
export function playBeep(kind: 'tick' | 'phase' | 'done' = 'phase') {
  try {
    const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
    const ctx = new Ctx();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);

    const now = ctx.currentTime;
    if (kind === 'tick') {
      osc.frequency.value = 880;
      gain.gain.setValueAtTime(0.08, now);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.08);
      osc.start(now);
      osc.stop(now + 0.09);
    } else if (kind === 'done') {
      osc.frequency.value = 523;
      gain.gain.setValueAtTime(0.2, now);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.45);
      osc.start(now);
      osc.stop(now + 0.5);
    } else {
      osc.frequency.value = 660;
      gain.gain.setValueAtTime(0.18, now);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.25);
      osc.start(now);
      osc.stop(now + 0.28);
    }

    window.setTimeout(() => void ctx.close(), 600);
  } catch {
    // Audio may be blocked until user gesture; ignore.
  }
}

export function tabataRounds(sets: number, warmupExtra: boolean): number {
  const base = Math.max(1, sets);
  return warmupExtra ? base + 1 : base;
}
