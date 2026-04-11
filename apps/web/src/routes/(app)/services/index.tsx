import { type FormEvent, useState } from 'react';
import type { CreateServiceInput, ServiceDTO } from '@petcontrol/shared-types';

import {
  useCreateServiceMutation,
  useDeleteServiceMutation,
  useServicesQuery,
  useUpdateServiceMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';

type ServiceFormState = CreateServiceInput;

const initialServiceForm: ServiceFormState = {
  type_name: 'Banho',
  title: '',
  description: '',
  notes: '',
  price: '',
  discount_rate: '0.00',
  image_url: '',
  is_active: true,
};

export function ServicesPage() {
  const servicesQuery = useServicesQuery();
  const createMutation = useCreateServiceMutation();
  const updateMutation = useUpdateServiceMutation();
  const deleteMutation = useDeleteServiceMutation();
  const [editingServiceId, setEditingServiceId] = useState<string | null>(null);
  const [form, setForm] = useState<ServiceFormState>(initialServiceForm);

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;
  const isPending =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  function resetForm() {
    setEditingServiceId(null);
    setForm(initialServiceForm);
  }

  function startEdit(service: ServiceDTO) {
    setEditingServiceId(service.id);
    setForm({
      type_name: service.type_name,
      title: service.title,
      description: service.description,
      notes: service.notes ?? '',
      price: service.price,
      discount_rate: service.discount_rate,
      image_url: service.image_url ?? '',
      is_active: service.is_active,
    });
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const payload = {
      ...form,
      notes: form.notes || undefined,
      image_url: form.image_url || undefined,
    };

    if (editingServiceId) {
      await updateMutation.mutateAsync({
        serviceId: editingServiceId,
        input: payload,
      });
      resetForm();
      return;
    }

    await createMutation.mutateAsync(payload);
    resetForm();
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
      <section className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
          Catálogo
        </p>
        <h2 className="mt-2 font-display text-3xl text-white">Serviços</h2>
        <p className="mt-2 text-sm text-slate-300">
          Catálogo ativo por tenant, usado diretamente pelos agendamentos.
        </p>

        <div className="mt-6 space-y-3">
          {servicesQuery.isLoading ? (
            <StateMessage message="Carregando serviços..." />
          ) : null}
          {servicesQuery.isError ? (
            <StateMessage message="Falha ao carregar serviços." tone="error" />
          ) : null}
          {!servicesQuery.isLoading &&
          (servicesQuery.data ?? []).length === 0 ? (
            <StateMessage message="Nenhum serviço cadastrado." />
          ) : null}
          {(servicesQuery.data ?? []).map((service) => (
            <article
              key={service.id}
              className="rounded-3xl border border-white/10 bg-white/5 p-4"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h3 className="font-display text-xl text-white">
                    {service.title}
                  </h3>
                  <p className="mt-1 text-sm text-slate-300">
                    {service.type_name} · R$ {service.price}
                  </p>
                  <p className="mt-1 text-xs text-slate-500">
                    {service.description}
                  </p>
                </div>
                <Actions
                  onEdit={() => startEdit(service)}
                  onDelete={() => void deleteMutation.mutateAsync(service.id)}
                />
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
          {editingServiceId ? 'Editar serviço' : 'Novo serviço'}
        </p>
        <h3 className="mt-2 font-display text-2xl text-white">
          {editingServiceId ? 'Atualizar serviço' : 'Criar serviço'}
        </h3>

        <form
          className="mt-6 space-y-4"
          onSubmit={(event) => void submit(event)}
        >
          <Field label="Tipo">
            <input
              required
              className={fieldClassName}
              value={form.type_name}
              onChange={(event) =>
                setForm({ ...form, type_name: event.target.value })
              }
            />
          </Field>
          <Field label="Título">
            <input
              required
              className={fieldClassName}
              value={form.title}
              onChange={(event) =>
                setForm({ ...form, title: event.target.value })
              }
            />
          </Field>
          <Field label="Descrição">
            <textarea
              required
              className={fieldClassName}
              rows={3}
              value={form.description}
              onChange={(event) =>
                setForm({ ...form, description: event.target.value })
              }
            />
          </Field>
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Preço">
              <input
                required
                className={fieldClassName}
                value={form.price}
                onChange={(event) =>
                  setForm({ ...form, price: event.target.value })
                }
              />
            </Field>
            <Field label="Desconto (%)">
              <input
                className={fieldClassName}
                value={form.discount_rate}
                onChange={(event) =>
                  setForm({ ...form, discount_rate: event.target.value })
                }
              />
            </Field>
          </div>
          <Field label="Notas">
            <textarea
              className={fieldClassName}
              rows={3}
              value={form.notes}
              onChange={(event) =>
                setForm({ ...form, notes: event.target.value })
              }
            />
          </Field>
          <MutationError error={mutationError} />
          <FormActions
            isPending={isPending}
            editing={Boolean(editingServiceId)}
            onReset={resetForm}
          />
        </form>
      </section>
    </div>
  );
}

const fieldClassName =
  'w-full rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white outline-none transition placeholder:text-slate-500 focus:border-primary/50 focus:ring-2 focus:ring-primary/20';

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <label className="block space-y-2">
      <span className="text-sm font-medium text-slate-200">{label}</span>
      {children}
    </label>
  );
}

function Actions({
  onEdit,
  onDelete,
}: {
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="flex gap-2">
      <button
        type="button"
        onClick={onEdit}
        className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-slate-200 transition hover:bg-white/10"
      >
        Editar
      </button>
      <button
        type="button"
        onClick={onDelete}
        className="rounded-xl border border-rose-400/40 bg-rose-500/10 px-3 py-1 text-xs text-rose-200 transition hover:bg-rose-500/20"
      >
        Excluir
      </button>
    </div>
  );
}

function StateMessage({
  message,
  tone = 'neutral',
}: {
  message: string;
  tone?: 'neutral' | 'error';
}) {
  return (
    <div
      className={`rounded-2xl border px-4 py-3 text-sm ${tone === 'error' ? 'border-rose-400/30 bg-rose-500/10 text-rose-100' : 'border-white/10 bg-white/5 text-slate-300'}`}
    >
      {message}
    </div>
  );
}

function MutationError({ error }: { error: unknown }) {
  if (!(error instanceof ApiError)) return null;
  return (
    <div className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
      {error.message}
    </div>
  );
}

function FormActions({
  isPending,
  editing,
  onReset,
}: {
  isPending: boolean;
  editing: boolean;
  onReset: () => void;
}) {
  return (
    <div className="flex gap-3">
      <button
        type="submit"
        disabled={isPending}
        className="rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-70"
      >
        {isPending ? 'Salvando...' : editing ? 'Atualizar' : 'Criar'}
      </button>
      <button
        type="button"
        onClick={onReset}
        className="rounded-2xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition hover:bg-white/10"
      >
        Limpar
      </button>
    </div>
  );
}
