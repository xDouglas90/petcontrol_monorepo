import { Settings, ShieldCheck } from 'lucide-react';

export function SettingsPage() {
  return (
    <div className="space-y-6">
      <section className="rounded-[1.75rem] border border-stone-200 bg-white p-6 shadow-[0_18px_45px_rgba(15,23,42,0.08)]">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.3em] text-stone-400">
              Configuracoes
            </p>
            <h2 className="mt-3 font-display text-3xl text-stone-900">
              Central de ajustes do tenant
            </h2>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-stone-500">
              Esta tela entra agora como placeholder da nova navegacao. Nas
              proximas fases ela sera preenchida com configuracoes operacionais,
              preferências visuais e controles do tenant.
            </p>
          </div>

          <div className="rounded-3xl bg-stone-100 p-3 text-stone-600">
            <Settings className="h-6 w-6" />
          </div>
        </div>
      </section>

      <section className="rounded-[1.75rem] border border-dashed border-stone-300 bg-stone-50 p-6">
        <div className="flex items-center gap-3">
          <ShieldCheck className="h-5 w-5 text-emerald-600" />
          <p className="text-sm font-medium text-stone-700">
            O shell novo ja pode direcionar o usuario para configuracoes sem
            quebrar a navegacao.
          </p>
        </div>
      </section>
    </div>
  );
}
