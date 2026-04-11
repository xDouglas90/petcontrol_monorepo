import { type FormEvent, useState } from 'react';
import type { CreatePetInput, PetDTO } from '@petcontrol/shared-types';

import {
  useClientsQuery,
  useCreatePetMutation,
  useDeletePetMutation,
  usePetsQuery,
  useUpdatePetMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';

type PetFormState = CreatePetInput;

const initialPetForm: PetFormState = {
  owner_id: '',
  name: '',
  size: 'medium',
  kind: 'dog',
  temperament: 'playful',
  image_url: '',
  birth_date: '',
  notes: '',
};

export function PetsPage() {
  const clientsQuery = useClientsQuery();
  const petsQuery = usePetsQuery();
  const createMutation = useCreatePetMutation();
  const updateMutation = useUpdatePetMutation();
  const deleteMutation = useDeletePetMutation();
  const [editingPetId, setEditingPetId] = useState<string | null>(null);
  const [form, setForm] = useState<PetFormState>(initialPetForm);

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;
  const isPending =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  function resetForm() {
    setEditingPetId(null);
    setForm(initialPetForm);
  }

  function startEdit(pet: PetDTO) {
    setEditingPetId(pet.id);
    setForm({
      owner_id: pet.owner_id,
      name: pet.name,
      size: pet.size,
      kind: pet.kind,
      temperament: pet.temperament,
      image_url: pet.image_url ?? '',
      birth_date: pet.birth_date ?? '',
      notes: pet.notes ?? '',
    });
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const payload = {
      ...form,
      image_url: form.image_url || undefined,
      birth_date: form.birth_date || undefined,
      notes: form.notes || undefined,
    };

    if (editingPetId) {
      await updateMutation.mutateAsync({ petId: editingPetId, input: payload });
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
          Pets
        </p>
        <h2 className="mt-2 font-display text-3xl text-white">Pets</h2>
        <p className="mt-2 text-sm text-slate-300">
          Pets associados a clientes do tenant, prontos para alimentar
          agendamentos.
        </p>

        <div className="mt-6 space-y-3">
          {petsQuery.isLoading ? (
            <StateMessage message="Carregando pets..." />
          ) : null}
          {petsQuery.isError ? (
            <StateMessage message="Falha ao carregar pets." tone="error" />
          ) : null}
          {!petsQuery.isLoading && (petsQuery.data?.data ?? []).length === 0 ? (
            <StateMessage message="Nenhum pet cadastrado." />
          ) : null}
          {(petsQuery.data?.data ?? []).map((pet) => (
            <article
              key={pet.id}
              className="rounded-3xl border border-white/10 bg-white/5 p-4"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h3 className="font-display text-xl text-white">
                    {pet.name}
                  </h3>
                  <p className="mt-1 text-sm text-slate-300">
                    {pet.kind} · {pet.size} · {pet.temperament}
                  </p>
                  <p className="mt-1 text-xs text-slate-500">
                    Tutor: {pet.owner_name ?? pet.owner_id}
                  </p>
                </div>
                <Actions
                  onEdit={() => startEdit(pet)}
                  onDelete={() => void deleteMutation.mutateAsync(pet.id)}
                />
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
          {editingPetId ? 'Editar pet' : 'Novo pet'}
        </p>
        <h3 className="mt-2 font-display text-2xl text-white">
          {editingPetId ? 'Atualizar pet' : 'Criar pet'}
        </h3>

        <form
          className="mt-6 space-y-4"
          onSubmit={(event) => void submit(event)}
        >
          <Field label="Tutor">
            <select
              required
              className={fieldClassName}
              value={form.owner_id}
              onChange={(event) =>
                setForm({ ...form, owner_id: event.target.value })
              }
            >
              <option value="">Selecione um cliente</option>
              {(clientsQuery.data?.data ?? []).map((client) => (
                <option key={client.id} value={client.id}>
                  {client.full_name}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Nome">
            <input
              required
              className={fieldClassName}
              value={form.name}
              onChange={(event) =>
                setForm({ ...form, name: event.target.value })
              }
            />
          </Field>
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Porte">
              <select
                className={fieldClassName}
                value={form.size}
                onChange={(event) =>
                  setForm({
                    ...form,
                    size: event.target.value as PetFormState['size'],
                  })
                }
              >
                <option value="small">Pequeno</option>
                <option value="medium">Médio</option>
                <option value="large">Grande</option>
                <option value="giant">Gigante</option>
              </select>
            </Field>
            <Field label="Espécie">
              <select
                className={fieldClassName}
                value={form.kind}
                onChange={(event) =>
                  setForm({
                    ...form,
                    kind: event.target.value as PetFormState['kind'],
                  })
                }
              >
                <option value="dog">Cachorro</option>
                <option value="cat">Gato</option>
                <option value="other">Outro</option>
              </select>
            </Field>
          </div>
          <Field label="Temperamento">
            <select
              className={fieldClassName}
              value={form.temperament}
              onChange={(event) =>
                setForm({
                  ...form,
                  temperament: event.target
                    .value as PetFormState['temperament'],
                })
              }
            >
              <option value="calm">Calmo</option>
              <option value="nervous">Nervoso</option>
              <option value="aggressive">Agressivo</option>
              <option value="playful">Brincalhão</option>
              <option value="loving">Carinhoso</option>
            </select>
          </Field>
          <Field label="Nascimento">
            <input
              type="date"
              className={fieldClassName}
              value={form.birth_date}
              onChange={(event) =>
                setForm({ ...form, birth_date: event.target.value })
              }
            />
          </Field>
          <Field label="Observações">
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
            editing={Boolean(editingPetId)}
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
