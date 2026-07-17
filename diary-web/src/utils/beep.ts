/**
 * Tabata audio: static WAV → AudioBuffer, scheduled on the Web Audio clock.
 *
 * Why this (not HTMLAudio from setInterval):
 * - HTMLAudio.play() from a timer is late and janky on iOS (main-thread + decode).
 * - Industry pattern (web.dev “A tale of two clocks”, metronomes, vekaev/tabata):
 *   pre-decode buffers, then source.start(audioCtx.currentTime + delay).
 * - Cues are scheduled when a phase starts, so 3/2/1 fire on the second boundary
 *   even if the JS tick is late.
 *
 * unlockAudio() must run inside the Start tap (iOS). Silent switch may still
 * mute Web Audio; we also prime a tiny HTMLAudio element as an unlock kick.
 */

type CueKind = 'pip' | 'phase' | 'done';

const SRC: Record<CueKind, string> = {
  pip: '/sounds/pip.wav?v=4',
  phase: '/sounds/phase.wav?v=4',
  done: '/sounds/done.wav?v=4',
};

const KINDS: CueKind[] = ['pip', 'phase', 'done'];

let ctx: AudioContext | null = null;
const buffers: Partial<Record<CueKind, AudioBuffer>> = {};
let loadPromise: Promise<void> | null = null;
let primed = false;

/** Active scheduled sources — stopped on pause / reschedule. */
let liveSources: AudioBufferSourceNode[] = [];
/** Fallback timers when AudioContext is unavailable. */
let liveTimers: number[] = [];
/** Bumps on cancel so in-flight async schedules are dropped. */
let scheduleGen = 0;

/** Tiny silent WAV (one sample) for HTMLAudio unlock kick on iOS. */
const SILENCE_WAV =
  'data:audio/wav;base64,UklGRiQAAABXQVZFZm10IBAAAAABAAEAESsAACJWAAACABAAZGF0YQAAAAA=';

let kickEl: HTMLAudioElement | null = null;

function getCtx(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  if (!ctx) {
    try {
      const Ctor =
        window.AudioContext ||
        (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
      if (!Ctor) return null;
      ctx = new Ctor();
    } catch {
      return null;
    }
  }
  return ctx;
}

async function ensureBuffers(): Promise<void> {
  if (buffers.pip && buffers.phase && buffers.done) return;
  if (loadPromise) return loadPromise;

  loadPromise = (async () => {
    const c = getCtx();
    if (!c) return;
    await Promise.all(
      KINDS.map(async (kind) => {
        if (buffers[kind]) return;
        const res = await fetch(SRC[kind]);
        const raw = await res.arrayBuffer();
        // decodeAudioData may detach the buffer — pass a copy
        buffers[kind] = await c.decodeAudioData(raw.slice(0));
      }),
    );
  })().catch(() => {
    loadPromise = null;
  });

  return loadPromise;
}

function stopAllScheduled(): void {
  scheduleGen += 1;
  for (const src of liveSources) {
    try {
      src.stop();
    } catch {
      /* already stopped */
    }
  }
  liveSources = [];
  for (const id of liveTimers) window.clearTimeout(id);
  liveTimers = [];
}

function playBufferNow(kind: CueKind, whenAudio: number): void {
  const c = getCtx();
  const buf = buffers[kind];
  if (!c || !buf) return;
  try {
    if (c.state === 'suspended') void c.resume();
    const src = c.createBufferSource();
    src.buffer = buf;
    src.connect(c.destination);
    const t = Math.max(whenAudio, c.currentTime);
    src.start(t);
    liveSources.push(src);
    src.onended = () => {
      liveSources = liveSources.filter((s) => s !== src);
    };
  } catch {
    /* ignore */
  }
}

/**
 * Schedule a cue at an absolute performance.now() timestamp.
 * Waits for decode if needed, then uses AudioContext.start(when).
 */
export function scheduleCueAt(kind: CueKind, atPerfMs: number): void {
  const gen = scheduleGen;
  void ensureBuffers().then(() => {
    if (gen !== scheduleGen) return;

    const c = getCtx();
    const buf = buffers[kind];
    const delayMs = atPerfMs - performance.now();

    // Missed by more than 80ms — skip (phase already past that second)
    if (delayMs < -80) return;

    if (c && buf) {
      if (c.state === 'suspended') void c.resume();
      const when = c.currentTime + Math.max(0, delayMs) / 1000;
      playBufferNow(kind, when);
      return;
    }

    const id = window.setTimeout(() => {
      if (gen !== scheduleGen) return;
      liveTimers = liveTimers.filter((x) => x !== id);
      playCueNow(kind);
    }, Math.max(0, delayMs));
    liveTimers.push(id);
  });
}

export function playCueNow(kind: CueKind): void {
  const c = getCtx();
  if (c && buffers[kind]) {
    if (c.state === 'suspended') void c.resume();
    playBufferNow(kind, c.currentTime);
    return;
  }
  scheduleCueAt(kind, performance.now());
}

export function cancelScheduledCues(): void {
  stopAllScheduled();
}

/** Warm decode before Start (safe anytime). */
export function preloadTabataAudio(): void {
  if (typeof window === 'undefined') return;
  getCtx();
  void ensureBuffers();
  if (!kickEl) {
    kickEl = new Audio(SILENCE_WAV);
    kickEl.preload = 'auto';
  }
}

/**
 * Must run synchronously inside a click/touch handler.
 */
export function unlockAudio(): void {
  if (typeof window === 'undefined') return;
  primed = true;
  const c = getCtx();
  if (c && c.state === 'suspended') void c.resume();
  void ensureBuffers();

  // HTMLAudio kick — helps some iOS builds unmute the audio pipeline
  try {
    if (!kickEl) {
      kickEl = new Audio(SILENCE_WAV);
      kickEl.preload = 'auto';
    }
    kickEl.volume = 0.01;
    kickEl.currentTime = 0;
    void kickEl.play().then(() => {
      kickEl?.pause();
      if (kickEl) kickEl.volume = 0;
    });
  } catch {
    /* ignore */
  }

  // Near-silent Web Audio blip in the same gesture
  if (c) {
    try {
      const osc = c.createOscillator();
      const gain = c.createGain();
      gain.gain.value = 0.0001;
      osc.connect(gain).connect(c.destination);
      osc.start();
      osc.stop(c.currentTime + 0.02);
    } catch {
      /* ignore */
    }
  }
}

export function cuePip(): void {
  playCueNow('pip');
}

export function cuePhase(): void {
  playCueNow('phase');
}

export function cueDone(): void {
  playCueNow('done');
}

export function playBeep(kind: 'tick' | 'phase' | 'done' = 'phase'): void {
  if (!primed) unlockAudio();
  if (kind === 'tick') cuePip();
  else if (kind === 'done') cueDone();
  else cuePhase();
}

export function tabataRounds(sets: number, warmupExtra: boolean): number {
  const base = Math.max(1, sets);
  return warmupExtra ? base + 1 : base;
}
