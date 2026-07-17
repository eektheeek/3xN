import { useEffect, useMemo, useRef, useState } from 'preact/hooks';
import type { IntervalProtocol } from '../types';
import {
  cancelScheduledCues,
  playCueNow,
  scheduleCueAt,
  tabataRounds,
  unlockAudio,
} from '../utils/beep';
import { formatDuration } from '../utils/dates';

type PhaseKind = 'prepare' | 'work' | 'rest' | 'done';

type Phase = {
  kind: PhaseKind;
  durationSec: number;
  /** 1-based work round; 0 for prepare/done */
  round: number;
};

type TimerStatus = 'running' | 'paused' | 'done';

interface TabataTimerProps {
  protocol: IntervalProtocol;
  sets: number;
  onClose: () => void;
}

const PIP_THRESHOLDS = [3, 2, 1] as const;

function buildSequence(protocol: IntervalProtocol, sets: number): Phase[] {
  const rounds = tabataRounds(sets, protocol.warmupExtra);
  const seq: Phase[] = [];

  if (protocol.prepareSec > 0) {
    seq.push({ kind: 'prepare', durationSec: protocol.prepareSec, round: 0 });
  }

  for (let r = 1; r <= rounds; r++) {
    seq.push({ kind: 'work', durationSec: protocol.workSec, round: r });
    if (r < rounds && protocol.restSec > 0) {
      seq.push({ kind: 'rest', durationSec: protocol.restSec, round: r });
    }
  }

  seq.push({ kind: 'done', durationSec: 0, round: 0 });
  return seq;
}

/**
 * Schedule audio for the active phase against the absolute deadline.
 * Pips fire at the exact second boundaries (display 3 → 2 → 1).
 */
function armPhaseAudio(phase: Phase, deadlinePerf: number): void {
  cancelScheduledCues();

  if (phase.kind === 'work' || phase.kind === 'rest') {
    playCueNow('phase');
  }

  for (const t of PIP_THRESHOLDS) {
    if (phase.durationSec >= t) {
      // When remaining hits exactly t seconds (same moment display shows t)
      scheduleCueAt('pip', deadlinePerf - t * 1000);
    }
  }
}

/**
 * Drift-free Tabata engine:
 * absolute performance.now() deadline per phase.
 * Audio is pre-scheduled on the Web Audio clock (not fired from the poll).
 * UI state updates only when the displayed second / phase changes.
 */
export function TabataTimer({ protocol, sets, onClose }: TabataTimerProps) {
  const sequence = useMemo(() => buildSequence(protocol, sets), [protocol, sets]);
  const totalRounds = tabataRounds(sets, protocol.warmupExtra);

  const [status, setStatus] = useState<TimerStatus>('running');
  const [phaseIndex, setPhaseIndex] = useState(0);
  const [remainingDisplay, setRemainingDisplay] = useState(
    () => sequence[0]?.durationSec ?? 0,
  );

  const deadlineRef = useRef(performance.now() + (sequence[0]?.durationSec ?? 0) * 1000);
  const indexRef = useRef(0);
  const pausedLeftRef = useRef(0);
  const sequenceRef = useRef(sequence);
  sequenceRef.current = sequence;
  const lastDisplayRef = useRef(sequence[0]?.durationSec ?? 0);

  // Arm cues for the first phase (prepare: only 3-2-1, no long buzz)
  useEffect(() => {
    const first = sequence[0];
    if (first && first.kind !== 'done') {
      armPhaseAudio(first, deadlineRef.current);
    }
    return () => cancelScheduledCues();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (status !== 'running') return;

    const tick = () => {
      const seq = sequenceRef.current;
      const now = performance.now();
      let index = indexRef.current;
      let remainingMs = deadlineRef.current - now;
      let phase = seq[index];
      if (!phase) return;

      if (phase.kind !== 'done') {
        const display = Math.max(0, Math.ceil(remainingMs / 1000 - 0.0001));
        if (display !== lastDisplayRef.current) {
          lastDisplayRef.current = display;
          setRemainingDisplay(display);
        }
      }

      // Catch up if tab was throttled
      while (remainingMs <= 0 && phase.kind !== 'done') {
        index += 1;
        phase = seq[index];
        if (!phase) break;

        if (phase.kind === 'done') {
          cancelScheduledCues();
          playCueNow('done');
          indexRef.current = index;
          lastDisplayRef.current = 0;
          setPhaseIndex(index);
          setRemainingDisplay(0);
          setStatus('done');
          return;
        }

        deadlineRef.current = now + phase.durationSec * 1000 + remainingMs;
        remainingMs = deadlineRef.current - now;
        indexRef.current = index;

        const display = Math.max(0, Math.ceil(remainingMs / 1000 - 0.0001));
        lastDisplayRef.current = display;
        setPhaseIndex(index);
        setRemainingDisplay(display);

        armPhaseAudio(phase, deadlineRef.current);
      }
    };

    const id = window.setInterval(tick, 50);
    tick();
    return () => window.clearInterval(id);
  }, [status]);

  const phase = sequence[phaseIndex] ?? sequence[sequence.length - 1];
  const isWarmupWork = phase?.kind === 'work' && protocol.warmupExtra && phase.round === 1;

  const phaseLabel =
    status === 'done' || phase?.kind === 'done'
      ? 'Готово'
      : phase?.kind === 'prepare'
        ? 'Подготовка'
        : isWarmupWork
          ? 'Разминка'
          : phase?.kind === 'work'
            ? 'Работа'
            : 'Отдых';

  const togglePause = () => {
    if (status === 'done') return;
    if (status === 'running') {
      pausedLeftRef.current = Math.max(0, deadlineRef.current - performance.now());
      cancelScheduledCues();
      setStatus('paused');
      return;
    }
    unlockAudio();
    deadlineRef.current = performance.now() + pausedLeftRef.current;
    const p = sequence[indexRef.current];
    if (p && p.kind !== 'done') {
      armPhaseAudio(p, deadlineRef.current);
    }
    setStatus('running');
  };

  const handleClose = () => {
    cancelScheduledCues();
    onClose();
  };

  return (
    <div
      class={`tabata-inline tabata-inline--${status === 'done' ? 'done' : (phase?.kind ?? 'done')}`}
      role="region"
      aria-label="Табата"
    >
      <div class="tabata-inline__head">
        <span class="tabata-inline__name">{protocol.name}</span>
        <span class="tabata-inline__phase">{phaseLabel}</span>
      </div>

      {phase?.kind === 'work' || phase?.kind === 'rest' ? (
        <p class="tabata-inline__round">
          Раунд {phase.round} / {totalRounds}
        </p>
      ) : null}
      {phase?.kind === 'prepare' && <p class="tabata-inline__round">Скоро старт</p>}

      <p class="tabata-inline__time">
        {status === 'done' || phase?.kind === 'done' ? '✓' : formatDuration(remainingDisplay)}
      </p>

      <div class="tabata-inline__actions">
        {status !== 'done' && (
          <button type="button" class="btn btn-secondary" onClick={togglePause}>
            {status === 'paused' ? 'Продолжить' : 'Пауза'}
          </button>
        )}
        <button type="button" class="btn btn-ghost" onClick={handleClose}>
          {status === 'done' ? 'Свернуть' : 'Стоп'}
        </button>
      </div>
    </div>
  );
}
