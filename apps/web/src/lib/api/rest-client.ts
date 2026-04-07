export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginSession {
  accessToken: string;
  tokenType: string;
  userId: string;
  companyId: string;
  role: string;
  kind: string;
}

interface LoginApiResponse {
  data: {
    access_token: string;
    token_type: string;
    user_id: string;
    company_id: string;
    role: string;
    kind: string;
  };
}

interface ErrorResponse {
  error?: string;
  message?: string;
}

const apiUrl = normalizeUrl(
  import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1',
);
const authMode = (import.meta.env.VITE_AUTH_MODE ?? 'api').toLowerCase();

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
  if (authMode === 'mock') {
    return mockLogin(credentials);
  }

  const response = await fetch(`${apiUrl}/auth/login`, {
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

  return mapLoginSession((payload as LoginApiResponse).data);
}

async function mockLogin(credentials: LoginCredentials): Promise<LoginSession> {
  await delay(380);

  const email = credentials.email.trim().toLowerCase();
  if (!email || !credentials.password.trim()) {
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

function mapLoginSession(payload: LoginApiResponse['data']): LoginSession {
  return {
    accessToken: payload.access_token,
    tokenType: payload.token_type,
    userId: payload.user_id,
    companyId: payload.company_id,
    role: payload.role,
    kind: payload.kind,
  };
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
    const errorPayload = payload as ErrorResponse;
    return errorPayload.error ?? errorPayload.message;
  }

  return undefined;
}

function normalizeUrl(value: string) {
  return value.replace(/\/$/, '');
}

function delay(milliseconds: number) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, milliseconds);
  });
}
