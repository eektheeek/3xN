import type { ExerciseKind, ExerciseStatsPoint } from '../types';

interface VolumeChartProps {
  points: ExerciseStatsPoint[];
  kind: ExerciseKind;
}

/** Y-range padded around data so small differences stay visible. */
export function chartYRange(values: number[]): { yMin: number; yMax: number } {
  if (values.length === 0) return { yMin: 0, yMax: 1 };
  const rawMin = Math.min(...values);
  const rawMax = Math.max(...values);
  const span = rawMax - rawMin;
  const pad = span > 0 ? span * 0.25 : Math.max(rawMax * 0.15, 2);
  let yMin = Math.max(0, Math.floor(rawMin - pad));
  let yMax = Math.ceil(rawMax + pad);
  if (yMax <= yMin) yMax = yMin + 1;
  return { yMin, yMax };
}

function shortDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
}

export function VolumeChart({ points, kind }: VolumeChartProps) {
  const width = 320;
  const height = 200;
  const padL = 36;
  const padR = 14;
  const padT = 28;
  const padB = 28;
  const plotW = width - padL - padR;
  const plotH = height - padT - padB;

  const volumes = points.map((p) => p.volume);
  const targets = points
    .map((p) => p.targetVolume)
    .filter((v): v is number => v != null);
  const { yMin, yMax } = chartYRange([...volumes, ...targets]);
  const ySpan = yMax - yMin;
  const n = points.length;

  const xAt = (i: number) => {
    if (n <= 1) return padL + plotW / 2;
    return padL + (i / (n - 1)) * plotW;
  };
  const yAt = (v: number) => padT + plotH - ((v - yMin) / ySpan) * plotH;

  const factPath = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${xAt(i).toFixed(1)} ${yAt(p.volume).toFixed(1)}`)
    .join(' ');

  const planPath = points
    .map((p, i) => {
      const tv = p.targetVolume;
      if (tv == null) return null;
      const prevMissing = i === 0 || points[i - 1]?.targetVolume == null;
      return `${prevMissing ? 'M' : 'L'} ${xAt(i).toFixed(1)} ${yAt(tv).toFixed(1)}`;
    })
    .filter(Boolean)
    .join(' ');

  const ticks = [yMin, Math.round((yMin + yMax) / 2), yMax];
  const labelFmt = (v: number) => (kind === 'hold' ? String(v) : String(v));

  return (
    <div class="volume-chart">
      <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="План и факт по тренировкам">
        {ticks.map((t) => (
          <g key={t}>
            <line
              class="volume-chart__grid"
              x1={padL}
              y1={yAt(t)}
              x2={width - padR}
              y2={yAt(t)}
            />
            <text class="volume-chart__tick" x={padL - 6} y={yAt(t) + 3} text-anchor="end">
              {t}
            </text>
          </g>
        ))}

        {planPath && <path class="volume-chart__target-line" d={planPath} fill="none" />}
        {factPath && <path class="volume-chart__volume-line" d={factPath} fill="none" />}

        {points.map((p, i) => {
          const planY = p.targetVolume != null ? yAt(p.targetVolume) : null;
          const factY = yAt(p.volume);
          const same = p.targetVolume != null && p.targetVolume === p.volume;
          return (
            <g key={p.sessionId}>
              {p.targetVolume != null && planY != null && (
                <>
                  <circle
                    class="volume-chart__dot volume-chart__dot--plan"
                    cx={xAt(i)}
                    cy={planY}
                    r={3}
                  />
                  {!same && (
                    <text
                      class="volume-chart__value volume-chart__value--plan"
                      x={xAt(i)}
                      y={planY - 8}
                      text-anchor="middle"
                    >
                      {labelFmt(p.targetVolume)}
                    </text>
                  )}
                </>
              )}
              <circle
                class={`volume-chart__dot${p.isDeload ? ' volume-chart__dot--deload' : ''}`}
                cx={xAt(i)}
                cy={factY}
                r={3.5}
              />
              <text
                class="volume-chart__value"
                x={xAt(i)}
                y={factY - 8}
                text-anchor="middle"
              >
                {same && p.targetVolume != null
                  ? `${labelFmt(p.volume)}`
                  : labelFmt(p.volume)}
              </text>
            </g>
          );
        })}

        {n > 0 && (
          <>
            <text
              class="volume-chart__xlabel"
              x={xAt(0)}
              y={height - 6}
              text-anchor={n === 1 ? 'middle' : 'start'}
            >
              {shortDate(points[0].performedAt)}
            </text>
            {n > 1 && (
              <text class="volume-chart__xlabel" x={xAt(n - 1)} y={height - 6} text-anchor="end">
                {shortDate(points[n - 1].performedAt)}
              </text>
            )}
          </>
        )}
      </svg>
      <div class="volume-chart__legend">
        <span class="volume-chart__legend-item volume-chart__legend-item--volume">Факт</span>
        <span class="volume-chart__legend-item volume-chart__legend-item--target">План</span>
      </div>
    </div>
  );
}
