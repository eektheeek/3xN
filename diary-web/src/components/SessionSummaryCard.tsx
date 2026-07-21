import type { WorkoutSessionSummary } from '../types';
import { exerciseCountLabel, formatDurationLabel, formatSessionDate } from '../utils/dates';

interface SessionSummaryCardProps {
  session: WorkoutSessionSummary;
  showDate?: boolean;
}

export function SessionSummaryCard({ session, showDate = true }: SessionSummaryCardProps) {
  const names = session.exerciseNames.join(', ');
  const duration = formatDurationLabel(session.durationSec);
  const cycleLabel =
    session.cycleName != null && session.cycleName !== ''
      ? session.cycleStep != null && session.cycleStep > 0
        ? `Цикл: ${session.cycleName} · шаг ${session.cycleStep}`
        : `Цикл: ${session.cycleName}`
      : null;

  return (
    <a href={`/diary/${session.id}`} class="list-item">
      {showDate && <span class="list-item__title">{formatSessionDate(session.performedAt)}</span>}
      {cycleLabel && <span class="list-item__meta list-item__meta--cycle">{cycleLabel}</span>}
      <span class="list-item__meta">
        {exerciseCountLabel(session.exerciseCount)}
        {duration ? ` · ${duration}` : ''}
      </span>
      {names && <span class="list-item__meta">{names}</span>}
      {session.isDeload && <span class="badge badge--muted">Разгрузка</span>}
    </a>
  );
}
