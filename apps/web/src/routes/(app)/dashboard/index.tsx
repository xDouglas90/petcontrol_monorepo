import {
  Activity,
  ArrowUpRight,
  CalendarDays,
  CircleAlert,
  CircleCheck,
  PawPrint,
  ReceiptText,
  Users,
} from 'lucide-react';
import {
  formatScheduleStatus,
  resolveAsyncViewState,
  scheduleStatusColorClass,
} from '@petcontrol/ui/web';

import { selectSession, useAuthStore } from '@/lib/auth/auth.store';
import {
  useCurrentCompanyQuery,
  useSchedulesQuery,
} from '@/lib/api/domain.queries';
import { useUIStore } from '@/stores/ui.store';

export function DashboardPage() {
  const session = useAuthStore(selectSession);
  const theme = useUIStore((state) => state.theme);
  const companyQuery = useCurrentCompanyQuery();
  const schedulesQuery = useSchedulesQuery();

  const schedules = schedulesQuery.data ?? [];
  const today = new Date();
  const todaySchedules = schedules.filter((item) => {
    const scheduleDate = new Date(item.scheduled_at);
    return (
      scheduleDate.getDate() === today.getDate() &&
      scheduleDate.getMonth() === today.getMonth() &&
      scheduleDate.getFullYear() === today.getFullYear()
    );
  });
  const confirmedCount = schedules.filter(
    (item) => item.current_status === 'confirmed',
  ).length;
  const waitingCount = schedules.filter(
    (item) => item.current_status === 'waiting',
  ).length;

  const stats = [
    {
      label: 'Agendamentos hoje',
      value: String(todaySchedules.length),
      detail: `${confirmedCount} confirmados`,
      icon: CalendarDays,
    },
    {
      label: 'Schedules no tenant',
      value: String(schedules.length),
      detail: `${waitingCount} aguardando`,
      icon: PawPrint,
    },
    {
      label: 'Usuário atual',
      value: session?.role ?? '-',
      detail: session?.kind ?? 'sem sessão',
      icon: Users,
    },
    {
      label: 'Pacote ativo',
      value: companyQuery.data?.active_package ?? '-',
      detail: companyQuery.data?.is_active ? 'empresa ativa' : 'inativa',
      icon: ReceiptText,
    },
  ];

  const viewState = resolveAsyncViewState({
    isLoading: schedulesQuery.isLoading,
    isError: schedulesQuery.isError,
    itemCount: schedules.length,
  });

  return (
    <div className="space-y-6">
      <section className="grid gap-4 lg:grid-cols-[1.45fr_0.95fr]">
        <div className="rounded-[1.75rem] border border-white/10 bg-gradient-to-br from-white/10 via-white/5 to-transparent p-6 shadow-glow">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-3">
              <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
                Visão geral
              </p>
              <h2 className="font-display text-4xl text-white">
                {companyQuery.data?.fantasy_name ?? 'Centro de operações'}
              </h2>
              <p className="max-w-2xl text-sm leading-7 text-slate-300">
                Dashboard conectado ao backend com dados reais de tenant,
                empresa e agendamentos.
              </p>
            </div>

            <div className="rounded-3xl border border-white/10 bg-slate-950/60 px-4 py-3 text-sm text-slate-300">
              <p className="text-xs uppercase tracking-[0.3em] text-secondary/70">
                Sessão
              </p>
              <div className="mt-2 space-y-1">
                <p className="font-medium text-white">
                  {companyQuery.data?.name ?? session?.companyId.slice(0, 8)}
                </p>
                <p>
                  {new Date().toLocaleDateString('pt-BR', {
                    dateStyle: 'full',
                  })}
                </p>
              </div>
            </div>
          </div>

          <div className="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {stats.map((stat) => (
              <article
                key={stat.label}
                className="rounded-3xl border border-white/10 bg-slate-950/60 p-4 backdrop-blur"
              >
                <div className="flex items-center justify-between gap-3 text-slate-300">
                  <span className="text-sm">{stat.label}</span>
                  <stat.icon className="h-4 w-4 text-primary" />
                </div>
                <p className="mt-4 font-display text-3xl text-white">
                  {stat.value}
                </p>
                <p className="mt-2 text-sm text-slate-400">{stat.detail}</p>
              </article>
            ))}
          </div>
        </div>

        <aside className="space-y-4 rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6 shadow-glow">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
                Próximas ações
              </p>
              <h3 className="mt-2 font-display text-2xl text-white">
                Rotina do dia
              </h3>
            </div>
            <ArrowUpRight className="h-5 w-5 text-primary" />
          </div>

          <div className="space-y-3">
            {[
              'Revisar schedules aguardando confirmação',
              'Atualizar status de atendimentos em andamento',
              'Validar consistência de agenda do tenant',
            ].map((item) => (
              <div
                key={item}
                className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-200"
              >
                {item}
              </div>
            ))}
          </div>

          <div className="rounded-3xl border border-white/10 bg-[rgba(245,158,11,0.08)] p-4">
            <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
              Tenant
            </p>
            <p className="mt-2 font-display text-xl text-white">
              tenant isolado por token no backend
            </p>
            <p className="mt-1 text-sm leading-6 text-slate-300">
              O frontend não injeta tenant manualmente em payloads; o backend já
              controla isso na camada de auth.
            </p>
          </div>
        </aside>
      </section>

      <section className="grid gap-4 lg:grid-cols-[0.95fr_1.05fr]">
        <div className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
                Status
              </p>
              <h3 className="mt-2 font-display text-2xl text-white">
                Estado do sistema
              </h3>
            </div>
            <Activity className="h-5 w-5 text-primary" />
          </div>

          <div className="mt-6 space-y-4">
            <StatusRow label="API" value="Conectada" />
            <StatusRow
              label="Auth"
              value={session?.accessToken ? 'JWT persistido' : 'sem sessão'}
            />
            <StatusRow label="UI" value="Zustand + Query" />
            <StatusRow label="Tema" value={theme} />
          </div>
        </div>

        <div className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
          <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
            Schedules recentes
          </p>
          <h3 className="mt-2 font-display text-2xl text-white">
            Últimas movimentações
          </h3>

          <div className="mt-6 space-y-4">
            {viewState === 'loading' ? (
              <StateRow
                icon={Activity}
                message="Carregando schedules do backend..."
              />
            ) : null}

            {viewState === 'error' ? (
              <StateRow
                icon={CircleAlert}
                message="Falha ao carregar schedules da API."
              />
            ) : null}

            {viewState === 'empty' ? (
              <StateRow
                icon={CircleCheck}
                message="Sem schedules cadastrados para este tenant."
              />
            ) : null}

            {viewState === 'ready'
              ? schedules.slice(0, 4).map((item) => (
              <div
                key={item.id}
                className="flex items-start gap-4 rounded-2xl border border-white/10 bg-white/5 p-4"
              >
                <div className="mt-1 h-2.5 w-2.5 rounded-full bg-primary" />
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="font-medium text-white">
                      {new Date(item.scheduled_at).toLocaleString('pt-BR', {
                        dateStyle: 'short',
                        timeStyle: 'short',
                      })}
                    </p>
                    <span
                      className={`rounded-full border px-2 py-1 text-xs ${scheduleStatusColorClass(item.current_status)}`}
                    >
                      {formatScheduleStatus(item.current_status)}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-slate-300">
                    {item.notes || 'Sem observações'}
                  </p>
                </div>
              </div>
                ))
              : null}
          </div>
        </div>
      </section>
    </div>
  );
}

function StatusRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm">
      <span className="text-slate-300">{label}</span>
      <span className="font-medium text-white">{value}</span>
    </div>
  );
}

function StateRow({
  icon: Icon,
  message,
}: {
  icon: typeof Activity;
  message: string;
}) {
  return (
    <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-slate-300">
      <Icon className="h-4 w-4 text-primary" />
      <span>{message}</span>
    </div>
  );
}
