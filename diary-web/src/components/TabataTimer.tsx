import { useEffect, useState } from 'preact/hooks';
import type { IntervalProtocol } from '../types';
import { playBeep, tabataRounds } from '../utils/beep';
import { formatDuration } from '../utils/dates';

type Phase = 'work' | 'rest' | 'done';

interface TabataTimerProps {
  protocol: IntervalProtocol;
  sets: number;
  exerciseName: string;
  onClose: () => void;
}

export function TabataTimer({ protocol, sets, exerciseName, onClose }: TabataTimerProps) {
  const totalRounds = tabataRounds(sets, protocol.warmupExtra);
  const [round, setRound] = useState(1);
  const [phase, setPhase] = useState<Phase>('work');
  const [remaining, setRemaining] = useState(protocol.workSec);
  const [running, setRunning] = useState(true);

  useEffect(() => {
    if (!running || phase === 'done') return;

    const id = window.setInterval(() => {
      setRemaining((prev) => {
        if (prev > 1) {
          if (prev <= 4) playBeep('tick');
          return prev - 1;
        }

        // Phase ended
        playBeep(phase === 'work' && round >= totalRounds ? 'done' : 'phase');

        if (phase === 'work') {
          if (round >= totalRounds) {
            setPhase('done');
            setRunning(false);
            return 0;
          }
          if (protocol.restSec <= 0) {
            setRound((r) => r + 1);
            setPhase('work');
            return protocol.workSec;
          }
          setPhase('rest');
          return protocol.restSec;
        }

        // rest -> next work
        setRound((r) => r + 1);
        setPhase('work');
        return protocol.workSec;
      });
    }, 1000);

    return () => window.clearInterval(id);
  }, [running, phase, round, totalRounds, protocol.workSec, protocol.restSec]);

  const phaseLabel =
    phase === 'done' ? 'Готово' : phase === 'work' ? (round === 1 && protocol.warmupExtra ? 'Разминка' : 'Работа') : 'Отдых';

  const isWarmupWork = phase === 'work' && protocol.warmupExtra && round === 1;

  return (
    <div class="tabata-overlay" role="dialog" aria-modal="true" aria-label="Табата">
      <div class={`tabata-panel tabata-panel--${phase}`}>
        <p class="tabata-panel__exercise">{exerciseName}</p>
        <p class="tabata-panel__protocol muted">{protocol.name}</p>

        {phase !== 'done' ? (
          <>
            <p class="tabata-panel__phase">{isWarmupWork ? 'Разминка' : phaseLabel}</p>
            <p class="tabata-panel__round">
              Раунд {round} / {totalRounds}
            </p>
            <p class="tabata-panel__time">{formatDuration(remaining)}</p>
          </>
        ) : (
          <>
            <p class="tabata-panel__phase">Готово</p>
            <p class="tabata-panel__time">✓</p>
          </>
        )}

        <div class="tabata-panel__actions">
          {phase !== 'done' && (
            <button type="button" class="btn btn-secondary" onClick={() => setRunning((v) => !v)}>
              {running ? 'Пауза' : 'Продолжить'}
            </button>
          )}
          <button
            type="button"
            class="btn btn-primary"
            onClick={() => {
              setRunning(false);
              onClose();
            }}
          >
            {phase === 'done' ? 'Закрыть' : 'Стоп'}
          </button>
        </div>
      </div>
    </div>
  );
}
