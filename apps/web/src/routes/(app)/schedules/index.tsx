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
    <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
      <section className="space-y-4 rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        <div>
          <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
            Schedules
          </p>
          <h2 className="mt-2 font-display text-3xl text-white">
            Agendamentos do tenant
          </h2>
          <p className="mt-2 text-sm text-slate-300">
            Lista real conectada em GET /api/v1/schedules, com isolamento por
            company_id no backend.
          </p>
        </div>

        <div className="mt-4">
          <SearchBar
            value={search}
            onChange={setSearch}
            placeholder="Buscar por cliente, pet ou serviço..."
            id="schedules-search"
          />
        </div>

        <div className="mt-4 overflow-hidden rounded-3xl border border-white/10">
          <table className="w-full border-collapse text-left text-sm">
            <thead className="bg-white/5 text-slate-300">
              <tr>
                <th className="px-4 py-3 font-medium">Data</th>
                <th className="px-4 py-3 font-medium">Client</th>
                <th className="px-4 py-3 font-medium">Pet</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Ações</th>
              </tr>
            </thead>
            <tbody>
              {viewState === 'loading' ? (
                <tr>
                  <td
                    colSpan={5}
                    className="px-4 py-6 text-center text-slate-300"
                  >
                    Carregando schedules...
                  </td>
                </tr>
              ) : null}

              {viewState === 'error' ? (
                <tr>
                  <td
                    colSpan={5}
                    className="px-4 py-6 text-center text-rose-200"
                  >
                    Falha ao buscar schedules.
                  </td>
                </tr>
              ) : null}

              {viewState === 'empty' ? (
                <tr>
                  <td
                    colSpan={5}
                    className="px-4 py-6 text-center text-slate-300"
                  >
                    Nenhum agendamento encontrado para este tenant.
                  </td>
                </tr>
              ) : null}

              {viewState === 'ready'
                ? schedules.map((schedule) => (
                    <tr key={schedule.id} className="border-t border-white/10">
                      <td className="px-4 py-3 text-slate-200">
                        {new Date(schedule.scheduled_at).toLocaleString(
                          'pt-BR',
                          {
                            dateStyle: 'short',
                            timeStyle: 'short',
                          },
                        )}
                      </td>
                      <td className="px-4 py-3 text-slate-300">
                        {schedule.client_name || schedule.client_id}
                      </td>
                      <td className="px-4 py-3 text-slate-300">
                        <div>{schedule.pet_name || schedule.pet_id}</div>
                        {schedule.service_titles?.length ? (
                          <div className="mt-1 text-xs text-slate-400">
                            {schedule.service_titles.join(', ')}
                          </div>
                        ) : null}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            'rounded-full border px-2 py-1 text-xs',
                            scheduleStatusColorClass(schedule.current_status),
                          )}
                        >
                          {formatScheduleStatus(schedule.current_status)}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex gap-2">
                          <button
                            type="button"
                            onClick={() => startEdit(schedule.id)}
                            className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-slate-200 transition hover:bg-white/10"
                          >
                            Editar
                          </button>
                          <button
                            type="button"
                            onClick={() => void removeSchedule(schedule.id)}
                            className="rounded-xl border border-rose-400/40 bg-rose-500/10 px-3 py-1 text-xs text-rose-200 transition hover:bg-rose-500/20"
                          >
                            Excluir
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                : null}
            </tbody>
          </table>
        </div>

        <PaginationBar
          meta={schedulesQuery.data?.meta}
          onPageChange={goToPage}
        />
      </section>

      <section className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
          {editingScheduleId ? 'Editar schedule' : 'Novo schedule'}
        </p>
        <h3 className="mt-2 font-display text-2xl text-white">
          {editingScheduleId ? 'Atualizar agendamento' : 'Criar agendamento'}
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
              className="rounded-2xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition hover:bg-white/10"
            >
              Limpar
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}

const fieldClassName =
  'w-full rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white outline-none transition placeholder:text-slate-500 focus:border-primary/50 focus:ring-2 focus:ring-primary/20';

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
      <span className="text-sm font-medium text-slate-200">{label}</span>
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
