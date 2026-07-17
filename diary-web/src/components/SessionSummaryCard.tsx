import type { WorkoutSessionSummary } from '../types';
import { exerciseCountLabel, formatSessionDate } from '../utils/dates';

interface SessionSummaryCardProps {
  session: WorkoutSessionSummary;
  showDate?: boolean;
}

export function SessionSummaryCard({ session, showDate = true }: SessionSummaryCardProps) {
  const names = session.exerciseNames.join(', ');

  return (
    <a href={`/diary/${session.id}`} class="list-item">
      {showDate && <span class="list-item__title">{formatSessionDate(session.performedAt)}</span>}
      <span class="list-item__meta">{exerciseCountLabel(session.exerciseCount)}</span>
      {names && <span class="list-item__meta">{names}</span>}
      {session.isDeload && <span class="badge badge--muted">Разгрузка</span>}
    </a>
  );
}
