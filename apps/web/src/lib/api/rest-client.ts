import { API_PATHS, AUTH_MODES } from '@petcontrol/shared-constants';
import type {
  AdminSystemChatMessageDTO,
  AdminSystemChatMessageListApiResponseDTO,
  ApiErrorPayloadDTO,
  ClientDTO,
  ClientApiResponseDTO,
  ClientListApiResponseDTO,
  CompanyDTO,
  CompanyUserDTO,
  CompanyUserListApiResponseDTO,
  CompanyUserPermissionDTO,
  CompanyUserPermissionsApiResponseDTO,
  CompanyUserPermissionsDTO,
  CompanySystemConfigDTO,
  CurrentUserApiResponseDTO,
  CurrentCompanySystemConfigApiResponseDTO,
  CurrentUserDTO,
  CompleteUploadApiResponseDTO,
  CompleteUploadDTO,
  CompleteUploadInput,
  CreateAdminSystemChatMessageInput,
  CreateClientInput,
  CreatePetInput,
  CreateScheduleInput,
  CreateServiceInput,
  ListQueryParams,
  PetDTO,
  PetApiResponseDTO,
  PetListApiResponseDTO,
  CurrentCompanyApiResponseDTO,
  LoginApiResponseDTO,
  LoginCredentials,
  LoginSession,
  ScheduleApiResponseDTO,
  ScheduleHistoryApiResponseDTO,
  ScheduleHistoryItemDTO,
  ScheduleDTO,
  ScheduleListApiResponseDTO,
  ServiceDTO,
  ServiceApiResponseDTO,
  ServiceListApiResponseDTO,
  UpdateClientInput,
  UpdateCompanyUserPermissionsInput,
  UpdateCurrentCompanyInput,
  UpdateCurrentCompanySystemConfigInput,
  UpdatePetInput,
  UpdateScheduleInput,
  UpdateServiceInput,
  CreateUploadIntentInput,
  UploadIntentDTO,
  UploadIntentApiResponseDTO,
  HealthDTO,
} from '@petcontrol/shared-types';
import {
  isNonEmptyTrimmed,
  normalizeUrl,
  safeLowerCase,
} from '@petcontrol/shared-utils';

const apiUrl = normalizeUrl(
  import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1',
);
const authMode = (import.meta.env.VITE_AUTH_MODE ?? 'api').toLowerCase();

const mockCompany: CompanyDTO = {
  id: '22222222-2222-2222-2222-222222222222',
  slug: 'petcontrol-dev',
  name: 'PetControl Dev Company',
  fantasy_name: 'PetControl Dev',
  cnpj: '12345678000195',
  active_package: 'starter',
  is_active: true,
};

let mockSchedules: ScheduleDTO[] = [
  {
    id: '33333333-3333-3333-3333-333333333333',
    company_id: mockCompany.id,
    client_id: '44444444-4444-4444-4444-444444444444',
    pet_id: '55555555-5555-5555-5555-555555555555',
    client_name: 'Maria Silva',
    pet_name: 'Thor',
    service_ids: ['66666666-6666-6666-6666-666666666666'],
    service_titles: ['Banho completo'],
    scheduled_at: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
    estimated_end: new Date(Date.now() + 3 * 60 * 60 * 1000).toISOString(),
    notes: 'Banho e hidratação',
    current_status: 'waiting',
  },
];

let mockClients: ClientDTO[] = [
  {
    id: '44444444-4444-4444-4444-444444444444',
    person_id: '77777777-7777-7777-7777-777777777777',
    company_id: mockCompany.id,
    full_name: 'Maria Silva',
    short_name: 'Maria',
    gender_identity: 'woman_cisgender',
    marital_status: 'single',
    birth_date: '1992-06-15',
    cpf: '12345678901',
    email: 'maria@petcontrol.local',
    cellphone: '+5511999990001',
    has_whatsapp: true,
    is_active: true,
  },
];

let mockPets: PetDTO[] = [
  {
    id: '55555555-5555-5555-5555-555555555555',
    owner_id: mockClients[0].id,
    company_id: mockCompany.id,
    owner_name: mockClients[0].full_name,
    name: 'Thor',
    size: 'medium',
    kind: 'dog',
    temperament: 'playful',
    is_active: true,
  },
];

let mockServices: ServiceDTO[] = [
  {
    id: '66666666-6666-6666-6666-666666666666',
    type_id: '88888888-8888-8888-8888-888888888888',
    type_name: 'Banho',
    title: 'Banho completo',
    description: 'Banho com secagem e perfume',
    price: '89.90',
    discount_rate: '0.00',
    is_active: true,
  },
];

let mockAdminSystemChatMessages: AdminSystemChatMessageDTO[] = [
  {
    id: 'chat-message-1',
    conversation_id: 'chat-conversation-1',
    company_id: mockCompany.id,
    sender_user_id: '11111111-1111-1111-1111-111111111111',
    sender_name: 'Maria',
    sender_role: 'admin',
    sender_image_url: null,
    body: 'Bom dia, preciso acompanhar a operação desta semana.',
    created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: 'chat-message-2',
    conversation_id: 'chat-conversation-1',
    company_id: mockCompany.id,
    sender_user_id: 'system-user-1',
    sender_name: 'System',
    sender_role: 'system',
    sender_image_url: null,
    body: 'Tudo certo. O monitoramento do tenant já está ativo.',
    created_at: new Date(Date.now() - 23 * 60 * 60 * 1000).toISOString(),
  },
];

export class ApiError extends Error {
  status: number;
  details: unknown;

  constructor(message: string, status: number, details: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

export function isUnauthorizedApiError(error: unknown): error is ApiError {
  return error instanceof ApiError && error.status === 401;
}

export function getApiUrl() {
  return apiUrl;
}

export function getAuthMode() {
  return authMode;
}

export async function login(
  credentials: LoginCredentials,
): Promise<LoginSession> {
  if (authMode === AUTH_MODES.mock) {
    return mockLogin(credentials);
  }

  const response = await fetch(`${apiUrl}${API_PATHS.authLogin}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  const payload = await readJson(response);
  if (!response.ok) {
    throw new ApiError(
      extractMessage(payload) ?? 'Falha ao autenticar',
      response.status,
      payload,
    );
  }

  return mapLoginSession((payload as LoginApiResponseDTO).data);
}

export async function checkHealth(): Promise<HealthDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(150);
    return {
      status: 'ok',
      timestamp: new Date().toISOString(),
      version: 'mock-1.0.0',
    };
  }

  try {
    const baseUrl = new URL(apiUrl).origin;
    const healthUrl = `${baseUrl}${API_PATHS.health}`;
    const response = await fetch(healthUrl);

    if (!response.ok) {
      throw new Error('API indisponível');
    }

    return response.json();
  } catch {
    throw new Error('Erro ao verificar saúde da API');
  }
}

export async function createUploadIntent(
  accessToken: string,
  input: CreateUploadIntentInput,
): Promise<UploadIntentDTO> {
  const payload = await request<UploadIntentApiResponseDTO>(
    API_PATHS.uploadsIntent,
    {
      method: 'POST',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function uploadToGCS(
  intent: Pick<UploadIntentDTO, 'upload_url' | 'method' | 'headers'>,
  file: File,
): Promise<void> {
  const headers = new Headers(intent.headers ?? {});
  if (!headers.has('Content-Type') && file.type) {
    headers.set('Content-Type', file.type);
  }

  const response = await fetch(intent.upload_url, {
    method: intent.method,
    headers,
    body: file,
  });

  if (!response.ok) {
    throw new Error('Falha ao enviar arquivo para o GCS');
  }
}

export async function completeUpload(
  accessToken: string,
  input: CompleteUploadInput,
): Promise<CompleteUploadDTO> {
  const payload = await request<CompleteUploadApiResponseDTO>(
    API_PATHS.uploadsComplete,
    {
      method: 'POST',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function getCurrentCompany(
  accessToken: string,
): Promise<CompanyDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    return mockCompany;
  }

  const payload = await request<CurrentCompanyApiResponseDTO>(
    API_PATHS.currentCompany,
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function updateCurrentCompany(
  accessToken: string,
  input: UpdateCurrentCompanyInput,
): Promise<CompanyDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    Object.assign(mockCompany, {
      ...mockCompany,
      ...(input.name !== undefined ? { name: input.name } : {}),
      ...(input.fantasy_name !== undefined
        ? { fantasy_name: input.fantasy_name }
        : {}),
      ...(input.logo_url !== undefined ? { logo_url: input.logo_url } : {}),
    });
    return mockCompany;
  }

  const payload = await request<CurrentCompanyApiResponseDTO>(
    API_PATHS.currentCompany,
    {
      method: 'PATCH',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function getCurrentCompanySystemConfig(
  accessToken: string,
): Promise<CompanySystemConfigDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(140);
    return {
      company_id: mockCompany.id,
      schedule_init_time: '08:00',
      schedule_pause_init_time: '12:00',
      schedule_pause_end_time: '13:00',
      schedule_end_time: '18:00',
      min_schedules_per_day: 4,
      max_schedules_per_day: 18,
      schedule_days: [
        'monday',
        'tuesday',
        'wednesday',
        'thursday',
        'friday',
        'saturday',
      ],
      dynamic_cages: false,
      total_small_cages: 8,
      total_medium_cages: 6,
      total_large_cages: 4,
      total_giant_cages: 2,
      whatsapp_notifications: true,
      whatsapp_conversation: true,
      whatsapp_business_phone: '+5511999990001',
    };
  }

  const payload = await request<CurrentCompanySystemConfigApiResponseDTO>(
    API_PATHS.currentCompanySystemConfig,
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function updateCurrentCompanySystemConfig(
  accessToken: string,
  input: UpdateCurrentCompanySystemConfigInput,
): Promise<CompanySystemConfigDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    return {
      company_id: mockCompany.id,
      schedule_init_time: input.schedule_init_time,
      schedule_pause_init_time: input.schedule_pause_init_time ?? null,
      schedule_pause_end_time: input.schedule_pause_end_time ?? null,
      schedule_end_time: input.schedule_end_time,
      min_schedules_per_day: input.min_schedules_per_day,
      max_schedules_per_day: input.max_schedules_per_day,
      schedule_days: input.schedule_days,
      dynamic_cages: input.dynamic_cages,
      total_small_cages: input.total_small_cages,
      total_medium_cages: input.total_medium_cages,
      total_large_cages: input.total_large_cages,
      total_giant_cages: input.total_giant_cages,
      whatsapp_notifications: input.whatsapp_notifications,
      whatsapp_conversation: input.whatsapp_conversation,
      whatsapp_business_phone: input.whatsapp_business_phone ?? null,
    };
  }

  const payload = await request<CurrentCompanySystemConfigApiResponseDTO>(
    API_PATHS.currentCompanySystemConfig,
    {
      method: 'PATCH',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function getCurrentUser(
  accessToken: string,
): Promise<CurrentUserDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    return {
      user_id: '11111111-1111-1111-1111-111111111111',
      company_id: mockCompany.id,
      person_id: '77777777-7777-7777-7777-777777777777',
      role: 'admin',
      kind: 'owner',
      full_name: 'Maria Silva',
      short_name: 'Maria',
      image_url: null,
      settings_access: {
        can_view: true,
        can_manage_permissions: true,
        active_permission_codes: [
          'company_settings:edit',
          'plan_settings:edit',
          'payment_settings:edit',
          'notification_settings:edit',
          'integration_settings:edit',
          'security_settings:edit',
        ],
        editable_permission_codes: [
          'company_settings:edit',
          'plan_settings:edit',
          'payment_settings:edit',
          'notification_settings:edit',
          'integration_settings:edit',
          'security_settings:edit',
        ],
      },
    };
  }

  const payload = await request<CurrentUserApiResponseDTO>(
    API_PATHS.currentUser,
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function listSchedules(
  accessToken: string,
  params?: ListQueryParams,
): Promise<ScheduleListApiResponseDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(160);
    const mockData = [...mockSchedules].sort((a, b) =>
      a.scheduled_at.localeCompare(b.scheduled_at),
    );
    return {
      data: mockData,
      meta: { total: mockData.length, page: 1, limit: 100, total_pages: 1 },
    };
  }

  const payload = await request<ScheduleListApiResponseDTO>(
    API_PATHS.schedules,
    {
      method: 'GET',
      accessToken,
      queryParams: params,
    },
  );
  return payload;
}

export async function getScheduleHistory(
  accessToken: string,
  scheduleId: string,
): Promise<ScheduleHistoryItemDTO[]> {
  if (authMode === AUTH_MODES.mock) {
    await delay(100);
    return [];
  }

  const payload = await request<ScheduleHistoryApiResponseDTO>(
    API_PATHS.scheduleHistory(scheduleId),
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function listCompanyUsers(
  accessToken: string,
): Promise<CompanyUserDTO[]> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    return [
      {
        id: 'company-user-system-1',
        company_id: mockCompany.id,
        user_id: 'system-user-1',
        kind: 'employee',
        role: 'system',
        is_owner: false,
        is_active: true,
        full_name: 'System PetControl',
        short_name: 'System',
        image_url: null,
        joined_at: new Date().toISOString(),
        left_at: null,
      },
    ];
  }

  const payload = await request<CompanyUserListApiResponseDTO>(
    API_PATHS.companyUsers,
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function getCompanyUserPermissions(
  accessToken: string,
  userId: string,
): Promise<CompanyUserPermissionsDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    const permissions: CompanyUserPermissionDTO[] = [
      {
        id: crypto.randomUUID(),
        code: 'company_settings:edit',
        description: 'Editar configurações gerais',
        default_roles: ['root', 'admin'],
        is_active: false,
        is_default_for_role: false,
        granted_by: null,
        granted_at: null,
      },
      {
        id: crypto.randomUUID(),
        code: 'plan_settings:edit',
        description: 'Editar configurações de plano',
        default_roles: ['root', 'admin', 'system'],
        is_active: true,
        is_default_for_role: userId === 'system-user-1',
        granted_by: '11111111-1111-1111-1111-111111111111',
        granted_at: new Date().toISOString(),
      },
    ];
    return {
      user_id: userId,
      company_id: mockCompany.id,
      active_package: mockCompany.active_package,
      role: userId === 'system-user-1' ? 'system' : 'admin',
      kind: 'employee',
      is_owner: false,
      is_active: true,
      managed_by: '11111111-1111-1111-1111-111111111111',
      scope: 'tenant_settings',
      permissions,
      permission_groups: [
        {
          module_code: 'CFG',
          module_name: 'Configurações',
          module_description:
            'Configurações institucionais, plano, pagamentos, notificações, integrações e segurança do tenant.',
          min_package: 'starter',
          permissions,
        },
      ],
    };
  }

  const payload = await request<CompanyUserPermissionsApiResponseDTO>(
    API_PATHS.companyUserPermissions(userId),
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function updateCompanyUserPermissions(
  accessToken: string,
  userId: string,
  input: UpdateCompanyUserPermissionsInput,
): Promise<CompanyUserPermissionsDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(140);
    const current = await getCompanyUserPermissions(accessToken, userId);
    const desired = new Set(input.permission_codes);
    const permissions = current.permissions.map((permission) => ({
      ...permission,
      is_active: desired.has(permission.code),
      granted_at: desired.has(permission.code)
        ? new Date().toISOString()
        : permission.granted_at,
    }));

    return {
      ...current,
      permissions,
      permission_groups: current.permission_groups.map((group) => ({
        ...group,
        permissions: group.permissions.map((permission) => {
          const updated = permissions.find((item) => item.code === permission.code);
          return updated ?? permission;
        }),
      })),
    };
  }

  const payload = await request<CompanyUserPermissionsApiResponseDTO>(
    API_PATHS.companyUserPermissions(userId),
    {
      method: 'PATCH',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function listAdminSystemChatMessages(
  accessToken: string,
  userId: string,
): Promise<AdminSystemChatMessageDTO[]> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    if (userId !== 'system-user-1') {
      return [];
    }
    return [...mockAdminSystemChatMessages];
  }

  const payload = await request<AdminSystemChatMessageListApiResponseDTO>(
    API_PATHS.adminSystemChatMessages(userId),
    {
      method: 'GET',
      accessToken,
    },
  );
  return payload.data;
}

export async function createAdminSystemChatMessage(
  accessToken: string,
  userId: string,
  input: CreateAdminSystemChatMessageInput,
): Promise<AdminSystemChatMessageDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(140);
    const created: AdminSystemChatMessageDTO = {
      id: crypto.randomUUID(),
      conversation_id: 'chat-conversation-1',
      company_id: mockCompany.id,
      sender_user_id: '11111111-1111-1111-1111-111111111111',
      sender_name: 'Maria',
      sender_role: 'admin',
      sender_image_url: null,
      body: input.message,
      created_at: new Date().toISOString(),
    };
    if (userId === 'system-user-1') {
      mockAdminSystemChatMessages = [...mockAdminSystemChatMessages, created];
    }
    return created;
  }

  const payload = await request<{ data: AdminSystemChatMessageDTO }>(
    API_PATHS.adminSystemChatMessages(userId),
    {
      method: 'POST',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function listClients(
  accessToken: string,
  params?: ListQueryParams,
): Promise<ClientListApiResponseDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    return {
      data: [...mockClients],
      meta: { total: mockClients.length, page: 1, limit: 100, total_pages: 1 },
    };
  }

  const payload = await request<ClientListApiResponseDTO>(API_PATHS.clients, {
    method: 'GET',
    accessToken,
    queryParams: params,
  });
  return payload;
}

export async function listPets(
  accessToken: string,
  params?: ListQueryParams,
): Promise<PetListApiResponseDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    return {
      data: [...mockPets],
      meta: { total: mockPets.length, page: 1, limit: 100, total_pages: 1 },
    };
  }

  const payload = await request<PetListApiResponseDTO>(API_PATHS.pets, {
    method: 'GET',
    accessToken,
    queryParams: params,
  });
  return payload;
}

export async function listServices(
  accessToken: string,
  params?: ListQueryParams,
): Promise<ServiceListApiResponseDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    return {
      data: [...mockServices],
      meta: { total: mockServices.length, page: 1, limit: 100, total_pages: 1 },
    };
  }

  const payload = await request<ServiceListApiResponseDTO>(API_PATHS.services, {
    method: 'GET',
    accessToken,
    queryParams: params,
  });
  return payload;
}

export async function createClient(
  accessToken: string,
  input: CreateClientInput,
): Promise<ClientDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    const client: ClientDTO = {
      id: crypto.randomUUID(),
      person_id: crypto.randomUUID(),
      company_id: mockCompany.id,
      full_name: input.full_name,
      short_name: input.short_name,
      gender_identity: input.gender_identity,
      marital_status: input.marital_status,
      birth_date: input.birth_date,
      cpf: input.cpf,
      email: input.email,
      phone: input.phone ?? null,
      cellphone: input.cellphone,
      has_whatsapp: input.has_whatsapp,
      client_since: input.client_since ?? null,
      notes: input.notes ?? null,
      is_active: true,
    };
    mockClients = [...mockClients, client];
    return client;
  }

  const payload = await request<ClientApiResponseDTO>(API_PATHS.clients, {
    method: 'POST',
    accessToken,
    body: input,
  });
  return payload.data;
}

export async function updateClient(
  accessToken: string,
  clientId: string,
  input: UpdateClientInput,
): Promise<ClientDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    const existing = mockClients.find((item) => item.id === clientId);
    if (!existing) {
      throw new ApiError('Cliente não encontrado', 404, { error: 'not found' });
    }
    const updated = { ...existing, ...input };
    mockClients = mockClients.map((item) =>
      item.id === clientId ? updated : item,
    );
    return updated;
  }

  const payload = await request<ClientApiResponseDTO>(
    `${API_PATHS.clients}/${clientId}`,
    {
      method: 'PUT',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function deleteClient(
  accessToken: string,
  clientId: string,
): Promise<void> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    mockClients = mockClients.filter((item) => item.id !== clientId);
    return;
  }

  await request<void>(`${API_PATHS.clients}/${clientId}`, {
    method: 'DELETE',
    accessToken,
  });
}

export async function createPet(
  accessToken: string,
  input: CreatePetInput,
): Promise<PetDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    const owner = mockClients.find((item) => item.id === input.owner_id);
    const pet: PetDTO = {
      id: crypto.randomUUID(),
      owner_id: input.owner_id,
      company_id: mockCompany.id,
      owner_name: owner?.full_name,
      name: input.name,
      size: input.size,
      kind: input.kind,
      temperament: input.temperament,
      image_url: input.image_url ?? null,
      birth_date: input.birth_date ?? null,
      is_active: true,
      notes: input.notes ?? null,
    };
    mockPets = [...mockPets, pet];
    return pet;
  }

  const payload = await request<PetApiResponseDTO>(API_PATHS.pets, {
    method: 'POST',
    accessToken,
    body: input,
  });
  return payload.data;
}

export async function updatePet(
  accessToken: string,
  petId: string,
  input: UpdatePetInput,
): Promise<PetDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    const existing = mockPets.find((item) => item.id === petId);
    if (!existing) {
      throw new ApiError('Pet não encontrado', 404, { error: 'not found' });
    }
    const ownerID = input.owner_id ?? existing.owner_id;
    const owner = mockClients.find((item) => item.id === ownerID);
    const updated: PetDTO = {
      ...existing,
      ...input,
      owner_id: ownerID,
      owner_name: owner?.full_name ?? existing.owner_name,
    };
    mockPets = mockPets.map((item) => (item.id === petId ? updated : item));
    return updated;
  }

  const payload = await request<PetApiResponseDTO>(
    `${API_PATHS.pets}/${petId}`,
    {
      method: 'PUT',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function deletePet(
  accessToken: string,
  petId: string,
): Promise<void> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    mockPets = mockPets.filter((item) => item.id !== petId);
    return;
  }

  await request<void>(`${API_PATHS.pets}/${petId}`, {
    method: 'DELETE',
    accessToken,
  });
}

export async function createService(
  accessToken: string,
  input: CreateServiceInput,
): Promise<ServiceDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    const item: ServiceDTO = {
      id: crypto.randomUUID(),
      type_id: crypto.randomUUID(),
      type_name: input.type_name,
      title: input.title,
      description: input.description,
      notes: input.notes ?? null,
      price: input.price,
      discount_rate: input.discount_rate ?? '0.00',
      image_url: input.image_url ?? null,
      is_active: input.is_active ?? true,
    };
    mockServices = [...mockServices, item];
    return item;
  }

  const payload = await request<ServiceApiResponseDTO>(API_PATHS.services, {
    method: 'POST',
    accessToken,
    body: input,
  });
  return payload.data;
}

export async function updateService(
  accessToken: string,
  serviceId: string,
  input: UpdateServiceInput,
): Promise<ServiceDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(180);
    const existing = mockServices.find((item) => item.id === serviceId);
    if (!existing) {
      throw new ApiError('Serviço não encontrado', 404, { error: 'not found' });
    }
    const updated = { ...existing, ...input };
    mockServices = mockServices.map((item) =>
      item.id === serviceId ? updated : item,
    );
    return updated;
  }

  const payload = await request<ServiceApiResponseDTO>(
    `${API_PATHS.services}/${serviceId}`,
    {
      method: 'PUT',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function deleteService(
  accessToken: string,
  serviceId: string,
): Promise<void> {
  if (authMode === AUTH_MODES.mock) {
    await delay(120);
    mockServices = mockServices.filter((item) => item.id !== serviceId);
    return;
  }

  await request<void>(`${API_PATHS.services}/${serviceId}`, {
    method: 'DELETE',
    accessToken,
  });
}

export async function createSchedule(
  accessToken: string,
  input: CreateScheduleInput,
): Promise<ScheduleDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(200);

    const schedule: ScheduleDTO = {
      id: crypto.randomUUID(),
      company_id: mockCompany.id,
      client_id: input.client_id,
      pet_id: input.pet_id,
      client_name:
        mockClients.find((item) => item.id === input.client_id)?.full_name ??
        null,
      pet_name: mockPets.find((item) => item.id === input.pet_id)?.name ?? null,
      service_ids: input.service_ids ?? [],
      service_titles: mockServices
        .filter((item) => (input.service_ids ?? []).includes(item.id))
        .map((item) => item.title),
      scheduled_at: input.scheduled_at,
      estimated_end: input.estimated_end ?? null,
      notes: input.notes ?? null,
      current_status: input.status ?? 'waiting',
    };
    mockSchedules = [...mockSchedules, schedule];
    return schedule;
  }

  const payload = await request<ScheduleApiResponseDTO>(API_PATHS.schedules, {
    method: 'POST',
    accessToken,
    body: input,
  });
  return payload.data;
}

export async function updateSchedule(
  accessToken: string,
  scheduleId: string,
  input: UpdateScheduleInput,
): Promise<ScheduleDTO> {
  if (authMode === AUTH_MODES.mock) {
    await delay(220);

    const existing = mockSchedules.find((item) => item.id === scheduleId);
    if (!existing) {
      throw new ApiError('Schedule não encontrado', 404, {
        error: 'not found',
      });
    }

    const updated: ScheduleDTO = {
      ...existing,
      client_id: input.client_id ?? existing.client_id,
      pet_id: input.pet_id ?? existing.pet_id,
      client_name:
        input.client_id == null
          ? existing.client_name
          : (mockClients.find((item) => item.id === input.client_id)
              ?.full_name ?? null),
      pet_name:
        input.pet_id == null
          ? existing.pet_name
          : (mockPets.find((item) => item.id === input.pet_id)?.name ?? null),
      service_ids: input.service_ids ?? existing.service_ids,
      service_titles:
        input.service_ids == null
          ? existing.service_titles
          : mockServices
              .filter((item) => input.service_ids?.includes(item.id))
              .map((item) => item.title),
      scheduled_at: input.scheduled_at ?? existing.scheduled_at,
      estimated_end: input.estimated_end ?? existing.estimated_end,
      notes: input.notes ?? existing.notes,
      current_status: input.status ?? existing.current_status,
    };

    mockSchedules = mockSchedules.map((item) =>
      item.id === scheduleId ? updated : item,
    );
    return updated;
  }

  const payload = await request<ScheduleApiResponseDTO>(
    `${API_PATHS.schedules}/${scheduleId}`,
    {
      method: 'PUT',
      accessToken,
      body: input,
    },
  );
  return payload.data;
}

export async function deleteSchedule(
  accessToken: string,
  scheduleId: string,
): Promise<void> {
  if (authMode === AUTH_MODES.mock) {
    await delay(140);
    mockSchedules = mockSchedules.filter((item) => item.id !== scheduleId);
    return;
  }

  await request<void>(`${API_PATHS.schedules}/${scheduleId}`, {
    method: 'DELETE',
    accessToken,
  });
}

async function mockLogin(credentials: LoginCredentials): Promise<LoginSession> {
  await delay(380);

  const email = safeLowerCase(credentials.email);
  if (!isNonEmptyTrimmed(email) || !isNonEmptyTrimmed(credentials.password)) {
    throw new ApiError('Credenciais inválidas', 422, {
      error: 'invalid payload',
    });
  }

  return {
    accessToken: `mock.${btoa(email)}.${Date.now().toString(36)}`,
    tokenType: 'Bearer',
    userId: '11111111-1111-1111-1111-111111111111',
    companyId: '22222222-2222-2222-2222-222222222222',
    role: 'admin',
    kind: 'owner',
  };
}

function mapLoginSession(payload: LoginApiResponseDTO['data']): LoginSession {
  return {
    accessToken: payload.access_token,
    tokenType: payload.token_type,
    userId: payload.user_id,
    companyId: payload.company_id,
    role: payload.role,
    kind: payload.kind,
  };
}

type RequestOptions = {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  accessToken: string;
  body?: unknown;
  queryParams?: ListQueryParams;
};

function buildQueryString(params?: ListQueryParams): string {
  if (!params) return '';
  const entries: string[] = [];
  if (params.page != null) entries.push(`page=${params.page}`);
  if (params.limit != null) entries.push(`limit=${params.limit}`);
  if (params.search)
    entries.push(`search=${encodeURIComponent(params.search)}`);
  return entries.length > 0 ? `?${entries.join('&')}` : '';
}

async function request<T>(path: string, options: RequestOptions): Promise<T> {
  const qs = buildQueryString(options.queryParams);
  const response = await fetch(`${apiUrl}${path}${qs}`, {
    method: options.method,
    headers: {
      Authorization: `Bearer ${options.accessToken}`,
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const payload = await readJson(response);
  if (!response.ok) {
    throw new ApiError(
      extractMessage(payload) ?? 'Falha na requisição',
      response.status,
      payload,
    );
  }

  return payload as T;
}

async function readJson(response: Response) {
  const raw = await response.text();
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as unknown;
  } catch {
    return raw;
  }
}

function extractMessage(payload: unknown): string | undefined {
  if (payload && typeof payload === 'object') {
    const errorPayload = payload as ApiErrorPayloadDTO;
    return errorPayload.error ?? errorPayload.message;
  }

  return undefined;
}

function delay(milliseconds: number) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, milliseconds);
  });
}
