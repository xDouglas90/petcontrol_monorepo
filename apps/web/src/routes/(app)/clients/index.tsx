import { type FormEvent, useState } from 'react';
import { User } from 'lucide-react';
import type { ClientDTO, CreateClientInput } from '@petcontrol/shared-types';

import {
  useClientsQuery,
  useCreateClientMutation,
  useDeleteClientMutation,
  useUpdateClientMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';
import { useListParams } from '@/hooks/use-list-params';
import { SearchBar } from '@/ui/search-bar';
import { PaginationBar } from '@/ui/pagination-bar';

type ClientFormState = CreateClientInput;

const initialClientForm: ClientFormState = {
  full_name: '',
  short_name: '',
  gender_identity: 'woman_cisgender',
  marital_status: 'single',
  birth_date: '1992-06-15',
  cpf: '',
  email: '',
  phone: '',
  cellphone: '',
  has_whatsapp: true,
  client_since: '',
  notes: '',
};

export function ClientsPage() {
  const { params, search, setSearch, goToPage } = useListParams();
  const clientsQuery = useClientsQuery(params);
  const createMutation = useCreateClientMutation();
  const updateMutation = useUpdateClientMutation();
  const deleteMutation = useDeleteClientMutation();
  const [editingClientId, setEditingClientId] = useState<string | null>(null);
  const [form, setForm] = useState<ClientFormState>(initialClientForm);

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;
  const isPending =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  function resetForm() {
    setEditingClientId(null);
    setForm(initialClientForm);
  }

  function startEdit(client: ClientDTO) {
    setEditingClientId(client.id);
    setForm({
      full_name: client.full_name,
      short_name: client.short_name,
      gender_identity: client.gender_identity,
      marital_status: client.marital_status,
      birth_date: client.birth_date,
      cpf: client.cpf,
      email: client.email,
      phone: client.phone ?? '',
      cellphone: client.cellphone,
      has_whatsapp: client.has_whatsapp,
      client_since: client.client_since ?? '',
      notes: client.notes ?? '',
    });
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const payload = {
      ...form,
      phone: form.phone || undefined,
      client_since: form.client_since || undefined,
      notes: form.notes || undefined,
    };

    if (editingClientId) {
      await updateMutation.mutateAsync({
        clientId: editingClientId,
        input: payload,
      });
      resetForm();
      return;
    }

    await createMutation.mutateAsync(payload);
    resetForm();
  }

  return (
    <main className="flex min-w-0 flex-col min-h-full">
      <div className="flex-1 grid grid-cols-1 divide-y divide-border/50 xl:grid-cols-[minmax(0,1.1fr)_26rem] xl:divide-x xl:divide-y-0">
        {/* ── List Section ── */}
        <section className="flex flex-col min-h-full">
          <header className="bg-[radial-gradient(circle_at_top_right,rgba(2,132,199,0.08),transparent_40%),radial-gradient(circle_at_bottom_left,rgba(16,185,129,0.05),transparent_35%)] px-6 py-8 lg:px-10">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <p className="app-eyebrow">Gestão</p>
                <h1 className="mt-3 font-display text-4xl text-foreground sm:text-5xl">
                  Clientes
                </h1>
                <p className="mt-4 max-w-2xl text-sm leading-6 text-muted">
                  Base de tutores e responsáveis. Gerencie o cadastro de
                  clientes, informações de contato e histórico de
                  relacionamento.
                </p>
              </div>
              <button
                type="button"
                onClick={resetForm}
                className="inline-flex h-12 items-center justify-center rounded-2xl bg-primary px-6 text-sm font-bold text-slate-950 transition hover:brightness-110 shadow-sm"
              >
                Inserir cliente
              </button>
            </div>
          </header>

          <div className="p-6 lg:p-10">
            <SearchBar
              value={search}
              onChange={setSearch}
              placeholder="Buscar por nome, email ou CPF..."
              id="clients-search"
            />

            <div className="mt-6 space-y-3">
              {clientsQuery.isLoading ? (
                <StateMessage message="Carregando clientes..." />
              ) : null}
              {clientsQuery.isError ? (
                <StateMessage
                  message="Falha ao carregar clientes."
                  tone="error"
                />
              ) : null}
              {!clientsQuery.isLoading &&
              (clientsQuery.data?.data ?? []).length === 0 ? (
                <StateMessage message="Nenhum cliente cadastrado." />
              ) : null}

              {(clientsQuery.data?.data ?? []).map((client) => (
                <article
                  key={client.id}
                  className={`group flex items-center justify-between gap-4 rounded-[1.8rem] border p-4 transition ${
                    editingClientId === client.id
                      ? 'border-primary/40 bg-primary/10'
                      : 'border-border/50 bg-surface/30 hover:border-border hover:bg-surface/60'
                  }`}
                >
                  <div className="flex w-full items-center justify-between gap-3">
                    <div className="flex items-center gap-4">
                      <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-surface border border-border/50 text-primary shadow-sm">
                        <User className="h-5 w-5" />
                      </div>
                      <div className="min-w-0">
                        <p className="font-medium text-foreground group-hover:text-primary transition">
                          {client.full_name}
                        </p>
                        <p className="mt-0.5 text-sm text-muted">
                          {client.email} · {client.cellphone}
                        </p>
                        <p className="text-[11px] text-muted/70">
                          CPF {client.cpf} · desde{' '}
                          {client.client_since ?? 'N/I'}
                        </p>
                      </div>
                    </div>
                    <Actions
                      onEdit={() => startEdit(client)}
                      onDelete={() =>
                        void deleteMutation.mutateAsync(client.id)
                      }
                    />
                  </div>
                </article>
              ))}
            </div>

            <PaginationBar
              meta={clientsQuery.data?.meta}
              onPageChange={goToPage}
            />
          </div>
        </section>

        {/* ── Form Aside ── */}
        <aside className="bg-surface/30 p-6 lg:p-10">
          <div className="xl:sticky xl:top-10">
            <p className="app-eyebrow">
              {editingClientId ? 'Editar cliente' : 'Novo cliente'}
            </p>
            <h3 className="mt-4 font-display text-3xl text-foreground">
              {editingClientId ? 'Atualizar cadastro' : 'Criar cliente'}
            </h3>

            <form
              className="mt-6 space-y-4"
              onSubmit={(event) => void submit(event)}
            >
              <Field label="Nome completo" htmlFor="client-full-name">
                <input
                  id="client-full-name"
                  title="Nome completo do cliente"
                  placeholder="Ex: João da Silva"
                  className={fieldClassName}
                  required
                  value={form.full_name}
                  onChange={(event) =>
                    setForm({ ...form, full_name: event.target.value })
                  }
                />
              </Field>
              <Field label="Nome curto" htmlFor="client-short-name">
                <input
                  id="client-short-name"
                  title="Nome curto ou apelido"
                  placeholder="Ex: João"
                  className={fieldClassName}
                  required
                  value={form.short_name}
                  onChange={(event) =>
                    setForm({ ...form, short_name: event.target.value })
                  }
                />
              </Field>
              <div className="grid gap-4 sm:grid-cols-2">
                <Field label="Nascimento" htmlFor="client-birth">
                  <input
                    id="client-birth"
                    title="Data de nascimento"
                    className={fieldClassName}
                    required
                    type="date"
                    value={form.birth_date}
                    onChange={(event) =>
                      setForm({ ...form, birth_date: event.target.value })
                    }
                  />
                </Field>
                <Field label="CPF" htmlFor="client-cpf">
                  <input
                    id="client-cpf"
                    title="Número do CPF"
                    placeholder="000.000.000-00"
                    className={fieldClassName}
                    required
                    value={form.cpf}
                    onChange={(event) =>
                      setForm({ ...form, cpf: event.target.value })
                    }
                  />
                </Field>
              </div>
              <Field label="E-mail" htmlFor="client-email">
                <input
                  id="client-email"
                  title="Endereço de e-mail"
                  placeholder="exemplo@email.com"
                  className={fieldClassName}
                  required
                  type="email"
                  value={form.email}
                  onChange={(event) =>
                    setForm({ ...form, email: event.target.value })
                  }
                />
              </Field>
              <div className="grid gap-4 sm:grid-cols-2">
                <Field label="Celular" htmlFor="client-cellphone">
                  <input
                    id="client-cellphone"
                    title="Número de celular"
                    placeholder="(00) 00000-0000"
                    className={fieldClassName}
                    required
                    value={form.cellphone}
                    onChange={(event) =>
                      setForm({ ...form, cellphone: event.target.value })
                    }
                  />
                </Field>
                <Field label="Cliente desde" htmlFor="client-since">
                  <input
                    id="client-since"
                    title="Data em que se tornou cliente"
                    className={fieldClassName}
                    type="date"
                    value={form.client_since}
                    onChange={(event) =>
                      setForm({ ...form, client_since: event.target.value })
                    }
                  />
                </Field>
              </div>
              <Field label="Observações" htmlFor="client-notes">
                <textarea
                  id="client-notes"
                  title="Observações adicionais"
                  placeholder="Ex: Preferências, restrições..."
                  className={fieldClassName}
                  rows={3}
                  value={form.notes}
                  onChange={(event) =>
                    setForm({ ...form, notes: event.target.value })
                  }
                />
              </Field>
              <label
                className="flex items-center gap-3 text-sm text-foreground"
                htmlFor="client-has-whatsapp"
              >
                <input
                  id="client-has-whatsapp"
                  title="O cliente possui WhatsApp?"
                  type="checkbox"
                  checked={form.has_whatsapp}
                  onChange={(event) =>
                    setForm({ ...form, has_whatsapp: event.target.checked })
                  }
                />
                Possui WhatsApp
              </label>
              <MutationError error={mutationError} />
              <FormActions
                isPending={isPending}
                editing={Boolean(editingClientId)}
                onReset={resetForm}
              />
            </form>
          </div>
        </aside>
      </div>
    </main>
  );
}

const fieldClassName =
  'w-full rounded-2xl border border-border/50 bg-surface/50 px-3 py-2 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-primary/50 focus:ring-2 focus:ring-primary/20';

function Field({
  label,
  children,
  htmlFor,
}: {
  label: string;
  children: React.ReactNode;
  htmlFor?: string;
}) {
  return (
    <label className="block space-y-2" htmlFor={htmlFor}>
      <span className="text-sm font-medium text-foreground">{label}</span>
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
        className="rounded-xl border border-border/50 bg-surface/50 px-3 py-1 text-xs text-foreground transition hover:bg-surface"
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
      className={`rounded-2xl border px-4 py-3 text-sm ${
        tone === 'error'
          ? 'border-rose-400/30 bg-rose-500/10 text-rose-100'
          : 'border-border/50 bg-surface/50 text-muted'
      }`}
    >
      {message}
    </div>
  );
}

function MutationError({ error }: { error: unknown }) {
  if (!(error instanceof ApiError)) {
    return null;
  }
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
        className="rounded-2xl border border-border/50 bg-surface/50 px-4 py-2 text-sm text-foreground transition hover:bg-surface"
      >
        Limpar
      </button>
    </div>
  );
}
