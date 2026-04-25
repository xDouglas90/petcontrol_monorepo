import type {
  CompanySystemConfigDTO,
  ScheduleHistoryItemDTO,
  ScheduleDTO,
  ScheduleStatus,
} from '@petcontrol/shared-types';
import {
  Activity,
  CalendarDays,
  CalendarRange,
  Clock3,
  PawPrint,
  TrendingDown,
  TrendingUp,
} from 'lucide-react';
import { useMemo, useState } from 'react';

import {
  useCurrentCompanyQuery,
  useCurrentCompanySystemConfigQuery,
  useCurrentUserQuery,
  useScheduleHistoriesQuery,
  useSchedulesQuery,
} from '@/lib/api/domain.queries';

const ACTIVE_SCHEDULE_STATUSES: ScheduleStatus[] = [
  'waiting',
  'confirmed',
  'in_progress',
];

const SHIFT_SCHEDULE_STATUSES: ScheduleStatus[] = [
  ...ACTIVE_SCHEDULE_STATUSES,
  'finished',
  'delivered',
];

export function DashboardPage() {
  const companyQuery = useCurrentCompanyQuery();
  const currentUserQuery = useCurrentUserQuery();
  const systemConfigQuery = useCurrentCompanySystemConfigQuery();
  const schedulesQuery = useSchedulesQuery();

  const company = companyQuery.data;
  const currentUser = currentUserQuery.data;
  const systemConfig = systemConfigQuery.data;
  const schedules = useMemo(
    () => schedulesQuery.data?.data ?? [],
    [schedulesQuery.data?.data],
  );
  const historyScheduleIds = useMemo(
    () =>
      schedules
        .filter((item) =>
          ['finished', 'delivered'].includes(item.current_status),
        )
        .map((item) => item.id),
    [schedules],
  );
  const scheduleHistoryQueries = useScheduleHistoriesQuery(historyScheduleIds);
  const now = new Date();
  const weekOptions = buildWeekOptions(now);
  const defaultWeekKey = resolveCurrentWeekKey(now);
  const [selectedWeekKey, setSelectedWeekKey] = useState(defaultWeekKey);

  const scheduleHistoryMap = useMemo(() => {
    const entries = historyScheduleIds.map(
      (scheduleId, index) =>
        [scheduleId, scheduleHistoryQueries[index]?.data ?? []] as const,
    );

    return new Map<string, ScheduleHistoryItemDTO[]>(entries);
  }, [historyScheduleIds, scheduleHistoryQueries]);

  if (
    companyQuery.isLoading ||
    currentUserQuery.isLoading ||
    systemConfigQuery.isLoading ||
    schedulesQuery.isLoading
  ) {
    return <DashboardSkeleton />;
  }

  if (
    companyQuery.isError ||
    currentUserQuery.isError ||
    systemConfigQuery.isError ||
    schedulesQuery.isError ||
    !company ||
    !currentUser ||
    !systemConfig
  ) {
    return (
      <section className="app-panel border-rose-500/30 bg-rose-500/10 p-6 text-rose-500">
        <p className="app-eyebrow text-rose-500">
          Dashboard indisponível
        </p>
        <h2 className="mt-3 font-display text-3xl text-rose-400">
          Não foi possível carregar os dados operacionais da empresa.
        </h2>
        <p className="mt-3 max-w-2xl text-sm leading-7">
          Verifique a disponibilidade da API, do perfil autenticado e da
          configuração operacional da empresa.
        </p>
      </section>
    );
  }

  if (currentUser.role !== 'admin') {
    return (
      <section className="app-panel p-6 shadow-premium">
        <p className="app-eyebrow">
          Home em preparação
        </p>
        <h2 className="mt-3 font-display text-3xl text-foreground">
          A experiência inicial para o perfil{' '}
          <span className="lowercase text-primary">{currentUser.role}</span> ainda será
          construída.
        </h2>
        <p className="mt-3 max-w-2xl text-sm leading-7 text-muted">
          Nesta etapa, o dashboard completo está sendo priorizado para o papel
          de administrador da empresa.
        </p>
      </section>
    );
  }

  const greetingName =
    currentUser.short_name || currentUser.full_name || company.fantasy_name;
  const todayCount = countSchedulesInDay(schedules, now);
  const previousDayCount = countSchedulesInDay(
    schedules,
    shiftDate(now, { days: -1 }),
  );
  const currentMonthCount = countSchedulesInMonth(schedules, now);
  const previousMonthCount = countSchedulesInMonth(
    schedules,
    shiftDate(now, { months: -1 }),
  );
  const monthlyTarget = calculateMonthlyTarget(systemConfig, now);
  const efficiencyPercentage =
    monthlyTarget > 0 ? (currentMonthCount / monthlyTarget) * 100 : 0;
  const completionPercentage =
    monthlyTarget > 0
      ? Math.min(100, Math.round((currentMonthCount / monthlyTarget) * 100))
      : 0;
  const shiftSchedules = resolveShiftSchedules(schedules, now);
  const currentShiftLabel =
    now.getHours() < 12 ? 'Turno da manhã' : 'Turno da tarde';
  const normalizedSelectedWeekKey = weekOptions.some(
    (item) => item.key === selectedWeekKey,
  )
    ? selectedWeekKey
    : defaultWeekKey;
  const weeklySeries = buildWeeklySeries(
    schedules,
    now,
    systemConfig,
    normalizedSelectedWeekKey,
  );
  const stats = [
    {
      label: 'Agendamentos/dia',
      value: String(todayCount),
      change: todayCount - previousDayCount,
      changeLabel: '-> ontem',
      description: 'Atualizado a cada novo agendamento criado no sistema.',
      icon: CalendarDays,
    },
    {
      label: 'Agendamentos/mês',
      value: String(currentMonthCount),
      change: currentMonthCount - previousMonthCount,
      changeLabel: '-> mês anterior',
      description: 'Volume operacional acumulado no mês corrente.',
      icon: CalendarRange,
    },
    {
      label: 'Eficiência (meta mensal)',
      value: `${Math.round(efficiencyPercentage)}%`,
      change: Math.round(efficiencyPercentage - 100),
      changeLabel: efficiencyPercentage >= 100 ? '%' : '%',
      description: `${currentMonthCount} de ${monthlyTarget} agendamentos mínimos previstos.`,
      icon: Activity,
    },
  ] as const;

  return (
    <main className="flex min-w-0 flex-col divide-y divide-border/50 min-h-full">
      <header className="bg-[radial-gradient(circle_at_top_right,rgba(2,132,199,0.08),transparent_40%),radial-gradient(circle_at_bottom_left,rgba(16,185,129,0.05),transparent_35%)] px-6 py-8 lg:px-10">
        <div className="flex flex-col gap-4">
          <div>
            <p className="app-eyebrow text-[11px]">
              VISÃO GERAL
            </p>
            <div className="mt-3 flex items-center justify-between gap-4">
              <h1 className="font-display text-4xl text-foreground sm:text-5xl">
                Olá, {greetingName}
              </h1>

              <div className="flex items-center gap-3 text-muted">
                <CalendarDays className="h-5 w-5" />
                <div className="hidden sm:block">
                  <p className="app-eyebrow text-[11px]">
                    Hoje
                  </p>
                  <span className="text-sm font-medium text-foreground">
                    {formatLongDate(now)}
                  </span>
                </div>
              </div>
            </div>
            <p className="mt-4 max-w-2xl text-sm leading-5 text-muted">
              Você está visualizando a operação de {company.fantasy_name}, com
              foco em agenda diária, comparação mensal e eficiência da meta.
            </p>
          </div>
        </div>
      </header>
      <section className="p-6 lg:p-10">
        <div className="grid gap-6 sm:grid-cols-3">
          {stats.map((stat) => (
            <AdminStatCard key={stat.label} {...stat} />
          ))}
        </div>
      </section>

      <section className="p-6 lg:p-10">
        <div>
          <p className="app-eyebrow">
            Performance
          </p>
          <div className="mt-2 flex flex-wrap items-center justify-between gap-3">
            <h3 className="font-display text-2xl text-foreground">
              Ocupação por horário operacional
            </h3>
            <div className="rounded-xl border border-border/50 bg-surface/50 px-3 py-1.5 text-xs font-medium text-foreground">
              <select
                id="dashboard-week-range"
                aria-label="Selecionar semana de performance"
                value={normalizedSelectedWeekKey}
                onChange={(event) => setSelectedWeekKey(event.target.value)}
                className="bg-transparent outline-none"
              >
                {weekOptions.map((option) => (
                  <option key={option.key} value={option.key}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <p className="mt-2 text-sm text-muted">
            Comparativo da semana selecionada com o mesmo recorte do mês
            anterior, respeitando a janela operacional da empresa.
          </p>
        </div>

        <div className="mt-8">
          <WeeklyPerformanceChart
            current={weeklySeries.current}
            previous={weeklySeries.previous}
            scheduleInitTime={systemConfig.schedule_init_time}
            scheduleEndTime={systemConfig.schedule_end_time}
          />
        </div>

        <div className="mt-6 flex flex-wrap gap-6 text-sm text-muted">
          <LegendDot color="bg-sky-400" label="Mês atual" />
          <LegendDot color="bg-amber-400" label="Mês anterior" />
        </div>
      </section>

      <section className="p-6 lg:p-10">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-col gap-1">
            <p className="app-eyebrow">
              Operação atual
            </p>
            <h3 className="font-display text-2xl text-foreground">
              Agendamentos em andamento
            </h3>
            <span className="text-xs font-medium text-muted">
              {currentShiftLabel}
            </span>
          </div>
          <div className="rounded-[1.4rem] border border-border/50 bg-surface/50 px-4 py-2 text-sm font-medium text-foreground">
            Meta mensal concluída: {completionPercentage}%
          </div>
        </div>

        <div className="mt-8 space-y-4">
          {shiftSchedules.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-border/50 p-8 text-center text-muted">
              Nenhum agendamento ocorrendo neste momento.
            </div>
          ) : (
            shiftSchedules.map((item) => (
              <article
                key={item.id}
                className="group flex items-center justify-between gap-4 rounded-[1.8rem] border border-border/50 bg-surface/30 p-4 transition hover:border-border hover:bg-surface/60"
              >
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-surface border border-border/50 text-primary shadow-sm">
                    <PawPrint className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="font-medium text-foreground group-hover:text-primary transition">
                      {item.pet_name ?? 'Pet sem nome'}
                    </p>
                    <p className="text-sm text-muted">
                      {item.client_name ?? 'Tutor não identificado'}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-8">
                  <div className="flex items-center gap-2">
                    <span
                      className={`h-2 w-2 rounded-full ${resolveScheduleStatusDotClass(item.current_status)}`}
                    />
                    <span className="text-sm font-medium text-foreground">
                      {formatScheduleStatus(item.current_status)}
                    </span>
                  </div>
                  <div className="hidden sm:flex items-center gap-1.5 text-sm text-muted">
                    <Clock3 className="h-4 w-4" />
                    {formatElapsedTime(
                      item,
                      now,
                      scheduleHistoryMap.get(item.id) ?? [],
                    )}
                  </div>
                </div>
              </article>
            ))
          )}
        </div>
      </section>
    </main>
  );
}

function AdminStatCard({
  label,
  value,
  change,
  changeLabel,
  description,
  icon: Icon,
}: {
  label: string;
  value: string;
  change: number;
  changeLabel: string;
  description: string;
  icon: typeof CalendarDays;
}) {
  const isNeutral = change === 0;
  const positive = change > 0;

  return (
    <article className="app-card transition hover:border-border hover:bg-surface/80 p-6 group relative">
      <p className="app-eyebrow">{label}</p>

      <div className="mt-4 flex items-center gap-4">
        <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl border border-border/50 bg-surface/50 text-muted shadow-sm transition-colors group-hover:border-primary/50 group-hover:bg-primary/10 group-hover:text-primary">
          <Icon className="h-6 w-6" />
        </div>
        <div className="min-w-0 flex flex-1 justify-center">
          <div className="flex flex-col items-center justify-center gap-1 text-center">
            <p className="font-display text-4xl text-foreground">{value}</p>
            <div
              className={`inline-flex items-center justify-center gap-1 text-xs font-bold ${
                isNeutral
                  ? 'text-muted'
                  : positive
                    ? 'text-emerald-500'
                    : 'text-rose-500'
              }`}
            >
              {!isNeutral ? (
                positive ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />
              ) : null}
              {isNeutral ? 'estável' : `${positive ? '+' : ''}${change}`}
              {!isNeutral && changeLabel ? ` ${changeLabel}` : ''}
            </div>
          </div>
        </div>
      </div>

      <div className="pointer-events-none absolute -top-12 right-0 z-10 w-64 translate-x-4 opacity-0 transition-all duration-300 group-hover:-translate-x-2 group-hover:opacity-100">
        <div className="rounded-2xl bg-surface border border-border p-4 shadow-glow ring-1 ring-primary/20">
          <p className="text-xs leading-relaxed text-muted">{description}</p>
        </div>
      </div>
    </article>
  );
}

function WeeklyPerformanceChart({
  current,
  previous,
  scheduleInitTime,
  scheduleEndTime,
}: {
  current: Array<{ label: string; value: number | null }>;
  previous: Array<{ label: string; value: number | null }>;
  scheduleInitTime: string;
  scheduleEndTime: string;
}) {
  const startHour = parseHourValue(scheduleInitTime);
  const endHour = parseHourValue(scheduleEndTime);
  const hourSpan = Math.max(1, endHour - startHour);
  const guideHours = buildHourGuides(startHour, endHour);

  const buildPoints = (values: Array<{ value: number | null }>) =>
    values
      .map((item, index) => {
        if (item.value === null) {
          return null;
        }
        const x = 36 + index * 44;
        const y = 160 - ((item.value - startHour) / hourSpan) * 120;
        return `${x},${y}`;
      })
      .filter(Boolean)
      .join(' ');

  return (
    <div className="relative h-[240px] w-full">
      <svg
        viewBox="0 0 360 200"
        className="h-full w-full"
        preserveAspectRatio="none"
      >
        {guideHours.map((hour) => {
          const y = 160 - ((hour - startHour) / hourSpan) * 120;
          return (
            <g key={hour}>
              <line
                x1="36"
                y1={y}
                x2="324"
                y2={y}
                className="stroke-border/50"
                strokeWidth="1"
              />
              <text
                x="24"
                y={y + 4}
                textAnchor="end"
                className="fill-muted text-[10px] font-medium"
              >
                {formatHourLabel(hour)}
              </text>
            </g>
          );
        })}

        <polyline
          fill="none"
          stroke="#fbbf24"
          strokeWidth="3"
          strokeLinejoin="round"
          strokeLinecap="round"
          points={buildPoints(previous)}
          className="opacity-40"
        />

        <polyline
          fill="none"
          stroke="#38bdf8"
          strokeWidth="4"
          strokeLinejoin="round"
          strokeLinecap="round"
          points={buildPoints(current)}
          className="drop-shadow-[0px_4px_8px_rgba(56,189,248,0.2)]"
        />

        {current.map((item, index) => {
          const x = 36 + index * 44;
          return (
            <text
              key={item.label}
              x={x}
              y="190"
              textAnchor="middle"
              className="fill-muted text-[10px] font-medium"
            >
              {item.label}
            </text>
          );
        })}
      </svg>
    </div>
  );
}

function LegendDot({ color, label }: { color: string; label: string }) {
  return (
    <div className="inline-flex items-center gap-2">
      <span className={`h-2.5 w-2.5 rounded-full ${color}`} />
      <span>{label}</span>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-4 p-6 lg:p-10">
      <div className="h-52 animate-pulse rounded-[1.85rem] bg-surface" />
      <div className="grid gap-4 xl:grid-cols-[1.4fr_0.9fr]">
        <div className="h-80 animate-pulse rounded-[1.85rem] bg-surface" />
        <div className="h-80 animate-pulse rounded-[1.85rem] bg-surface" />
      </div>
    </div>
  );
}

function countSchedulesInDay(schedules: ScheduleDTO[], date: Date) {
  return schedules.filter((item) =>
    isSameDay(new Date(item.scheduled_at), date),
  ).length;
}

function countSchedulesInMonth(schedules: ScheduleDTO[], date: Date) {
  return schedules.filter((item) => {
    const value = new Date(item.scheduled_at);
    return (
      value.getFullYear() === date.getFullYear() &&
      value.getMonth() === date.getMonth()
    );
  }).length;
}

function calculateMonthlyTarget(config: CompanySystemConfigDTO, date: Date) {
  const daysInMonth = new Date(
    date.getFullYear(),
    date.getMonth() + 1,
    0,
  ).getDate();
  let totalWorkDays = 0;

  for (let day = 1; day <= daysInMonth; day += 1) {
    const value = new Date(date.getFullYear(), date.getMonth(), day);
    if (config.schedule_days.includes(resolveWeekDay(value))) {
      totalWorkDays += 1;
    }
  }

  return totalWorkDays * config.min_schedules_per_day;
}

function resolveShiftSchedules(schedules: ScheduleDTO[], now: Date) {
  const morningShift = now.getHours() < 12;

  return schedules
    .filter((item) => SHIFT_SCHEDULE_STATUSES.includes(item.current_status))
    .filter((item) => {
      const value = new Date(item.scheduled_at);
      if (!isSameDay(value, now)) {
        return false;
      }
      return morningShift ? value.getHours() < 12 : value.getHours() >= 12;
    })
    .sort((left, right) => left.scheduled_at.localeCompare(right.scheduled_at));
}

function buildWeeklySeries(
  schedules: ScheduleDTO[],
  now: Date,
  config: CompanySystemConfigDTO,
  selectedWeekKey: string,
) {
  const lastDayOfMonth = new Date(
    now.getFullYear(),
    now.getMonth() + 1,
    0,
  ).getDate();
  const parsedWeek = parseWeekKey(selectedWeekKey);
  const weekStart = clamp(parsedWeek.start, 1, lastDayOfMonth);
  const weekEnd = clamp(parsedWeek.end, weekStart, lastDayOfMonth);
  const current = [];

  for (let day = weekStart; day <= weekEnd; day += 1) {
    const currentDate = new Date(now.getFullYear(), now.getMonth(), day);
    current.push({
      label: String(day).padStart(2, '0'),
      value: averageScheduledHourInDay(schedules, currentDate, config),
    });
  }

  const previousMonthDate = shiftDate(now, { months: -1 });
  const previousMonthLastDay = new Date(
    previousMonthDate.getFullYear(),
    previousMonthDate.getMonth() + 1,
    0,
  ).getDate();
  const previous = current.map((item, index) => {
    const previousDay = Math.min(weekStart + index, previousMonthLastDay);
    const targetDate = new Date(
      previousMonthDate.getFullYear(),
      previousMonthDate.getMonth(),
      previousDay,
    );

    return {
      label: item.label,
      value: averageScheduledHourInDay(schedules, targetDate, config),
    };
  });

  return {
    current,
    previous,
  };
}

function buildWeekOptions(date: Date) {
  const lastDayOfMonth = new Date(
    date.getFullYear(),
    date.getMonth() + 1,
    0,
  ).getDate();
  const options = [];

  for (let start = 1; start <= lastDayOfMonth; start += 7) {
    const end = Math.min(start + 6, lastDayOfMonth);
    options.push({
      key: `${start}-${end}`,
      label:
        `${String(start).padStart(2, '0')}-${String(end).padStart(2, '0')} ${date.toLocaleDateString('pt-BR', { month: 'short' })}`.replace(
          '.',
          '',
        ),
    });
  }

  return options;
}

function resolveCurrentWeekKey(date: Date) {
  const lastDayOfMonth = new Date(
    date.getFullYear(),
    date.getMonth() + 1,
    0,
  ).getDate();
  const start = Math.floor((date.getDate() - 1) / 7) * 7 + 1;
  const end = Math.min(start + 6, lastDayOfMonth);

  return `${start}-${end}`;
}

function formatLongDate(date: Date) {
  return date.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
  });
}

function formatScheduleStatus(status: ScheduleStatus) {
  switch (status) {
    case 'waiting':
      return 'Aguardando';
    case 'confirmed':
      return 'Confirmado';
    case 'in_progress':
      return 'Em andamento';
    case 'finished':
      return 'Finalizado';
    case 'delivered':
      return 'Entregue';
    case 'canceled':
      return 'Cancelado';
    default:
      return status;
  }
}

function resolveScheduleStatusDotClass(status: ScheduleStatus) {
  switch (status) {
    case 'waiting':
      return 'bg-amber-400';
    case 'confirmed':
      return 'bg-sky-400';
    case 'in_progress':
      return 'bg-violet-400';
    case 'finished':
      return 'bg-emerald-400';
    case 'delivered':
      return 'bg-stone-400';
    case 'canceled':
      return 'bg-rose-400';
    default:
      return 'bg-stone-300';
  }
}

function formatElapsedTime(
  item: ScheduleDTO,
  now: Date,
  historyItems: ScheduleHistoryItemDTO[],
) {
  const startedAt = new Date(item.scheduled_at);
  const finishedAt = resolveScheduleFinishedAt(item, historyItems);
  const diffMs = Math.max(
    0,
    (finishedAt ?? now).getTime() - startedAt.getTime(),
  );
  const totalMinutes = Math.floor(diffMs / 60000);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;

  if (hours === 0) {
    return `${minutes}min`;
  }

  return `${hours}h ${String(minutes).padStart(2, '0')}min`;
}

function resolveScheduleFinishedAt(
  item: ScheduleDTO,
  historyItems: ScheduleHistoryItemDTO[],
) {
  if (
    item.current_status !== 'finished' &&
    item.current_status !== 'delivered'
  ) {
    return null;
  }

  const historyEntry = historyItems.find(
    (historyItem) => historyItem.status === item.current_status,
  );

  if (historyEntry) {
    return new Date(historyEntry.changed_at);
  }

  if (item.estimated_end) {
    return new Date(item.estimated_end);
  }

  return new Date(item.scheduled_at);
}

function averageScheduledHourInDay(
  schedules: ScheduleDTO[],
  date: Date,
  config: CompanySystemConfigDTO,
) {
  const start = parseHourValue(config.schedule_init_time);
  const end = parseHourValue(config.schedule_end_time);
  const values = schedules
    .filter((item) => isSameDay(new Date(item.scheduled_at), date))
    .map((item) => {
      const value = new Date(item.scheduled_at);
      return value.getHours() + value.getMinutes() / 60;
    })
    .filter((hour) => hour >= start && hour <= end);

  if (values.length === 0) {
    return null;
  }

  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function resolveWeekDay(
  date: Date,
): CompanySystemConfigDTO['schedule_days'][number] {
  return [
    'sunday',
    'monday',
    'tuesday',
    'wednesday',
    'thursday',
    'friday',
    'saturday',
  ][date.getDay()] as CompanySystemConfigDTO['schedule_days'][number];
}

function shiftDate(date: Date, shift: { days?: number; months?: number }) {
  const value = new Date(date);
  if (shift.days) {
    value.setDate(value.getDate() + shift.days);
  }
  if (shift.months) {
    value.setMonth(value.getMonth() + shift.months);
  }
  return value;
}

function parseWeekKey(value: string) {
  const [startRaw, endRaw] = value.split('-');
  const start = Number.parseInt(startRaw ?? '', 10);
  const end = Number.parseInt(endRaw ?? '', 10);

  return {
    start: Number.isFinite(start) ? start : 1,
    end: Number.isFinite(end) ? end : 7,
  };
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function isSameDay(left: Date, right: Date) {
  return (
    left.getDate() === right.getDate() &&
    left.getMonth() === right.getMonth() &&
    left.getFullYear() === right.getFullYear()
  );
}

function parseHourValue(time: string) {
  const [hoursRaw, minutesRaw] = time.split(':');
  const hours = Number.parseInt(hoursRaw ?? '0', 10);
  const minutes = Number.parseInt(minutesRaw ?? '0', 10);

  return hours + minutes / 60;
}

function buildHourGuides(startHour: number, endHour: number) {
  const totalSlots = Math.max(
    2,
    Math.min(5, Math.round(endHour - startHour) + 1),
  );
  const step = totalSlots === 1 ? 1 : (endHour - startHour) / (totalSlots - 1);

  return Array.from({ length: totalSlots }, (_, index) => {
    const value = startHour + step * index;
    return Math.round(value * 2) / 2;
  });
}

function formatHourLabel(hourValue: number) {
  const hours = Math.floor(hourValue);
  const minutes = hourValue % 1 === 0.5 ? '30' : '00';
  return `${String(hours).padStart(2, '0')}h${minutes === '30' ? '30' : ''}`;
}
