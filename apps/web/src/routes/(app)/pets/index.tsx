import {
  type Dispatch,
  type FormEvent,
  type ReactNode,
  type SetStateAction,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { Pencil, Trash2 } from 'lucide-react';
import {
  completeUpload,
  createUploadIntent,
  uploadToGCS,
} from '@/lib/api/rest-client';
import type {
  CreatePetInput,
  PetDTO,
  PetDetailDTO,
  PetKind,
  PetSize,
  PetTemperament,
} from '@petcontrol/shared-types';

import {
  useClientsQuery,
  useCreatePetMutation,
  useDeletePetMutation,
  usePetQuery,
  usePeopleQuery,
  usePetsQuery,
  useUpdatePetMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';
import { useListParams } from '@/hooks/use-list-params';
import { PaginationBar } from '@/ui/pagination-bar';
import { ImageUpload } from '@/ui/image-upload';
import { useAuthStore, selectSession } from '@/lib/auth/auth.store';
import { useToastStore } from '@/stores/toast.store';

type PetPanelMode = 'view' | 'create' | 'edit';
type PetStatusFilter = 'all' | 'active' | 'inactive';

type PetFormState = CreatePetInput;

type PetFilterState = {
  size: 'all' | PetSize;
  kind: 'all' | PetKind;
  temperament: 'all' | PetTemperament;
  is_active: PetStatusFilter;
};

const INITIAL_FORM: PetFormState = {
  owner_id: '',
  guardian_ids: undefined,
  name: '',
  race: '',
  color: '',
  sex: '',
  size: 'medium',
  kind: 'dog',
  temperament: 'playful',
  image_url: '',
  upload_object_key: '',
  birth_date: '',
  is_active: true,
  is_deceased: false,
  is_vaccinated: false,
  is_neutered: false,
  is_microchipped: false,
  microchip_number: '',
  microchip_expiration_date: '',
  notes: '',
};

const INITIAL_FILTERS: PetFilterState = {
  size: 'all',
  kind: 'all',
  temperament: 'all',
  is_active: 'all',
};

const PET_SIZE_OPTIONS: Array<{ value: 'all' | PetSize; label: string }> = [
  { value: 'all', label: 'Todos os portes' },
  { value: 'small', label: 'Pequeno' },
  { value: 'medium', label: 'Médio' },
  { value: 'large', label: 'Grande' },
  { value: 'giant', label: 'Gigante' },
];

const PET_KIND_OPTIONS: Array<{ value: 'all' | PetKind; label: string }> = [
  { value: 'all', label: 'Todos os tipos' },
  { value: 'dog', label: 'Cachorro' },
  { value: 'cat', label: 'Gato' },
  { value: 'bird', label: 'Ave' },
  { value: 'fish', label: 'Peixe' },
  { value: 'reptile', label: 'Réptil' },
  { value: 'rodent', label: 'Roedor' },
  { value: 'rabbit', label: 'Coelho' },
  { value: 'other', label: 'Outro' },
];

const PET_TEMPERAMENT_OPTIONS: Array<{
  value: 'all' | PetTemperament;
  label: string;
}> = [
  { value: 'all', label: 'Todos os temperamentos' },
  { value: 'calm', label: 'Calmo' },
  { value: 'nervous', label: 'Nervoso' },
  { value: 'aggressive', label: 'Agressivo' },
  { value: 'playful', label: 'Brincalhão' },
  { value: 'loving', label: 'Carinhoso' },
];

const PET_STATUS_OPTIONS: Array<{ value: PetStatusFilter; label: string }> = [
  { value: 'all', label: 'Todos' },
  { value: 'active', label: 'Ativos' },
  { value: 'inactive', label: 'Inativos' },
];

const GCS_PUBLIC_URL = (
  import.meta.env.VITE_GCS_PUBLIC_URL ??
  import.meta.env.VITE_GCS_PUBLIC_BASE_URL ??
  ''
).replace(/\/$/, '');

export function PetsPage() {
  const session = useAuthStore(selectSession);
  const pushToast = useToastStore((state) => state.pushToast);
  const clientsQuery = useClientsQuery();
  const guardiansQuery = usePeopleQuery({
    kind: 'guardian',
    page: 1,
    limit: 200,
  });
  const createMutation = useCreatePetMutation();
  const updateMutation = useUpdatePetMutation();
  const deleteMutation = useDeletePetMutation();
  const [selectedPetId, setSelectedPetId] = useState<string | null>(
    readSelectedPetIdFromLocation(),
  );
  const [panelMode, setPanelMode] = useState<PetPanelMode>(
    readPanelModeFromLocation(),
  );
  const [filters, setFilters] = useState<PetFilterState>(() =>
    readFiltersFromLocation(),
  );
  const [form, setForm] = useState<PetFormState>(INITIAL_FORM);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const { page, params, search, setSearch, goToPage } = useListParams(
    undefined,
    readSearchFromLocation(),
    readPageFromLocation(),
  );

  useEffect(() => {
    function handleOpenCreateForm() {
      startCreate();
    }

    window.addEventListener('open-pets-create-form', handleOpenCreateForm);
    return () => {
      window.removeEventListener('open-pets-create-form', handleOpenCreateForm);
    };
  }, []);

  useEffect(() => {
    syncPetLocation({
      page,
      search,
      selectedPetId,
      panelMode,
      filters,
    });
  }, [filters, page, panelMode, search, selectedPetId]);

  const petsQueryParams = useMemo(
    () => ({
      ...params,
      ...(filters.size !== 'all' ? { size: filters.size } : {}),
      ...(filters.kind !== 'all' ? { kind: filters.kind } : {}),
      ...(filters.temperament !== 'all'
        ? { temperament: filters.temperament }
        : {}),
      ...(filters.is_active !== 'all'
        ? { is_active: filters.is_active === 'active' ? 'true' : 'false' }
        : {}),
    }),
    [filters, params],
  );

  const petsQuery = usePetsQuery(petsQueryParams);

  const visiblePets = useMemo(
    () => petsQuery.data?.data ?? [],
    [petsQuery.data],
  );
  const activeSelectedPetId = useMemo(() => {
    if (panelMode === 'create') {
      return null;
    }

    if (selectedPetId && visiblePets.some((pet) => pet.id === selectedPetId)) {
      return selectedPetId;
    }

    return visiblePets[0]?.id ?? null;
  }, [panelMode, selectedPetId, visiblePets]);

  const selectedPetSummary =
    visiblePets.find((pet) => pet.id === activeSelectedPetId) ?? null;
  const petDetailQuery = usePetQuery(activeSelectedPetId ?? undefined);
  const selectedPetDetail =
    petDetailQuery.data?.data?.id === activeSelectedPetId
      ? petDetailQuery.data.data
      : null;
  const selectedPet = selectedPetDetail ?? selectedPetSummary;

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;
  const isPending =
    isUploading ||
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  useEffect(() => {
    if (panelMode !== 'create' && activeSelectedPetId !== selectedPetId) {
      setSelectedPetId(activeSelectedPetId);
    }
  }, [activeSelectedPetId, panelMode, selectedPetId]);

  useEffect(() => {
    if (panelMode === 'create') {
      return;
    }
    if (!selectedPet) {
      return;
    }
    setForm(buildFormFromPet(selectedPet));
  }, [panelMode, selectedPet]);

  function startCreate() {
    setPanelMode('create');
    setSelectedPetId(null);
    setForm(INITIAL_FORM);
    setSelectedFile(null);
  }

  function startEdit(pet: PetDTO) {
    setPanelMode('edit');
    setSelectedPetId(pet.id);
    setForm(buildFormFromPet(pet));
    setSelectedFile(null);
  }

  function selectPet(pet: PetDTO) {
    setPanelMode('view');
    setSelectedPetId(pet.id);
    setForm(buildFormFromPet(pet));
  }

  function resetPanelToView() {
    if (selectedPetSummary) {
      setPanelMode('view');
      setSelectedPetId(selectedPetSummary.id);
      setForm(buildFormFromPet(selectedPetSummary));
      setSelectedFile(null);
      return;
    }
    startCreate();
  }

  function updateFilter<K extends keyof PetFilterState>(
    key: K,
    value: PetFilterState[K],
  ) {
    setFilters((current) => ({ ...current, [key]: value }));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!form.owner_id.trim()) {
      pushToast('Selecione o tutor antes de continuar.', 'error');
      return;
    }

    if (!form.name.trim()) {
      pushToast('Informe o nome do pet.', 'error');
      return;
    }

    let uploadObjectKey = form.upload_object_key;

    if (selectedFile && session?.accessToken) {
      try {
        setIsUploading(true);
        const intent = await createUploadIntent(session.accessToken, {
          resource: 'pets',
          field: 'image_url',
          file_name: selectedFile.name,
          content_type: selectedFile.type,
          size_bytes: selectedFile.size,
        });

        await uploadToGCS(intent, selectedFile);
        const completed = await completeUpload(session.accessToken, {
          resource: 'pets',
          field: 'image_url',
          object_key: intent.object_key,
        });
        uploadObjectKey = completed.object_key;
      } catch (error) {
        console.error('Failed to upload image:', error);
        pushToast('Falha ao enviar imagem. Tente novamente.', 'error');
        setIsUploading(false);
        return;
      } finally {
        setIsUploading(false);
      }
    }

    const payload: CreatePetInput = {
      ...form,
      owner_id: form.owner_id,
      guardian_ids: form.guardian_ids ?? [],
      name: form.name.trim(),
      race: form.race?.trim() || undefined,
      color: form.color?.trim() || undefined,
      sex: form.sex?.trim() || undefined,
      image_url: form.image_url?.trim() || undefined,
      upload_object_key: uploadObjectKey || undefined,
      birth_date: form.birth_date || undefined,
      microchip_number: form.microchip_number?.trim() || undefined,
      microchip_expiration_date: form.microchip_expiration_date || undefined,
      notes: form.notes?.trim() || undefined,
    };

    try {
      if (panelMode === 'edit' && activeSelectedPetId) {
        const updated = await updateMutation.mutateAsync({
          petId: activeSelectedPetId,
          input: payload,
        });
        setSelectedPetId(updated.id);
        setPanelMode('view');
        pushToast('Pet atualizado com sucesso.', 'success');
        return;
      }

      const created = await createMutation.mutateAsync(payload);
      setSelectedPetId(created.id);
      setPanelMode('view');
      pushToast('Pet criado com sucesso.', 'success');
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Falha ao salvar o pet.';
      pushToast(message, 'error');
    }
  }

  const listIsEmpty =
    !petsQuery.isLoading && !petsQuery.isError && visiblePets.length === 0;
  return (
    <div className="overflow-hidden bg-white/75 shadow-[0_20px_50px_rgba(15,23,42,0.05)]">
      <div className="grid min-h-full grid-cols-1 divide-y divide-stone-100 xl:grid-cols-[minmax(0,1.1fr)_26rem] xl:divide-x xl:divide-y-0">
        <section>
          <header className="bg-[radial-gradient(circle_at_top_left,_rgba(249,115,22,0.12),_transparent_40%),linear-gradient(145deg,#fffef8,#f5f5f4)] px-6 py-8 lg:px-10">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.34em] text-stone-400">
                  Gestão de Pets
                </p>
                <h1 className="mt-3 font-display text-4xl text-stone-950 sm:text-5xl">
                  Pets
                </h1>
                <p className="mt-4 max-w-2xl text-sm leading-6 text-stone-500">
                  Listagem de pets com filtros estruturados, acesso rápido ao
                  detalhe e composição visual centrada na imagem do pet.
                </p>
              </div>

              <button
                type="button"
                onClick={startCreate}
                className="inline-flex h-12 items-center justify-center rounded-2xl bg-sky-600 px-6 text-sm font-bold text-white shadow-sm shadow-sky-200 transition hover:bg-sky-700"
              >
                Inserir pet
              </button>
            </div>
          </header>

          <div className="px-6 py-6 lg:px-10 lg:py-10">
            <div>
              <input
                id="pets-search"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Buscar por nome do pet, nome do cliente, raça ou tipo..."
                className="h-12 w-full rounded-2xl border border-stone-200 bg-stone-50 px-4 text-sm text-stone-900 outline-none transition placeholder:text-stone-400 focus:border-sky-300 focus:bg-white"
              />
            </div>

            <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <select
                aria-label="Filtrar por porte"
                value={filters.size}
                onChange={(event) =>
                  updateFilter(
                    'size',
                    event.target.value as PetFilterState['size'],
                  )
                }
                className="h-12 rounded-2xl border border-stone-200 bg-white px-4 text-sm text-stone-700 outline-none transition focus:border-sky-300"
              >
                {PET_SIZE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>

              <select
                aria-label="Filtrar por tipo"
                value={filters.kind}
                onChange={(event) =>
                  updateFilter(
                    'kind',
                    event.target.value as PetFilterState['kind'],
                  )
                }
                className="h-12 rounded-2xl border border-stone-200 bg-white px-4 text-sm text-stone-700 outline-none transition focus:border-sky-300"
              >
                {PET_KIND_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>

              <select
                aria-label="Filtrar por temperamento"
                value={filters.temperament}
                onChange={(event) =>
                  updateFilter(
                    'temperament',
                    event.target.value as PetFilterState['temperament'],
                  )
                }
                className="h-12 rounded-2xl border border-stone-200 bg-white px-4 text-sm text-stone-700 outline-none transition focus:border-sky-300"
              >
                {PET_TEMPERAMENT_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>

              <select
                aria-label="Filtrar por status"
                value={filters.is_active}
                onChange={(event) =>
                  updateFilter(
                    'is_active',
                    event.target.value as PetFilterState['is_active'],
                  )
                }
                className="h-12 rounded-2xl border border-stone-200 bg-white px-4 text-sm text-stone-700 outline-none transition focus:border-sky-300"
              >
                {PET_STATUS_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="mt-8 space-y-px divide-y divide-stone-100 border-y border-stone-100">
              {petsQuery.isLoading ? (
                <StateMessage message="Carregando pets..." />
              ) : null}
              {petsQuery.isError ? (
                <StateMessage message="Falha ao carregar pets." tone="error" />
              ) : null}
              {listIsEmpty ? (
                <StateMessage message="Nenhum pet encontrado com os filtros atuais." />
              ) : null}

              <div>
                {visiblePets.map((pet) => {
                  const isSelected = pet.id === activeSelectedPetId;
                  return (
                    <PetListCard
                      key={pet.id}
                      pet={pet}
                      selected={isSelected}
                      onSelect={() => selectPet(pet)}
                    />
                  );
                })}
              </div>
            </div>

            <div className="border-t border-stone-100 pt-4">
              <PaginationBar
                meta={petsQuery.data?.meta}
                onPageChange={goToPage}
              />
            </div>
          </div>
        </section>

        <aside className="bg-white p-6 xl:sticky xl:top-0 xl:h-screen xl:overflow-y-auto xl:p-10">
          {panelMode === 'create' || panelMode === 'edit' ? (
            <PetFormPanel
              mode={panelMode}
              form={form}
              setForm={setForm}
              onSubmit={submit}
              onReset={resetPanelToView}
              mutationError={mutationError}
              isPending={isPending}
              setSelectedFile={setSelectedFile}
              clients={clientsQuery.data?.data ?? []}
              guardians={guardiansQuery.data?.data ?? []}
            />
          ) : selectedPet ? (
            <PetDetailPanel
              pet={selectedPet}
              isLoading={petDetailQuery.isLoading}
              onEdit={() => startEdit(selectedPet)}
              onDelete={() => void deleteMutation.mutateAsync(selectedPet.id)}
            />
          ) : (
            <EmptyAside
              title="Selecione um pet"
              description="O detalhe, tutor e guardiões aparecem aqui. Crie um pet novo para começar."
            />
          )}
        </aside>
      </div>
    </div>
  );
}

function PetFormPanel({
  mode,
  form,
  setForm,
  onSubmit,
  onReset,
  mutationError,
  isPending,
  setSelectedFile,
  clients,
  guardians,
}: {
  mode: PetPanelMode;
  form: PetFormState;
  setForm: Dispatch<SetStateAction<PetFormState>>;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onReset: () => void;
  mutationError: unknown;
  isPending: boolean;
  setSelectedFile: (file: File | null) => void;
  clients: Array<{ id: string; full_name: string; short_name?: string | null }>;
  guardians: Array<{
    id: string;
    full_name?: string | null;
    short_name?: string | null;
  }>;
}) {
  const hasGuardianSelection = form.guardian_ids !== undefined;
  return (
    <div className="xl:sticky xl:top-10">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.28em] text-stone-400">
            {mode === 'edit' ? 'Editar pet' : 'Novo pet'}
          </p>
          <h2 className="mt-4 font-display text-3xl text-stone-950">
            {mode === 'edit' ? 'Atualizar cadastro' : 'Formulário de cadastro'}
          </h2>
        </div>

        <button
          type="button"
          onClick={onReset}
          className="rounded-2xl border border-stone-200 px-4 py-2 text-sm font-semibold text-stone-600 transition hover:bg-stone-50"
        >
          Cancelar
        </button>
      </div>

      <p className="mt-4 text-xs font-semibold uppercase tracking-[0.2em] text-stone-400">
        Campos com * sao obrigatórios.
      </p>

      <form className="mt-8 space-y-8" onSubmit={onSubmit}>
        <DetailSection title="Vinculo">
          <Field label="Tutor" htmlFor="pet-owner" required>
            <select
              title="Tutor"
              id="pet-owner"
              required
              className={fieldClassName}
              value={form.owner_id}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  owner_id: event.target.value,
                }))
              }
            >
              <option value="">Selecione um cliente</option>
              {clients.map((client) => (
                <option key={client.id} value={client.id}>
                  {client.full_name}
                </option>
              ))}
            </select>
          </Field>

          <ToggleField
            label="Inserir guardião"
            checked={hasGuardianSelection}
            onChange={(checked) =>
              setForm((current) => ({
                ...current,
                guardian_ids: checked
                  ? (current.guardian_ids ?? [])
                  : undefined,
              }))
            }
          />

          {hasGuardianSelection ? (
            <Field label="Guardião" htmlFor="pet-guardian">
              <select
                title="Guardião"
                id="pet-guardian"
                className={fieldClassName}
                value={form.guardian_ids?.[0] ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    guardian_ids: event.target.value
                      ? [event.target.value]
                      : [],
                  }))
                }
              >
                <option value="">Selecione um guardião</option>
                {guardians.map((guardian) => (
                  <option key={guardian.id} value={guardian.id}>
                    {guardian.full_name || guardian.short_name || guardian.id}
                  </option>
                ))}
              </select>
            </Field>
          ) : null}
        </DetailSection>

        <DetailSection title="Mídia">
          <ImageUpload
            label="Foto do pet"
            module="pets"
            value={form.image_url}
            onFileSelect={setSelectedFile}
          />
        </DetailSection>

        <DetailSection title="Identificação">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Nome" htmlFor="pet-name" required>
              <input
                id="pet-name"
                required
                className={fieldClassName}
                value={form.name}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    name: event.target.value,
                  }))
                }
                placeholder="Ex: Thor"
              />
            </Field>
            <Field label="Raca" htmlFor="pet-race">
              <input
                id="pet-race"
                className={fieldClassName}
                value={form.race ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    race: event.target.value,
                  }))
                }
                placeholder="Ex: Labrador"
              />
            </Field>
            <Field label="Cor" htmlFor="pet-color">
              <input
                id="pet-color"
                className={fieldClassName}
                value={form.color ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    color: event.target.value,
                  }))
                }
                placeholder="Ex: Caramelo"
              />
            </Field>
            <Field label="Sexo" htmlFor="pet-sex">
              <input
                id="pet-sex"
                className={fieldClassName}
                value={form.sex ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    sex: event.target.value,
                  }))
                }
                placeholder="Ex: M"
              />
            </Field>
          </div>
        </DetailSection>

        <DetailSection title="Perfil">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Porte" htmlFor="pet-size">
              <select
                title="Porte"
                id="pet-size"
                className={fieldClassName}
                value={form.size}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    size: event.target.value as PetSize,
                  }))
                }
              >
                <option value="small">Pequeno</option>
                <option value="medium">Médio</option>
                <option value="large">Grande</option>
                <option value="giant">Gigante</option>
              </select>
            </Field>
            <Field label="Tipo" htmlFor="pet-kind">
              <select
                title="Tipo"
                id="pet-kind"
                className={fieldClassName}
                value={form.kind}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    kind: event.target.value as PetKind,
                  }))
                }
              >
                <option value="dog">Cachorro</option>
                <option value="cat">Gato</option>
                <option value="bird">Ave</option>
                <option value="fish">Peixe</option>
                <option value="reptile">Reptil</option>
                <option value="rodent">Roedor</option>
                <option value="rabbit">Coelho</option>
                <option value="other">Outro</option>
              </select>
            </Field>
            <Field label="Temperamento" htmlFor="pet-temperament">
              <select
                title="Temperamento"
                id="pet-temperament"
                className={fieldClassName}
                value={form.temperament}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    temperament: event.target.value as PetTemperament,
                  }))
                }
              >
                <option value="calm">Calmo</option>
                <option value="nervous">Nervoso</option>
                <option value="aggressive">Agressivo</option>
                <option value="playful">Brincalhão</option>
                <option value="loving">Carinhoso</option>
              </select>
            </Field>
            <Field label="Nascimento" htmlFor="pet-birth">
              <input
                title="Nascimento"
                id="pet-birth"
                type="date"
                className={fieldClassName}
                value={form.birth_date ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    birth_date: event.target.value,
                  }))
                }
              />
            </Field>
          </div>
        </DetailSection>

        <DetailSection title="Saúde">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Numero do microchip" htmlFor="pet-microchip">
              <input
                title="Número do microchip"
                id="pet-microchip"
                className={fieldClassName}
                value={form.microchip_number ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    microchip_number: event.target.value,
                  }))
                }
                placeholder="Opcional"
              />
            </Field>
            <Field
              label="Validade do microchip"
              htmlFor="pet-microchip-expiration"
            >
              <input
                title="Validade do microchip"
                id="pet-microchip-expiration"
                type="date"
                className={fieldClassName}
                value={form.microchip_expiration_date ?? ''}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    microchip_expiration_date: event.target.value,
                  }))
                }
              />
            </Field>
          </div>

          <div className="mt-4 grid gap-3 sm:grid-cols-2">
            <ToggleField
              label="Ativo"
              checked={Boolean(form.is_active)}
              onChange={(checked) =>
                setForm((current) => ({ ...current, is_active: checked }))
              }
            />
            <ToggleField
              label="Falecido"
              checked={Boolean(form.is_deceased)}
              onChange={(checked) =>
                setForm((current) => ({ ...current, is_deceased: checked }))
              }
            />
            <ToggleField
              label="Vacinado"
              checked={Boolean(form.is_vaccinated)}
              onChange={(checked) =>
                setForm((current) => ({ ...current, is_vaccinated: checked }))
              }
            />
            <ToggleField
              label="Castrado"
              checked={Boolean(form.is_neutered)}
              onChange={(checked) =>
                setForm((current) => ({ ...current, is_neutered: checked }))
              }
            />
            <ToggleField
              label="Microchipado"
              checked={Boolean(form.is_microchipped)}
              onChange={(checked) =>
                setForm((current) => ({ ...current, is_microchipped: checked }))
              }
            />
          </div>
        </DetailSection>

        <DetailSection title="Observações">
          <Field label="Notas" htmlFor="pet-notes">
            <textarea
              id="pet-notes"
              rows={5}
              className={`${fieldClassName} min-h-28 resize-y`}
              value={form.notes ?? ''}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  notes: event.target.value,
                }))
              }
              placeholder="Cuidados, alergias, comportamento..."
            />
          </Field>
        </DetailSection>

        <MutationError error={mutationError} />

        <div className="flex flex-wrap gap-3 pt-1">
          <button
            type="submit"
            disabled={isPending}
            className="inline-flex h-12 items-center justify-center rounded-2xl bg-sky-600 px-6 text-sm font-bold text-white shadow-sm shadow-sky-200 transition hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isPending
              ? 'Salvando...'
              : mode === 'edit'
                ? 'Atualizar'
                : 'Criar'}
          </button>
          <button
            type="button"
            onClick={onReset}
            className="inline-flex h-12 items-center justify-center rounded-2xl border border-stone-200 px-6 text-sm font-semibold text-stone-600 transition hover:bg-stone-50"
          >
            Cancelar
          </button>
        </div>
      </form>
    </div>
  );
}

function PetDetailPanel({
  pet,
  onEdit,
  onDelete,
  isLoading,
}: {
  pet: PetDetailDTO | PetDTO;
  onEdit: () => void;
  onDelete: () => void;
  isLoading: boolean;
}) {
  const imageUrl = resolvePetImageUrl(pet.kind, pet.image_url ?? undefined);
  const guardians = 'guardians' in pet ? (pet.guardians ?? []) : [];

  return (
    <div className="space-y-5">
      <div>
        <p className="text-xs uppercase tracking-[0.3em] text-stone-400">
          Detalhe
        </p>
        <h2 className="mt-2 font-display text-3xl text-stone-950">
          {pet.name}
        </h2>
        <p className="mt-1 text-sm text-stone-500">
          {resolvePetKindLabel(pet.kind)} · {resolvePetSizeLabel(pet.size)}
        </p>
      </div>

      <div className="space-y-4 border-y border-stone-200 py-4">
        <div className="aspect-[4/3] w-full rounded-2xl overflow-hidden bg-stone-100">
          <img
            src={imageUrl}
            alt={pet.name}
            className="h-full w-full object-center"
          />
        </div>
        <div className="flex items-center justify-between gap-3">
          <div>
            <p className="text-xs uppercase tracking-[0.28em] text-stone-400">
              Status
            </p>
            <div className="mt-2 flex flex-wrap gap-2">
              <StatusBadge active={pet.is_active} />
              {pet.is_deceased ? (
                <StatusBadge label="Falecido" tone="neutral" />
              ) : null}
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={onEdit}
              title="Editar"
              aria-label="Editar"
              className="inline-flex h-10 items-center justify-center rounded-2xl border border-stone-200 px-4 py-2 text-sm font-semibold text-stone-700 transition hover:bg-stone-50"
            >
              <Pencil className="h-4 w-4" aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={onDelete}
              title="Excluir"
              aria-label="Excluir"
              className="inline-flex h-10 items-center justify-center rounded-2xl border border-rose-200 px-6 text-sm font-semibold text-rose-600 transition hover:bg-rose-50"
            >
              <Trash2 className="h-4 w-4" aria-hidden="true" />
            </button>
          </div>
        </div>
      </div>

      {isLoading ? <StateMessage message="Carregando detalhe..." /> : null}

      <div className="grid gap-0 sm:grid-cols-2">
        <InfoCard label="Tutor" value={pet.owner_name ?? pet.owner_id} />
        <InfoCard label="Raça" value={pet.race || 'Não informada'} />
        <InfoCard label="Cor" value={pet.color || 'Não informada'} />
        <InfoCard label="Sexo" value={pet.sex || 'Não informado'} />
        <InfoCard
          label="Temperamento"
          value={resolvePetTemperamentLabel(pet.temperament)}
        />
        <InfoCard
          label="Nascimento"
          value={pet.birth_date ?? 'Não informado'}
        />
      </div>

      <div className="grid gap-0 sm:grid-cols-2">
        <InfoCard label="Vacinado" value={boolToLabel(pet.is_vaccinated)} />
        <InfoCard label="Castrado" value={boolToLabel(pet.is_neutered)} />
        <InfoCard
          label="Microchipado"
          value={boolToLabel(pet.is_microchipped)}
        />
        <InfoCard
          label="Nº microchip"
          value={pet.microchip_number ?? 'Não informado'}
        />
      </div>

      <section className="border-t border-stone-200 pt-4">
        <p className="text-xs uppercase tracking-[0.28em] text-stone-400">
          Guardiões
        </p>
        <div className="mt-3 space-y-0">
          {guardians.length > 0 ? (
            guardians.map((guardian) => (
              <div
                key={guardian.guardian_id}
                className="border-b border-stone-200 py-3 last:border-b-0"
              >
                <p className="font-medium text-stone-950">
                  {guardian.full_name}
                </p>
                <p className="text-xs text-stone-500">
                  {guardian.email} · {guardian.cellphone}
                </p>
              </div>
            ))
          ) : (
            <p className="text-sm text-stone-500">
              Nenhum guardião vinculado a este pet.
            </p>
          )}
        </div>
      </section>

      {pet.notes ? (
        <section className="border-t border-stone-200 pt-4">
          <p className="text-xs uppercase tracking-[0.28em] text-stone-400">
            Observações
          </p>
          <p className="mt-3 text-sm leading-6 text-stone-700">{pet.notes}</p>
        </section>
      ) : null}
    </div>
  );
}

function EmptyAside({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="flex min-h-[24rem] flex-col justify-center border-t border-stone-200 py-8 text-center">
      <p className="font-display text-3xl text-stone-950">{title}</p>
      <p className="mt-3 text-sm leading-6 text-stone-600">{description}</p>
    </div>
  );
}

function PetListCard({
  pet,
  selected,
  onSelect,
}: {
  pet: PetDTO;
  selected: boolean;
  onSelect: () => void;
}) {
  const imageUrl = resolvePetImageUrl(pet.kind, pet.image_url ?? undefined);

  return (
    <article
      className={`transition ${selected ? 'bg-sky-50/50 ring-1 ring-inset ring-sky-100' : 'bg-white hover:bg-stone-50'}`}
    >
      <button
        type="button"
        onClick={onSelect}
        className="flex w-full flex-col gap-4 p-4 text-lef lg:flex-row items-center"
      >
        <div className="relative h-24 w-24 shrink-0 overflow-hidden rounded-2xl bg-stone-100 lg:h-20 lg:w-20">
          <img
            src={imageUrl}
            alt={pet.name}
            className="h-full w-full object-cover"
          />
        </div>
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <div className="flex items-start justify-between border-b gap-3">
            <div className="min-w-0">
              <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                <h3 className="font-display text-xl text-stone-950">
                  {pet.name}
                </h3>
                <span className="text-sm text-stone-400">
                  ({pet.race || 'sem raça'})
                </span>
              </div>
              <p className="mt-1 text-sm text-stone-500">
                Tutor: {pet.owner_name ?? pet.owner_id}
              </p>
            </div>
            <div className="flex flex-col items-end gap-2">
              <StatusBadge active={pet.is_active} />
              {pet.is_deceased ? (
                <StatusBadge label="Falecido" tone="neutral" />
              ) : null}
            </div>
          </div>

          <div className="grid text-xs text-stone-600 sm:grid-cols-3">
            <MetaRow label="Tipo" value={resolvePetKindLabel(pet.kind)} />
            <MetaRow label="Porte" value={resolvePetSizeLabel(pet.size)} />
            <MetaRow
              label="Temperamento"
              value={resolvePetTemperamentLabel(pet.temperament)}
            />
            <MetaRow label="Cor" value={pet.color || 'Não informada'} />
            <MetaRow label="Sexo" value={pet.sex || 'Não informado'} />
            <MetaRow label="Idade" value={resolvePetAgeLabel(pet.birth_date)} />
          </div>
        </div>
      </button>
    </article>
  );
}

function Field({
  label,
  children,
  htmlFor,
  required = false,
}: {
  label: string;
  children: ReactNode;
  htmlFor?: string;
  required?: boolean;
}) {
  return (
    <label className="block space-y-2" htmlFor={htmlFor}>
      <span className="text-[10px] font-bold uppercase tracking-[0.24em] text-stone-400">
        {label}
        {required ? <span className="ml-1 text-rose-500">*</span> : null}
      </span>
      {children}
    </label>
  );
}

function DetailSection({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section>
      <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-stone-400">
        {title}
      </p>
      <div className="mt-3 space-y-4">{children}</div>
    </section>
  );
}

function ToggleField({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label className="flex items-center justify-between gap-4 rounded-2xl border border-stone-200 px-4 py-3">
      <span className="text-sm font-medium text-stone-700">{label}</span>
      <input
        type="checkbox"
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
        className="h-4 w-4 rounded border-stone-300 text-sky-600 focus:ring-sky-500"
      />
    </label>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="border-b border-stone-200 py-3 pr-4">
      <p className="text-xs uppercase tracking-[0.24em] text-stone-400">
        {label}
      </p>
      <p className="mt-1 text-sm font-medium text-stone-950">{value}</p>
    </div>
  );
}

function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-left gap-2 py-1.5 pr-2">
      <span className="text-[10px] uppercase tracking-[0.2em] text-stone-400">
        {label}:
      </span>
      <span className="text-right text-xs font-medium text-stone-800">
        {value}
      </span>
    </div>
  );
}

function StatusBadge({
  active,
  label,
  tone = 'active',
}: {
  active?: boolean;
  label?: string;
  tone?: 'active' | 'neutral';
}) {
  const text = label ?? (active ? 'Ativo' : 'Inativo');
  const activeTone = active ?? tone === 'active';
  return (
    <span
      className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] ${activeTone ? 'bg-emerald-100 text-emerald-700' : 'bg-stone-100 text-stone-600'}`}
    >
      {text}
    </span>
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
      className={`border-l-2 pl-3 text-sm ${tone === 'error' ? 'border-rose-300 text-rose-700' : 'border-stone-300 text-stone-600'}`}
    >
      {message}
    </div>
  );
}

function MutationError({ error }: { error: unknown }) {
  if (!(error instanceof ApiError)) return null;

  return (
    <div className="border-l-2 border-rose-300 pl-3 text-sm text-rose-700">
      {error.message}
    </div>
  );
}

function resolvePetImageUrl(kind: string, imageUrl?: string | null) {
  if (imageUrl) {
    return imageUrl;
  }

  const normalizedKind = kind?.trim().toLowerCase() || 'other';
  const fallback = `${GCS_PUBLIC_URL || ''}/assets/images/${normalizedKind}-default-image.png`;
  return fallback.startsWith('//') ? fallback.slice(1) : fallback;
}

function buildFormFromPet(pet: PetDTO): PetFormState {
  const guardianIDs =
    pet.guardians && pet.guardians.length > 0
      ? pet.guardians.map((item) => item.guardian_id)
      : undefined;

  return {
    owner_id: pet.owner_id,
    guardian_ids: guardianIDs,
    name: pet.name,
    race: pet.race ?? '',
    color: pet.color ?? '',
    sex: pet.sex ?? '',
    size: pet.size,
    kind: pet.kind,
    temperament: pet.temperament,
    image_url: pet.image_url ?? '',
    upload_object_key: '',
    birth_date: pet.birth_date ?? '',
    is_active: pet.is_active,
    is_deceased: pet.is_deceased,
    is_vaccinated: pet.is_vaccinated,
    is_neutered: pet.is_neutered,
    is_microchipped: pet.is_microchipped,
    microchip_number: pet.microchip_number ?? '',
    microchip_expiration_date: pet.microchip_expiration_date ?? '',
    notes: pet.notes ?? '',
  };
}

function readSearchFromLocation(): string {
  if (typeof window === 'undefined') {
    return '';
  }

  return new URLSearchParams(window.location.search).get('search') ?? '';
}

function readPageFromLocation(): number {
  if (typeof window === 'undefined') {
    return 1;
  }

  const raw = new URLSearchParams(window.location.search).get('page');
  const parsed = Number(raw);
  if (!Number.isInteger(parsed) || parsed < 1) {
    return 1;
  }

  return parsed;
}

function readPanelModeFromLocation(): PetPanelMode {
  if (typeof window === 'undefined') {
    return 'view';
  }

  const panel = new URLSearchParams(window.location.search).get('panel');
  if (panel === 'create' || panel === 'edit') {
    return panel;
  }
  return 'view';
}

function readSelectedPetIdFromLocation(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const params = new URLSearchParams(window.location.search);
  return params.get('id') ?? params.get('pet');
}

function readFiltersFromLocation(): PetFilterState {
  if (typeof window === 'undefined') {
    return INITIAL_FILTERS;
  }

  const params = new URLSearchParams(window.location.search);
  const status = params.get('is_active');
  return {
    size: isPetSize(params.get('size'))
      ? (params.get('size') as PetSize)
      : 'all',
    kind: isPetKind(params.get('kind'))
      ? (params.get('kind') as PetKind)
      : 'all',
    temperament: isPetTemperament(params.get('temperament'))
      ? (params.get('temperament') as PetTemperament)
      : 'all',
    is_active:
      status === 'true' ? 'active' : status === 'false' ? 'inactive' : 'all',
  };
}

function syncPetLocation({
  page,
  search,
  selectedPetId,
  panelMode,
  filters,
}: {
  page: number;
  search: string;
  selectedPetId: string | null;
  panelMode: PetPanelMode;
  filters: PetFilterState;
}) {
  if (typeof window === 'undefined') {
    return;
  }

  const url = new URL(window.location.href);

  if (page <= 1) url.searchParams.delete('page');
  else url.searchParams.set('page', String(page));

  const trimmedSearch = search.trim();
  if (trimmedSearch === '') url.searchParams.delete('search');
  else url.searchParams.set('search', trimmedSearch);

  if (filters.size === 'all') url.searchParams.delete('size');
  else url.searchParams.set('size', filters.size);

  if (filters.kind === 'all') url.searchParams.delete('kind');
  else url.searchParams.set('kind', filters.kind);

  if (filters.temperament === 'all') url.searchParams.delete('temperament');
  else url.searchParams.set('temperament', filters.temperament);

  if (filters.is_active === 'all') url.searchParams.delete('is_active');
  else
    url.searchParams.set(
      'is_active',
      filters.is_active === 'active' ? 'true' : 'false',
    );

  if (panelMode === 'create') {
    url.searchParams.set('panel', 'create');
    url.searchParams.delete('id');
  } else if (panelMode === 'edit') {
    url.searchParams.set('panel', 'edit');
    if (selectedPetId) {
      url.searchParams.set('id', selectedPetId);
    } else {
      url.searchParams.delete('id');
    }
  } else {
    url.searchParams.delete('panel');
    if (selectedPetId) {
      url.searchParams.set('id', selectedPetId);
    } else {
      url.searchParams.delete('id');
    }
  }
  url.searchParams.delete('pet');

  window.history.replaceState(
    {},
    '',
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function isPetSize(value: string | null): value is PetSize {
  return (
    value === 'small' ||
    value === 'medium' ||
    value === 'large' ||
    value === 'giant'
  );
}

function isPetKind(value: string | null): value is PetKind {
  return (
    value === 'dog' ||
    value === 'cat' ||
    value === 'bird' ||
    value === 'fish' ||
    value === 'reptile' ||
    value === 'rodent' ||
    value === 'rabbit' ||
    value === 'other'
  );
}

function isPetTemperament(value: string | null): value is PetTemperament {
  return (
    value === 'calm' ||
    value === 'nervous' ||
    value === 'aggressive' ||
    value === 'playful' ||
    value === 'loving'
  );
}

function resolvePetAgeLabel(birthDate?: string | null): string {
  if (!birthDate) {
    return 'Não informada';
  }

  const parts = birthDate.split('-').map((part) => Number(part));
  if (parts.length !== 3 || parts.some((part) => Number.isNaN(part))) {
    return 'Não informada';
  }

  const [year, month, day] = parts;
  const today = new Date();
  const nowYear = today.getFullYear();
  const nowMonth = today.getMonth() + 1;
  const nowDay = today.getDate();

  let years = nowYear - year;
  let months = nowMonth - month;
  let days = nowDay - day;

  if (days < 0) {
    months -= 1;
    const previousMonthLastDay = new Date(nowYear, nowMonth - 1, 0).getDate();
    days += previousMonthLastDay;
  }

  if (months < 0) {
    years -= 1;
    months += 12;
  }

  if (years < 0) {
    return 'Não informada';
  }

  if (years > 0) {
    return years === 1 ? '1 ano' : `${years} anos`;
  }

  if (months > 0) {
    return months === 1 ? '1 mês' : `${months} meses`;
  }

  return days === 1 ? '1 dia' : `${days} dias`;
}

function resolvePetKindLabel(kind: PetKind) {
  switch (kind) {
    case 'dog':
      return 'Cachorro';
    case 'cat':
      return 'Gato';
    case 'bird':
      return 'Ave';
    case 'fish':
      return 'Peixe';
    case 'reptile':
      return 'Réptil';
    case 'rodent':
      return 'Roedor';
    case 'rabbit':
      return 'Coelho';
    case 'other':
      return 'Outro';
    default:
      return kind;
  }
}

function resolvePetSizeLabel(size: PetSize) {
  switch (size) {
    case 'small':
      return 'Pequeno';
    case 'medium':
      return 'Médio';
    case 'large':
      return 'Grande';
    case 'giant':
      return 'Gigante';
    default:
      return size;
  }
}

function resolvePetTemperamentLabel(temperament: PetTemperament) {
  switch (temperament) {
    case 'calm':
      return 'Calmo';
    case 'nervous':
      return 'Nervoso';
    case 'aggressive':
      return 'Agressivo';
    case 'playful':
      return 'Brincalhão';
    case 'loving':
      return 'Carinhoso';
    default:
      return temperament;
  }
}

function boolToLabel(value: boolean) {
  return value ? 'Sim' : 'Não';
}

const fieldClassName =
  'w-full rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm text-stone-900 outline-none transition placeholder:text-stone-400 focus:border-sky-300 focus:bg-white';
