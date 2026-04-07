export type UserRole = 'admin' | 'manager' | 'staff' | string;

export type UserKind = 'owner' | 'employee' | string;

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginSession {
  accessToken: string;
  tokenType: string;
  userId: string;
  companyId: string;
  role: UserRole;
  kind: UserKind;
}

export interface LoginApiResponseDTO {
  data: {
    access_token: string;
    token_type: string;
    user_id: string;
    company_id: string;
    role: UserRole;
    kind: UserKind;
  };
}

export interface ApiErrorPayloadDTO {
  error?: string;
  message?: string;
}
