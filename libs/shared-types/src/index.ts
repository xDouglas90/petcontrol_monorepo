export type UserRole =
  | 'root'
  | 'admin'
  | 'manager'
  | 'employee'
  | 'aux'
  | 'general';

export type UserKind = 'internal' | 'owner' | 'staff' | 'free';

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
