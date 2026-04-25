import { useEffect, useMemo, useState, type ReactNode } from 'react';
import { Navigate, useParams } from '@tanstack/react-router';
import { Pencil, User } from 'lucide-react';
import { buildCompanyRoute, DEFAULT_PAGE } from '@petcontrol/shared-constants';
import type {
  BankAccountKind,
  CreatePersonInput,
  GenderIdentity,
  MaritalStatus,
  PetDTO,
  PixKeyKind,
  PersonAddressInput,
  PersonDetailDTO,
  PersonFinanceInput,
  PersonKind,
} from '@petcontrol/shared-types';

import { useListParams } from '@/hooks/use-list-params';
import {
  useCreatePersonMutation,
  useCurrentUserQuery,
  usePeopleQuery,
  usePetsQuery,
  usePersonQuery,
  useUpdatePersonMutation,
} from '@/lib/api/domain.queries';
import { useToastStore } from '@/stores/toast.store';
import { PaginationBar } from '@/ui/pagination-bar';
import { SearchBar } from '@/ui/search-bar';

const PERSON_KIND_OPTIONS: Array<{ value: 'all' | PersonKind; label: string }> =
  [
    { value: 'all', label: 'Todos' },
    { value: 'client', label: 'Clientes' },
    { value: 'employee', label: 'Funcionários' },
    { value: 'outsourced_employee', label: 'Terceirizados' },
    { value: 'supplier', label: 'Fornecedores' },
    { value: 'guardian', label: 'Guardiões' },
    { value: 'responsible', label: 'Responsáveis' },
  ];

const GENDER_OPTIONS: Array<{ value: GenderIdentity; label: string }> = [
  { value: 'woman_cisgender', label: 'Mulher cisgênero' },
  { value: 'man_cisgender', label: 'Homem cisgênero' },
  { value: 'transgender', label: 'Transgênero' },
  { value: 'non_binary', label: 'Não-binário' },
  { value: 'gender_fluid', label: 'Gênero fluido' },
  { value: 'gender_queer', label: 'Gênero queer' },
  { value: 'agender', label: 'Agênero' },
  { value: 'gender_non_conforming', label: 'Não-conforme' },
  { value: 'not_to_expose', label: 'Prefere não expor' },
];

const MARITAL_OPTIONS: Array<{ value: MaritalStatus; label: string }> = [
  { value: 'single', label: 'Solteiro(a)' },
  { value: 'married', label: 'Casado(a)' },
  { value: 'divorced', label: 'Divorciado(a)' },
  { value: 'widowed', label: 'Viúvo(a)' },
  { value: 'separated', label: 'Separado(a)' },
];

const GRADUATION_OPTIONS = [
  { value: 'high_complete', label: 'Ensino médio completo' },
  { value: 'college_complete', label: 'Superior completo' },
  { value: 'postgraduate_complete', label: 'Pós-graduação completa' },
  { value: 'master_complete', label: 'Mestrado completo' },
  { value: 'doctorate_complete', label: 'Doutorado completo' },
] as const;

type PanelMode = 'view' | 'create' | 'edit';

type PersonEmploymentInput = {
  role: string;
  admission_date: string;
  resignation_date?: string;
  salary: string;
};

type PersonEmployeeDocumentsInput = {
  rg: string;
  issuing_body: string;
  issuing_date: string;
  ctps: string;
  ctps_series: string;
  ctps_state: string;
  pis: string;
  graduation: string;
};

type PersonEmployeeBenefitsInput = {
  meal_ticket: boolean;
  meal_ticket_value: string;
  transport_voucher: boolean;
  transport_voucher_qty: number;
  transport_voucher_value: string;
  valid_from: string;
  valid_until?: string;
};

type PersonFinanceFormInput = PersonFinanceInput;

type PeopleFormState = CreatePersonInput & {
  address_enabled: boolean;
  address: PersonAddressInput;
  employment: PersonEmploymentInput;
  finance_enabled: boolean;
  finance: PersonFinanceFormInput;
  employee_documents: PersonEmployeeDocumentsInput;
  employee_benefits_enabled: boolean;
  employee_benefits: PersonEmployeeBenefitsInput;
  pet_ids: string[];
};

type FormSectionKey =
  | 'identification'
  | 'contact'
  | 'address'
  | 'client'
  | 'employment'
  | 'finance'
  | 'employee_documents'
  | 'employee_benefits';

type FormValidationErrors = Partial<Record<FormSectionKey, string[]>>;

const initialAddress: PersonAddressInput = {
  zip_code: '',
  street: '',
  number: '',
  complement: '',
  district: '',
  city: '',
  state: '',
  country: 'Brasil',
  label: '',
};

const initialEmployment: PersonEmploymentInput = {
  role: '',
  admission_date: '',
  resignation_date: '',
  salary: '',
};

const initialEmployeeDocuments: PersonEmployeeDocumentsInput = {
  rg: '',
  issuing_body: '',
  issuing_date: '',
  ctps: '',
  ctps_series: '',
  ctps_state: '',
  pis: '',
  graduation: 'high_complete',
};

const initialEmployeeBenefits: PersonEmployeeBenefitsInput = {
  meal_ticket: false,
  meal_ticket_value: '0.00',
  transport_voucher: false,
  transport_voucher_qty: 0,
  transport_voucher_value: '0.00',
  valid_from: '',
  valid_until: '',
};

const BANK_ACCOUNT_TYPE_OPTIONS: Array<{
  value: BankAccountKind;
  label: string;
}> = [
  { value: 'checking', label: 'Conta corrente' },
  { value: 'savings', label: 'Conta poupança' },
  { value: 'salary', label: 'Conta salário' },
];

const PIX_KEY_TYPE_OPTIONS: Array<{ value: PixKeyKind; label: string }> = [
  { value: 'cpf', label: 'CPF' },
  { value: 'cnpj', label: 'CNPJ' },
  { value: 'email', label: 'Email' },
  { value: 'phone', label: 'Telefone' },
  { value: 'random', label: 'Chave aleatória' },
];

const initialFinance: PersonFinanceFormInput = {
  bank_name: '',
  bank_code: '',
  bank_branch: '',
  bank_account: '',
  bank_account_digit: '',
  bank_account_type: 'checking',
  has_pix: false,
  pix_key: '',
  pix_key_type: 'cpf',
};

function buildInitialForm(kind: PersonKind): PeopleFormState {
  return {
    kind,
    full_name: '',
    short_name: '',
    gender_identity: 'woman_cisgender',
    marital_status: 'single',
    birth_date: '',
    cpf: '',
    email: '',
    phone: '',
    cellphone: '',
    has_whatsapp: true,
    has_system_user: false,
    is_active: true,
    address_enabled: true,
    address: { ...initialAddress },
    client_since: '',
    notes: '',
    employment: { ...initialEmployment },
    finance_enabled: false,
    finance: { ...initialFinance },
    employee_documents: { ...initialEmployeeDocuments },
    employee_benefits_enabled: false,
    employee_benefits: { ...initialEmployeeBenefits },
    pet_ids: [],
  };
}

function buildFormFromDetail(detail: PersonDetailDTO): PeopleFormState {
  return {
    kind: detail.kind,
    full_name: detail.full_name ?? '',
    short_name: detail.short_name ?? '',
    gender_identity: detail.gender_identity ?? 'woman_cisgender',
    marital_status: detail.marital_status ?? 'single',
    birth_date: detail.birth_date ?? '',
    cpf: detail.cpf ?? '',
    email: detail.contact?.email ?? '',
    phone: detail.contact?.phone ?? '',
    cellphone: detail.contact?.cellphone ?? '',
    has_whatsapp: detail.contact?.has_whatsapp ?? false,
    has_system_user: detail.has_system_user,
    is_active: detail.is_active,
    address_enabled: Boolean(detail.address),
    address: {
      zip_code: detail.address?.zip_code ?? '',
      street: detail.address?.street ?? '',
      number: detail.address?.number ?? '',
      complement: detail.address?.complement ?? '',
      district: detail.address?.district ?? '',
      city: detail.address?.city ?? '',
      state: detail.address?.state ?? '',
      country: detail.address?.country ?? 'Brasil',
      label: detail.address?.label ?? '',
    },
    client_since: detail.client_details?.client_since ?? '',
    notes: detail.client_details?.notes ?? '',
    employment: {
      role: detail.employee_details?.role ?? '',
      admission_date: detail.employee_details?.admission_date ?? '',
      resignation_date: detail.employee_details?.resignation_date ?? '',
      salary: detail.employee_details?.salary ?? '',
    },
    finance_enabled: Boolean(detail.finance),
    finance: {
      bank_name: detail.finance?.bank_name ?? '',
      bank_code: detail.finance?.bank_code ?? '',
      bank_branch: detail.finance?.bank_branch ?? '',
      bank_account: detail.finance?.bank_account ?? '',
      bank_account_digit: detail.finance?.bank_account_digit ?? '',
      bank_account_type: detail.finance?.bank_account_type ?? 'checking',
      has_pix: detail.finance?.has_pix ?? false,
      pix_key: detail.finance?.pix_key ?? '',
      pix_key_type: detail.finance?.pix_key_type ?? 'cpf',
    },
    employee_documents: {
      rg: detail.employee_documents?.rg ?? '',
      issuing_body: detail.employee_documents?.issuing_body ?? '',
      issuing_date: '',
      ctps: detail.employee_documents?.ctps ?? '',
      ctps_series: '',
      ctps_state: '',
      pis: detail.employee_documents?.pis ?? '',
      graduation: detail.employee_documents?.graduation ?? 'high_complete',
    },
    employee_benefits_enabled: Boolean(detail.employee_benefits),
    employee_benefits: {
      meal_ticket: detail.employee_benefits?.meal_ticket ?? false,
      meal_ticket_value: detail.employee_benefits?.meal_ticket_value ?? '0.00',
      transport_voucher: detail.employee_benefits?.transport_voucher ?? false,
      transport_voucher_qty:
        detail.employee_benefits?.transport_voucher_qty ?? 0,
      transport_voucher_value:
        detail.employee_benefits?.transport_voucher_value ?? '0.00',
      valid_from: detail.employee_benefits?.valid_from ?? '',
      valid_until: detail.employee_benefits?.valid_until ?? '',
    },
    pet_ids: detail.guardian_pets?.map((pet) => pet.pet_id) ?? [],
  };
}

function readSelectedKindFromLocation(): 'all' | PersonKind {
  if (typeof window === 'undefined') {
    return 'all';
  }

  const kind = new URLSearchParams(window.location.search).get('kind');
  if (
    kind === 'client' ||
    kind === 'employee' ||
    kind === 'outsourced_employee' ||
    kind === 'supplier' ||
    kind === 'guardian' ||
    kind === 'responsible'
  ) {
    return kind;
  }

  return 'all';
}

function readSearchFromLocation(): string {
  if (typeof window === 'undefined') {
    return '';
  }

  return new URLSearchParams(window.location.search).get('search') ?? '';
}

function readPageFromLocation(): number {
  if (typeof window === 'undefined') {
    return DEFAULT_PAGE;
  }

  const raw = new URLSearchParams(window.location.search).get('page');
  const parsed = Number(raw);
  if (!Number.isInteger(parsed) || parsed < DEFAULT_PAGE) {
    return DEFAULT_PAGE;
  }

  return parsed;
}

function readSelectedPersonIdFromLocation(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const params = new URLSearchParams(window.location.search);
  return params.get('id') ?? params.get('person');
}

function syncPersonSelectionToLocation({
  selectedPersonId,
  panelMode,
}: {
  selectedPersonId: string | null;
  panelMode: PanelMode;
}) {
  if (typeof window === 'undefined') {
    return;
  }

  const url = new URL(window.location.href);
  if (panelMode === 'create') {
    url.searchParams.delete('id');
  } else if (selectedPersonId) {
    url.searchParams.set('id', selectedPersonId);
  } else {
    url.searchParams.delete('id');
  }
  url.searchParams.delete('person');

  window.history.replaceState(
    {},
    '',
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function syncSelectedKindToLocation(kind: 'all' | PersonKind) {
  if (typeof window === 'undefined') {
    return;
  }

  const url = new URL(window.location.href);
  if (kind === 'all') {
    url.searchParams.delete('kind');
  } else {
    url.searchParams.set('kind', kind);
  }

  window.history.replaceState(
    {},
    '',
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function syncSearchToLocation(search: string) {
  if (typeof window === 'undefined') {
    return;
  }

  const url = new URL(window.location.href);
  const trimmed = search.trim();
  if (trimmed === '') {
    url.searchParams.delete('search');
  } else {
    url.searchParams.set('search', trimmed);
  }

  window.history.replaceState(
    {},
    '',
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function syncPageToLocation(page: number) {
  if (typeof window === 'undefined') {
    return;
  }

  const url = new URL(window.location.href);
  if (page <= DEFAULT_PAGE) {
    url.searchParams.delete('page');
  } else {
    url.searchParams.set('page', String(page));
  }

  window.history.replaceState(
    {},
    '',
    `${url.pathname}${url.search}${url.hash}`,
  );
}

function normalizeAddressInput(
  form: PeopleFormState,
): PersonAddressInput | undefined {
  if (!form.address_enabled) {
    return undefined;
  }

  const hasValue = Object.values(form.address).some(
    (value) => String(value ?? '').trim().length > 0,
  );
  if (!hasValue) {
    return undefined;
  }

  return {
    zip_code: form.address.zip_code.trim(),
    street: form.address.street.trim(),
    number: form.address.number.trim(),
    complement: form.address.complement?.trim() || undefined,
    district: form.address.district.trim(),
    city: form.address.city.trim(),
    state: form.address.state.trim(),
    country: form.address.country.trim() || 'Brasil',
    label: form.address.label?.trim() || undefined,
  };
}

function normalizeFinanceInput(
  form: PeopleFormState,
): PersonFinanceInput | undefined {
  const supportsFinance =
    form.kind === 'employee' || form.kind === 'outsourced_employee';
  if (!supportsFinance || !form.finance_enabled) {
    return undefined;
  }

  return {
    bank_name: form.finance.bank_name.trim(),
    bank_code: form.finance.bank_code?.trim() || undefined,
    bank_branch: form.finance.bank_branch.trim(),
    bank_account: form.finance.bank_account.trim(),
    bank_account_digit: form.finance.bank_account_digit.trim(),
    bank_account_type: form.finance.bank_account_type,
    has_pix: form.finance.has_pix,
    pix_key: form.finance.pix_key?.trim() || undefined,
    pix_key_type: form.finance.has_pix ? form.finance.pix_key_type : undefined,
  };
}

function hasTrimmedValue(value: string | undefined | null) {
  return String(value ?? '').trim().length > 0;
}

function validatePeopleForm(form: PeopleFormState): FormValidationErrors {
  const errors: FormValidationErrors = {};

  const identificationErrors: string[] = [];
  if (!hasTrimmedValue(form.full_name)) {
    identificationErrors.push('Informe o nome completo.');
  }
  if (!hasTrimmedValue(form.short_name)) {
    identificationErrors.push('Informe o nome curto.');
  }
  if (!hasTrimmedValue(form.birth_date)) {
    identificationErrors.push('Informe a data de nascimento.');
  }
  if (!hasTrimmedValue(form.cpf)) {
    identificationErrors.push('Informe o CPF.');
  }
  if (!hasTrimmedValue(form.gender_identity)) {
    identificationErrors.push('Informe o gênero.');
  }
  if (!hasTrimmedValue(form.marital_status)) {
    identificationErrors.push('Informe o estado civil.');
  }
  if (identificationErrors.length > 0) {
    errors.identification = identificationErrors;
  }

  const contactErrors: string[] = [];
  if (!hasTrimmedValue(form.email)) {
    contactErrors.push('Informe o email.');
  }
  if (!hasTrimmedValue(form.cellphone)) {
    contactErrors.push('Informe o celular.');
  }
  if (contactErrors.length > 0) {
    errors.contact = contactErrors;
  }

  if (form.address_enabled) {
    const addressErrors: string[] = [];
    if (!hasTrimmedValue(form.address.zip_code)) {
      addressErrors.push('Informe o CEP.');
    }
    if (!hasTrimmedValue(form.address.street)) {
      addressErrors.push('Informe o logradouro.');
    }
    if (!hasTrimmedValue(form.address.number)) {
      addressErrors.push('Informe o número.');
    }
    if (!hasTrimmedValue(form.address.district)) {
      addressErrors.push('Informe o bairro.');
    }
    if (!hasTrimmedValue(form.address.city)) {
      addressErrors.push('Informe a cidade.');
    }
    if (!hasTrimmedValue(form.address.state)) {
      addressErrors.push('Informe a UF.');
    }
    if (!hasTrimmedValue(form.address.country)) {
      addressErrors.push('Informe o país.');
    }
    if (addressErrors.length > 0) {
      errors.address = addressErrors;
    }
  }

  if (form.kind === 'employee' || form.kind === 'outsourced_employee') {
    const employmentErrors: string[] = [];
    if (!hasTrimmedValue(form.employment.role)) {
      employmentErrors.push('Informe o cargo.');
    }
    if (!hasTrimmedValue(form.employment.admission_date)) {
      employmentErrors.push('Informe a data de admissão.');
    }
    if (!hasTrimmedValue(form.employment.salary)) {
      employmentErrors.push('Informe o salário.');
    }
    if (employmentErrors.length > 0) {
      errors.employment = employmentErrors;
    }

    if (form.finance_enabled) {
      const financeErrors: string[] = [];
      if (!hasTrimmedValue(form.finance.bank_name)) {
        financeErrors.push('Informe o banco.');
      }
      if (!hasTrimmedValue(form.finance.bank_branch)) {
        financeErrors.push('Informe a agência.');
      }
      if (!hasTrimmedValue(form.finance.bank_account)) {
        financeErrors.push('Informe a conta.');
      }
      if (!hasTrimmedValue(form.finance.bank_account_digit)) {
        financeErrors.push('Informe o dígito da conta.');
      }
      if (form.finance.has_pix && !hasTrimmedValue(form.finance.pix_key)) {
        financeErrors.push('Informe a chave PIX.');
      }
      if (financeErrors.length > 0) {
        errors.finance = financeErrors;
      }
    }

    const employeeDocumentErrors: string[] = [];
    if (!hasTrimmedValue(form.employee_documents.rg)) {
      employeeDocumentErrors.push('Informe o RG.');
    }
    if (!hasTrimmedValue(form.employee_documents.issuing_body)) {
      employeeDocumentErrors.push('Informe o órgão emissor.');
    }
    if (!hasTrimmedValue(form.employee_documents.issuing_date)) {
      employeeDocumentErrors.push('Informe a data de emissão.');
    }
    if (!hasTrimmedValue(form.employee_documents.ctps)) {
      employeeDocumentErrors.push('Informe a CTPS.');
    }
    if (!hasTrimmedValue(form.employee_documents.ctps_series)) {
      employeeDocumentErrors.push('Informe a série da CTPS.');
    }
    if (!hasTrimmedValue(form.employee_documents.ctps_state)) {
      employeeDocumentErrors.push('Informe a UF da CTPS.');
    }
    if (!hasTrimmedValue(form.employee_documents.pis)) {
      employeeDocumentErrors.push('Informe o PIS.');
    }
    if (!hasTrimmedValue(form.employee_documents.graduation)) {
      employeeDocumentErrors.push('Informe a escolaridade.');
    }
    if (employeeDocumentErrors.length > 0) {
      errors.employee_documents = employeeDocumentErrors;
    }

    if (form.employee_benefits_enabled) {
      const employeeBenefitsErrors: string[] = [];
      if (!hasTrimmedValue(form.employee_benefits.valid_from)) {
        employeeBenefitsErrors.push(
          'Informe a data inicial de vigência dos benefícios.',
        );
      }
      if (employeeBenefitsErrors.length > 0) {
        errors.employee_benefits = employeeBenefitsErrors;
      }
    }
  }

  return errors;
}

const VISIBLE_KINDS_FOR_SYSTEM: PersonKind[] = ['client', 'supplier'];

export function PeoplePage() {
  const { companySlug } = useParams({ strict: false });
  const { page, params, search, setSearch, goToPage } = useListParams(
    100,
    readSearchFromLocation(),
    readPageFromLocation(),
  );
  const [selectedKind, setSelectedKind] = useState<'all' | PersonKind>(() =>
    readSelectedKindFromLocation(),
  );
  const currentUserQuery = useCurrentUserQuery();
  const currentUser = currentUserQuery.data;
  const peopleQuery = usePeopleQuery({
    ...params,
    ...(selectedKind !== 'all' ? { kind: selectedKind } : {}),
  });
  const petsQuery = usePetsQuery({ page: 1, limit: 200 });
  const [selectedPersonId, setSelectedPersonId] = useState<string | null>(() =>
    readSelectedPersonIdFromLocation(),
  );
  const [panelMode, setPanelMode] = useState<PanelMode>('view');

  const editableKinds = useMemo(
    () =>
      currentUser?.role === 'system'
        ? (['client', 'supplier'] as PersonKind[])
        : ([
            'client',
            'employee',
            'outsourced_employee',
            'supplier',
            'guardian',
            'responsible',
          ] as PersonKind[]),
    [currentUser?.role],
  );

  const [form, setForm] = useState<PeopleFormState>(() =>
    buildInitialForm(editableKinds[0] ?? 'client'),
  );
  const [showValidation, setShowValidation] = useState(false);

  // Ouve evento customizado para abrir o formulário de criação
  useEffect(() => {
    function handleOpenCreateForm() {
      setPanelMode('create');
      setForm(buildInitialForm(editableKinds[0] ?? 'client'));
      setShowValidation(false);
    }
    window.addEventListener('open-people-create-form', handleOpenCreateForm);
    return () => {
      window.removeEventListener(
        'open-people-create-form',
        handleOpenCreateForm,
      );
    };
  }, [editableKinds]);
  const canAccessPeople =
    currentUser?.role === 'admin' || currentUser?.role === 'system';
  const createMutation = useCreatePersonMutation();
  const updateMutation = useUpdatePersonMutation();
  const pushToast = useToastStore((state) => state.pushToast);

  const [accessProvisionFeedback, setAccessProvisionFeedback] = useState<{
    personId: string;
    email: string;
  } | null>(null);
  const [recentlyProvisionedPeople, setRecentlyProvisionedPeople] = useState<
    Record<string, string>
  >({});

  const filteredPeople = useMemo(() => {
    const items = peopleQuery.data?.data ?? [];

    return items.filter((person) => {
      if (
        currentUser?.role === 'system' &&
        !VISIBLE_KINDS_FOR_SYSTEM.includes(person.kind)
      ) {
        return false;
      }

      if (!search.trim()) {
        return true;
      }

      const haystack = [
        person.full_name ?? '',
        person.short_name ?? '',
        person.cpf ?? '',
        person.kind,
      ]
        .join(' ')
        .toLowerCase();

      return haystack.includes(search.trim().toLowerCase());
    });
  }, [currentUser?.role, peopleQuery.data?.data, search]);

  const activeSelectedPersonId = useMemo(() => {
    if (panelMode === 'create') {
      return null;
    }

    if (
      selectedPersonId &&
      filteredPeople.some((person) => person.id === selectedPersonId)
    ) {
      return selectedPersonId;
    }

    return filteredPeople[0]?.id ?? null;
  }, [filteredPeople, panelMode, selectedPersonId]);

  const personQuery = usePersonQuery(activeSelectedPersonId ?? undefined);

  useEffect(() => {
    syncSelectedKindToLocation(selectedKind);
  }, [selectedKind]);

  useEffect(() => {
    syncSearchToLocation(search);
  }, [search]);

  useEffect(() => {
    syncPageToLocation(page);
  }, [page]);

  useEffect(() => {
    syncPersonSelectionToLocation({
      selectedPersonId: activeSelectedPersonId,
      panelMode,
    });
  }, [activeSelectedPersonId, panelMode]);

  const selectedPerson =
    filteredPeople.find((person) => person.id === activeSelectedPersonId) ??
    null;
  const personDetail = personQuery.data ?? null;
  const selectedPersonSupportsCurrentForm = selectedPerson
    ? editableKinds.includes(selectedPerson.kind)
    : false;
  const isSaving = createMutation.isPending || updateMutation.isPending;
  const validationErrors = useMemo(() => validatePeopleForm(form), [form]);
  const canManageSystemUser =
    (currentUser?.role === 'admin' &&
      (form.kind === 'employee' || form.kind === 'outsourced_employee')) ||
    (currentUser?.role === 'system' && form.kind === 'client');
  const guardianPets = petsQuery.data?.data ?? [];

  if (
    currentUserQuery.isSuccess &&
    !canAccessPeople &&
    typeof companySlug === 'string'
  ) {
    return (
      <Navigate to={buildCompanyRoute(companySlug, 'dashboard')} replace />
    );
  }

  async function submitPersonForm() {
    setShowValidation(true);
    if (Object.keys(validationErrors).length > 0) {
      pushToast(
        'Revise os campos obrigatórios destacados no formulário.',
        'error',
      );
      return;
    }

    try {
      const address = normalizeAddressInput(form);
      const finance = normalizeFinanceInput(form);
      const requestedSystemUser =
        canManageSystemUser &&
        form.has_system_user &&
        (panelMode === 'create' || !personDetail?.has_system_user);
      const feedbackEmail = form.email.trim();

      if (panelMode === 'edit' && activeSelectedPersonId) {
        const input = {
          full_name: form.full_name.trim(),
          short_name: form.short_name.trim(),
          gender_identity: form.gender_identity,
          marital_status: form.marital_status,
          birth_date: form.birth_date || undefined,
          cpf: form.cpf.trim(),
          email: form.email.trim(),
          phone: form.phone?.trim() || undefined,
          cellphone: form.cellphone.trim(),
          has_whatsapp: form.has_whatsapp,
          has_system_user: canManageSystemUser
            ? form.has_system_user
            : undefined,
          is_active: form.is_active ?? true,
          address,
          client_since:
            form.kind === 'client' ? form.client_since || undefined : undefined,
          notes:
            form.kind === 'client'
              ? form.notes?.trim() || undefined
              : undefined,
          pet_ids: form.kind === 'guardian' ? form.pet_ids : undefined,
          employment:
            form.kind === 'employee' || form.kind === 'outsourced_employee'
              ? form.employment
              : undefined,
          finance,
          employee_documents:
            form.kind === 'employee' || form.kind === 'outsourced_employee'
              ? form.employee_documents
              : undefined,
          employee_benefits:
            (form.kind === 'employee' || form.kind === 'outsourced_employee') &&
            form.employee_benefits_enabled
              ? form.employee_benefits
              : undefined,
        };

        const updated = await updateMutation.mutateAsync({
          personId: activeSelectedPersonId,
          input,
        });
        setSelectedPersonId(updated.id);
        if (requestedSystemUser) {
          setAccessProvisionFeedback({
            personId: updated.id,
            email: feedbackEmail,
          });
          setRecentlyProvisionedPeople((current) => ({
            ...current,
            [updated.id]: feedbackEmail,
          }));
        } else {
          setAccessProvisionFeedback(null);
        }
        setPanelMode('view');
        pushToast('Pessoa atualizada com sucesso.', 'success');
        return;
      }

      const input = {
        kind: form.kind,
        full_name: form.full_name.trim(),
        short_name: form.short_name.trim(),
        gender_identity: form.gender_identity,
        marital_status: form.marital_status,
        birth_date: form.birth_date,
        cpf: form.cpf.trim(),
        email: form.email.trim(),
        phone: form.phone?.trim() || undefined,
        cellphone: form.cellphone.trim(),
        has_whatsapp: form.has_whatsapp,
        has_system_user: canManageSystemUser ? form.has_system_user : undefined,
        is_active: form.is_active ?? true,
        address,
        client_since:
          form.kind === 'client' ? form.client_since || undefined : undefined,
        notes:
          form.kind === 'client' ? form.notes?.trim() || undefined : undefined,
        pet_ids: form.kind === 'guardian' ? form.pet_ids : undefined,
        employment:
          form.kind === 'employee' || form.kind === 'outsourced_employee'
            ? form.employment
            : undefined,
        finance,
        employee_documents:
          form.kind === 'employee' || form.kind === 'outsourced_employee'
            ? form.employee_documents
            : undefined,
        employee_benefits:
          (form.kind === 'employee' || form.kind === 'outsourced_employee') &&
          form.employee_benefits_enabled
            ? form.employee_benefits
            : undefined,
      };

      const created = await createMutation.mutateAsync(input);
      setSelectedPersonId(created.id);
      if (requestedSystemUser) {
        setAccessProvisionFeedback({
          personId: created.id,
          email: feedbackEmail,
        });
        setRecentlyProvisionedPeople((current) => ({
          ...current,
          [created.id]: feedbackEmail,
        }));
      } else {
        setAccessProvisionFeedback(null);
      }
      setPanelMode('view');
      pushToast('Pessoa criada com sucesso.', 'success');
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Falha ao salvar a pessoa.';
      pushToast(message, 'error');
    }
  }

  function startCreate() {
    setPanelMode('create');
    setForm(buildInitialForm(editableKinds[0] ?? 'client'));
    setShowValidation(false);
  }

  function startEdit() {
    if (!personDetail || !selectedPersonSupportsCurrentForm) {
      return;
    }
    setPanelMode('edit');
    setForm(buildFormFromDetail(personDetail));
    setShowValidation(false);
  }

  function cancelForm() {
    setPanelMode('view');
    setShowValidation(false);
    if (personDetail) {
      setForm(buildFormFromDetail(personDetail));
    } else {
      setForm(buildInitialForm(editableKinds[0] ?? 'client'));
    }
  }

  return (
    <main className="flex min-w-0 flex-col min-h-full">
      <div className="flex-1 grid grid-cols-1 divide-y divide-border/50 xl:grid-cols-[minmax(0,1.1fr)_26rem] xl:divide-x xl:divide-y-0">
        <section className="flex flex-col min-h-full">
          <header className="bg-[radial-gradient(circle_at_top_right,rgba(2,132,199,0.08),transparent_40%),radial-gradient(circle_at_bottom_left,rgba(16,185,129,0.05),transparent_35%)] px-6 py-8 lg:px-10">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <p className="app-eyebrow">Gestão de Pessoas</p>
                <h1 className="mt-3 font-display text-4xl text-foreground sm:text-5xl">
                  Pessoas
                </h1>
                <p className="mt-4 max-w-2xl text-sm leading-6 text-muted">
                  Base de pessoas do sistema. Visualize, gerencie e evolua o
                  cadastro de clientes, funcionários, fornecedores e
                  responsáveis em um único lugar.
                </p>
              </div>

              <button
                type="button"
                onClick={startCreate}
                className="inline-flex h-12 items-center justify-center rounded-2xl bg-primary px-6 text-sm font-bold text-slate-950 transition hover:brightness-110 shadow-sm"
              >
                Inserir pessoa
              </button>
            </div>
          </header>

          <div className="p-6 lg:p-10">
            <div className="flex flex-col gap-4 lg:flex-row">
              <SearchBar
                id="people-search"
                value={search}
                onChange={setSearch}
                placeholder="Buscar por nome, CPF ou tipo..."
              />

              <select
                aria-label="Filtrar por tipo de pessoa"
                value={selectedKind}
                onChange={(event) =>
                  setSelectedKind(event.target.value as 'all' | PersonKind)
                }
                className="h-12 rounded-2xl border border-border/50 bg-surface px-4 text-sm text-foreground outline-none transition focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
              >
                {PERSON_KIND_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="mt-8 space-y-px divide-y divide-border/50 border-y border-border/50">
              {peopleQuery.isLoading ? (
                <PeopleStateMessage message="Carregando pessoas..." />
              ) : null}
              {peopleQuery.isError ? (
                <PeopleStateMessage
                  message="Falha ao carregar a listagem de pessoas."
                  tone="error"
                />
              ) : null}
              {!peopleQuery.isLoading && filteredPeople.length === 0 ? (
                <PeopleStateMessage message="Nenhuma pessoa encontrada para este filtro." />
              ) : null}

              {filteredPeople.map((person) => {
                const isSelected =
                  panelMode !== 'create' &&
                  activeSelectedPersonId === person.id;
                const wasRecentlyProvisioned = Boolean(
                  recentlyProvisionedPeople[person.id],
                );

                return (
                  <button
                    key={person.id}
                    type="button"
                    onClick={() => {
                      setSelectedPersonId(person.id);
                      setPanelMode('view');
                    }}
                    className={`group flex w-full items-center justify-between gap-4 rounded-[1.8rem] border p-4 text-left transition ${
                      isSelected
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
                          <div className="flex flex-wrap items-center gap-2">
                            <p className="font-medium text-foreground group-hover:text-primary transition">
                              {person.full_name ?? 'Pessoa sem identificação'}
                            </p>
                            <span className="text-[11px] text-muted">
                              ({resolveKindLabel(person.kind)})
                            </span>
                            {wasRecentlyProvisioned ? (
                              <span className="inline-flex rounded-full border border-sky-400/30 bg-sky-500/10 px-2 py-0.5 text-[10px] font-bold uppercase tracking-[0.16em] text-sky-300">
                                Credenciais provisionadas
                              </span>
                            ) : null}
                          </div>
                          <p className="mt-0.5 text-sm text-muted">
                            {person.email ?? 'Email não informado'}
                          </p>
                          <p className="text-[11px] text-muted/70">
                            {person.cpf ? `CPF ${maskCpf(person.cpf)} · ` : ''}
                            {wasRecentlyProvisioned
                              ? 'Usuário provisionado agora'
                              : person.has_system_user
                                ? 'Usuário de sistema'
                                : 'Sem acesso'}
                          </p>
                        </div>
                      </div>

                      <span
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-[0.22em] border ${
                          person.is_active
                            ? 'border-emerald-400/30 bg-emerald-500/10 text-emerald-100'
                            : 'border-stone-400/30 bg-stone-500/10 text-stone-100'
                        }`}
                      >
                        {person.is_active ? 'Ativo' : 'Inativo'}
                      </span>
                    </div>
                  </button>
                );
              })}
            </div>

            <PaginationBar
              meta={peopleQuery.data?.meta}
              onPageChange={goToPage}
            />
          </div>
        </section>

        <aside className="bg-surface/30 p-6 lg:p-10">
          {panelMode === 'create' || panelMode === 'edit' ? (
            <div className="sticky top-10">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="app-eyebrow">
                    {panelMode === 'create' ? 'Nova pessoa' : 'Editar pessoa'}
                  </p>
                  <h2 className="mt-4 font-display text-3xl text-foreground">
                    {panelMode === 'create'
                      ? 'Formulário de cadastro'
                      : form.full_name || 'Atualizar cadastro'}
                  </h2>
                </div>

                <button
                  type="button"
                  onClick={cancelForm}
                  className="rounded-2xl border border-border px-4 py-2 text-sm font-semibold text-foreground transition hover:bg-surface/50"
                >
                  Cancelar
                </button>
              </div>

              <p className="mt-4 text-xs font-semibold uppercase tracking-[0.2em] text-muted">
                Campos com * são obrigatórios.
              </p>

              <form
                className="mt-8 space-y-8"
                noValidate
                onSubmit={(event) => {
                  event.preventDefault();
                  void submitPersonForm();
                }}
              >
                {panelMode === 'create' ? (
                  <Field label="Tipo de pessoa" htmlFor="person-kind">
                    <select
                      aria-label="Tipo de pessoa"
                      id="person-kind"
                      value={form.kind}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          kind: event.target.value as PersonKind,
                        }))
                      }
                      className={fieldClassName}
                    >
                      {editableKinds.map((kind) => (
                        <option key={kind} value={kind}>
                          {resolveKindLabel(kind)}
                        </option>
                      ))}
                    </select>
                  </Field>
                ) : (
                  <div className="rounded-2xl border border-border bg-surface/50 px-4 py-3 text-sm text-foreground">
                    Tipo atual: {resolveKindLabel(form.kind)}
                  </div>
                )}

                <DetailSection
                  title="Identificação"
                  errorMessages={
                    showValidation ? validationErrors.identification : undefined
                  }
                >
                  <Field
                    label="Nome completo"
                    htmlFor="person-full-name"
                    required
                  >
                    <input
                      aria-label="Nome completo"
                      id="person-full-name"
                      className={fieldClassName}
                      required
                      value={form.full_name}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          full_name: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field
                    label="Nome curto"
                    htmlFor="person-short-name"
                    required
                  >
                    <input
                      aria-label="Nome curto"
                      id="person-short-name"
                      className={fieldClassName}
                      required
                      value={form.short_name}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          short_name: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field
                    label="Nascimento"
                    htmlFor="person-birth-date"
                    required
                  >
                    <input
                      aria-label="Nascimento"
                      id="person-birth-date"
                      className={fieldClassName}
                      type="date"
                      required
                      value={form.birth_date}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          birth_date: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="CPF" htmlFor="person-cpf" required>
                    <input
                      aria-label="CPF"
                      id="person-cpf"
                      className={fieldClassName}
                      required
                      value={form.cpf}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          cpf: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="Gênero" htmlFor="person-gender" required>
                    <select
                      aria-label="Gênero"
                      id="person-gender"
                      value={form.gender_identity}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          gender_identity: event.target.value as GenderIdentity,
                        }))
                      }
                      className={fieldClassName}
                    >
                      {GENDER_OPTIONS.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </Field>
                  <Field label="Estado civil" htmlFor="person-marital" required>
                    <select
                      aria-label="Estado civil"
                      id="person-marital"
                      value={form.marital_status}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          marital_status: event.target.value as MaritalStatus,
                        }))
                      }
                      className={fieldClassName}
                    >
                      {MARITAL_OPTIONS.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </Field>
                </DetailSection>

                <DetailSection
                  title="Contato"
                  errorMessages={
                    showValidation ? validationErrors.contact : undefined
                  }
                >
                  <Field label="Email" htmlFor="person-email" required>
                    <input
                      aria-label="Email"
                      id="person-email"
                      className={fieldClassName}
                      type="email"
                      required
                      value={form.email}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          email: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="Celular" htmlFor="person-cellphone" required>
                    <input
                      aria-label="Celular"
                      id="person-cellphone"
                      className={fieldClassName}
                      required
                      value={form.cellphone}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          cellphone: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="Telefone" htmlFor="person-phone">
                    <input
                      aria-label="Telefone"
                      id="person-phone"
                      className={fieldClassName}
                      value={form.phone ?? ''}
                      onChange={(event) =>
                        setForm((current: PeopleFormState) => ({
                          ...current,
                          phone: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <ToggleRow
                    label="WhatsApp no celular"
                    checked={form.has_whatsapp}
                    onChange={(checked) =>
                      setForm((current: PeopleFormState) => ({
                        ...current,
                        has_whatsapp: checked,
                      }))
                    }
                  />
                  <ToggleRow
                    label="Pessoa ativa"
                    checked={form.is_active ?? true}
                    onChange={(checked) =>
                      setForm((current: PeopleFormState) => ({
                        ...current,
                        is_active: checked,
                      }))
                    }
                  />
                  {canManageSystemUser ? (
                    <div className="space-y-3 rounded-2xl border border-sky-400/30 bg-sky-500/10 px-4 py-4">
                      <ToggleRow
                        label="Criar usuário de sistema"
                        checked={form.has_system_user ?? false}
                        disabled={
                          panelMode === 'edit' &&
                          Boolean(personDetail?.has_system_user)
                        }
                        onChange={(checked) =>
                          setForm((current: PeopleFormState) => ({
                            ...current,
                            has_system_user: checked,
                          }))
                        }
                      />
                      <p className="text-xs leading-5 text-sky-300">
                        {panelMode === 'edit' && personDetail?.has_system_user
                          ? 'Esta pessoa já possui usuário vinculado. A desativação desse vínculo fica para um fluxo administrativo específico.'
                          : currentUser?.role === 'system'
                            ? 'Ao finalizar o cadastro, a pessoa também receberá um usuário simples de sistema e credenciais enviadas por email.'
                            : 'Ao finalizar o cadastro, a pessoa também receberá um usuário de sistema e credenciais enviadas por email.'}
                      </p>
                    </div>
                  ) : null}
                </DetailSection>

                <DetailSection
                  title="Endereço principal"
                  errorMessages={
                    showValidation ? validationErrors.address : undefined
                  }
                >
                  <ToggleRow
                    label="Cadastrar endereço agora"
                    checked={form.address_enabled}
                    onChange={(checked) =>
                      setForm((current: PeopleFormState) => ({
                        ...current,
                        address_enabled: checked,
                      }))
                    }
                  />

                  {form.address_enabled ? (
                    <>
                      <Field label="CEP" htmlFor="person-address-zip" required>
                        <input
                          aria-label="CEP"
                          id="person-address-zip"
                          className={fieldClassName}
                          value={form.address.zip_code}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                zip_code: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Logradouro"
                        htmlFor="person-address-street"
                        required
                      >
                        <input
                          aria-label="Logradouro"
                          id="person-address-street"
                          className={fieldClassName}
                          value={form.address.street}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                street: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Número"
                        htmlFor="person-address-number"
                        required
                      >
                        <input
                          aria-label="Número"
                          id="person-address-number"
                          className={fieldClassName}
                          value={form.address.number}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                number: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Complemento"
                        htmlFor="person-address-complement"
                      >
                        <input
                          aria-label="Complemento"
                          id="person-address-complement"
                          className={fieldClassName}
                          value={form.address.complement ?? ''}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                complement: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Bairro"
                        htmlFor="person-address-district"
                        required
                      >
                        <input
                          aria-label="Bairro"
                          id="person-address-district"
                          className={fieldClassName}
                          value={form.address.district}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                district: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Cidade"
                        htmlFor="person-address-city"
                        required
                      >
                        <input
                          aria-label="Cidade"
                          id="person-address-city"
                          className={fieldClassName}
                          value={form.address.city}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                city: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field label="UF" htmlFor="person-address-state" required>
                        <input
                          aria-label="UF"
                          id="person-address-state"
                          className={fieldClassName}
                          value={form.address.state}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                state: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="País"
                        htmlFor="person-address-country"
                        required
                      >
                        <input
                          aria-label="País"
                          id="person-address-country"
                          className={fieldClassName}
                          value={form.address.country}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                country: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field label="Rótulo" htmlFor="person-address-label">
                        <input
                          aria-label="Rótulo"
                          id="person-address-label"
                          className={fieldClassName}
                          value={form.address.label ?? ''}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              address: {
                                ...current.address,
                                label: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                    </>
                  ) : null}
                </DetailSection>

                {form.kind === 'client' ? (
                  <DetailSection title="Dados de cliente">
                    <Field label="Cliente desde" htmlFor="person-client-since">
                      <input
                        aria-label="Cliente desde"
                        id="person-client-since"
                        className={fieldClassName}
                        type="date"
                        value={form.client_since ?? ''}
                        onChange={(event) =>
                          setForm((current: PeopleFormState) => ({
                            ...current,
                            client_since: event.target.value,
                          }))
                        }
                      />
                    </Field>
                    <Field label="Observações" htmlFor="person-client-notes">
                      <textarea
                        aria-label="Observações"
                        id="person-client-notes"
                        className={`${fieldClassName} min-h-28 resize-y`}
                        value={form.notes ?? ''}
                        onChange={(event) =>
                          setForm((current: PeopleFormState) => ({
                            ...current,
                            notes: event.target.value,
                          }))
                        }
                      />
                    </Field>
                  </DetailSection>
                ) : null}

                {form.kind === 'guardian' ? (
                  <DetailSection title="Pets vinculados">
                    {petsQuery.isLoading ? (
                      <PeopleStateMessage message="Carregando pets disponíveis..." />
                    ) : null}
                    {petsQuery.isError ? (
                      <PeopleStateMessage
                        message="Falha ao carregar os pets disponíveis."
                        tone="error"
                      />
                    ) : null}
                    {!petsQuery.isLoading && guardianPets.length === 0 ? (
                      <PeopleStateMessage message="Nenhum pet disponível para vínculo neste tenant." />
                    ) : null}
                    {guardianPets.length > 0 ? (
                      <>
                        <Field
                          label="Selecione os pets"
                          htmlFor="person-guardian-pets"
                        >
                          <select
                            aria-label="Selecione os pets"
                            id="person-guardian-pets"
                            multiple
                            className={`${fieldClassName} min-h-44`}
                            value={form.pet_ids}
                            onChange={(event) =>
                              setForm((current: PeopleFormState) => ({
                                ...current,
                                pet_ids: Array.from(
                                  event.target.selectedOptions,
                                  (option) => option.value,
                                ),
                              }))
                            }
                          >
                            {guardianPets.map((pet) => (
                              <option key={pet.id} value={pet.id}>
                                {pet.name} · {resolvePetKindLabel(pet.kind)} ·{' '}
                                {pet.owner_name ?? 'Cliente sem nome'}
                              </option>
                            ))}
                          </select>
                        </Field>
                        <p className="text-xs leading-5 text-muted">
                          Segure <strong>Ctrl</strong> ou <strong>Cmd</strong>{' '}
                          para selecionar mais de um pet.
                        </p>
                      </>
                    ) : null}
                  </DetailSection>
                ) : null}

                {form.kind === 'employee' ||
                form.kind === 'outsourced_employee' ? (
                  <>
                    <DetailSection
                      title="Vínculo empregatício"
                      errorMessages={
                        showValidation ? validationErrors.employment : undefined
                      }
                    >
                      <Field
                        label="Cargo"
                        htmlFor="person-employment-role"
                        required
                      >
                        <input
                          aria-label="Cargo"
                          id="person-employment-role"
                          className={fieldClassName}
                          value={form.employment.role}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employment: {
                                ...current.employment,
                                role: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Data de admissão"
                        htmlFor="person-employment-admission"
                        required
                      >
                        <input
                          aria-label="Data de admissão"
                          id="person-employment-admission"
                          type="date"
                          className={fieldClassName}
                          value={form.employment.admission_date}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employment: {
                                ...current.employment,
                                admission_date: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Data de desligamento"
                        htmlFor="person-employment-resignation"
                      >
                        <input
                          aria-label="Data de desligamento"
                          id="person-employment-resignation"
                          type="date"
                          className={fieldClassName}
                          value={form.employment.resignation_date ?? ''}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employment: {
                                ...current.employment,
                                resignation_date: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Salário"
                        htmlFor="person-employment-salary"
                        required
                      >
                        <input
                          aria-label="Salário"
                          id="person-employment-salary"
                          className={fieldClassName}
                          value={form.employment.salary}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employment: {
                                ...current.employment,
                                salary: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                    </DetailSection>

                    <DetailSection
                      title="Documentos do funcionário"
                      errorMessages={
                        showValidation
                          ? validationErrors.employee_documents
                          : undefined
                      }
                    >
                      <Field label="RG" htmlFor="person-doc-rg" required>
                        <input
                          aria-label="RG"
                          id="person-doc-rg"
                          className={fieldClassName}
                          value={form.employee_documents.rg}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                rg: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Órgão emissor"
                        htmlFor="person-doc-issuing-body"
                        required
                      >
                        <input
                          aria-label="Órgão emissor"
                          id="person-doc-issuing-body"
                          className={fieldClassName}
                          value={form.employee_documents.issuing_body}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                issuing_body: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Data de emissão"
                        htmlFor="person-doc-issuing-date"
                        required
                      >
                        <input
                          aria-label="Data de emissão"
                          id="person-doc-issuing-date"
                          type="date"
                          className={fieldClassName}
                          value={form.employee_documents.issuing_date}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                issuing_date: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field label="CTPS" htmlFor="person-doc-ctps" required>
                        <input
                          aria-label="CTPS"
                          id="person-doc-ctps"
                          className={fieldClassName}
                          value={form.employee_documents.ctps}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                ctps: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Série CTPS"
                        htmlFor="person-doc-ctps-series"
                        required
                      >
                        <input
                          aria-label="Série CTPS"
                          id="person-doc-ctps-series"
                          className={fieldClassName}
                          value={form.employee_documents.ctps_series}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                ctps_series: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="UF CTPS"
                        htmlFor="person-doc-ctps-state"
                        required
                      >
                        <input
                          aria-label="UF CTPS"
                          id="person-doc-ctps-state"
                          className={fieldClassName}
                          value={form.employee_documents.ctps_state}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                ctps_state: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field label="PIS" htmlFor="person-doc-pis" required>
                        <input
                          aria-label="PIS"
                          id="person-doc-pis"
                          className={fieldClassName}
                          value={form.employee_documents.pis}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                pis: event.target.value,
                              },
                            }))
                          }
                        />
                      </Field>
                      <Field
                        label="Escolaridade"
                        htmlFor="person-doc-graduation"
                        required
                      >
                        <select
                          aria-label="Escolaridade"
                          id="person-doc-graduation"
                          className={fieldClassName}
                          value={form.employee_documents.graduation}
                          onChange={(event) =>
                            setForm((current: PeopleFormState) => ({
                              ...current,
                              employee_documents: {
                                ...current.employee_documents,
                                graduation: event.target.value,
                              },
                            }))
                          }
                        >
                          {GRADUATION_OPTIONS.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                      </Field>
                    </DetailSection>

                    <DetailSection
                      title="Financeiro"
                      errorMessages={
                        showValidation ? validationErrors.finance : undefined
                      }
                    >
                      <ToggleRow
                        label="Cadastrar conta bancária"
                        checked={form.finance_enabled}
                        onChange={(checked) =>
                          setForm((current: PeopleFormState) => ({
                            ...current,
                            finance_enabled: checked,
                          }))
                        }
                      />

                      {form.finance_enabled ? (
                        <>
                          <Field
                            label="Banco"
                            htmlFor="person-finance-bank-name"
                            required
                          >
                            <input
                              aria-label="Banco"
                              id="person-finance-bank-name"
                              className={fieldClassName}
                              value={form.finance.bank_name}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  finance: {
                                    ...current.finance,
                                    bank_name: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Código do banco"
                            htmlFor="person-finance-bank-code"
                          >
                            <input
                              aria-label="Código do banco"
                              id="person-finance-bank-code"
                              className={fieldClassName}
                              value={form.finance.bank_code ?? ''}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  finance: {
                                    ...current.finance,
                                    bank_code: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Agência"
                            htmlFor="person-finance-bank-branch"
                            required
                          >
                            <input
                              aria-label="Agência"
                              id="person-finance-bank-branch"
                              className={fieldClassName}
                              value={form.finance.bank_branch}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  finance: {
                                    ...current.finance,
                                    bank_branch: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Conta"
                            htmlFor="person-finance-bank-account"
                            required
                          >
                            <input
                              aria-label="Conta"
                              id="person-finance-bank-account"
                              className={fieldClassName}
                              value={form.finance.bank_account}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  finance: {
                                    ...current.finance,
                                    bank_account: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Dígito"
                            htmlFor="person-finance-bank-account-digit"
                            required
                          >
                            <input
                              aria-label="Dígito"
                              id="person-finance-bank-account-digit"
                              className={fieldClassName}
                              value={form.finance.bank_account_digit}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  finance: {
                                    ...current.finance,
                                    bank_account_digit: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Tipo de conta"
                            htmlFor="person-finance-bank-account-type"
                          >
                            <select
                              aria-label="Tipo de conta"
                              id="person-finance-bank-account-type"
                              className={fieldClassName}
                              value={form.finance.bank_account_type}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  finance: {
                                    ...current.finance,
                                    bank_account_type: event.target
                                      .value as BankAccountKind,
                                  },
                                }))
                              }
                            >
                              {BANK_ACCOUNT_TYPE_OPTIONS.map((option) => (
                                <option key={option.value} value={option.value}>
                                  {option.label}
                                </option>
                              ))}
                            </select>
                          </Field>
                          <ToggleRow
                            label="Possui PIX"
                            checked={form.finance.has_pix}
                            onChange={(checked) =>
                              setForm((current: PeopleFormState) => ({
                                ...current,
                                finance: {
                                  ...current.finance,
                                  has_pix: checked,
                                },
                              }))
                            }
                          />
                          {form.finance.has_pix ? (
                            <>
                              <Field
                                label="Chave PIX"
                                htmlFor="person-finance-pix-key"
                                required
                              >
                                <input
                                  aria-label="Chave PIX"
                                  id="person-finance-pix-key"
                                  className={fieldClassName}
                                  value={form.finance.pix_key ?? ''}
                                  onChange={(event) =>
                                    setForm((current: PeopleFormState) => ({
                                      ...current,
                                      finance: {
                                        ...current.finance,
                                        pix_key: event.target.value,
                                      },
                                    }))
                                  }
                                />
                              </Field>
                              <Field
                                label="Tipo de chave PIX"
                                htmlFor="person-finance-pix-key-type"
                              >
                                <select
                                  aria-label="Tipo de chave PIX"
                                  id="person-finance-pix-key-type"
                                  className={fieldClassName}
                                  value={form.finance.pix_key_type ?? 'cpf'}
                                  onChange={(event) =>
                                    setForm((current: PeopleFormState) => ({
                                      ...current,
                                      finance: {
                                        ...current.finance,
                                        pix_key_type: event.target
                                          .value as PixKeyKind,
                                      },
                                    }))
                                  }
                                >
                                  {PIX_KEY_TYPE_OPTIONS.map((option) => (
                                    <option
                                      key={option.value}
                                      value={option.value}
                                    >
                                      {option.label}
                                    </option>
                                  ))}
                                </select>
                              </Field>
                            </>
                          ) : null}
                        </>
                      ) : null}
                    </DetailSection>

                    <DetailSection
                      title="Benefícios"
                      errorMessages={
                        showValidation
                          ? validationErrors.employee_benefits
                          : undefined
                      }
                    >
                      <ToggleRow
                        label="Cadastrar benefícios agora"
                        checked={form.employee_benefits_enabled}
                        onChange={(checked) =>
                          setForm((current: PeopleFormState) => ({
                            ...current,
                            employee_benefits_enabled: checked,
                          }))
                        }
                      />

                      {form.employee_benefits_enabled ? (
                        <>
                          <ToggleRow
                            label="Vale refeição"
                            checked={form.employee_benefits.meal_ticket}
                            onChange={(checked) =>
                              setForm((current: PeopleFormState) => ({
                                ...current,
                                employee_benefits: {
                                  ...current.employee_benefits,
                                  meal_ticket: checked,
                                },
                              }))
                            }
                          />
                          <Field
                            label="Valor vale refeição"
                            htmlFor="person-benefits-meal-value"
                          >
                            <input
                              aria-label="Valor vale refeição"
                              id="person-benefits-meal-value"
                              className={fieldClassName}
                              value={form.employee_benefits.meal_ticket_value}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  employee_benefits: {
                                    ...current.employee_benefits,
                                    meal_ticket_value: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <ToggleRow
                            label="Vale transporte"
                            checked={form.employee_benefits.transport_voucher}
                            onChange={(checked) =>
                              setForm((current: PeopleFormState) => ({
                                ...current,
                                employee_benefits: {
                                  ...current.employee_benefits,
                                  transport_voucher: checked,
                                },
                              }))
                            }
                          />
                          <Field
                            label="Quantidade VT"
                            htmlFor="person-benefits-voucher-qty"
                          >
                            <input
                              aria-label="Quantidade VT"
                              id="person-benefits-voucher-qty"
                              type="number"
                              className={fieldClassName}
                              value={
                                form.employee_benefits.transport_voucher_qty
                              }
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  employee_benefits: {
                                    ...current.employee_benefits,
                                    transport_voucher_qty: Number(
                                      event.target.value,
                                    ),
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Valor VT"
                            htmlFor="person-benefits-voucher-value"
                          >
                            <input
                              aria-label="Valor VT"
                              id="person-benefits-voucher-value"
                              className={fieldClassName}
                              value={
                                form.employee_benefits.transport_voucher_value
                              }
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  employee_benefits: {
                                    ...current.employee_benefits,
                                    transport_voucher_value: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Válido a partir de"
                            htmlFor="person-benefits-valid-from"
                            required
                          >
                            <input
                              aria-label="Válido a partir de"
                              id="person-benefits-valid-from"
                              type="date"
                              className={fieldClassName}
                              value={form.employee_benefits.valid_from}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  employee_benefits: {
                                    ...current.employee_benefits,
                                    valid_from: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                          <Field
                            label="Válido até"
                            htmlFor="person-benefits-valid-until"
                          >
                            <input
                              aria-label="Válido até"
                              id="person-benefits-valid-until"
                              type="date"
                              className={fieldClassName}
                              value={form.employee_benefits.valid_until ?? ''}
                              onChange={(event) =>
                                setForm((current: PeopleFormState) => ({
                                  ...current,
                                  employee_benefits: {
                                    ...current.employee_benefits,
                                    valid_until: event.target.value,
                                  },
                                }))
                              }
                            />
                          </Field>
                        </>
                      ) : null}
                    </DetailSection>
                  </>
                ) : null}

                <div className="flex flex-wrap gap-3">
                  <button
                    type="submit"
                    disabled={isSaving}
                    className="inline-flex h-12 items-center justify-center rounded-2xl bg-primary px-6 text-sm font-bold text-slate-950 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isSaving
                      ? 'Salvando...'
                      : panelMode === 'create'
                        ? 'Criar pessoa'
                        : 'Salvar alterações'}
                  </button>
                  <button
                    type="button"
                    onClick={cancelForm}
                    className="inline-flex h-12 items-center justify-center rounded-2xl border border-border px-6 text-sm font-semibold text-foreground transition hover:bg-surface/50"
                  >
                    Voltar
                  </button>
                </div>
              </form>
            </div>
          ) : selectedPerson ? (
            <div className="sticky top-10">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="app-eyebrow">Seleção atual</p>
                  <h2 className="mt-4 font-display text-3xl text-foreground">
                    {selectedPerson.full_name ?? 'Pessoa sem identificação'}
                  </h2>
                </div>

                {selectedPersonSupportsCurrentForm ? (
                  <button
                    type="button"
                    onClick={startEdit}
                    title="Editar"
                    aria-label="Editar"
                    className="rounded-2xl border border-border px-4 py-2 text-sm font-semibold text-foreground transition hover:bg-surface/50"
                  >
                    <Pencil className="h-4 w-4" aria-hidden="true" />
                  </button>
                ) : null}
              </div>

              {!selectedPersonSupportsCurrentForm ? (
                <div className="mt-6 rounded-2xl border border-amber-400/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
                  A edição deste tipo de pessoa será concluída no próximo slice
                  do módulo.
                </div>
              ) : null}

              {personQuery.isLoading ? (
                <div className="mt-8">
                  <PeopleStateMessage message="Carregando detalhes da pessoa..." />
                </div>
              ) : null}

              {personQuery.isError ? (
                <div className="mt-8">
                  <PeopleStateMessage
                    message="Falha ao carregar os detalhes da pessoa."
                    tone="error"
                  />
                </div>
              ) : null}

              {personDetail ? (
                <div className="mt-8 space-y-8">
                  {accessProvisionFeedback?.personId === personDetail.id ? (
                    <div className="rounded-2xl border border-emerald-400/30 bg-emerald-500/10 px-4 py-4 text-sm text-emerald-200">
                      <p className="font-semibold uppercase tracking-[0.16em] text-emerald-100">
                        Credenciais enviadas
                      </p>
                      <p className="mt-2 leading-6">
                        O acesso foi provisionado e o envio das credenciais foi
                        iniciado para{' '}
                        <strong>{accessProvisionFeedback.email}</strong>.
                      </p>
                    </div>
                  ) : null}

                  <dl className="space-y-4">
                    <SummaryRow
                      label="Tipo"
                      value={resolveKindLabel(personDetail.kind)}
                    />
                    <SummaryRow
                      label="Nome curto"
                      value={personDetail.short_name ?? 'Não informado'}
                    />
                    <SummaryRow label="CPF" value={maskCpf(personDetail.cpf)} />
                    <SummaryRow
                      label="Usuário do sistema"
                      value={personDetail.has_system_user ? 'Sim' : 'Não'}
                    />
                    <SummaryRow
                      label="Status"
                      value={personDetail.is_active ? 'Ativo' : 'Inativo'}
                    />
                  </dl>

                  <DetailSection title="Contato">
                    <SummaryRow
                      label="Email"
                      value={personDetail.contact?.email ?? 'Não informado'}
                    />
                    <SummaryRow
                      label="Celular"
                      value={personDetail.contact?.cellphone ?? 'Não informado'}
                    />
                    <SummaryRow
                      label="Telefone"
                      value={personDetail.contact?.phone ?? 'Não informado'}
                    />
                    <SummaryRow
                      label="WhatsApp"
                      value={personDetail.contact?.has_whatsapp ? 'Sim' : 'Não'}
                    />
                  </DetailSection>

                  {personDetail.linked_user ? (
                    <DetailSection title="Usuário vinculado">
                      <SummaryRow
                        label="Email de acesso"
                        value={personDetail.linked_user.email}
                      />
                      <SummaryRow
                        label="Papel técnico"
                        value={personDetail.linked_user.role}
                      />
                      <SummaryRow
                        label="Tipo de vínculo"
                        value={personDetail.linked_user.kind}
                      />
                      <SummaryRow
                        label="Status do usuário"
                        value={
                          personDetail.linked_user.is_active
                            ? 'Ativo'
                            : 'Inativo'
                        }
                      />
                      <SummaryRow
                        label="Desde"
                        value={personDetail.linked_user.joined_at}
                      />
                    </DetailSection>
                  ) : null}

                  <DetailSection title="Endereço principal">
                    <SummaryRow
                      label="Logradouro"
                      value={
                        personDetail.address
                          ? `${personDetail.address.street}, ${personDetail.address.number}`
                          : 'Não informado'
                      }
                    />
                    <SummaryRow
                      label="Complemento"
                      value={
                        personDetail.address?.complement ?? 'Não informado'
                      }
                    />
                    <SummaryRow
                      label="Bairro"
                      value={personDetail.address?.district ?? 'Não informado'}
                    />
                    <SummaryRow
                      label="Cidade/UF"
                      value={
                        personDetail.address
                          ? `${personDetail.address.city}/${personDetail.address.state}`
                          : 'Não informado'
                      }
                    />
                    <SummaryRow
                      label="CEP"
                      value={personDetail.address?.zip_code ?? 'Não informado'}
                    />
                  </DetailSection>

                  {personDetail.client_details ? (
                    <DetailSection title="Cliente">
                      <SummaryRow
                        label="Cliente desde"
                        value={
                          personDetail.client_details.client_since ??
                          'Não informado'
                        }
                      />
                      <SummaryRow
                        label="Observações"
                        value={
                          personDetail.client_details.notes ?? 'Não informado'
                        }
                      />
                    </DetailSection>
                  ) : null}

                  {personDetail.guardian_pets &&
                  personDetail.guardian_pets.length > 0 ? (
                    <DetailSection title="Pets vinculados">
                      {personDetail.guardian_pets.map((pet) => (
                        <SummaryRow
                          key={pet.pet_id}
                          label={pet.name}
                          value={`${resolvePetKindLabel(pet.kind)} · ${pet.owner_name}`}
                        />
                      ))}
                    </DetailSection>
                  ) : null}

                  {personDetail.employee_details ? (
                    <DetailSection title="Vínculo empregatício">
                      <SummaryRow
                        label="Cargo"
                        value={
                          personDetail.employee_details.role ?? 'Não informado'
                        }
                      />
                      <SummaryRow
                        label="Admissão"
                        value={
                          personDetail.employee_details.admission_date ??
                          'Não informado'
                        }
                      />
                      <SummaryRow
                        label="Rescisão"
                        value={
                          personDetail.employee_details.resignation_date ??
                          'Não informado'
                        }
                      />
                      <SummaryRow
                        label="Salário"
                        value={
                          personDetail.employee_details.salary ??
                          'Não informado'
                        }
                      />
                    </DetailSection>
                  ) : null}

                  {personDetail.finance ? (
                    <DetailSection title="Financeiro">
                      <SummaryRow
                        label="Banco"
                        value={personDetail.finance.bank_name}
                      />
                      <SummaryRow
                        label="Código"
                        value={
                          personDetail.finance.bank_code ?? 'Não informado'
                        }
                      />
                      <SummaryRow
                        label="Agência"
                        value={personDetail.finance.bank_branch}
                      />
                      <SummaryRow
                        label="Conta"
                        value={`${personDetail.finance.bank_account}-${personDetail.finance.bank_account_digit}`}
                      />
                      <SummaryRow
                        label="Tipo de conta"
                        value={resolveBankAccountTypeLabel(
                          personDetail.finance.bank_account_type,
                        )}
                      />
                      <SummaryRow
                        label="PIX"
                        value={
                          personDetail.finance.has_pix
                            ? `${resolvePixKeyTypeLabel(personDetail.finance.pix_key_type)}: ${personDetail.finance.pix_key ?? 'Não informado'}`
                            : 'Não'
                        }
                      />
                    </DetailSection>
                  ) : null}

                  {personDetail.employee_documents ? (
                    <DetailSection title="Documentos do funcionário">
                      <SummaryRow
                        label="RG"
                        value={personDetail.employee_documents.rg}
                      />
                      <SummaryRow
                        label="Órgão emissor"
                        value={personDetail.employee_documents.issuing_body}
                      />
                      <SummaryRow
                        label="CTPS"
                        value={personDetail.employee_documents.ctps}
                      />
                      <SummaryRow
                        label="PIS"
                        value={personDetail.employee_documents.pis}
                      />
                      <SummaryRow
                        label="Escolaridade"
                        value={personDetail.employee_documents.graduation}
                      />
                    </DetailSection>
                  ) : null}

                  {personDetail.employee_benefits ? (
                    <DetailSection title="Benefícios">
                      <SummaryRow
                        label="Vale refeição"
                        value={
                          personDetail.employee_benefits.meal_ticket
                            ? `Sim (${personDetail.employee_benefits.meal_ticket_value})`
                            : 'Não'
                        }
                      />
                      <SummaryRow
                        label="Vale transporte"
                        value={
                          personDetail.employee_benefits.transport_voucher
                            ? `Sim (${personDetail.employee_benefits.transport_voucher_qty} x ${personDetail.employee_benefits.transport_voucher_value})`
                            : 'Não'
                        }
                      />
                      <SummaryRow
                        label="Vigência inicial"
                        value={personDetail.employee_benefits.valid_from}
                      />
                      <SummaryRow
                        label="Vigência final"
                        value={
                          personDetail.employee_benefits.valid_until ??
                          'Não informado'
                        }
                      />
                    </DetailSection>
                  ) : null}
                </div>
              ) : null}
            </div>
          ) : (
            <div className="flex h-full min-h-[18rem] items-center justify-center text-center text-sm text-muted">
              <div className="max-w-[200px]">
                <p>
                  Selecione uma pessoa da lista para abrir o painel lateral.
                </p>
              </div>
            </div>
          )}
        </aside>
      </div>
    </main>
  );
}

function SummaryRow({ label, value }: { label: string; value: string }) {
  return (
    <dl className="border-b border-border/50 pb-3 last:border-0">
      <dt className="text-[10px] font-bold uppercase tracking-[0.24em] text-muted">
        {label}
      </dt>
      <dd className="mt-1 text-sm font-medium text-foreground">{value}</dd>
    </dl>
  );
}

function DetailSection({
  title,
  errorMessages,
  children,
}: {
  title: string;
  errorMessages?: string[];
  children: ReactNode;
}) {
  return (
    <section>
      <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-muted">
        {title}
      </p>
      {errorMessages && errorMessages.length > 0 ? (
        <div className="mt-3 rounded-2xl border border-rose-400/40 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
          {errorMessages.map((message) => (
            <p key={`${title}-${message}`}>{message}</p>
          ))}
        </div>
      ) : null}
      <div className="mt-3 space-y-4">{children}</div>
    </section>
  );
}

function Field({
  label,
  htmlFor,
  required = false,
  children,
}: {
  label: string;
  htmlFor: string;
  required?: boolean;
  children: ReactNode;
}) {
  return (
    <label className="block space-y-2" htmlFor={htmlFor}>
      <span className="text-[10px] font-bold uppercase tracking-[0.24em] text-muted">
        {label}
        {required ? <span className="ml-1 text-rose-500">*</span> : null}
      </span>
      {children}
    </label>
  );
}

function ToggleRow({
  label,
  checked,
  disabled = false,
  onChange,
}: {
  label: string;
  checked: boolean;
  disabled?: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label
      className={`flex items-center justify-between gap-4 rounded-2xl border border-border px-4 py-3 ${
        disabled ? 'cursor-not-allowed bg-surface/30 opacity-70' : ''
      }`}
    >
      <span className="text-sm font-medium text-foreground">{label}</span>
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        className="h-4 w-4 rounded border-border/50 bg-surface/50 text-primary focus:ring-primary/20"
      />
    </label>
  );
}

function PeopleStateMessage({
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
          ? 'border-rose-400/40 bg-rose-500/10 text-rose-200'
          : 'border-border bg-surface/50 text-muted'
      }`}
    >
      {message}
    </div>
  );
}

const fieldClassName =
  'w-full rounded-2xl border border-border bg-surface/50 px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-primary/50 focus:bg-surface focus:ring-2 focus:ring-primary/20';

function resolveKindLabel(kind: PersonKind) {
  switch (kind) {
    case 'client':
      return 'Cliente';
    case 'employee':
      return 'Funcionário';
    case 'outsourced_employee':
      return 'Terceirizado';
    case 'supplier':
      return 'Fornecedor';
    case 'guardian':
      return 'Guardião';
    case 'responsible':
      return 'Responsável';
    default:
      return kind;
  }
}

function resolveBankAccountTypeLabel(kind: BankAccountKind) {
  switch (kind) {
    case 'checking':
      return 'Conta corrente';
    case 'savings':
      return 'Conta poupança';
    case 'salary':
      return 'Conta salário';
    default:
      return kind;
  }
}

function resolvePixKeyTypeLabel(kind?: PixKeyKind | null) {
  switch (kind) {
    case 'cpf':
      return 'CPF';
    case 'cnpj':
      return 'CNPJ';
    case 'email':
      return 'Email';
    case 'phone':
      return 'Telefone';
    case 'random':
      return 'Chave aleatória';
    default:
      return 'Tipo não informado';
  }
}

function maskCpf(cpf?: string | null) {
  const digits = String(cpf ?? '').replace(/\D/g, '');
  if (digits.length === 0) {
    return 'Não informado';
  }

  if (digits.length <= 4) {
    return digits;
  }

  const visiblePrefix = digits.slice(0, 3);
  const visibleSuffix = digits.slice(-2);
  const maskedMiddle = '*'.repeat(Math.max(digits.length - 5, 0));

  return `${visiblePrefix}${maskedMiddle}${visibleSuffix}`;
}

function resolvePetKindLabel(kind: PetDTO['kind']) {
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
