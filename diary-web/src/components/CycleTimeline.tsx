import { useState } from 'preact/hooks';
import type { Cycle, CycleStep } from '../types';
import { formatSessionDate } from '../utils/dates';

type StepStatus = 'done' | 'current' | 'upcoming';

/** Show full list when short; window when longer. */
const SHORT_MAX = 5;
const DONE_TAIL = 2;
const UPCOMING_HEAD = 2;

function statusFor(cycle: Cycle, position: number): StepStatus {
  if (cycle.completed || position < cycle.currentStep) return 'done';
  if (position === cycle.currentStep) return 'current';
  return 'upcoming';
}

type Row =
  | { kind: 'summary'; key: string; count: number }
  | { kind: 'more'; key: string; count: number }
  | { kind: 'collapse'; key: string }
  | { kind: 'step'; key: string; step: CycleStep; status: StepStatus; finish: boolean };

function needsWindow(cycle: Cycle): boolean {
  return cycle.steps.length > SHORT_MAX;
}

function buildRows(cycle: Cycle, expanded: boolean): Row[] {
  const steps = cycle.steps;
  const n = steps.length;
  if (n === 0) return [];

  const allSteps = (): Row[] =>
    steps.map((step) => {
      const status = statusFor(cycle, step.position);
      return {
        kind: 'step' as const,
        key: step.id,
        step,
        status,
        finish: step.position === n && status === 'done',
      };
    });

  if (!needsWindow(cycle) || expanded) {
    const rows = allSteps();
    if (needsWindow(cycle) && expanded) {
      rows.push({ kind: 'collapse', key: 'collapse' });
    }
    return rows;
  }

  if (cycle.completed) {
    const rows: Row[] = [];
    const hidden = Math.max(0, n - DONE_TAIL);
    if (hidden > 0) {
      rows.push({ kind: 'summary', key: 'summary', count: hidden });
    }
    for (let i = hidden; i < n; i++) {
      const step = steps[i]!;
      rows.push({
        kind: 'step',
        key: step.id,
        step,
        status: 'done',
        finish: i === n - 1,
      });
    }
    return rows;
  }

  const currentIdx = Math.max(
    0,
    steps.findIndex((s) => s.position === cycle.currentStep),
  );
  const windowStart = Math.max(0, currentIdx - DONE_TAIL);
  const windowEnd = Math.min(n - 1, currentIdx + UPCOMING_HEAD);

  const rows: Row[] = [];
  if (windowStart > 0) {
    rows.push({ kind: 'summary', key: 'summary', count: windowStart });
  }
  for (let i = windowStart; i <= windowEnd; i++) {
    const step = steps[i]!;
    const status = statusFor(cycle, step.position);
    rows.push({
      kind: 'step',
      key: step.id,
      step,
      status,
      finish: i === n - 1 && status === 'done',
    });
  }
  const remaining = n - 1 - windowEnd;
  if (remaining > 0) {
    rows.push({ kind: 'more', key: 'more', count: remaining });
  }
  return rows;
}

function progressLabel(cycle: Cycle): string {
  const n = cycle.steps.length;
  if (cycle.completed) return `Завершён · ${n} шагов`;
  return `Шаг ${cycle.currentStep} из ${n}`;
}

function lineAfterIsDone(row: Row): boolean {
  return row.kind === 'summary' || (row.kind === 'step' && row.status === 'done');
}

function StepLabel({
  step,
  status,
  canStart,
  onStartCurrent,
}: {
  step: CycleStep;
  status: StepStatus;
  canStart: boolean;
  onStartCurrent?: (step: CycleStep) => void;
}) {
  const label = step.workoutPlanName || `Шаг ${step.position}`;
  const date =
    step.completedPerformedAt != null && step.completedPerformedAt !== ''
      ? formatSessionDate(step.completedPerformedAt)
      : null;

  if (canStart && onStartCurrent) {
    return (
      <button
        type="button"
        class="cycle-timeline__label cycle-timeline__label--action"
        onClick={() => onStartCurrent(step)}
      >
        {label}
        <span class="cycle-timeline__hint">Начать</span>
      </button>
    );
  }

  if (status === 'done' && step.completedSessionId) {
    return (
      <a href={`/diary/${step.completedSessionId}`} class="cycle-timeline__label cycle-timeline__label--done-link">
        <span class="cycle-timeline__name">{label}</span>
        {date && <span class="cycle-timeline__meta">{date}</span>}
      </a>
    );
  }

  return (
    <span class="cycle-timeline__label">
      <span class="cycle-timeline__name">{label}</span>
      {status === 'done' && date && <span class="cycle-timeline__meta">{date}</span>}
    </span>
  );
}

interface CycleTimelineProps {
  cycle: Cycle;
  /** When set, the current step row is tappable to start. */
  onStartCurrent?: (step: CycleStep) => void;
}

export function CycleTimeline({ cycle, onStartCurrent }: CycleTimelineProps) {
  const [expanded, setExpanded] = useState(false);

  if (cycle.steps.length === 0) return null;

  const rows = buildRows(cycle, expanded);
  const toggle = () => setExpanded((v) => !v);

  return (
    <div class="cycle-timeline-wrap">
      <p class="cycle-timeline__progress">{progressLabel(cycle)}</p>
      <ol class="cycle-timeline">
        {rows.map((row, index) => {
          const isLastRow = index === rows.length - 1;
          const showLine = !isLastRow && row.kind !== 'collapse' && row.kind !== 'more';
          const lineDone = showLine && lineAfterIsDone(row);

          if (row.kind === 'summary') {
            return (
              <li key={row.key} class="cycle-timeline__item cycle-timeline__item--summary">
                <div class="cycle-timeline__rail" aria-hidden="true">
                  <span class="cycle-timeline__node cycle-timeline__node--summary">✓</span>
                  {showLine && (
                    <span class={`cycle-timeline__line${lineDone ? ' cycle-timeline__line--done' : ''}`} />
                  )}
                </div>
                <button
                  type="button"
                  class="cycle-timeline__label cycle-timeline__label--toggle"
                  onClick={toggle}
                >
                  {row.count} пройдено
                  <span class="cycle-timeline__hint">показать все</span>
                </button>
              </li>
            );
          }

          if (row.kind === 'more') {
            return (
              <li key={row.key} class="cycle-timeline__item cycle-timeline__item--more">
                <div class="cycle-timeline__rail" aria-hidden="true">
                  <span class="cycle-timeline__node cycle-timeline__node--more">…</span>
                </div>
                <button
                  type="button"
                  class="cycle-timeline__label cycle-timeline__label--toggle"
                  onClick={toggle}
                >
                  ещё {row.count}
                  <span class="cycle-timeline__hint">развернуть</span>
                </button>
              </li>
            );
          }

          if (row.kind === 'collapse') {
            return (
              <li key={row.key} class="cycle-timeline__item cycle-timeline__item--more">
                <div class="cycle-timeline__rail" aria-hidden="true">
                  <span class="cycle-timeline__node cycle-timeline__node--more">↑</span>
                </div>
                <button
                  type="button"
                  class="cycle-timeline__label cycle-timeline__label--toggle"
                  onClick={toggle}
                >
                  свернуть
                </button>
              </li>
            );
          }

          const { step, status, finish } = row;
          const canStart = status === 'current' && Boolean(onStartCurrent);

          return (
            <li
              key={row.key}
              class={`cycle-timeline__item cycle-timeline__item--${status}${finish ? ' cycle-timeline__item--finish' : ''}`}
            >
              <div class="cycle-timeline__rail" aria-hidden="true">
                <span
                  class={`cycle-timeline__node${finish ? ' cycle-timeline__node--finish' : ''}`}
                  aria-label={finish ? 'Финиш' : undefined}
                >
                  {finish ? '' : status === 'done' ? '✓' : step.position}
                </span>
                {showLine && (
                  <span class={`cycle-timeline__line${lineDone ? ' cycle-timeline__line--done' : ''}`} />
                )}
              </div>
              <StepLabel
                step={step}
                status={status}
                canStart={canStart}
                onStartCurrent={onStartCurrent}
              />
            </li>
          );
        })}
      </ol>
    </div>
  );
}
