import { Navigate, useParams } from '@tanstack/react-router';
import type {
  CompanyUserPermissionDTO,
  CompanyUserPermissionsDTO,
  WeekDay,
} from '@petcontrol/shared-types';
import { WEEK_DAYS } from '@petcontrol/shared-types';
import { buildCompanyRoute } from '@petcontrol/shared-constants';
import { cn } from '@petcontrol/ui/web';
import {
  Building2,
  Clock3,
  LoaderCircle,
  Settings2,
  ShieldCheck,
  UsersRound,
} from 'lucide-react';
import type { ReactNode } from 'react';
import { useEffect, useMemo, useState } from 'react';

import {
  useCompanyUserPermissionsQuery,
  useCompanyUsersQuery,
  useCurrentCompanyQuery,
  useCurrentCompanySystemConfigQuery,
  useCurrentUserQuery,
  useUpdateCompanyUserPermissionsMutation,
  useUpdateCurrentCompanyMutation,
  useUpdateCurrentCompanySystemConfigMutation,
} from '@/lib/api/domain.queries';
import { useToastStore } from '@/stores/toast.store';

type CompanyFormState = {
  name: string;
  fantasy_name: string;
  logo_url: string;
};

type BusinessFormState = {
  schedule_init_time: string;
  schedule_pause_init_time: string;
  schedule_pause_end_time: string;
  schedule_end_time: string;
  min_schedules_per_day: string;
  max_schedules_per_day: string;
  schedule_days: WeekDay[];
  dynamic_cages: boolean;
  total_small_cages: string;
  total_medium_cages: string;
  total_large_cages: string;
  total_giant_cages: string;
  whatsapp_notifications: boolean;
  whatsapp_conversation: boolean;
  whatsapp_business_phone: string;
};

export function SettingsPage() {
  const params = useParams({ strict: false });
  const companySlug =
    typeof params.companySlug === 'string' ? params.companySlug : '';

  const currentUserQuery = useCurrentUserQuery();
  const companyQuery = useCurrentCompanyQuery();
  const systemConfigQuery = useCurrentCompanySystemConfigQuery();
  const companyUsersQuery = useCompanyUsersQuery();

  const currentUser = currentUserQuery.data;
  const settingsAccess = currentUser?.settings_access;
  const canViewSettings =
    settingsAccess?.can_view ?? currentUser?.role === 'admin';
  const canManagePermissions =
    settingsAccess?.can_manage_permissions ?? currentUser?.role === 'admin';
  const editablePermissionCodes =
    settingsAccess?.editable_permission_codes ?? [];
  const canEditCompanySettings = currentUser?.role === 'admin';
  const canEditBusinessSettings =
    currentUser?.role === 'admin' ||
    editablePermissionCodes.includes('company_settings:edit');

  if (
    currentUserQuery.isLoading ||
    companyQuery.isLoading ||
    systemConfigQuery.isLoading ||
    companyUsersQuery.isLoading
  ) {
    return <SettingsPageLoading />;
  }

  if (!currentUser || !companyQuery.data || !systemConfigQuery.data) {
    return <SettingsPageLoading />;
  }

  if (!canViewSettings && companySlug) {
    return (
      <Navigate to={buildCompanyRoute(companySlug, 'dashboard')} replace />
    );
  }

  const isReadOnly =
    !canManagePermissions &&
    !canEditCompanySettings &&
    !canEditBusinessSettings;

  return (
    <div>
      <section className="overflow-hidden bg-white/75 shadow-[0_20px_50px_rgba(15,23,42,0.05)]">
        <div className="divide-y divide-stone-200">
          <section className="bg-[radial-gradient(circle_at_top_right,rgba(2,132,199,0.08),transparent_40%),radial-gradient(circle_at_bottom_left,rgba(16,185,129,0.05),transparent_35%)] px-6 py-7 md:px-7">
            <div className="flex flex-wrap items-start justify-between gap-5">
              <div className="max-w-3xl">
                <p className="text-xs font-semibold uppercase tracking-[0.32em] text-stone-400">
                  Configurações
                </p>
                <h2 className="mt-3 font-display text-3xl text-stone-950">
                  Central de ajustes
                </h2>
                <p className="mt-3 text-sm leading-7 text-stone-600">
                  Esta área reúne os dados institucionais da empresa, as regras
                  operacionais do negócio e, para perfis administradores, a
                  gestão das permissões dos usuários.
                </p>
              </div>

              <div className="rounded-3xl border border-white/80 bg-white/70 p-3 text-stone-700 shadow-sm">
                <Settings2 className="h-6 w-6" />
              </div>
            </div>

            <div className="mt-6 grid gap-3 md:grid-cols-3">
              <SettingsHeadlineCard
                title="Empresa"
                description={
                  companyQuery.data.fantasy_name || companyQuery.data.name
                }
                icon={Building2}
              />
              <SettingsHeadlineCard
                title="Negócios"
                description={`${systemConfigQuery.data.schedule_init_time} - ${systemConfigQuery.data.schedule_end_time}`}
                icon={Clock3}
              />
              <SettingsHeadlineCard
                title="Acesso"
                description={
                  canManagePermissions
                    ? 'Edição completa e gestão de permissões'
                    : isReadOnly
                      ? 'Modo somente leitura'
                      : `Edição parcial: ${editablePermissionCodes.length} permissões`
                }
                icon={ShieldCheck}
              />
            </div>
          </section>

          <CompanySettingsForm
            initialData={companyQuery.data}
            disabled={!canEditCompanySettings}
          />

          <BusinessSettingsForm
            initialData={systemConfigQuery.data}
            disabled={!canEditBusinessSettings}
          />

          {canManagePermissions && (
            <UserPermissionsManager
              companyUsers={companyUsersQuery.data ?? []}
            />
          )}
        </div>
      </section>
    </div>
  );
}

function CompanySettingsForm({
  initialData,
  disabled,
}: {
  initialData: {
    name: string;
    fantasy_name: string;
    logo_url?: string | null;
    cnpj: string;
    active_package: string;
  };
  disabled: boolean;
}) {
  const [form, setForm] = useState<CompanyFormState>({
    name: initialData.name,
    fantasy_name: initialData.fantasy_name,
    logo_url: initialData.logo_url ?? '',
  });

  const mutation = useUpdateCurrentCompanyMutation();
  const pushToast = useToastStore((state) => state.pushToast);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await mutation.mutateAsync({
      name: form.name.trim(),
      fantasy_name: form.fantasy_name.trim(),
      logo_url: form.logo_url.trim() || null,
    });
  };

  useEffect(() => {
    if (mutation.isSuccess) {
      pushToast('Configurações da empresa atualizadas.', 'success');
    }
  }, [mutation.isSuccess, pushToast]);

  useEffect(() => {
    if (mutation.isError) {
      pushToast(mutation.error.message, 'error');
    }
  }, [mutation.error, mutation.isError, pushToast]);

  return (
    <SettingsCard
      eyebrow="Empresa"
      title="Configurações da empresa"
      description="Dados institucionais da empresa usados no shell, na identificação interna e na comunicação visual."
    >
      <form className="space-y-6" onSubmit={handleSubmit}>
        <div className="grid gap-4 md:grid-cols-2">
          <Field
            label="Nome jurídico"
            value={form.name}
            onChange={(value) => setForm((c) => ({ ...c, name: value }))}
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Nome fantasia"
            value={form.fantasy_name}
            onChange={(value) =>
              setForm((c) => ({ ...c, fantasy_name: value }))
            }
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Logo URL"
            value={form.logo_url}
            onChange={(value) => setForm((c) => ({ ...c, logo_url: value }))}
            placeholder="https://cdn.exemplo.com/logo.png"
            disabled={disabled || mutation.isPending}
          />
          <ReadOnlyField
            label="CNPJ"
            value={initialData.cnpj}
            helpText="Para alterar, contate o suporte."
          />
          <ReadOnlyField
            label="Plano ativo"
            value={initialData.active_package}
            helpText="A configuração de plano será detalhada em um recorte próprio."
          />
        </div>

        <div className="flex flex-wrap items-center justify-end gap-3">
          <button
            type="submit"
            disabled={disabled || mutation.isPending}
            className="inline-flex items-center justify-center rounded-2xl bg-sky-100 px-5 py-3 text-sm font-bold text-sky-600 shadow-sm transition hover:bg-sky-200 disabled:cursor-not-allowed disabled:bg-stone-200 disabled:text-stone-400"
          >
            {mutation.isPending ? 'Salvando...' : 'Salvar empresa'}
          </button>
        </div>
      </form>
    </SettingsCard>
  );
}

function BusinessSettingsForm({
  initialData,
  disabled,
}: {
  initialData: {
    schedule_init_time: string;
    schedule_pause_init_time?: string | null;
    schedule_pause_end_time?: string | null;
    schedule_end_time: string;
    min_schedules_per_day: number;
    max_schedules_per_day: number;
    schedule_days: WeekDay[];
    dynamic_cages: boolean;
    total_small_cages: number;
    total_medium_cages: number;
    total_large_cages: number;
    total_giant_cages: number;
    whatsapp_notifications: boolean;
    whatsapp_conversation: boolean;
    whatsapp_business_phone?: string | null;
  };
  disabled: boolean;
}) {
  const [form, setForm] = useState<BusinessFormState>({
    schedule_init_time: initialData.schedule_init_time,
    schedule_pause_init_time: initialData.schedule_pause_init_time ?? '',
    schedule_pause_end_time: initialData.schedule_pause_end_time ?? '',
    schedule_end_time: initialData.schedule_end_time,
    min_schedules_per_day: String(initialData.min_schedules_per_day),
    max_schedules_per_day: String(initialData.max_schedules_per_day),
    schedule_days: initialData.schedule_days,
    dynamic_cages: initialData.dynamic_cages,
    total_small_cages: String(initialData.total_small_cages),
    total_medium_cages: String(initialData.total_medium_cages),
    total_large_cages: String(initialData.total_large_cages),
    total_giant_cages: String(initialData.total_giant_cages),
    whatsapp_notifications: initialData.whatsapp_notifications,
    whatsapp_conversation: initialData.whatsapp_conversation,
    whatsapp_business_phone: initialData.whatsapp_business_phone ?? '',
  });

  const mutation = useUpdateCurrentCompanySystemConfigMutation();
  const pushToast = useToastStore((state) => state.pushToast);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await mutation.mutateAsync({
      schedule_init_time: form.schedule_init_time,
      schedule_pause_init_time: form.schedule_pause_init_time || null,
      schedule_pause_end_time: form.schedule_pause_end_time || null,
      schedule_end_time: form.schedule_end_time,
      min_schedules_per_day: Number(form.min_schedules_per_day),
      max_schedules_per_day: Number(form.max_schedules_per_day),
      schedule_days: form.schedule_days,
      dynamic_cages: form.dynamic_cages,
      total_small_cages: Number(form.total_small_cages),
      total_medium_cages: Number(form.total_medium_cages),
      total_large_cages: Number(form.total_large_cages),
      total_giant_cages: Number(form.total_giant_cages),
      whatsapp_notifications: form.whatsapp_notifications,
      whatsapp_conversation: form.whatsapp_conversation,
      whatsapp_business_phone: form.whatsapp_business_phone || null,
    });
  };

  useEffect(() => {
    if (mutation.isSuccess) {
      pushToast('Configurações de negócios atualizadas.', 'success');
    }
  }, [mutation.isSuccess, pushToast]);

  useEffect(() => {
    if (mutation.isError) {
      pushToast(mutation.error.message, 'error');
    }
  }, [mutation.error, mutation.isError, pushToast]);

  return (
    <SettingsCard
      eyebrow="Negócios"
      title="Configurações de negócios"
      description="Regras operacionais, capacidade física e canais que dirigem a rotina diária da empresa."
    >
      <form className="space-y-6" onSubmit={handleSubmit}>
        <div className="grid gap-4 md:grid-cols-4">
          <Field
            label="Início do expediente"
            type="time"
            value={form.schedule_init_time}
            onChange={(v) => setForm((c) => ({ ...c, schedule_init_time: v }))}
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Início da pausa"
            type="time"
            value={form.schedule_pause_init_time}
            onChange={(v) =>
              setForm((c) => ({ ...c, schedule_pause_init_time: v }))
            }
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Fim da pausa"
            type="time"
            value={form.schedule_pause_end_time}
            onChange={(v) =>
              setForm((c) => ({ ...c, schedule_pause_end_time: v }))
            }
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Fim do expediente"
            type="time"
            value={form.schedule_end_time}
            onChange={(v) => setForm((c) => ({ ...c, schedule_end_time: v }))}
            disabled={disabled || mutation.isPending}
          />
        </div>

        <div>
          <p className="text-sm font-semibold text-stone-800">
            Dias de atendimento
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            {WEEK_DAYS.map((day) => {
              const active = form.schedule_days.includes(day);
              return (
                <button
                  key={day}
                  type="button"
                  onClick={() =>
                    setForm((c) => ({
                      ...c,
                      schedule_days: active
                        ? c.schedule_days.filter((i) => i !== day)
                        : [...c.schedule_days, day],
                    }))
                  }
                  disabled={disabled || mutation.isPending}
                  className={cn(
                    'rounded-full border px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] transition',
                    active
                      ? 'border-stone-950 bg-stone-950 text-white'
                      : 'border-stone-200 bg-white text-stone-500',
                  )}
                >
                  {weekDayLabel(day)}
                </button>
              );
            })}
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Field
            label="Mín. agendamentos/dia"
            type="number"
            value={form.min_schedules_per_day}
            onChange={(v) =>
              setForm((c) => ({ ...c, min_schedules_per_day: v }))
            }
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Máx. agendamentos/dia"
            type="number"
            value={form.max_schedules_per_day}
            onChange={(v) =>
              setForm((c) => ({ ...c, max_schedules_per_day: v }))
            }
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="WhatsApp Business"
            value={form.whatsapp_business_phone}
            onChange={(v) =>
              setForm((c) => ({ ...c, whatsapp_business_phone: v }))
            }
            placeholder="+5511999990001"
            disabled={disabled || mutation.isPending}
          />
          <ToggleField
            label="Gaiolas dinâmicas"
            checked={form.dynamic_cages}
            onChange={(v) => setForm((c) => ({ ...c, dynamic_cages: v }))}
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Gaiolas pequenas"
            type="number"
            value={form.total_small_cages}
            onChange={(v) => setForm((c) => ({ ...c, total_small_cages: v }))}
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Gaiolas médias"
            type="number"
            value={form.total_medium_cages}
            onChange={(v) => setForm((c) => ({ ...c, total_medium_cages: v }))}
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Gaiolas grandes"
            type="number"
            value={form.total_large_cages}
            onChange={(v) => setForm((c) => ({ ...c, total_large_cages: v }))}
            disabled={disabled || mutation.isPending}
          />
          <Field
            label="Gaiolas gigantes"
            type="number"
            value={form.total_giant_cages}
            onChange={(v) => setForm((c) => ({ ...c, total_giant_cages: v }))}
            disabled={disabled || mutation.isPending}
          />
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <ToggleField
            label="Notificações por WhatsApp"
            checked={form.whatsapp_notifications}
            onChange={(v) =>
              setForm((c) => ({ ...c, whatsapp_notifications: v }))
            }
            disabled={disabled || mutation.isPending}
          />
          <ToggleField
            label="Conversa por WhatsApp"
            checked={form.whatsapp_conversation}
            onChange={(v) =>
              setForm((c) => ({ ...c, whatsapp_conversation: v }))
            }
            disabled={disabled || mutation.isPending}
          />
        </div>

        <div className="flex flex-wrap items-center justify-end gap-3">
          <button
            type="submit"
            disabled={disabled || mutation.isPending}
            className="inline-flex items-center justify-center rounded-2xl bg-sky-100 px-5 py-3 text-sm font-bold text-sky-600 shadow-sm transition hover:bg-sky-200 disabled:cursor-not-allowed disabled:bg-stone-200 disabled:text-stone-400"
          >
            {mutation.isPending ? 'Salvando...' : 'Salvar negócio'}
          </button>
        </div>
      </form>
    </SettingsCard>
  );
}

function UserPermissionsManager({
  companyUsers,
}: {
  companyUsers: {
    user_id: string;
    full_name?: string | null;
    short_name?: string | null;
    role: string;
    kind: string;
  }[];
}) {
  const filteredUsers = useMemo(
    () => companyUsers.filter((u) => u.role === 'admin' || u.role === 'system'),
    [companyUsers],
  );

  const [selectedUserId, setSelectedUserId] = useState(() => {
    const preferred = filteredUsers.find((u) => u.role === 'system');
    return preferred?.user_id ?? (filteredUsers[0]?.user_id || '');
  });

  const selectedCompanyUser = useMemo(
    () => filteredUsers.find((u) => u.user_id === selectedUserId),
    [filteredUsers, selectedUserId],
  );

  const permissionsQuery = useCompanyUserPermissionsQuery(selectedUserId);

  return (
    <SettingsCard
      eyebrow="Permissões"
      title="Permissões por usuário"
      description="Selecione um usuário e ajuste as permissões agrupadas por módulos disponíveis no plano atual da empresa."
    >
      <div className="space-y-6">
        <div className="grid gap-4 lg:grid-cols-[minmax(0,280px)_1fr]">
          <div className="space-y-3">
            <label
              htmlFor="tenant-user-select"
              className="text-sm font-semibold text-stone-800"
            >
              Usuários
            </label>
            <select
              id="tenant-user-select"
              value={selectedUserId}
              onChange={(e) => setSelectedUserId(e.target.value)}
              className="w-full rounded-2xl border border-stone-200 bg-white px-4 py-3 text-sm text-stone-700 outline-none transition focus:border-stone-400"
            >
              {filteredUsers.map((u) => (
                <option key={u.user_id} value={u.user_id}>
                  {u.full_name || u.short_name || u.user_id}
                </option>
              ))}
            </select>
            {selectedCompanyUser ? (
              <div className="rounded-2xl border border-stone-200 bg-stone-50 p-4 text-sm text-stone-600">
                <p className="font-semibold text-stone-800">
                  {selectedCompanyUser.full_name ||
                    selectedCompanyUser.short_name ||
                    selectedCompanyUser.user_id}
                </p>
                <p className="mt-1 uppercase tracking-[0.18em] text-stone-400">
                  {selectedCompanyUser.role} · {selectedCompanyUser.kind}
                </p>
              </div>
            ) : null}
          </div>

          <div className="rounded-[1.5rem] border border-stone-200 bg-stone-50 p-5">
            {permissionsQuery.isLoading ? (
              <LoadingInline message="Carregando permissões do usuário..." />
            ) : permissionsQuery.data ? (
              <UserPermissionsForm
                key={selectedUserId}
                userId={selectedUserId}
                initialPermissions={permissionsQuery.data}
              />
            ) : (
              <p className="text-sm text-stone-500">
                Selecione um usuário para visualizar as permissões gerenciáveis.
              </p>
            )}
          </div>
        </div>
      </div>
    </SettingsCard>
  );
}

function UserPermissionsForm({
  userId,
  initialPermissions,
}: {
  userId: string;
  initialPermissions: CompanyUserPermissionsDTO;
}) {
  const [selectedPermissionCodes, setSelectedPermissionCodes] = useState<
    string[]
  >(() => getActivePermissionCodes(initialPermissions));

  const mutation = useUpdateCompanyUserPermissionsMutation(userId);
  const pushToast = useToastStore((state) => state.pushToast);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await mutation.mutateAsync({
      permission_codes: selectedPermissionCodes,
    });
  };

  useEffect(() => {
    if (mutation.isSuccess) {
      pushToast('Permissões do usuário atualizadas.', 'success');
    }
  }, [mutation.isSuccess, pushToast]);

  useEffect(() => {
    if (mutation.isError) {
      pushToast(mutation.error.message, 'error');
    }
  }, [mutation.error, mutation.isError, pushToast]);

  return (
    <form className="space-y-6" onSubmit={handleSubmit}>
      <PermissionsChecklist
        snapshot={initialPermissions}
        selectedPermissionCodes={selectedPermissionCodes}
        onToggle={(code) =>
          setSelectedPermissionCodes((prev) =>
            prev.includes(code)
              ? prev.filter((c) => c !== code)
              : [...prev, code],
          )
        }
        disabled={mutation.isPending}
      />

      <div className="flex flex-wrap items-center justify-end gap-3">
        <button
          type="submit"
          disabled={mutation.isPending}
          className="inline-flex items-center justify-center rounded-2xl bg-sky-100 px-5 py-3 text-sm font-bold text-sky-600 shadow-sm transition hover:bg-sky-200 disabled:cursor-not-allowed disabled:bg-stone-200 disabled:text-stone-400"
        >
          {mutation.isPending ? 'Salvando...' : 'Salvar permissões'}
        </button>
      </div>
    </form>
  );
}

function SettingsPageLoading() {
  return (
    <section className="rounded-[2rem] bg-white/75 p-6 shadow-[0_20px_50px_rgba(15,23,42,0.05)]">
      <LoadingInline message="Carregando central de configurações..." />
    </section>
  );
}

function SettingsHeadlineCard({
  title,
  description,
  icon: Icon,
}: {
  title: string;
  description: string;
  icon: typeof Building2;
}) {
  return (
    <div className="rounded-[1.5rem] border border-white/80 bg-white/70 p-4 shadow-sm">
      <div className="flex items-center gap-3">
        <div className="rounded-2xl bg-stone-100 p-2 text-stone-700">
          <Icon className="h-4 w-4" />
        </div>
        <div>
          <p className="text-sm font-semibold text-stone-900">{title}</p>
          <p className="text-sm text-stone-500">{description}</p>
        </div>
      </div>
    </div>
  );
}

function SettingsCard({
  eyebrow,
  title,
  description,
  children,
}: {
  eyebrow: string;
  title: string;
  description: string;
  children: ReactNode;
}) {
  return (
    <section className="px-6 py-7 md:px-7">
      <div className="mb-6">
        <p className="text-xs font-semibold uppercase tracking-[0.3em] text-stone-400">
          {eyebrow}
        </p>
        <h3 className="mt-3 font-display text-2xl text-stone-950">{title}</h3>
        <p className="mt-2 max-w-3xl text-sm leading-7 text-stone-500">
          {description}
        </p>
      </div>

      {children}
    </section>
  );
}

function Field({
  label,
  value,
  onChange,
  disabled,
  placeholder,
  type = 'text',
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  placeholder?: string;
  type?: React.HTMLInputTypeAttribute;
}) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-semibold text-stone-800">
        {label}
      </span>
      <input
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        disabled={disabled}
        placeholder={placeholder}
        className="w-full rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm text-stone-700 outline-none transition focus:border-stone-400 disabled:cursor-not-allowed disabled:bg-stone-100 disabled:text-stone-400"
      />
    </label>
  );
}

function ReadOnlyField({
  label,
  value,
  helpText,
}: {
  label: string;
  value: string;
  helpText?: string;
}) {
  return (
    <div>
      <span className="mb-2 block text-sm font-semibold text-stone-800">
        {label}
      </span>
      <div className="rounded-2xl border border-stone-200 bg-stone-100 px-4 py-3 text-sm text-stone-600">
        {value}
      </div>
      {helpText ? (
        <p className="mt-2 text-xs text-stone-400">{helpText}</p>
      ) : null}
    </div>
  );
}

function ToggleField({
  label,
  checked,
  onChange,
  disabled,
}: {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <label className="flex min-h-[50px] items-center justify-between rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3">
      <span className="text-sm font-semibold text-stone-800">{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked ? 'true' : 'false'}
        onClick={() => onChange(!checked)}
        disabled={disabled}
        className={cn(
          'relative h-7 w-12 rounded-full transition',
          checked ? 'bg-emerald-500' : 'bg-stone-300',
          disabled ? 'cursor-not-allowed opacity-60' : '',
        )}
      >
        <span
          className={cn(
            'absolute top-1 h-5 w-5 rounded-full bg-white transition',
            checked ? 'left-6' : 'left-1',
          )}
        />
      </button>
    </label>
  );
}

function LoadingInline({ message }: { message: string }) {
  return (
    <div className="flex items-center gap-3 text-sm text-stone-500">
      <LoaderCircle className="h-4 w-4 animate-spin" />
      <span>{message}</span>
    </div>
  );
}

function PermissionsChecklist({
  snapshot,
  selectedPermissionCodes,
  onToggle,
  disabled,
}: {
  snapshot: CompanyUserPermissionsDTO;
  selectedPermissionCodes: string[];
  onToggle: (permissionCode: string) => void;
  disabled?: boolean;
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-stone-700">
        <UsersRound className="h-4 w-4" />
        <p className="text-sm font-semibold">
          Permissões gerenciáveis por módulo
        </p>
      </div>

      <div className="space-y-3">
        {getPermissionGroups(snapshot).map((group) => (
          <section
            key={group.module_code}
            className="overflow-hidden rounded-[1.75rem] border border-stone-200 bg-white"
          >
            <div className="border-b border-stone-100 bg-stone-50/80 px-5 py-4">
              <div className="flex flex-wrap items-center gap-2">
                <p className="text-sm font-semibold text-stone-900">
                  {group.module_name}
                </p>
                <span className="rounded-full bg-sky-100 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-sky-700">
                  {group.module_code}
                </span>
                <span className="rounded-full bg-stone-100 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-stone-500">
                  pacote {group.min_package}
                </span>
              </div>
              <p className="mt-2 text-sm text-stone-500">
                {group.module_description}
              </p>
            </div>

            <div className="space-y-0.5 px-4 py-3">
              {group.permissions.map((permission) => (
                <PermissionChecklistItem
                  key={permission.id}
                  permission={permission}
                  checked={selectedPermissionCodes.includes(permission.code)}
                  onToggle={onToggle}
                  disabled={disabled}
                />
              ))}
            </div>
          </section>
        ))}
      </div>
    </div>
  );
}

function PermissionChecklistItem({
  permission,
  checked,
  onToggle,
  disabled,
}: {
  permission: CompanyUserPermissionDTO;
  checked: boolean;
  onToggle: (permissionCode: string) => void;
  disabled?: boolean;
}) {
  return (
    <label className="flex items-start gap-3 rounded-2xl px-3 py-1 transition hover:bg-stone-50/80">
      <input
        type="checkbox"
        checked={checked}
        onChange={() => onToggle(permission.code)}
        disabled={disabled}
        className="mt-1 h-4 w-4 rounded border-stone-300 text-stone-950"
      />
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-sm font-semibold text-stone-900">
            {formatPermissionLabel(permission.code)}
          </p>
          {permission.is_default_for_role ? (
            <span className="rounded-full bg-stone-100 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-stone-500">
              padrão
            </span>
          ) : (
            <span className="rounded-full bg-amber-100 px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.16em] text-amber-700">
              customizado
            </span>
          )}
        </div>
      </div>
    </label>
  );
}

function getActivePermissionCodes(snapshot: CompanyUserPermissionsDTO) {
  const source = getPermissionGroups(snapshot).flatMap(
    (group) => group.permissions,
  );

  return source.filter((permission) => permission.is_active).map((p) => p.code);
}

function getPermissionGroups(snapshot: CompanyUserPermissionsDTO) {
  if (snapshot.permission_groups.length > 0) {
    return snapshot.permission_groups;
  }

  return [
    {
      module_code: 'LEGACY',
      module_name: 'Permissões',
      module_description: 'Permissões carregadas em formato legado.',
      min_package: snapshot.active_package,
      permissions: snapshot.permissions,
    },
  ];
}

function formatPermissionLabel(code: string) {
  const permissionLabelMap: Record<string, string> = {
    'company_settings:edit': 'Configurações de negócios',
    'plan_settings:edit': 'Operação e regras de agenda',
    'payment_settings:edit': 'Pagamentos e cobrança',
    'notification_settings:edit': 'Notificações automáticas',
    'integration_settings:edit': 'Integrações externas',
    'security_settings:edit': 'Segurança e acesso',
  };

  if (permissionLabelMap[code]) {
    return permissionLabelMap[code];
  }

  const [resource, action] = code.split(':');
  if (!resource || !action) {
    return code;
  }

  const resourceLabel = resource
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');

  const actionMap: Record<string, string> = {
    create: 'Criar',
    view: 'Visualizar',
    update: 'Atualizar',
    delete: 'Deletar',
    restore: 'Restaurar',
    deactivate: 'Desativar',
    reactivate: 'Reativar',
    edit: 'Editar',
    block: 'Bloquear',
    unblock: 'Desbloquear',
  };

  return `${actionMap[action] ?? action} ${resourceLabel}`;
}

function weekDayLabel(day: WeekDay) {
  const labels: Record<WeekDay, string> = {
    sunday: 'Dom',
    monday: 'Seg',
    tuesday: 'Ter',
    wednesday: 'Qua',
    thursday: 'Qui',
    friday: 'Sex',
    saturday: 'Sab',
  };

  return labels[day];
}
