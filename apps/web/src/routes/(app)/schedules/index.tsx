import { zodResolver } from '@hookform/resolvers/zod';
import {
  formatScheduleStatus,
  resolveAsyncViewState,
  scheduleStatusColorClass,
  cn,
} from '@petcontrol/ui/web';
import type { ScheduleStatus } from '@petcontrol/shared-types';
import { useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';
import { type ReactNode, useMemo, useState } from 'react';

import {
  useClientsQuery,
  useCreateScheduleMutation,
  usePetsQuery,
  useDeleteScheduleMutation,
  useSchedulesQuery,
  useServicesQuery,
  useUpdateScheduleMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';
import { useListParams } from '@/hooks/use-list-params';
import { SearchBar } from '@/ui/search-bar';
import { PaginationBar } from '@/ui/pagination-bar';

const scheduleSchema = z
  .object({
    clientId: z.string().min(1, 'Selecione um cliente'),
    petId: z.string().min(1, 'Selecione um pet'),
    serviceIds: z
      .array(z.string().uuid('Selecione serviços válidos'))
      .default([]),
    scheduledAt: z.string().min(1, 'Informe data/hora do agendamento'),
    estimatedEnd: z.string().optional(),
    notes: z.string().max(500, 'Máximo de 500 caracteres').optional(),
    status: z.enum([
      'waiting',
      'confirmed',
      'canceled',
      'in_progress',
      'finished',
      'delivered',
    ]),
  })
  .refine(
    (values) => {
      if (!values.estimatedEnd) {
        return true;
      }
      return new Date(values.estimatedEnd) > new Date(values.scheduledAt);
    },
    {
      path: ['estimatedEnd'],
      message: 'estimated_end deve ser maior que scheduled_at',
    },
  );

type ScheduleFormInput = z.input<typeof scheduleSchema>;
type ScheduleFormValues = z.output<typeof scheduleSchema>;

const scheduleStatusOptions: Array<{ value: ScheduleStatus; label: string }> = [
  { value: 'waiting', label: 'Aguardando' },
  { value: 'confirmed', label: 'Confirmado' },
  { value: 'in_progress', label: 'Em andamento' },
  { value: 'finished', label: 'Finalizado' },
  { value: 'delivered', label: 'Entregue' },
  { value: 'canceled', label: 'Cancelado' },
];

export function SchedulesPage() {
  const { params, search, setSearch, goToPage } = useListParams();
  const clientsQuery = useClientsQuery();
  const petsQuery = usePetsQuery();
  const servicesQuery = useServicesQuery();
  const schedulesQuery = useSchedulesQuery(params);
  const createMutation = useCreateScheduleMutation();
  const updateMutation = useUpdateScheduleMutation();
  const deleteMutation = useDeleteScheduleMutation();
  const [editingScheduleId, setEditingScheduleId] = useState<string | null>(
    null,
  );

  const form = useForm<ScheduleFormInput, unknown, ScheduleFormValues>({
    resolver: zodResolver(scheduleSchema),
    defaultValues: {
      clientId: '',
      petId: '',
      serviceIds: [],
      scheduledAt: '',
      estimatedEnd: '',
      notes: '',
      status: 'waiting',
    },
  });

  const schedules = useMemo(
    () =>
      [...(schedulesQuery.data?.data ?? [])].sort((a, b) =>
        a.scheduled_at.localeCompare(b.scheduled_at),
      ),
    [schedulesQuery.data],
  );

  const selectedClientId = useWatch({
    control: form.control,
    name: 'clientId',
  });
  const availablePets = useMemo(
    () =>
      (petsQuery.data?.data ?? []).filter(
        (pet) => !selectedClientId || pet.owner_id === selectedClientId,
      ),
    [petsQuery.data, selectedClientId],
  );

  const viewState = resolveAsyncViewState({
    isLoading: schedulesQuery.isLoading,
    isError: schedulesQuery.isError,
    itemCount: schedules.length,
  });

  const pendingMutation =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  function resetForm() {
    setEditingScheduleId(null);
    form.reset({
      clientId: '',
      petId: '',
      serviceIds: [],
      scheduledAt: '',
      estimatedEnd: '',
      notes: '',
      status: 'waiting',
    });
  }

  async function onSubmit(values: ScheduleFormValues) {
    const payload = {
      client_id: values.clientId,
      pet_id: values.petId,
      service_ids: values.serviceIds,
      scheduled_at: new Date(values.scheduledAt).toISOString(),
      estimated_end: values.estimatedEnd
        ? new Date(values.estimatedEnd).toISOString()
        : undefined,
      notes: values.notes?.trim() || undefined,
      status: values.status,
    };

    if (!editingScheduleId) {
      await createMutation.mutateAsync(payload);
      resetForm();
      return;
    }

    await updateMutation.mutateAsync({
      scheduleId: editingScheduleId,
      input: payload,
    });
    resetForm();
  }

  function startEdit(scheduleId: string) {
    const schedule = schedules.find((item) => item.id === scheduleId);
    if (!schedule) {
      return;
    }

    setEditingScheduleId(schedule.id);
    form.reset({
      clientId: schedule.client_id,
      petId: schedule.pet_id,
      serviceIds: schedule.service_ids ?? [],
      scheduledAt: toLocalDateTimeInput(schedule.scheduled_at),
      estimatedEnd: schedule.estimated_end
        ? toLocalDateTimeInput(schedule.estimated_end)
        : '',
      notes: schedule.notes ?? '',
      status: schedule.current_status,
    });
  }

  async function removeSchedule(scheduleId: string) {
    await deleteMutation.mutateAsync(scheduleId);
    if (editingScheduleId === scheduleId) {
      resetForm();
    }
  }

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;

  return (
    <main className="flex min-w-0 flex-col min-h-full">
      <div className="flex-1 grid grid-cols-1 divide-y divide-border/50 xl:grid-cols-[minmax(0,1.1fr)_26rem] xl:divide-x xl:divide-y-0">
        <section className="flex flex-col min-h-full">
          <header className="bg-[radial-gradient(circle_at_top_right,rgba(2,132,199,0.08),transparent_40%),radial-gradient(circle_at_bottom_left,rgba(16,185,129,0.05),transparent_35%)] px-6 py-8 lg:px-10">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <p className="app-eyebrow">Operação</p>
                <h1 className="mt-3 font-display text-4xl text-foreground sm:text-5xl">
                  Agendamentos
                </h1>
                <p className="mt-4 max-w-2xl text-sm leading-6 text-muted">
                  Gestão completa da agenda do tenant. Monitore o status dos
                  serviços, tempos de execução e entregas programadas.
                </p>
              </div>
              <button
                type="button"
                onClick={resetForm}
                className="inline-flex h-12 items-center justify-center rounded-2xl bg-primary px-6 text-sm font-bold text-slate-950 transition hover:brightness-110 shadow-sm"
              >
                Novo agendamento
              </button>
            </div>
          </header>

          <div className="p-6 lg:p-10">
            <div className="mt-4">
              <SearchBar
                value={search}
                onChange={setSearch}
                placeholder="Buscar por cliente, pet ou serviço..."
                id="schedules-search"
              />
            </div>

            <div className="mt-6 space-y-3">
              {viewState === 'loading' ? (
                <div className="rounded-2xl border border-border/50 bg-surface/30 p-8 text-center text-muted">
                  Carregando agendamentos...
                </div>
              ) : null}

              {viewState === 'error' ? (
                <div className="rounded-2xl border border-rose-400/30 bg-rose-500/10 p-8 text-center text-rose-200">
                  Falha ao buscar agendamentos.
                </div>
              ) : null}

              {viewState === 'empty' ? (
                <div className="rounded-2xl border border-border/50 bg-surface/30 p-8 text-center text-muted">
                  Nenhum agendamento encontrado para este tenant.
                </div>
              ) : null}

              {viewState === 'ready'
                ? schedules.map((schedule) => (
                    <article
                      key={schedule.id}
                      className={`group flex items-center justify-between gap-4 rounded-[1.8rem] border p-5 transition ${editingScheduleId === schedule.id ? 'border-primary/40 bg-primary/10' : 'border-border/50 bg-surface/30 hover:border-border hover:bg-surface/60'}`}
                    >
                      <div className="flex w-full items-center justify-between gap-4">
                        <div className="flex items-center gap-5">
                          <div className="flex flex-col items-center justify-center rounded-2xl bg-surface border border-border/50 px-3 py-2 text-center shadow-sm">
                            <span className="text-[10px] uppercase font-bold text-muted">
                              {new Date(
                                schedule.scheduled_at,
                              ).toLocaleDateString('pt-BR', { month: 'short' })}
                            </span>
                            <span className="text-xl font-display font-bold text-primary">
                              {new Date(schedule.scheduled_at).getDate()}
                            </span>
                          </div>
                          <div className="min-w-0">
                            <div className="flex items-center gap-2">
                              <p className="font-medium text-foreground group-hover:text-primary transition">
                                {schedule.pet_name || 'Pet N/I'}
                              </p>
                              <span
                                className={cn(
                                  'rounded-full border px-2 py-0.5 text-[10px] font-bold uppercase',
                                  scheduleStatusColorClass(
                                    schedule.current_status,
                                  ),
                                )}
                              >
                                {formatScheduleStatus(schedule.current_status)}
                              </span>
                            </div>
                            <p className="mt-0.5 text-sm text-muted">
                              Tutor: {schedule.client_name || 'N/I'} ·{' '}
                              {new Date(
                                schedule.scheduled_at,
                              ).toLocaleTimeString('pt-BR', {
                                hour: '2-digit',
                                minute: '2-digit',
                              })}
                            </p>
                            {schedule.service_titles?.length ? (
                              <p className="mt-1 text-[11px] text-muted/70 truncate max-w-xs">
                                {schedule.service_titles.join(', ')}
                              </p>
                            ) : null}
                          </div>
                        </div>

                        <div className="flex gap-2 shrink-0">
                          <button
                            type="button"
                            onClick={() => startEdit(schedule.id)}
                            className="rounded-xl border border-border/50 bg-surface/50 px-3 py-1.5 text-xs text-foreground transition hover:bg-surface"
                          >
                            Editar
                          </button>
                          <button
                            type="button"
                            onClick={() => void removeSchedule(schedule.id)}
                            className="rounded-xl border border-rose-400/40 bg-rose-500/10 px-3 py-1.5 text-xs text-rose-200 transition hover:bg-rose-500/20"
                          >
                            Excluir
                          </button>
                        </div>
                      </div>
                    </article>
                  ))
                : null}
            </div>

            <PaginationBar
              meta={schedulesQuery.data?.meta}
              onPageChange={goToPage}
            />
          </div>
        </section>

        <aside className="bg-surface/30 p-6 lg:p-10">
          <div className="xl:sticky xl:top-10">
            <p className="app-eyebrow">
              {editingScheduleId ? 'Editar agendamento' : 'Novo agendamento'}
            </p>
            <h3 className="mt-4 font-display text-3xl text-foreground">
              {editingScheduleId
                ? 'Atualizar agendamento'
                : 'Criar agendamento'}
            </h3>

            <form
              className="mt-6 space-y-4"
              onSubmit={form.handleSubmit((values) => {
                void onSubmit(values);
              })}
            >
              <FormField
                label="Cliente"
                htmlFor="schedule-client"
                error={form.formState.errors.clientId?.message}
              >
                <select
                  id="schedule-client"
                  title="Selecione o cliente"
                  {...form.register('clientId')}
                  className={fieldClassName}
                >
                  <option value="">Selecione um cliente</option>
                  {(clientsQuery.data?.data ?? []).map((client) => (
                    <option key={client.id} value={client.id}>
                      {client.full_name}
                    </option>
                  ))}
                </select>
              </FormField>

              <FormField
                label="Pet"
                htmlFor="schedule-pet"
                error={form.formState.errors.petId?.message}
              >
                <select
                  id="schedule-pet"
                  title="Selecione o pet"
                  {...form.register('petId')}
                  className={fieldClassName}
                >
                  <option value="">Selecione um pet</option>
                  {availablePets.map((pet) => (
                    <option key={pet.id} value={pet.id}>
                      {pet.name}
                    </option>
                  ))}
                </select>
              </FormField>

              <FormField
                label="Serviços"
                htmlFor="schedule-services"
                error={form.formState.errors.serviceIds?.message}
              >
                <select
                  id="schedule-services"
                  title="Selecione os serviços (segure Ctrl para múltiplos)"
                  {...form.register('serviceIds')}
                  multiple
                  className={`${fieldClassName} min-h-32`}
                >
                  {(servicesQuery.data?.data ?? []).map((service) => (
                    <option key={service.id} value={service.id}>
                      {service.title}
                    </option>
                  ))}
                </select>
              </FormField>

              <FormField
                label="Data/Hora"
                htmlFor="schedule-at"
                error={form.formState.errors.scheduledAt?.message}
              >
                <input
                  id="schedule-at"
                  title="Data e hora do agendamento"
                  {...form.register('scheduledAt')}
                  type="datetime-local"
                  className={fieldClassName}
                />
              </FormField>

              <FormField
                label="Fim estimado"
                htmlFor="schedule-end"
                error={form.formState.errors.estimatedEnd?.message}
              >
                <input
                  id="schedule-end"
                  title="Previsão de término"
                  {...form.register('estimatedEnd')}
                  type="datetime-local"
                  className={fieldClassName}
                />
              </FormField>

              <FormField
                label="Status"
                htmlFor="schedule-status"
                error={form.formState.errors.status?.message}
              >
                <select
                  id="schedule-status"
                  title="Status do agendamento"
                  {...form.register('status')}
                  className={fieldClassName}
                >
                  {scheduleStatusOptions.map((item) => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </select>
              </FormField>

              <FormField
                label="Observações"
                htmlFor="schedule-notes"
                error={form.formState.errors.notes?.message}
              >
                <textarea
                  id="schedule-notes"
                  title="Observações do agendamento"
                  {...form.register('notes')}
                  className={fieldClassName}
                  rows={3}
                  placeholder="Ex: Trazer toalha própria, alérgico a tal produto..."
                />
              </FormField>

              {mutationError instanceof ApiError ? (
                <div className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
                  {mutationError.message}
                </div>
              ) : null}

              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={pendingMutation}
                  className="rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-70"
                >
                  {pendingMutation
                    ? 'Salvando...'
                    : editingScheduleId
                      ? 'Atualizar'
                      : 'Criar'}
                </button>
                <button
                  type="button"
                  onClick={resetForm}
                  className="rounded-2xl border border-border/50 bg-surface/50 px-4 py-2 text-sm text-foreground transition hover:bg-surface"
                >
                  Limpar
                </button>
              </div>
            </form>
          </div>
        </aside>
      </div>
    </main>
  );
}

const fieldClassName =
  'w-full rounded-2xl border border-border/50 bg-surface/50 px-3 py-2 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-primary/50 focus:ring-2 focus:ring-primary/20';

function FormField({
  label,
  error,
  children,
  htmlFor,
}: {
  label: string;
  error?: string;
  children: ReactNode;
  htmlFor?: string;
}) {
  return (
    <label className="block space-y-2" htmlFor={htmlFor}>
      <span className="text-sm font-medium text-foreground">{label}</span>
      {children}
      {error ? <span className="text-sm text-rose-300">{error}</span> : null}
    </label>
  );
}

function toLocalDateTimeInput(value: string) {
  const date = new Date(value);
  const timezoneOffset = date.getTimezoneOffset() * 60 * 1000;
  const localDate = new Date(date.getTime() - timezoneOffset);
  return localDate.toISOString().slice(0, 16);
}
