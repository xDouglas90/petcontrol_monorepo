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

export const MODULE_PACKAGES = [
  'internal',
  'starter',
  'basic',
  'essential',
  'premium',
] as const;

export type ModulePackage = (typeof MODULE_PACKAGES)[number];

export const MODULE_CODES = ['SCH', 'CRM', 'FIN'] as const;

export type ModuleCode = (typeof MODULE_CODES)[number];

export interface CompanyDTO {
  id: UUID;
  slug: string;
  name: string;
  fantasy_name: string;
  cnpj: string;
  active_package: ModulePackage;
  is_active: boolean;
}

export interface CurrentCompanyApiResponseDTO {
  data: CompanyDTO;
}

export interface ClientDTO {
  id: UUID;
  person_id: UUID;
  company_id: UUID;
  full_name: string;
  short_name: string;
  cpf: string;
  email: string;
  cellphone: string;
  has_whatsapp: boolean;
  client_since?: string | null;
  notes?: string | null;
  is_active: boolean;
}

export interface ClientListApiResponseDTO {
  data: ClientDTO[];
}

export interface ClientApiResponseDTO {
  data: ClientDTO;
}

export const PET_SIZES = ['small', 'medium', 'large', 'giant'] as const;

export type PetSize = (typeof PET_SIZES)[number];

export const PET_KINDS = [
  'dog',
  'cat',
  'bird',
  'fish',
  'reptile',
  'rodent',
  'rabbit',
  'other',
] as const;

export type PetKind = (typeof PET_KINDS)[number];

export const PET_TEMPERAMENTS = [
  'calm',
  'nervous',
  'aggressive',
  'playful',
  'loving',
] as const;

export type PetTemperament = (typeof PET_TEMPERAMENTS)[number];

export interface PetDTO {
  id: UUID;
  owner_id: UUID;
  name: string;
  size: PetSize;
  kind: PetKind;
  temperament: PetTemperament;
  image_url?: string | null;
  birth_date?: string | null;
  is_active: boolean;
  notes?: string | null;
}

export interface PetListApiResponseDTO {
  data: PetDTO[];
}

export interface PetApiResponseDTO {
  data: PetDTO;
}

export interface ServiceTypeDTO {
  id: UUID;
  name: string;
  description?: string | null;
}

export interface ServiceDTO {
  id: UUID;
  type_id: UUID;
  type_name: string;
  title: string;
  description: string;
  notes?: string | null;
  price: string;
  discount_rate: string;
  image_url?: string | null;
  is_active: boolean;
}

export interface ServiceListApiResponseDTO {
  data: ServiceDTO[];
}

export interface ServiceApiResponseDTO {
  data: ServiceDTO;
}

export const SCHEDULE_STATUSES = [
  'waiting',
  'confirmed',
  'canceled',
  'in_progress',
  'finished',
  'delivered',
] as const;

export type ScheduleStatus = (typeof SCHEDULE_STATUSES)[number];

export interface ScheduleDTO {
  id: UUID;
  company_id: UUID;
  client_id: UUID;
  pet_id: UUID;
  client_name?: string | null;
  pet_name?: string | null;
  service_ids?: UUID[];
  service_titles?: string[];
  scheduled_at: string;
  estimated_end?: string | null;
  notes?: string | null;
  current_status: ScheduleStatus;
}

export interface ScheduleListApiResponseDTO {
  data: ScheduleDTO[];
}

export interface ScheduleApiResponseDTO {
  data: ScheduleDTO;
}

export interface CreateScheduleInput {
  client_id: UUID;
  pet_id: UUID;
  scheduled_at: string;
  estimated_end?: string;
  notes?: string;
  status?: ScheduleStatus;
  status_notes?: string;
}

export interface UpdateScheduleInput {
  client_id?: UUID;
  pet_id?: UUID;
  scheduled_at?: string;
  estimated_end?: string;
  notes?: string;
  status?: ScheduleStatus;
  status_notes?: string;
}
