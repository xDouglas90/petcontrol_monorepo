import { API_PATHS, AUTH_MODES } from '@petcontrol/shared-constants';
import type {
  ApiErrorPayloadDTO,
  CompanyDTO,
  CreateScheduleInput,
  CurrentCompanyApiResponseDTO,
  LoginApiResponseDTO,
  LoginCredentials,
  LoginSession,
  ScheduleApiResponseDTO,
  ScheduleDTO,
  ScheduleListApiResponseDTO,
  UpdateScheduleInput,
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
    scheduled_at: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
    estimated_end: new Date(Date.now() + 3 * 60 * 60 * 1000).toISOString(),
    notes: 'Banho e hidratação',
    current_status: 'waiting',
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

export async function getCurrentCompany(accessToken: string): Promise<CompanyDTO> {
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

export async function listSchedules(accessToken: string): Promise<ScheduleDTO[]> {
  if (authMode === AUTH_MODES.mock) {
    await delay(160);
    return [...mockSchedules].sort((a, b) =>
      a.scheduled_at.localeCompare(b.scheduled_at),
    );
  }

  const payload = await request<ScheduleListApiResponseDTO>(API_PATHS.schedules, {
    method: 'GET',
    accessToken,
  });
  return payload.data;
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
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  accessToken: string;
  body?: unknown;
};

async function request<T>(path: string, options: RequestOptions): Promise<T> {
  const response = await fetch(`${apiUrl}${path}`, {
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
