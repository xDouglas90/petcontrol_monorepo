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
  MessageSquareText,
  PawPrint,
  ShieldCheck,
  TrendingDown,
  TrendingUp,
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useInternalChatSocket } from '@/hooks/use-internal-chat-socket';

import {
  useAdminSystemChatMessagesQuery,
  useCreateAdminSystemChatMessageMutation,
  useCompanyUsersQuery,
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
  const companyUsersQuery = useCompanyUsersQuery();

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
        .filter((item) => ['finished', 'delivered'].includes(item.current_status))
        .map((item) => item.id),
    [schedules],
  );
  const scheduleHistoryQueries = useScheduleHistoriesQuery(historyScheduleIds);
  const now = new Date();
  const weekOptions = buildWeekOptions(now);
  const defaultWeekKey = resolveCurrentWeekKey(now);
  const [selectedWeekKey, setSelectedWeekKey] = useState(defaultWeekKey);
  const [selectedSystemContactId, setSelectedSystemContactId] = useState(
    'contract-pending',
  );
  const [chatDraft, setChatDraft] = useState('');
  const chatMessagesContainerRef = useRef<HTMLDivElement | null>(null);

  // Derive the effective system user ID for chat hooks (must be before early returns)
  const preliminarySystemUsers = useMemo(
    () =>
      (companyUsersQuery.data ?? []).filter(
        (user) => user.role === 'system' && user.user_id !== currentUser?.user_id,
      ),
    [companyUsersQuery.data, currentUser?.user_id],
  );
  const effectiveSystemContactId = preliminarySystemUsers.some(
    (user) => user.user_id === selectedSystemContactId,
  )
    ? selectedSystemContactId
    : (preliminarySystemUsers[0]?.user_id ?? undefined);
  const chatMessagesQuery = useAdminSystemChatMessagesQuery(effectiveSystemContactId);
  const sendChatMessageMutation =
    useCreateAdminSystemChatMessageMutation(effectiveSystemContactId);

  const { presenceMap } = useInternalChatSocket(effectiveSystemContactId);

  useEffect(() => {
    const container = chatMessagesContainerRef.current;
    if (!container) {
      return;
    }

    container.scrollTo({
      top: container.scrollHeight,
      behavior: 'smooth',
    });
  }, [chatMessagesQuery.data, effectiveSystemContactId]);

  const scheduleHistoryMap = useMemo(() => {
    const entries = historyScheduleIds.map((scheduleId, index) => [
      scheduleId,
      scheduleHistoryQueries[index]?.data ?? [],
    ] as const);

    return new Map<string, ScheduleHistoryItemDTO[]>(entries);
  }, [historyScheduleIds, scheduleHistoryQueries]);

  if (
    companyQuery.isLoading ||
    currentUserQuery.isLoading ||
    systemConfigQuery.isLoading ||
    schedulesQuery.isLoading ||
    companyUsersQuery.isLoading
  ) {
    return <DashboardSkeleton />;
  }

  if (
    companyQuery.isError ||
    currentUserQuery.isError ||
    systemConfigQuery.isError ||
    schedulesQuery.isError ||
    companyUsersQuery.isError ||
    !company ||
    !currentUser ||
    !systemConfig
  ) {
    return (
      <section className="rounded-[1.75rem] border border-rose-200 bg-rose-50 p-6 text-rose-700">
        <p className="text-sm font-semibold uppercase tracking-[0.24em]">
          Dashboard indisponível
        </p>
        <h2 className="mt-3 font-display text-3xl text-rose-900">
          Não foi possível carregar os dados operacionais do tenant.
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
      <section className="rounded-[1.75rem] border border-stone-200 bg-white p-6 shadow-[0_18px_45px_rgba(15,23,42,0.06)]">
        <p className="text-sm font-semibold uppercase tracking-[0.24em] text-stone-400">
          Home em preparação
        </p>
        <h2 className="mt-3 font-display text-3xl text-stone-900">
          A experiência inicial para o perfil{' '}
          <span className="lowercase">{currentUser.role}</span> ainda será
          construída.
        </h2>
        <p className="mt-3 max-w-2xl text-sm leading-7 text-stone-500">
          Nesta etapa, o dashboard completo está sendo priorizado para o papel
          de administrador do tenant.
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
  const chatContacts = buildSystemContactOptions(
    companyUsersQuery.data ?? [],
    currentUser.user_id,
  );
  const normalizedSelectedSystemContactId = chatContacts.some(
    (contact) => contact.id === selectedSystemContactId,
  )
    ? selectedSystemContactId
    : (chatContacts[0]?.id ?? 'contract-pending');
  const selectedSystemContact =
    chatContacts.find((contact) => contact.id === normalizedSelectedSystemContactId) ??
    chatContacts[0];

  const contactPresence = effectiveSystemContactId ? presenceMap[effectiveSystemContactId] : undefined;
  const isContactOnline = contactPresence?.status === 'online';

  const stats = [
    {
      label: 'Agendamentos/dia',
      value: String(todayCount),
      change: todayCount - previousDayCount,
      changeLabel: 'vs ontem',
      description: 'Atualizado a cada novo agendamento criado no tenant.',
      icon: CalendarDays,
    },
    {
      label: 'Agendamentos/mês',
      value: String(currentMonthCount),
      change: currentMonthCount - previousMonthCount,
      changeLabel: 'vs mês anterior',
      description: 'Volume operacional acumulado no mês corrente.',
      icon: CalendarRange,
    },
    {
      label: 'Eficiência',
      value: `${Math.round(efficiencyPercentage)}%`,
      change: Math.round(efficiencyPercentage - 100),
      changeLabel:
        efficiencyPercentage >= 100 ? 'acima da meta' : 'abaixo da meta',
      description: `${currentMonthCount} de ${monthlyTarget} agendamentos mínimos previstos.`,
      icon: Activity,
    },
  ] as const;

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_24rem]">
      <div className="flex min-w-0 flex-col gap-6">
        <header className="rounded-[2.5rem] border border-white/70 bg-white/85 px-6 py-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur-xl lg:px-8">
          <div className="flex flex-wrap items-start justify-between gap-6">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.34em] text-stone-400">
                Dashboard admin
              </p>
              <h1 className="mt-3 font-display text-4xl text-stone-950 sm:text-5xl">
                Olá, {greetingName}
              </h1>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-stone-500">
                Você está visualizando a operação do tenant{' '}
                {company.fantasy_name}, com foco em agenda diária, comparação
                mensal e eficiência da meta mínima configurada.
              </p>
            </div>

            <div className="flex items-center gap-3 rounded-[1.6rem] border border-stone-200 bg-stone-50/80 px-4 py-3 shadow-sm">
              <div className="flex h-10 w-10 items-center justify-center rounded-2xl border border-stone-200 bg-white text-stone-500">
                <CalendarDays className="h-4 w-4" />
              </div>
              <div>
                <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-stone-400">
                  Hoje
                </p>
                <span className="text-sm font-medium text-stone-700">
                  {formatLongDate(now)}
                </span>
              </div>
            </div>
          </div>
        </header>

        <section className="grid gap-4">
          <div className="grid gap-4 sm:grid-cols-3">
            {stats.map((stat) => (
              <AdminStatCard key={stat.label} {...stat} />
            ))}
          </div>
        </section>

        <section className="rounded-[2.5rem] border border-stone-100 bg-white p-6 shadow-[0_20px_50px_rgba(0,0,0,0.04)] lg:p-8">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.28em] text-stone-400">
                Performance
              </p>
              <h3 className="mt-2 font-display text-2xl text-stone-950">
                Ocupação por horário operacional
              </h3>
              <p className="mt-2 text-sm text-stone-500">
                Comparativo da semana selecionada com o mesmo recorte do mês
                anterior, respeitando a janela operacional do tenant.
              </p>
            </div>
            <div className="rounded-xl border border-stone-200 bg-stone-50 px-3 py-1.5 text-xs font-medium text-stone-600">
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

          <div className="mt-8">
            <WeeklyPerformanceChart
              current={weeklySeries.current}
              previous={weeklySeries.previous}
              scheduleInitTime={systemConfig.schedule_init_time}
              scheduleEndTime={systemConfig.schedule_end_time}
            />
          </div>

          <div className="mt-6 flex flex-wrap gap-6 text-sm text-stone-500">
            <LegendDot color="bg-sky-400" label="Mês atual" />
            <LegendDot color="bg-amber-400" label="Mês anterior" />
          </div>
        </section>

        <section className="rounded-[2.5rem] border border-stone-100 bg-white p-6 shadow-[0_20px_50px_rgba(0,0,0,0.04)] lg:p-8">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex flex-col gap-1">
              <p className="text-xs font-semibold uppercase tracking-[0.28em] text-stone-400">
                Operação atual
              </p>
              <h3 className="font-display text-2xl text-stone-950">
                Agendamentos em andamento
              </h3>
              <span className="text-xs font-medium text-stone-400">
                {currentShiftLabel}
              </span>
            </div>
            <div className="rounded-[1.4rem] border border-stone-200 bg-stone-50 px-4 py-2 text-sm font-medium text-stone-600">
              Meta mensal concluída: {completionPercentage}%
            </div>
          </div>

          <div className="mt-8 space-y-4">
            {shiftSchedules.length === 0 ? (
              <div className="rounded-2xl border border-dashed border-stone-200 p-8 text-center text-stone-400">
                Nenhum atendimento registrado para o turno atual.
              </div>
            ) : (
              shiftSchedules.map((item) => (
                <article
                  key={item.id}
                  className="group flex items-center justify-between gap-4 rounded-[1.8rem] border border-stone-100 bg-stone-50/60 p-4 transition hover:border-stone-200 hover:bg-stone-50"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-white text-sky-500 shadow-sm">
                      <PawPrint className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="font-medium text-stone-900 group-hover:text-sky-600 transition">
                        {item.pet_name ?? 'Pet sem nome'}
                      </p>
                      <p className="text-sm text-stone-400">
                        {item.client_name ?? 'Tutor não identificado'}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-8">
                    <div className="flex items-center gap-2">
                      <span
                        className={`h-2 w-2 rounded-full ${resolveScheduleStatusDotClass(item.current_status)}`}
                      />
                      <span className="text-sm font-medium text-stone-600">
                        {formatScheduleStatus(item.current_status)}
                      </span>
                    </div>
                    <div className="hidden sm:flex items-center gap-1.5 text-sm text-stone-400">
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
      </div>

      <aside className="flex w-full shrink-0 flex-col gap-6 xl:w-[24rem]">
        <section className="rounded-[2.5rem] border border-stone-100 bg-white p-8 shadow-[0_20px_50px_rgba(0,0,0,0.04)]">
          <div className="flex flex-col items-center text-center">
            <div className="relative">
              <div className="h-24 w-24 rounded-full border-4 border-stone-50 bg-stone-100 p-1 shadow-sm">
                <img
                  src={
                    currentUser.image_url ||
                    `https://ui-avatars.com/api/?name=${greetingName}&background=0D1117&color=fff`
                  }
                  alt={greetingName}
                  className="h-full w-full rounded-full object-cover"
                />
              </div>
              <div className="absolute bottom-1 right-1 h-4 w-4 rounded-full border-2 border-white bg-emerald-500" />
            </div>
            <h4 className="mt-4 font-display text-xl text-stone-950">
              {greetingName}
            </h4>
            <p className="mt-1 text-sm text-stone-400">
              Administrador do tenant
            </p>

            <div className="mt-6 grid w-full grid-cols-3 gap-3">
              <MiniBadge icon={ShieldCheck} label="Admin" />
              <MiniBadge icon={CalendarDays} label={formatCompactDate(now)} />
              <MiniBadge icon={Activity} label={`${todayCount} hoje`} />
            </div>
          </div>
        </section>

        <section className="flex flex-1 flex-col rounded-[2.5rem] border border-stone-100 bg-white p-6 shadow-[0_20px_50px_rgba(0,0,0,0.04)]">
          <div className="border-b border-stone-100 pb-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.28em] text-stone-400">
                  Chat do sistema
                </p>
                <h5 className="mt-2 font-display text-lg text-stone-950">
                  Suporte ao administrador
                </h5>
              </div>
              <div className="rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">
                Histórico persistido
              </div>
            </div>
            <p className="mt-3 text-sm leading-6 text-stone-500">
              Esta conversa agora persiste mensagens de texto entre o
              administrador do tenant e usuários do tipo system, com suporte a
              sincronização em tempo real e presença.
            </p>
          </div>

          <div className="mt-5">
            <label
              htmlFor="dashboard-system-contact"
              className="text-xs font-semibold uppercase tracking-[0.24em] text-stone-400"
            >
              Usuário system
            </label>
            <div className="mt-2 rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3">
              <select
                id="dashboard-system-contact"
                aria-label="Selecionar usuário system"
                value={normalizedSelectedSystemContactId}
                onChange={(event) => {
                  setSelectedSystemContactId(event.target.value);
                  setChatDraft('');
                }}
                className="w-full bg-transparent text-sm text-stone-700 outline-none"
              >
                {chatContacts.map((contact) => (
                  <option key={contact.id} value={contact.id}>
                    {contact.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="mt-6 flex items-center gap-3 rounded-[1.8rem] border border-stone-100 bg-stone-50/70 p-4">
            <div className="relative">
              {selectedSystemContact.imageUrl ? (
                <img
                  src={selectedSystemContact.imageUrl}
                  alt={selectedSystemContact.name}
                  className="h-12 w-12 rounded-full object-cover"
                />
              ) : (
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-sky-600 text-sm font-semibold uppercase tracking-[0.16em] text-white">
                  {selectedSystemContact.avatar}
                </div>
              )}
              <div
                className={`absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-2 border-white ${
                  isContactOnline ? 'bg-emerald-500' : 'bg-stone-300'
                }`}
              />
            </div>
            <div className="min-w-0">
              <p className="truncate font-medium text-stone-900">
                {selectedSystemContact.name}
              </p>
              <p className="truncate text-sm text-stone-400">
                {selectedSystemContact.subtitle}
              </p>
            </div>
          </div>

          <div
            ref={chatMessagesContainerRef}
            className="mt-6 h-[22rem] space-y-5 overflow-y-auto pr-2"
          >
            {!effectiveSystemContactId ? (
              <div className="rounded-[1.6rem] border border-dashed border-stone-200 bg-stone-50 px-4 py-6 text-sm leading-6 text-stone-500">
                Vincule um usuário do tipo <strong>system</strong> ao tenant
                para iniciar uma conversa persistida com o administrador.
              </div>
            ) : chatMessagesQuery.isLoading ? (
              <div className="rounded-[1.6rem] border border-stone-100 bg-stone-50 px-4 py-6 text-sm text-stone-500">
                Carregando histórico da conversa...
              </div>
            ) : chatMessagesQuery.isError ? (
              <div className="rounded-[1.6rem] border border-rose-100 bg-rose-50 px-4 py-6 text-sm text-rose-600">
                Não foi possível carregar o histórico persistido desta conversa.
              </div>
            ) : (chatMessagesQuery.data?.length ?? 0) === 0 ? (
              <div className="rounded-[1.6rem] border border-dashed border-stone-200 bg-stone-50 px-4 py-6 text-sm leading-6 text-stone-500">
                Ainda não existem mensagens persistidas entre este admin e o
                contato system selecionado.
              </div>
            ) : (
              chatMessagesQuery.data?.map((message) => {
                const isOwnMessage = message.sender_user_id === currentUser.user_id;

                return (
                  <div
                    key={message.id}
                    className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'}`}
                  >
                    <div
                      className={`max-w-[88%] rounded-[1.6rem] px-4 py-3 text-sm leading-6 ${
                        isOwnMessage
                          ? 'bg-sky-500 text-white'
                          : 'border border-stone-100 bg-stone-50 text-stone-600'
                      }`}
                    >
                      <p
                        className={`text-[11px] font-semibold uppercase tracking-[0.18em] ${
                          isOwnMessage ? 'text-white/70' : 'text-stone-400'
                        }`}
                      >
                        {message.sender_name}
                      </p>
                      <p className="mt-2 whitespace-pre-wrap">{message.body}</p>
                      <p
                        className={`mt-2 text-[11px] ${
                          isOwnMessage ? 'text-white/70' : 'text-stone-400'
                        }`}
                      >
                        {formatChatTimestamp(message.created_at)}
                      </p>
                    </div>
                  </div>
                );
              })
            )}
          </div>

          <form
            className="mt-6 rounded-[1.6rem] border border-stone-200 bg-stone-50 px-4 py-4"
            onSubmit={(event) => {
              event.preventDefault();
              const message = chatDraft.trim();
              if (!effectiveSystemContactId || !message || sendChatMessageMutation.isPending) {
                return;
              }

              sendChatMessageMutation.mutate(
                { message },
                {
                  onSuccess: () => {
                    setChatDraft('');
                  },
                },
              );
            }}
          >
            <div className="flex items-center gap-3">
              <MessageSquareText className="h-4 w-4 text-stone-500" />
              <input
                id="dashboard-chat-message"
                name="message"
                type="text"
                autoComplete="off"
                aria-label="Escrever mensagem para usuário system"
                value={chatDraft}
                onChange={(event) => setChatDraft(event.target.value)}
                placeholder={
                  effectiveSystemContactId
                    ? 'Escreva uma mensagem...'
                    : 'Selecione um usuário system para conversar'
                }
                disabled={!effectiveSystemContactId || sendChatMessageMutation.isPending}
                className="w-full bg-transparent text-sm text-stone-700 outline-none placeholder:text-stone-400 disabled:cursor-not-allowed"
              />
              <button
                type="submit"
                disabled={
                  !effectiveSystemContactId ||
                  !chatDraft.trim() ||
                  sendChatMessageMutation.isPending
                }
                className="inline-flex items-center justify-center rounded-xl bg-sky-600 px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white transition hover:bg-sky-700 disabled:cursor-not-allowed disabled:bg-stone-300"
              >
                {sendChatMessageMutation.isPending ? 'Enviando' : 'Enviar'}
              </button>
            </div>
            {sendChatMessageMutation.isError ? (
              <p className="mt-3 text-sm text-rose-600">
                Não foi possível persistir a mensagem desta conversa.
              </p>
            ) : null}
          </form>
        </section>
      </aside>
    </div>
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
  const ChangeIcon = positive ? TrendingUp : TrendingDown;

  return (
    <article className="group relative rounded-[2rem] border border-stone-100 bg-white p-6 shadow-[0_20px_50px_rgba(0,0,0,0.04)] transition hover:shadow-[0_20px_50px_rgba(0,0,0,0.08)]">
      {/* Title at the top */}
      <p className="text-sm font-medium text-stone-400">{label}</p>

      <div className="mt-4 flex items-center gap-4">
        {/* Icon and Value side-by-side */}
        <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl border border-stone-100 bg-stone-50 text-stone-900 shadow-sm transition-colors group-hover:bg-sky-50 group-hover:text-sky-600">
          <Icon className="h-6 w-6" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-baseline gap-3">
            <p className="font-display text-3xl text-stone-950">{value}</p>
            <div
              className={`inline-flex items-center gap-1 text-xs font-bold ${
                isNeutral
                  ? 'text-stone-400'
                  : positive
                    ? 'text-emerald-500'
                    : 'text-rose-500'
              }`}
            >
              {!isNeutral ? <ChangeIcon className="h-3.5 w-3.5 stroke-[3]" /> : null}
              {isNeutral ? 'estável' : `${positive ? '+' : ''}${change}`}
              {!isNeutral ? ` ${changeLabel}` : ''}
            </div>
          </div>
        </div>
      </div>

      {/* Absolute Hover Box */}
      <div className="pointer-events-none absolute inset-x-0 bottom-0 z-10 translate-y-2 px-4 pb-6 opacity-0 transition-all duration-300 group-hover:pointer-events-auto group-hover:translate-y-0 group-hover:opacity-100">
        <div className="rounded-2xl bg-sky-600 p-4 shadow-xl ring-1 ring-sky-300/40">
          <p className="text-xs leading-relaxed text-white/90">
            {description}
          </p>
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
                className="stroke-stone-100"
                strokeWidth="1"
              />
              <text
                x="24"
                y={y + 4}
                textAnchor="end"
                className="fill-stone-300 text-[10px] font-medium"
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
          style={{ filter: 'drop-shadow(0px 4px 8px rgba(56, 189, 248, 0.2))' }}
        />

        {current.map((item, index) => {
          const x = 36 + index * 44;
          return (
            <text
              key={item.label}
              x={x}
              y="190"
              textAnchor="middle"
              className="fill-stone-400 text-[10px] font-medium"
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

function MiniBadge({
  icon: Icon,
  label,
}: {
  icon: typeof CalendarDays;
  label: string;
}) {
  return (
    <div className="rounded-2xl border border-stone-100 bg-stone-50 px-3 py-3 text-center">
      <Icon className="mx-auto h-4 w-4 text-stone-500" />
      <p className="mt-2 text-xs font-medium text-stone-500">{label}</p>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-4">
      <div className="h-52 animate-pulse rounded-[1.85rem] bg-stone-200/70" />
      <div className="grid gap-4 xl:grid-cols-[1.4fr_0.9fr]">
        <div className="h-80 animate-pulse rounded-[1.85rem] bg-stone-200/70" />
        <div className="h-80 animate-pulse rounded-[1.85rem] bg-stone-200/70" />
      </div>
    </div>
  );
}

function countSchedulesInDay(schedules: ScheduleDTO[], date: Date) {
  return schedules.filter((item) => isSameDay(new Date(item.scheduled_at), date))
    .length;
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
      label: `${String(start).padStart(2, '0')}-${String(end).padStart(2, '0')} ${date.toLocaleDateString('pt-BR', { month: 'short' })}`.replace(
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

function formatCompactDate(date: Date) {
  return date.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
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
  if (item.current_status !== 'finished' && item.current_status !== 'delivered') {
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

function resolveWeekDay(date: Date): CompanySystemConfigDTO['schedule_days'][number] {
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

function shiftDate(
  date: Date,
  shift: { days?: number; months?: number },
) {
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
  const totalSlots = Math.max(2, Math.min(5, Math.round(endHour - startHour) + 1));
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

function resolveInitials(value: string) {
  const parts = value
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  if (parts.length === 0) {
    return 'PC';
  }

  return parts.map((part) => part[0]?.toUpperCase() ?? '').join('');
}

function buildSystemContactOptions(
  companyUsers: Array<{
    user_id: string;
    role: string;
    short_name?: string | null;
    full_name?: string | null;
    image_url?: string | null;
  }>,
  currentUserId: string,
) {
  const systemUsers = companyUsers.filter(
    (item) => item.role === 'system' && item.user_id !== currentUserId,
  );

  if (systemUsers.length === 0) {
    return [
      {
        id: 'contract-pending',
        label: 'Nenhum usuário system vinculado',
        name: 'Usuários system',
        subtitle: 'Vincule um contato system ao tenant para habilitar o seletor',
        avatar: 'SY',
        imageUrl: null,
        statusClass: 'bg-stone-400',
      },
    ];
  }

  return systemUsers.map((item) => {
    const name = item.short_name || item.full_name || 'Usuário system';
    return {
      id: item.user_id,
      label: name,
      name,
      subtitle: 'Contato system do tenant',
      avatar: resolveInitials(name),
      imageUrl: item.image_url ?? null,
      statusClass: 'bg-emerald-500',
    };
  });
}

function formatChatTimestamp(value: string) {
  const date = new Date(value);

  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}
