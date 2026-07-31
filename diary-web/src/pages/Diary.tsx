import { useEffect, useMemo, useState } from 'preact/hooks';
import type { RoutableProps } from 'preact-router';
import { api } from '../api/client';
import type { WorkoutSessionSummary } from '../types';
import { ErrorBanner } from '../components/ErrorBanner';
import { SessionSummaryCard } from '../components/SessionSummaryCard';
import { formatMonthYear, localDateKey } from '../utils/dates';

type ViewMode = 'calendar' | 'list';

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

function groupByDate(sessions: WorkoutSessionSummary[]): Map<string, WorkoutSessionSummary[]> {
  const map = new Map<string, WorkoutSessionSummary[]>();
  for (const s of sessions) {
    const key = localDateKey(s.performedAt);
    const list = map.get(key) ?? [];
    list.push(s);
    map.set(key, list);
  }
  return map;
}

function buildCalendarDays(year: number, month: number): (number | null)[] {
  const first = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0).getDate();
  // Monday-first: Mon=0 … Sun=6
  const startOffset = (first.getDay() + 6) % 7;
  const cells: (number | null)[] = Array.from({ length: startOffset }, () => null);
  for (let d = 1; d <= lastDay; d++) cells.push(d);
  while (cells.length % 7 !== 0) cells.push(null);
  return cells;
}

export function Diary(_props: RoutableProps) {
  const [sessions, setSessions] = useState<WorkoutSessionSummary[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState<ViewMode>('calendar');
  const [cursor, setCursor] = useState(() => {
    const now = new Date();
    return { year: now.getFullYear(), month: now.getMonth() };
  });
  const [selectedDay, setSelectedDay] = useState<string | null>(null);

  useEffect(() => {
    api
      .listWorkoutSessions()
      .then(setSessions)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const byDate = useMemo(() => groupByDate(sessions), [sessions]);
  const calendarDays = useMemo(
    () => buildCalendarDays(cursor.year, cursor.month),
    [cursor.year, cursor.month],
  );

  const selectedSessions = selectedDay ? (byDate.get(selectedDay) ?? []) : [];

  const shiftMonth = (delta: number) => {
    setCursor((prev) => {
      const d = new Date(prev.year, prev.month + delta, 1);
      return { year: d.getFullYear(), month: d.getMonth() };
    });
    setSelectedDay(null);
  };

  return (
    <div class="page page--with-tabs">
      <header class="page-header">
        <h1>Дневник</h1>
      </header>

      <ErrorBanner message={error} />

      <div class="view-toggle" role="tablist" aria-label="Вид дневника">
        <button
          type="button"
          role="tab"
          class={`view-toggle__btn${view === 'calendar' ? ' view-toggle__btn--active' : ''}`}
          aria-selected={view === 'calendar'}
          onClick={() => setView('calendar')}
        >
          Календарь
        </button>
        <button
          type="button"
          role="tab"
          class={`view-toggle__btn${view === 'list' ? ' view-toggle__btn--active' : ''}`}
          aria-selected={view === 'list'}
          onClick={() => setView('list')}
        >
          Список
        </button>
      </div>

      {loading && <p class="muted">Загрузка…</p>}

      {!loading && sessions.length === 0 && (
        <p class="muted">Пока нет записей. Завершите тренировку — она появится здесь.</p>
      )}

      {!loading && view === 'list' && sessions.length > 0 && (
        <ul class="list">
          {sessions.map((session) => (
            <li key={session.id}>
              <SessionSummaryCard session={session} />
            </li>
          ))}
        </ul>
      )}

      {!loading && view === 'calendar' && sessions.length > 0 && (
        <section class="calendar">
          <div class="calendar__nav">
            <button type="button" class="btn btn-ghost" onClick={() => shiftMonth(-1)} aria-label="Предыдущий месяц">
              ‹
            </button>
            <h2 class="calendar__title">{formatMonthYear(cursor.year, cursor.month)}</h2>
            <button type="button" class="btn btn-ghost" onClick={() => shiftMonth(1)} aria-label="Следующий месяц">
              ›
            </button>
          </div>

          <div class="calendar__weekdays">
            {WEEKDAYS.map((d) => (
              <span key={d} class="calendar__weekday">
                {d}
              </span>
            ))}
          </div>

          <div class="calendar__grid">
            {calendarDays.map((day, i) => {
              if (day == null) {
                return <span key={`empty-${i}`} class="calendar__cell calendar__cell--empty" />;
              }
              const key = `${cursor.year}-${String(cursor.month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
              const daySessions = byDate.get(key) ?? [];
              const hasWorkout = daySessions.length > 0;
              const isSelected = selectedDay === key;

              return (
                <button
                  key={key}
                  type="button"
                  class={`calendar__cell${hasWorkout ? ' calendar__cell--active' : ''}${isSelected ? ' calendar__cell--selected' : ''}`}
                  disabled={!hasWorkout}
                  onClick={() => setSelectedDay(isSelected ? null : key)}
                >
                  <span class="calendar__day">{day}</span>
                  {hasWorkout && <span class="calendar__dot" aria-hidden="true" />}
                </button>
              );
            })}
          </div>

          {selectedDay && (
            <div class="calendar__day-list">
              <h3 class="section-title">Тренировки за день</h3>
              {selectedSessions.length === 0 ? (
                <p class="muted">Нет записей</p>
              ) : (
                <ul class="list">
                  {selectedSessions.map((session) => (
                    <li key={session.id}>
                      <SessionSummaryCard session={session} />
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </section>
      )}
    </div>
  );
}
