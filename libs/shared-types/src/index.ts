export const USER_ROLES = [
  'root',
  'admin',
  'manager',
  'employee',
  'aux',
  'general',
] as const;

export type UserRole = (typeof USER_ROLES)[number];

export const USER_KINDS = ['internal', 'owner', 'staff', 'free'] as const;

export type UserKind = (typeof USER_KINDS)[number];

export const TOKEN_TYPES = ['Bearer'] as const;

export type TokenType = (typeof TOKEN_TYPES)[number];

export type UUID = string;

export interface AuthAccessClaims {
  user_id: UUID;
  company_id: UUID;
  role: UserRole;
  kind: UserKind;
  sub: UUID;
  iat: number;
  exp: number;
}

export interface TenantContext {
  companyId: UUID;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginSession {
  accessToken: string;
  tokenType: TokenType;
  userId: UUID;
  companyId: UUID;
  role: UserRole;
  kind: UserKind;
}

export interface LoginApiResponseDTO {
  data: {
    access_token: string;
    token_type: TokenType;
    user_id: UUID;
    company_id: UUID;
    role: UserRole;
    kind: UserKind;
  };
}

export interface ApiErrorPayloadDTO {
  error?: string;
  message?: string;
}
