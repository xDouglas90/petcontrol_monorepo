import { Navigate, useParams } from '@tanstack/react-router';
import { buildCompanyRoute } from '@petcontrol/shared-constants';
import { Settings, ShieldCheck } from 'lucide-react';

import { useCurrentUserQuery } from '@/lib/api/domain.queries';

export function SettingsPage() {
  const currentUserQuery = useCurrentUserQuery();
  const params = useParams({ strict: false });
  const companySlug =
    typeof params.companySlug === 'string' ? params.companySlug : '';

  if (currentUserQuery.isLoading || !currentUserQuery.data) {
    return (
      <section className="rounded-[1.75rem] border border-stone-200 bg-white p-6 shadow-[0_18px_45px_rgba(15,23,42,0.08)]">
        <p className="text-sm text-stone-500">Carregando contexto de configurações...</p>
      </section>
    );
  }

  const settingsAccess = currentUserQuery.data.settings_access;
  const canViewSettings =
    settingsAccess?.can_view ?? currentUserQuery.data.role === 'admin';
  const canManagePermissions =
    settingsAccess?.can_manage_permissions ?? currentUserQuery.data.role === 'admin';
  const editablePermissionCodes = settingsAccess?.editable_permission_codes ?? [];
  const isReadOnly = !canManagePermissions && editablePermissionCodes.length === 0;

  if (!canViewSettings && companySlug) {
    return <Navigate to={buildCompanyRoute(companySlug, 'dashboard')} replace />;
  }

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
              O acesso a esta area agora respeita as permissoes efetivas do
              usuario autenticado. A tela completa entra na proxima fase, com
              formularios reais para empresa, negocio e gestao de permissoes.
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
            {canManagePermissions
              ? 'Seu perfil pode editar configuracoes e tambem gerenciar permissoes de usuarios.'
              : isReadOnly
                ? 'Seu perfil pode visualizar a area de configuracoes, mas ainda esta em modo somente leitura.'
                : `Seu perfil tem edicao parcial nesta area: ${editablePermissionCodes.join(', ')}.`}
          </p>
        </div>
      </section>
    </div>
  );
}
