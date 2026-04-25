import {
  type FormEvent,
  type MouseEvent,
  useEffect,
  useMemo,
  useState,
} from 'react';
import type {
  CreateServiceAverageTimeInput,
  CreateServiceInput,
  CreateServiceSubServiceInput,
  PetKind,
  PetSize,
  PetTemperament,
  ServiceDTO,
} from '@petcontrol/shared-types';

import {
  useCreateServiceMutation,
  useDeleteServiceMutation,
  useServiceQuery,
  useServicesQuery,
  useUpdateServiceMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';
import { useListParams } from '@/hooks/use-list-params';
import { SearchBar } from '@/ui/search-bar';
import { PaginationBar } from '@/ui/pagination-bar';
import { selectSession, useAuthStore } from '@/lib/auth/auth.store';

type ServiceFormState = CreateServiceInput;
type ServicePanelMode = 'create' | 'detail' | 'edit';
type ServiceStatusFilter = 'all' | 'active' | 'inactive';

type ServiceFilterState = {
  type_name: string;
  is_active: ServiceStatusFilter;
  min_price: string;
  max_price: string;
};

const petSizeOptions: Array<{ value: PetSize; label: string }> = [
  { value: 'small', label: 'Pequeno' },
  { value: 'medium', label: 'Médio' },
  { value: 'large', label: 'Grande' },
  { value: 'giant', label: 'Gigante' },
];

const petKindOptions: Array<{ value: PetKind; label: string }> = [
  { value: 'dog', label: 'Cão' },
  { value: 'cat', label: 'Gato' },
  { value: 'bird', label: 'Ave' },
  { value: 'other', label: 'Outro' },
];

const temperamentOptions: Array<{ value: PetTemperament; label: string }> = [
  { value: 'calm', label: 'Calmo' },
  { value: 'playful', label: 'Brincalhão' },
  { value: 'nervous', label: 'Nervoso' },
  { value: 'aggressive', label: 'Agressivo' },
  { value: 'loving', label: 'Carinhoso' },
];

const initialAverageTime: CreateServiceAverageTimeInput = {
  pet_size: 'medium',
  pet_kind: 'dog',
  pet_temperament: 'playful',
  average_time_minutes: 60,
};

function createAverageTime(): CreateServiceAverageTimeInput {
  return { ...initialAverageTime };
}

function createSubService(): CreateServiceSubServiceInput {
  return {
    type_name: '',
    title: '',
    description: '',
    notes: '',
    price: '',
    discount_rate: '0.00',
    image_url: '',
    is_active: true,
    average_times: [createAverageTime()],
  };
}

const initialServiceForm: ServiceFormState = {
  type_name: 'Banho',
  title: '',
  description: '',
  notes: '',
  price: '',
  discount_rate: '0.00',
  image_url: '',
  is_active: true,
  sub_services: [createSubService()],
};

const initialServiceFilters: ServiceFilterState = {
  type_name: '',
  is_active: 'all',
  min_price: '',
  max_price: '',
};

export function ServicesPage() {
  const session = useAuthStore(selectSession);
  const canManageServices = session?.role === 'admin';
  const { page, params, search, setSearch, goToPage } = useListParams(
    undefined,
    readSearchFromLocation(),
    readPageFromLocation(),
  );
  const createMutation = useCreateServiceMutation();
  const updateMutation = useUpdateServiceMutation();
  const deleteMutation = useDeleteServiceMutation();
  const [editingServiceId, setEditingServiceId] = useState<string | null>(null);
  const [selectedServiceId, setSelectedServiceId] = useState<string | null>(
    readSelectedServiceIdFromLocation(),
  );
  const [panelMode, setPanelMode] = useState<ServicePanelMode>('create');
  const [filters, setFilters] = useState<ServiceFilterState>(() =>
    readServiceFiltersFromLocation(),
  );
  const [form, setForm] = useState<ServiceFormState>(initialServiceForm);

  useEffect(() => {
    setPanelMode(readPanelModeFromLocation(canManageServices));
  }, [canManageServices]);

  const servicesQueryParams = useMemo(
    () => ({
      ...params,
      ...(filters.type_name ? { type_name: filters.type_name } : {}),
      ...(filters.is_active !== 'all'
        ? { is_active: filters.is_active === 'active' ? 'true' : 'false' }
        : {}),
      ...(filters.min_price ? { min_price: filters.min_price } : {}),
      ...(filters.max_price ? { max_price: filters.max_price } : {}),
    }),
    [filters, params],
  );

  const servicesQuery = useServicesQuery(servicesQueryParams);

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;
  const isPending =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  function resetForm() {
    if (!canManageServices) {
      return;
    }
    setEditingServiceId(null);
    setSelectedServiceId(null);
    setPanelMode('create');
    setForm(cloneInitialForm());
  }

  useEffect(() => {
    const openCreateForm = () => {
      if (!canManageServices) {
        return;
      }
      setEditingServiceId(null);
      setSelectedServiceId(null);
      setPanelMode('create');
      setForm(cloneInitialForm());
    };

    window.addEventListener('open-services-create-form', openCreateForm);
    return () => {
      window.removeEventListener('open-services-create-form', openCreateForm);
    };
  }, [canManageServices]);

  useEffect(() => {
    syncServiceLocation({
      page,
      search,
      selectedServiceId,
      panelMode,
      filters,
    });
  }, [filters, page, panelMode, search, selectedServiceId]);

  function startEdit(service: ServiceDTO) {
    if (!canManageServices) {
      return;
    }
    setEditingServiceId(service.id);
    setSelectedServiceId(service.id);
    setPanelMode('edit');
    setForm({
      type_name: service.type_name,
      title: service.title,
      description: service.description,
      notes: service.notes ?? '',
      price: service.price,
      discount_rate: service.discount_rate,
      image_url: service.image_url ?? '',
      is_active: service.is_active,
      sub_services:
        service.sub_services && service.sub_services.length > 0
          ? service.sub_services.map((subService) => ({
              type_name: service.type_name,
              title: subService.title,
              description: subService.description,
              notes: subService.notes ?? '',
              price: subService.price,
              discount_rate: subService.discount_rate,
              image_url: subService.image_url ?? '',
              is_active: subService.is_active,
              average_times:
                subService.average_times.length > 0
                  ? subService.average_times.map((averageTime) => ({
                      pet_size: averageTime.pet_size,
                      pet_kind: averageTime.pet_kind,
                      pet_temperament: averageTime.pet_temperament,
                      average_time_minutes: averageTime.average_time_minutes,
                    }))
                  : [createAverageTime()],
            }))
          : [
              {
                ...createSubService(),
                type_name: service.type_name,
                title: service.title,
                description: service.description,
                price: service.price,
              },
            ],
    });
  }

  function selectService(service: ServiceDTO) {
    setSelectedServiceId(service.id);
    setEditingServiceId(null);
    setPanelMode('detail');
  }

  function updateFilter<K extends keyof ServiceFilterState>(
    key: K,
    value: ServiceFilterState[K],
  ) {
    setFilters((current) => ({ ...current, [key]: value }));
  }

  function addSubService() {
    setForm((current) => ({
      ...current,
      sub_services: [...current.sub_services, createSubService()],
    }));
  }

  function removeSubService(index: number) {
    setForm((current) => ({
      ...current,
      sub_services:
        current.sub_services.length === 1
          ? current.sub_services
          : current.sub_services.filter((_, itemIndex) => itemIndex !== index),
    }));
  }

  function updateSubService(
    index: number,
    patch: Partial<CreateServiceSubServiceInput>,
  ) {
    setForm((current) => ({
      ...current,
      sub_services: current.sub_services.map((item, itemIndex) =>
        itemIndex === index ? { ...item, ...patch } : item,
      ),
    }));
  }

  function addAverageTime(subServiceIndex: number) {
    setForm((current) => ({
      ...current,
      sub_services: current.sub_services.map((subService, itemIndex) =>
        itemIndex === subServiceIndex
          ? {
              ...subService,
              average_times: [...subService.average_times, createAverageTime()],
            }
          : subService,
      ),
    }));
  }

  function removeAverageTime(
    subServiceIndex: number,
    averageTimeIndex: number,
  ) {
    setForm((current) => ({
      ...current,
      sub_services: current.sub_services.map((subService, itemIndex) =>
        itemIndex === subServiceIndex && subService.average_times.length > 1
          ? {
              ...subService,
              average_times: subService.average_times.filter(
                (_, timeIndex) => timeIndex !== averageTimeIndex,
              ),
            }
          : subService,
      ),
    }));
  }

  function updateAverageTime(
    subServiceIndex: number,
    averageTimeIndex: number,
    patch: Partial<CreateServiceAverageTimeInput>,
  ) {
    setForm((current) => ({
      ...current,
      sub_services: current.sub_services.map((subService, itemIndex) =>
        itemIndex === subServiceIndex
          ? {
              ...subService,
              average_times: subService.average_times.map(
                (averageTime, timeIndex) =>
                  timeIndex === averageTimeIndex
                    ? { ...averageTime, ...patch }
                    : averageTime,
              ),
            }
          : subService,
      ),
    }));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canManageServices) {
      return;
    }
    const payload = normalizePayload(form);

    if (editingServiceId) {
      await updateMutation.mutateAsync({
        serviceId: editingServiceId,
        input: payload,
      });
      setSelectedServiceId(editingServiceId);
      setEditingServiceId(null);
      setPanelMode('detail');
      return;
    }

    const created = await createMutation.mutateAsync(payload);
    setSelectedServiceId(created.id);
    setEditingServiceId(null);
    setPanelMode('detail');
  }

  const services = useMemo(
    () => servicesQuery.data?.data ?? [],
    [servicesQuery.data?.data],
  );
  const typeOptions = useMemo(
    () =>
      Array.from(
        new Set(
          [filters.type_name, ...services.map((service) => service.type_name)]
            .map((value) => value.trim())
            .filter(Boolean),
        ),
      ).sort((a, b) => a.localeCompare(b)),
    [filters.type_name, services],
  );

  const activeSelectedServiceId =
    panelMode === 'create'
      ? null
      : selectedServiceId &&
          services.some((service) => service.id === selectedServiceId)
        ? selectedServiceId
        : (services[0]?.id ?? null);
  const serviceDetailQuery = useServiceQuery(
    activeSelectedServiceId ?? undefined,
  );
  const selectedServiceSummary =
    services.find((service) => service.id === selectedServiceId) ??
    services[0] ??
    null;
  const selectedService =
    serviceDetailQuery.data?.data?.id === activeSelectedServiceId
      ? serviceDetailQuery.data.data
      : selectedServiceSummary;
  const showDetail = panelMode === 'detail' && selectedService !== null;

  useEffect(() => {
    if (
      panelMode !== 'create' &&
      activeSelectedServiceId !== selectedServiceId
    ) {
      setSelectedServiceId(activeSelectedServiceId);
    }
  }, [activeSelectedServiceId, panelMode, selectedServiceId]);

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px] h-full p-6">
      <section className="app-panel p-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="app-eyebrow">Catálogo</p>
            <h2 className="mt-2 font-display text-3xl text-foreground">
              Serviços
            </h2>
            <p className="mt-2 max-w-2xl text-sm text-muted">
              Serviços, subserviços e tempos médios usados pelos agendamentos do
              tenant.
            </p>
          </div>
          <button
            type="button"
            onClick={resetForm}
            disabled={!canManageServices}
            className="rounded-2xl border border-border/50 bg-surface/50 px-4 py-2 text-sm text-foreground transition hover:bg-surface/50 disabled:cursor-not-allowed disabled:opacity-40"
          >
            Novo
          </button>
        </div>

        <div className="mt-4">
          <SearchBar
            value={search}
            onChange={setSearch}
            placeholder="Buscar por título ou descrição..."
            id="services-search"
          />
        </div>

        <div className="mt-4 grid gap-3 md:grid-cols-4">
          <FilterField label="Tipo" htmlFor="services-type-filter">
            <select
              title="Tipo"
              id="services-type-filter"
              className={fieldClassName}
              value={filters.type_name}
              onChange={(event) =>
                updateFilter('type_name', event.target.value)
              }
            >
              <option value="">Todos</option>
              {typeOptions.map((typeName) => (
                <option key={typeName} value={typeName}>
                  {typeName}
                </option>
              ))}
            </select>
          </FilterField>
          <FilterField label="Status" htmlFor="services-status-filter">
            <select
              title="Status"
              id="services-status-filter"
              className={fieldClassName}
              value={filters.is_active}
              onChange={(event) =>
                updateFilter(
                  'is_active',
                  event.target.value as ServiceStatusFilter,
                )
              }
            >
              <option value="all">Todos</option>
              <option value="active">Ativos</option>
              <option value="inactive">Inativos</option>
            </select>
          </FilterField>
          <FilterField label="Preço mín." htmlFor="services-min-price-filter">
            <input
              id="services-min-price-filter"
              className={fieldClassName}
              inputMode="decimal"
              placeholder="0.00"
              value={filters.min_price}
              onChange={(event) =>
                updateFilter('min_price', event.target.value)
              }
            />
          </FilterField>
          <FilterField label="Preço máx." htmlFor="services-max-price-filter">
            <input
              id="services-max-price-filter"
              className={fieldClassName}
              inputMode="decimal"
              placeholder="0.00"
              value={filters.max_price}
              onChange={(event) =>
                updateFilter('max_price', event.target.value)
              }
            />
          </FilterField>
        </div>

        <div className="mt-4 space-y-3">
          {servicesQuery.isLoading ? (
            <StateMessage message="Carregando serviços..." />
          ) : null}
          {servicesQuery.isError ? (
            <StateMessage message="Falha ao carregar serviços." tone="error" />
          ) : null}
          {!servicesQuery.isLoading && services.length === 0 ? (
            <StateMessage message="Nenhum serviço cadastrado." />
          ) : null}
          {services.map((service) => (
            <article
              key={service.id}
              className={`rounded-3xl border p-4 transition ${selectedServiceId === service.id ? 'border-primary/40 bg-primary/10' : 'border-border/50 bg-surface/50 hover:bg-surface/50'}`}
            >
              <div
                className="flex cursor-pointer flex-wrap items-start justify-between gap-3"
                role="button"
                tabIndex={0}
                onClick={() => selectService(service)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault();
                    selectService(service);
                  }
                }}
              >
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-display text-xl text-foreground">
                      {service.title}
                    </h3>
                    <span className="rounded-full border border-emerald-400/30 bg-emerald-500/10 px-2 py-1 text-xs text-emerald-100">
                      {service.is_active ? 'Ativo' : 'Inativo'}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-muted">
                    {service.type_name} · R$ {service.price}
                  </p>
                  <p className="mt-1 text-xs text-muted">
                    {service.sub_services_count ?? 0} subserviço(s) ·{' '}
                    {service.average_times_count ?? 0} tempo(s) médio(s)
                  </p>
                </div>
                {canManageServices ? (
                  <Actions
                    onEdit={(event) => {
                      event.stopPropagation();
                      startEdit(service);
                    }}
                    onDelete={(event) => {
                      event.stopPropagation();
                      void deleteMutation.mutateAsync(service.id);
                    }}
                  />
                ) : null}
              </div>
            </article>
          ))}
        </div>

        <PaginationBar
          meta={servicesQuery.data?.meta}
          onPageChange={goToPage}
        />
      </section>

      <aside className="app-panel p-6">
        {showDetail ? (
          <ServiceDetailPanel
            service={selectedService}
            canManage={canManageServices}
            onCreate={resetForm}
            onEdit={() => startEdit(selectedService)}
            onDelete={() => void deleteMutation.mutateAsync(selectedService.id)}
          />
        ) : !canManageServices ? (
          <StateMessage message="Selecione um serviço para visualizar os detalhes." />
        ) : (
          <>
            <p className="text-xs uppercase tracking-[0.3em] text-muted">
              {editingServiceId ? 'Editar serviço' : 'Novo serviço'}
            </p>
            <h3 className="mt-2 font-display text-2xl text-foreground">
              {editingServiceId ? 'Atualizar catálogo' : 'Criar catálogo'}
            </h3>

            <form
              className="mt-6 space-y-4"
              onSubmit={(event) => void submit(event)}
            >
              <Field label="Tipo" htmlFor="service-type">
                <input
                  title="Tipo"
                  id="service-type"
                  required
                  className={fieldClassName}
                  value={form.type_name}
                  onChange={(event) =>
                    setForm({ ...form, type_name: event.target.value })
                  }
                />
              </Field>
              <Field label="Título" htmlFor="service-title">
                <input
                  title="Título"
                  id="service-title"
                  required
                  className={fieldClassName}
                  value={form.title}
                  onChange={(event) =>
                    setForm({ ...form, title: event.target.value })
                  }
                />
              </Field>
              <Field label="Descrição" htmlFor="service-description">
                <textarea
                  title="Descrição"
                  id="service-description"
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
                <Field label="Preço base" htmlFor="service-price">
                  <input
                    id="service-price"
                    required
                    className={fieldClassName}
                    placeholder="0.00"
                    value={form.price}
                    onChange={(event) =>
                      setForm({ ...form, price: event.target.value })
                    }
                  />
                </Field>
                <Field label="Desconto (%)" htmlFor="service-discount">
                  <input
                    id="service-discount"
                    className={fieldClassName}
                    placeholder="0.00"
                    value={form.discount_rate}
                    onChange={(event) =>
                      setForm({ ...form, discount_rate: event.target.value })
                    }
                  />
                </Field>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between gap-3">
                  <p className="text-sm font-semibold text-foreground">
                    Subserviços
                  </p>
                  <button
                    type="button"
                    onClick={addSubService}
                    className="rounded-xl border border-border/50 bg-surface/50 px-3 py-1 text-xs text-foreground transition hover:bg-surface/50"
                  >
                    Adicionar
                  </button>
                </div>

                {form.sub_services.map((subService, subServiceIndex) => (
                  <div
                    key={subServiceIndex}
                    className="rounded-3xl border border-border/50 bg-surface/50 p-4"
                  >
                    <div className="flex items-center justify-between gap-3">
                      <p className="text-sm font-semibold text-foreground">
                        Subserviço {subServiceIndex + 1}
                      </p>
                      <button
                        type="button"
                        onClick={() => removeSubService(subServiceIndex)}
                        disabled={form.sub_services.length === 1}
                        className="rounded-xl border border-rose-400/40 bg-rose-500/10 px-3 py-1 text-xs text-rose-200 transition hover:bg-rose-500/20 disabled:cursor-not-allowed disabled:opacity-40"
                      >
                        Remover
                      </button>
                    </div>
                    <div className="mt-4 space-y-4">
                      <Field
                        label="Título"
                        htmlFor={`sub-service-title-${subServiceIndex}`}
                      >
                        <input
                          id={`sub-service-title-${subServiceIndex}`}
                          required
                          className={fieldClassName}
                          value={subService.title}
                          onChange={(event) =>
                            updateSubService(subServiceIndex, {
                              title: event.target.value,
                            })
                          }
                        />
                      </Field>
                      <Field
                        label="Descrição"
                        htmlFor={`sub-service-description-${subServiceIndex}`}
                      >
                        <textarea
                          id={`sub-service-description-${subServiceIndex}`}
                          required
                          className={fieldClassName}
                          rows={2}
                          value={subService.description}
                          onChange={(event) =>
                            updateSubService(subServiceIndex, {
                              description: event.target.value,
                            })
                          }
                        />
                      </Field>
                      <Field
                        label="Preço"
                        htmlFor={`sub-service-price-${subServiceIndex}`}
                      >
                        <input
                          id={`sub-service-price-${subServiceIndex}`}
                          required
                          className={fieldClassName}
                          placeholder="0.00"
                          value={subService.price}
                          onChange={(event) =>
                            updateSubService(subServiceIndex, {
                              price: event.target.value,
                            })
                          }
                        />
                      </Field>

                      <div className="space-y-3">
                        <div className="flex items-center justify-between gap-3">
                          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted">
                            Tempos médios
                          </p>
                          <button
                            type="button"
                            onClick={() => addAverageTime(subServiceIndex)}
                            className="rounded-xl border border-border/50 bg-surface/50 px-3 py-1 text-xs text-foreground transition hover:bg-surface/50"
                          >
                            Adicionar
                          </button>
                        </div>

                        {subService.average_times.map(
                          (averageTime, averageTimeIndex) => (
                            <div
                              key={averageTimeIndex}
                              className="rounded-2xl border border-border/50 bg-surface/50 p-3"
                            >
                              <div className="flex items-center justify-between gap-3">
                                <p className="text-sm text-foreground">
                                  Tempo {averageTimeIndex + 1}
                                </p>
                                <button
                                  type="button"
                                  onClick={() =>
                                    removeAverageTime(
                                      subServiceIndex,
                                      averageTimeIndex,
                                    )
                                  }
                                  disabled={
                                    subService.average_times.length === 1
                                  }
                                  className="rounded-xl border border-rose-400/40 bg-rose-500/10 px-3 py-1 text-xs text-rose-200 transition hover:bg-rose-500/20 disabled:cursor-not-allowed disabled:opacity-40"
                                >
                                  Remover
                                </button>
                              </div>
                              <div className="mt-3 grid gap-4 sm:grid-cols-2">
                                <Field
                                  label="Porte"
                                  htmlFor={`average-time-size-${subServiceIndex}-${averageTimeIndex}`}
                                >
                                  <Select
                                    id={`average-time-size-${subServiceIndex}-${averageTimeIndex}`}
                                    value={averageTime.pet_size}
                                    options={petSizeOptions}
                                    onChange={(value) =>
                                      updateAverageTime(
                                        subServiceIndex,
                                        averageTimeIndex,
                                        { pet_size: value as PetSize },
                                      )
                                    }
                                  />
                                </Field>
                                <Field
                                  label="Espécie"
                                  htmlFor={`average-time-kind-${subServiceIndex}-${averageTimeIndex}`}
                                >
                                  <Select
                                    id={`average-time-kind-${subServiceIndex}-${averageTimeIndex}`}
                                    value={averageTime.pet_kind}
                                    options={petKindOptions}
                                    onChange={(value) =>
                                      updateAverageTime(
                                        subServiceIndex,
                                        averageTimeIndex,
                                        { pet_kind: value as PetKind },
                                      )
                                    }
                                  />
                                </Field>
                                <Field
                                  label="Temperamento"
                                  htmlFor={`average-time-temperament-${subServiceIndex}-${averageTimeIndex}`}
                                >
                                  <Select
                                    id={`average-time-temperament-${subServiceIndex}-${averageTimeIndex}`}
                                    value={averageTime.pet_temperament}
                                    options={temperamentOptions}
                                    onChange={(value) =>
                                      updateAverageTime(
                                        subServiceIndex,
                                        averageTimeIndex,
                                        {
                                          pet_temperament:
                                            value as PetTemperament,
                                        },
                                      )
                                    }
                                  />
                                </Field>
                                <Field
                                  label="Minutos"
                                  htmlFor={`average-time-minutes-${subServiceIndex}-${averageTimeIndex}`}
                                >
                                  <input
                                    id={`average-time-minutes-${subServiceIndex}-${averageTimeIndex}`}
                                    required
                                    min={1}
                                    type="number"
                                    className={fieldClassName}
                                    value={averageTime.average_time_minutes}
                                    onChange={(event) =>
                                      updateAverageTime(
                                        subServiceIndex,
                                        averageTimeIndex,
                                        {
                                          average_time_minutes: Number(
                                            event.target.value,
                                          ),
                                        },
                                      )
                                    }
                                  />
                                </Field>
                              </div>
                            </div>
                          ),
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              <Field label="Notas" htmlFor="service-notes">
                <textarea
                  title="Notas"
                  id="service-notes"
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
          </>
        )}
      </aside>
    </div>
  );
}

const fieldClassName =
  'w-full rounded-2xl border border-border/50 bg-surface/50 px-3 py-2 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-primary/50 focus:ring-2 focus:ring-primary/20';

function cloneInitialForm(): ServiceFormState {
  return {
    ...initialServiceForm,
    sub_services: [createSubService()],
  };
}

function normalizePayload(form: ServiceFormState): CreateServiceInput {
  return {
    ...form,
    notes: form.notes || undefined,
    image_url: form.image_url || undefined,
    sub_services: form.sub_services.map((subService) => ({
      ...subService,
      type_name: subService.type_name || form.type_name,
      notes: subService.notes || undefined,
      image_url: subService.image_url || undefined,
      discount_rate: subService.discount_rate || '0.00',
      average_times: subService.average_times.map((averageTime) => ({
        ...averageTime,
        average_time_minutes: Number(averageTime.average_time_minutes),
      })),
    })),
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

function readPanelModeFromLocation(
  canManageServices: boolean,
): ServicePanelMode {
  if (!canManageServices || typeof window === 'undefined') {
    return 'detail';
  }

  const panel = new URLSearchParams(window.location.search).get('panel');
  if (panel === 'edit') {
    return 'edit';
  }
  if (panel === 'detail') {
    return 'detail';
  }
  return 'create';
}

function readSelectedServiceIdFromLocation(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const params = new URLSearchParams(window.location.search);
  return params.get('id') ?? params.get('service');
}

function readServiceFiltersFromLocation(): ServiceFilterState {
  if (typeof window === 'undefined') {
    return initialServiceFilters;
  }

  const params = new URLSearchParams(window.location.search);
  const isActive = params.get('is_active');
  return {
    type_name: params.get('type_name') ?? '',
    is_active:
      isActive === 'true'
        ? 'active'
        : isActive === 'false'
          ? 'inactive'
          : 'all',
    min_price: params.get('min_price') ?? '',
    max_price: params.get('max_price') ?? '',
  };
}

function syncServiceLocation({
  page,
  search,
  selectedServiceId,
  panelMode,
  filters,
}: {
  page: number;
  search: string;
  selectedServiceId: string | null;
  panelMode: ServicePanelMode;
  filters: ServiceFilterState;
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

  const trimmedType = filters.type_name.trim();
  if (trimmedType === '') url.searchParams.delete('type_name');
  else url.searchParams.set('type_name', trimmedType);

  if (filters.is_active === 'all') url.searchParams.delete('is_active');
  else
    url.searchParams.set(
      'is_active',
      filters.is_active === 'active' ? 'true' : 'false',
    );

  const trimmedMinPrice = filters.min_price.trim();
  if (trimmedMinPrice === '') url.searchParams.delete('min_price');
  else url.searchParams.set('min_price', trimmedMinPrice);

  const trimmedMaxPrice = filters.max_price.trim();
  if (trimmedMaxPrice === '') url.searchParams.delete('max_price');
  else url.searchParams.set('max_price', trimmedMaxPrice);

  if (panelMode === 'create') {
    url.searchParams.set('panel', 'create');
    url.searchParams.delete('id');
  } else if (panelMode === 'edit') {
    url.searchParams.set('panel', 'edit');
    if (selectedServiceId) {
      url.searchParams.set('id', selectedServiceId);
    } else {
      url.searchParams.delete('id');
    }
  } else {
    url.searchParams.delete('panel');
    if (selectedServiceId) {
      url.searchParams.set('id', selectedServiceId);
    } else {
      url.searchParams.delete('id');
    }
  }
  url.searchParams.delete('service');

  window.history.replaceState(
    {},
    '',
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function FilterField({
  label,
  children,
  htmlFor,
}: {
  label: string;
  children: React.ReactNode;
  htmlFor: string;
}) {
  return (
    <label className="block space-y-2" htmlFor={htmlFor}>
      <span className="text-xs font-medium uppercase tracking-[0.16em] text-muted">
        {label}
      </span>
      {children}
    </label>
  );
}

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

function Select({
  id,
  value,
  options,
  onChange,
}: {
  id: string;
  value: string;
  options: Array<{ value: string; label: string }>;
  onChange: (value: string) => void;
}) {
  return (
    <select
      id={id}
      className={fieldClassName}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}

function ServiceDetailPanel({
  service,
  canManage,
  onCreate,
  onEdit,
  onDelete,
}: {
  service: ServiceDTO;
  canManage: boolean;
  onCreate: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const subServices = service.sub_services ?? [];

  return (
    <div>
      <p className="app-eyebrow">Detalhe</p>
      <div className="mt-2 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h3 className="font-display text-2xl text-foreground">
            {service.title}
          </h3>
          <p className="mt-1 text-sm text-muted">{service.type_name}</p>
        </div>
        <span
          className={`rounded-full border px-3 py-1 text-xs ${
            service.is_active
              ? 'border-emerald-400/30 bg-emerald-500/10 text-emerald-100'
              : 'border-slate-400/30 bg-slate-500/10 text-foreground'
          }`}
        >
          {service.is_active ? 'Ativo' : 'Inativo'}
        </span>
      </div>

      <p className="mt-4 text-sm leading-6 text-muted">{service.description}</p>

      <div className="mt-5 grid gap-3 sm:grid-cols-2">
        <DetailMetric label="Preço base" value={`R$ ${service.price}`} />
        <DetailMetric label="Desconto" value={`${service.discount_rate}%`} />
        <DetailMetric
          label="Subserviços"
          value={String(service.sub_services_count ?? subServices.length)}
        />
        <DetailMetric
          label="Tempos"
          value={String(
            service.average_times_count ??
              subServices.reduce(
                (total, item) => total + item.average_times.length,
                0,
              ),
          )}
        />
      </div>

      {service.notes ? (
        <div className="mt-5 rounded-2xl border border-border/50 bg-surface/50 p-4">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted">
            Notas
          </p>
          <p className="mt-2 text-sm text-muted">{service.notes}</p>
        </div>
      ) : null}

      <div className="mt-6 space-y-4">
        <div className="flex items-center justify-between gap-3">
          <p className="text-sm font-semibold text-foreground">Subserviços</p>
          <span className="text-xs text-muted">{subServices.length}</span>
        </div>

        {subServices.length === 0 ? (
          <StateMessage message="Nenhum subserviço detalhado na resposta." />
        ) : null}

        {subServices.map((subService) => (
          <div
            key={subService.id}
            className="rounded-3xl border border-border/50 bg-surface/50 p-4"
          >
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-display text-lg text-foreground">
                  {subService.title}
                </p>
                <p className="mt-1 text-sm text-muted">R$ {subService.price}</p>
              </div>
              <span className="rounded-full border border-border/50 px-2 py-1 text-xs text-muted">
                {subService.average_times.length} tempo(s)
              </span>
            </div>
            <p className="mt-2 text-sm text-muted">{subService.description}</p>

            {subService.average_times.length > 0 ? (
              <div className="mt-3 space-y-2">
                {subService.average_times.map((averageTime) => (
                  <div
                    key={averageTime.id}
                    className="rounded-2xl border border-border/50 bg-surface/50 px-3 py-2 text-xs text-muted"
                  >
                    {formatOptionLabel(petSizeOptions, averageTime.pet_size)} ·{' '}
                    {formatOptionLabel(petKindOptions, averageTime.pet_kind)} ·{' '}
                    {formatOptionLabel(
                      temperamentOptions,
                      averageTime.pet_temperament,
                    )}{' '}
                    · {averageTime.average_time_minutes} min
                  </div>
                ))}
              </div>
            ) : null}
          </div>
        ))}
      </div>

      {canManage ? (
        <div className="mt-6 flex flex-wrap gap-3">
          <button
            type="button"
            onClick={onEdit}
            className="rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-slate-950 transition hover:brightness-110"
          >
            Editar
          </button>
          <button
            type="button"
            onClick={onCreate}
            className="rounded-2xl border border-border/50 bg-surface/50 px-4 py-2 text-sm text-foreground transition hover:bg-surface/50"
          >
            Novo
          </button>
          <button
            type="button"
            onClick={onDelete}
            className="rounded-2xl border border-rose-400/40 bg-rose-500/10 px-4 py-2 text-sm text-rose-200 transition hover:bg-rose-500/20"
          >
            Excluir
          </button>
        </div>
      ) : null}
    </div>
  );
}

function DetailMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-border/50 bg-surface/50 p-3">
      <p className="text-xs uppercase tracking-[0.16em] text-muted">{label}</p>
      <p className="mt-1 text-sm font-semibold text-foreground">{value}</p>
    </div>
  );
}

function formatOptionLabel<T extends string>(
  options: Array<{ value: T; label: string }>,
  value: T,
) {
  return options.find((option) => option.value === value)?.label ?? value;
}

function Actions({
  onEdit,
  onDelete,
}: {
  onEdit: (event: MouseEvent<HTMLButtonElement>) => void;
  onDelete: (event: MouseEvent<HTMLButtonElement>) => void;
}) {
  return (
    <div className="flex gap-2">
      <button
        type="button"
        onClick={onEdit}
        className="rounded-xl border border-border/50 bg-surface/50 px-3 py-1 text-xs text-foreground transition hover:bg-surface/50"
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
      className={`rounded-2xl border px-4 py-3 text-sm ${tone === 'error' ? 'border-rose-400/30 bg-rose-500/10 text-rose-100' : 'border-border/50 bg-surface/50 text-muted'}`}
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
        className="rounded-2xl border border-border/50 bg-surface/50 px-4 py-2 text-sm text-foreground transition hover:bg-surface/50"
      >
        Limpar
      </button>
    </div>
  );
}
