import {
  Activity,
  ArrowUpRight,
  CalendarDays,
  PawPrint,
  ReceiptText,
  Users,
} from 'lucide-react';

import { selectSession, useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const stats = [
  {
    label: 'Agendamentos hoje',
    value: '28',
    detail: '+12% vs ontem',
    icon: CalendarDays,
  },
  {
    label: 'Pets ativos',
    value: '146',
    detail: '8 check-ins pendentes',
    icon: PawPrint,
  },
  {
    label: 'Equipe online',
    value: '09',
    detail: '3 em atendimento',
    icon: Users,
  },
  {
    label: 'Receita parcial',
    value: 'R$ 18.240',
    detail: 'Meta 64% atingida',
    icon: ReceiptText,
  },
];

const activity = [
  {
    title: 'Banho concluído',
    time: 'há 6 min',
    note: 'Golden retriever • João A.',
  },
  {
    title: 'Vacina reagendada',
    time: 'há 18 min',
    note: 'Clínica • retorno em 15/04',
  },
  {
    title: 'Novo cliente importado',
    time: 'há 41 min',
    note: 'Origem: cadastro público',
  },
  {
    title: 'Cobrança confirmada',
    time: 'há 1h',
    note: 'Plano mensal com aditivo',
  },
];

export function DashboardPage() {
  const session = useAuthStore(selectSession);
  const theme = useUIStore((state) => state.theme);

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
                O centro de operações está pronto.
              </h2>
              <p className="max-w-2xl text-sm leading-7 text-slate-300">
                Esta tela já reserva espaço para cards de SLA, pipeline de
                atendimento, funil de vendas e fila de WhatsApp.
              </p>
            </div>

            <div className="rounded-3xl border border-white/10 bg-slate-950/60 px-4 py-3 text-sm text-slate-300">
              <p className="text-xs uppercase tracking-[0.3em] text-secondary/70">
                Sessão
              </p>
              <div className="mt-2 space-y-1">
                <p className="font-medium text-white">
                  {session?.companyId.slice(0, 8)}…
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
              'Confirmar banho das 10h',
              'Liberar acesso do novo módulo',
              'Enviar lembrete de retorno',
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
              company_id visível via middleware
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
            <StatusRow label="Auth" value="JWT persistido" />
            <StatusRow label="UI" value="Zustand + Query" />
            <StatusRow label="Tema" value={theme} />
          </div>
        </div>

        <div className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
          <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
            Atividade recente
          </p>
          <h3 className="mt-2 font-display text-2xl text-white">
            Últimos eventos
          </h3>

          <div className="mt-6 space-y-4">
            {activity.map((item) => (
              <div
                key={item.title}
                className="flex items-start gap-4 rounded-2xl border border-white/10 bg-white/5 p-4"
              >
                <div className="mt-1 h-2.5 w-2.5 rounded-full bg-primary" />
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <p className="font-medium text-white">{item.title}</p>
                    <span className="text-xs uppercase tracking-[0.24em] text-slate-400">
                      {item.time}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-slate-300">{item.note}</p>
                </div>
              </div>
            ))}
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
