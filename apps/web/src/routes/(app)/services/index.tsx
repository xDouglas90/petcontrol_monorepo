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
  useServicesQuery,
  useUpdateServiceMutation,
} from '@/lib/api/domain.queries';
import { ApiError } from '@/lib/api/rest-client';
import { useListParams } from '@/hooks/use-list-params';
import { SearchBar } from '@/ui/search-bar';
import { PaginationBar } from '@/ui/pagination-bar';

type ServiceFormState = CreateServiceInput;
type ServicePanelMode = 'create' | 'detail' | 'edit';

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

export function ServicesPage() {
  const { params, search, setSearch, goToPage } = useListParams();
  const servicesQuery = useServicesQuery(params);
  const createMutation = useCreateServiceMutation();
  const updateMutation = useUpdateServiceMutation();
  const deleteMutation = useDeleteServiceMutation();
  const [editingServiceId, setEditingServiceId] = useState<string | null>(null);
  const [selectedServiceId, setSelectedServiceId] = useState<string | null>(
    null,
  );
  const [panelMode, setPanelMode] = useState<ServicePanelMode>('create');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>(
    'all',
  );
  const [minPriceFilter, setMinPriceFilter] = useState('');
  const [maxPriceFilter, setMaxPriceFilter] = useState('');
  const [form, setForm] = useState<ServiceFormState>(initialServiceForm);

  const mutationError =
    createMutation.error || updateMutation.error || deleteMutation.error;
  const isPending =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  function resetForm() {
    setEditingServiceId(null);
    setSelectedServiceId(null);
    setPanelMode('create');
    setForm(cloneInitialForm());
  }

  useEffect(() => {
    const openCreateForm = () => {
      setEditingServiceId(null);
      setSelectedServiceId(null);
      setPanelMode('create');
      setForm(cloneInitialForm());
    };

    window.addEventListener('open-services-create-form', openCreateForm);
    return () => {
      window.removeEventListener('open-services-create-form', openCreateForm);
    };
  }, []);

  function startEdit(service: ServiceDTO) {
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
              average_times: [
                ...subService.average_times,
                createAverageTime(),
              ],
            }
          : subService,
      ),
    }));
  }

  function removeAverageTime(subServiceIndex: number, averageTimeIndex: number) {
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

    await createMutation.mutateAsync(payload);
    resetForm();
  }

  const services = servicesQuery.data?.data ?? [];
  const typeOptions = useMemo(
    () =>
      Array.from(new Set(services.map((service) => service.type_name))).sort(
        (a, b) => a.localeCompare(b),
      ),
    [services],
  );
  const visibleServices = useMemo(
    () =>
      services.filter((service) => {
        const price = Number(service.price);
        const minPrice = Number(minPriceFilter);
        const maxPrice = Number(maxPriceFilter);
        const matchesType = !typeFilter || service.type_name === typeFilter;
        const matchesStatus =
          statusFilter === 'all' ||
          (statusFilter === 'active' && service.is_active) ||
          (statusFilter === 'inactive' && !service.is_active);
        const matchesMin =
          !minPriceFilter || (Number.isFinite(price) && price >= minPrice);
        const matchesMax =
          !maxPriceFilter || (Number.isFinite(price) && price <= maxPrice);

        return matchesType && matchesStatus && matchesMin && matchesMax;
      }),
    [maxPriceFilter, minPriceFilter, services, statusFilter, typeFilter],
  );
  const selectedService =
    services.find((service) => service.id === selectedServiceId) ??
    visibleServices[0] ??
    null;
  const showDetail = panelMode === 'detail' && selectedService !== null;

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]">
      <section className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
              Catálogo
            </p>
            <h2 className="mt-2 font-display text-3xl text-white">Serviços</h2>
            <p className="mt-2 max-w-2xl text-sm text-slate-300">
              Serviços, subserviços e tempos médios usados pelos agendamentos do
              tenant.
            </p>
          </div>
          <button
            type="button"
            onClick={resetForm}
            className="rounded-2xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition hover:bg-white/10"
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
              id="services-type-filter"
              className={fieldClassName}
              value={typeFilter}
              onChange={(event) => setTypeFilter(event.target.value)}
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
              id="services-status-filter"
              className={fieldClassName}
              value={statusFilter}
              onChange={(event) =>
                setStatusFilter(
                  event.target.value as 'all' | 'active' | 'inactive',
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
              value={minPriceFilter}
              onChange={(event) => setMinPriceFilter(event.target.value)}
            />
          </FilterField>
          <FilterField label="Preço máx." htmlFor="services-max-price-filter">
            <input
              id="services-max-price-filter"
              className={fieldClassName}
              inputMode="decimal"
              placeholder="0.00"
              value={maxPriceFilter}
              onChange={(event) => setMaxPriceFilter(event.target.value)}
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
          {!servicesQuery.isLoading && visibleServices.length === 0 ? (
            <StateMessage message="Nenhum serviço cadastrado." />
          ) : null}
          {visibleServices.map((service) => (
            <article
              key={service.id}
              className={`rounded-3xl border p-4 transition ${selectedServiceId === service.id ? 'border-primary/40 bg-primary/10' : 'border-white/10 bg-white/5 hover:bg-white/10'}`}
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
                    <h3 className="font-display text-xl text-white">
                      {service.title}
                    </h3>
                    <span className="rounded-full border border-emerald-400/30 bg-emerald-500/10 px-2 py-1 text-xs text-emerald-100">
                      {service.is_active ? 'Ativo' : 'Inativo'}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-slate-300">
                    {service.type_name} · R$ {service.price}
                  </p>
                  <p className="mt-1 text-xs text-slate-500">
                    {service.sub_services_count ?? 0} subserviço(s) ·{' '}
                    {service.average_times_count ?? 0} tempo(s) médio(s)
                  </p>
                </div>
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
              </div>
            </article>
          ))}
        </div>

        <PaginationBar meta={servicesQuery.data?.meta} onPageChange={goToPage} />
      </section>

      <aside className="rounded-[1.75rem] border border-white/10 bg-slate-950/60 p-6">
        {showDetail ? (
          <ServiceDetailPanel
            service={selectedService}
            onCreate={resetForm}
            onEdit={() => startEdit(selectedService)}
            onDelete={() => void deleteMutation.mutateAsync(selectedService.id)}
          />
        ) : (
          <>
            <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
              {editingServiceId ? 'Editar serviço' : 'Novo serviço'}
            </p>
            <h3 className="mt-2 font-display text-2xl text-white">
              {editingServiceId ? 'Atualizar catálogo' : 'Criar catálogo'}
            </h3>

            <form
              className="mt-6 space-y-4"
              onSubmit={(event) => void submit(event)}
            >
          <Field label="Tipo" htmlFor="service-type">
            <input
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
              <p className="text-sm font-semibold text-white">Subserviços</p>
              <button
                type="button"
                onClick={addSubService}
                className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-slate-200 transition hover:bg-white/10"
              >
                Adicionar
              </button>
            </div>

            {form.sub_services.map((subService, subServiceIndex) => (
              <div
                key={subServiceIndex}
                className="rounded-3xl border border-white/10 bg-white/5 p-4"
              >
                <div className="flex items-center justify-between gap-3">
                  <p className="text-sm font-semibold text-white">
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
                      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">
                        Tempos médios
                      </p>
                      <button
                        type="button"
                        onClick={() => addAverageTime(subServiceIndex)}
                        className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-slate-200 transition hover:bg-white/10"
                      >
                        Adicionar
                      </button>
                    </div>

                    {subService.average_times.map(
                      (averageTime, averageTimeIndex) => (
                        <div
                          key={averageTimeIndex}
                          className="rounded-2xl border border-white/10 bg-slate-950/40 p-3"
                        >
                          <div className="flex items-center justify-between gap-3">
                            <p className="text-sm text-slate-200">
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
                              disabled={subService.average_times.length === 1}
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
  'w-full rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white outline-none transition placeholder:text-slate-500 focus:border-primary/50 focus:ring-2 focus:ring-primary/20';

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
      <span className="text-xs font-medium uppercase tracking-[0.16em] text-slate-400">
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
      <span className="text-sm font-medium text-slate-200">{label}</span>
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
  onCreate,
  onEdit,
  onDelete,
}: {
  service: ServiceDTO;
  onCreate: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const subServices = service.sub_services ?? [];

  return (
    <div>
      <p className="text-xs uppercase tracking-[0.3em] text-secondary/80">
        Detalhe
      </p>
      <div className="mt-2 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h3 className="font-display text-2xl text-white">{service.title}</h3>
          <p className="mt-1 text-sm text-slate-300">{service.type_name}</p>
        </div>
        <span
          className={`rounded-full border px-3 py-1 text-xs ${
            service.is_active
              ? 'border-emerald-400/30 bg-emerald-500/10 text-emerald-100'
              : 'border-slate-400/30 bg-slate-500/10 text-slate-200'
          }`}
        >
          {service.is_active ? 'Ativo' : 'Inativo'}
        </span>
      </div>

      <p className="mt-4 text-sm leading-6 text-slate-300">
        {service.description}
      </p>

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
        <div className="mt-5 rounded-2xl border border-white/10 bg-white/5 p-4">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">
            Notas
          </p>
          <p className="mt-2 text-sm text-slate-300">{service.notes}</p>
        </div>
      ) : null}

      <div className="mt-6 space-y-4">
        <div className="flex items-center justify-between gap-3">
          <p className="text-sm font-semibold text-white">Subserviços</p>
          <span className="text-xs text-slate-500">{subServices.length}</span>
        </div>

        {subServices.length === 0 ? (
          <StateMessage message="Nenhum subserviço detalhado na resposta." />
        ) : null}

        {subServices.map((subService) => (
          <div
            key={subService.id}
            className="rounded-3xl border border-white/10 bg-white/5 p-4"
          >
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-display text-lg text-white">
                  {subService.title}
                </p>
                <p className="mt-1 text-sm text-slate-300">
                  R$ {subService.price}
                </p>
              </div>
              <span className="rounded-full border border-white/10 px-2 py-1 text-xs text-slate-300">
                {subService.average_times.length} tempo(s)
              </span>
            </div>
            <p className="mt-2 text-sm text-slate-400">
              {subService.description}
            </p>

            {subService.average_times.length > 0 ? (
              <div className="mt-3 space-y-2">
                {subService.average_times.map((averageTime) => (
                  <div
                    key={averageTime.id}
                    className="rounded-2xl border border-white/10 bg-slate-950/40 px-3 py-2 text-xs text-slate-300"
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
          className="rounded-2xl border border-white/10 bg-white/5 px-4 py-2 text-sm text-slate-200 transition hover:bg-white/10"
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
    </div>
  );
}

function DetailMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/5 p-3">
      <p className="text-xs uppercase tracking-[0.16em] text-slate-500">
        {label}
      </p>
      <p className="mt-1 text-sm font-semibold text-white">{value}</p>
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
