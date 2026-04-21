export const USER_ROLES = [
  'root',
  'internal',
  'admin',
  'system',
  'common',
  'free',
] as const;

export type UserRole = (typeof USER_ROLES)[number];

export const USER_KINDS = [
  'owner',
  'employee',
  'client',
  'supplier',
  'outsourced_employee',
] as const;

export type UserKind = (typeof USER_KINDS)[number];

export const TOKEN_TYPES = ['Bearer'] as const;

export type TokenType = (typeof TOKEN_TYPES)[number];

export type UUID = string;

export interface PaginationMeta {
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export interface ListQueryParams {
  page?: number;
  limit?: number;
  search?: string;
}

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
  'trial',
] as const;

export type ModulePackage = (typeof MODULE_PACKAGES)[number];

export const MODULE_CODES = ['SCH', 'CRM', 'FIN'] as const;

export type ModuleCode = (typeof MODULE_CODES)[number];

export const GENDER_IDENTITIES = [
  'man_cisgender',
  'woman_cisgender',
  'transgender',
  'non_binary',
  'gender_fluid',
  'gender_queer',
  'agender',
  'gender_non_conforming',
  'not_to_expose',
] as const;

export type GenderIdentity = (typeof GENDER_IDENTITIES)[number];

export const MARITAL_STATUSES = [
  'single',
  'married',
  'divorced',
  'widowed',
  'separated',
] as const;

export type MaritalStatus = (typeof MARITAL_STATUSES)[number];

export interface CompanyDTO {
  id: UUID;
  slug: string;
  name: string;
  fantasy_name: string;
  cnpj: string;
  active_package: ModulePackage;
  is_active: boolean;
  logo_url?: string | null;
  upload_object_key?: string;
}

export interface CurrentCompanyApiResponseDTO {
  data: CompanyDTO;
}

export interface UpdateCurrentCompanyInput {
  name?: string;
  fantasy_name?: string;
  logo_url?: string | null;
  upload_object_key?: string;
}

export interface CurrentUserDTO {
  user_id: UUID;
  company_id: UUID;
  person_id: UUID;
  role: UserRole;
  kind: UserKind;
  full_name?: string | null;
  short_name?: string | null;
  image_url?: string | null;
  settings_access?: CurrentUserSettingsAccessDTO;
}

export interface CurrentUserApiResponseDTO {
  data: CurrentUserDTO;
}

export interface CurrentUserSettingsAccessDTO {
  can_view: boolean;
  can_manage_permissions: boolean;
  active_permission_codes: string[];
  editable_permission_codes: string[];
}

export const WEEK_DAYS = [
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
] as const;

export type WeekDay = (typeof WEEK_DAYS)[number];

export interface CompanySystemConfigDTO {
  company_id: UUID;
  schedule_init_time: string;
  schedule_pause_init_time?: string | null;
  schedule_pause_end_time?: string | null;
  schedule_end_time: string;
  min_schedules_per_day: number;
  max_schedules_per_day: number;
  schedule_days: WeekDay[];
  dynamic_cages: boolean;
  total_small_cages: number;
  total_medium_cages: number;
  total_large_cages: number;
  total_giant_cages: number;
  whatsapp_notifications: boolean;
  whatsapp_conversation: boolean;
  whatsapp_business_phone?: string | null;
}

export interface CurrentCompanySystemConfigApiResponseDTO {
  data: CompanySystemConfigDTO;
}

export interface UpdateCurrentCompanySystemConfigInput {
  schedule_init_time: string;
  schedule_pause_init_time?: string | null;
  schedule_pause_end_time?: string | null;
  schedule_end_time: string;
  min_schedules_per_day: number;
  max_schedules_per_day: number;
  schedule_days: WeekDay[];
  dynamic_cages: boolean;
  total_small_cages: number;
  total_medium_cages: number;
  total_large_cages: number;
  total_giant_cages: number;
  whatsapp_notifications: boolean;
  whatsapp_conversation: boolean;
  whatsapp_business_phone?: string | null;
}

export interface CompanyUserDTO {
  id: UUID;
  company_id: UUID;
  user_id: UUID;
  kind: UserKind;
  role: UserRole;
  is_owner: boolean;
  is_active: boolean;
  full_name?: string | null;
  short_name?: string | null;
  image_url?: string | null;
  joined_at: string;
  left_at?: string | null;
}

export interface CompanyUserListApiResponseDTO {
  data: CompanyUserDTO[];
}

export interface CompanyUserPermissionDTO {
  id: UUID;
  code: string;
  description?: string | null;
  default_roles: UserRole[];
  is_active: boolean;
  is_default_for_role: boolean;
  granted_by?: UUID | null;
  granted_at?: string | null;
}

export interface CompanyUserPermissionsDTO {
  user_id: UUID;
  company_id: UUID;
  role: UserRole;
  kind: UserKind;
  is_owner: boolean;
  is_active: boolean;
  managed_by: UUID;
  scope: string;
  permissions: CompanyUserPermissionDTO[];
}

export interface CompanyUserPermissionsApiResponseDTO {
  data: CompanyUserPermissionsDTO;
}

export interface UpdateCompanyUserPermissionsInput {
  permission_codes: string[];
}

export interface AdminSystemChatMessageDTO {
  id: UUID;
  conversation_id: UUID;
  company_id: UUID;
  sender_user_id: UUID;
  sender_name: string;
  sender_role: UserRole;
  sender_image_url?: string | null;
  body: string;
  created_at: string;
}

export interface AdminSystemChatMessageListApiResponseDTO {
  data: AdminSystemChatMessageDTO[];
}

export interface CreateAdminSystemChatMessageInput {
  message: string;
}

export const INTERNAL_CHAT_PRESENCE_STATUSES = [
  'online',
  'offline',
] as const;

export type InternalChatPresenceStatus =
  (typeof INTERNAL_CHAT_PRESENCE_STATUSES)[number];

export const INTERNAL_CHAT_SOCKET_EVENT_TYPES = [
  'chat.connected',
  'chat.message.created',
  'chat.presence.snapshot',
  'chat.presence.updated',
  'chat.error',
] as const;

export type InternalChatSocketEventType =
  (typeof INTERNAL_CHAT_SOCKET_EVENT_TYPES)[number];

export interface InternalChatSocketEnvelopeBase {
  type: InternalChatSocketEventType;
  company_id: UUID;
  counterpart_user_id: UUID;
  emitted_at: string;
}

export interface InternalChatPresenceDTO {
  user_id: UUID;
  status: InternalChatPresenceStatus;
  connections: number;
  last_changed_at: string;
}

export interface InternalChatSocketConnectedEvent
  extends InternalChatSocketEnvelopeBase {
  type: 'chat.connected';
  connection_id: string;
  viewer_user_id: UUID;
  viewer_role: UserRole;
}

export interface InternalChatSocketMessageCreatedEvent
  extends InternalChatSocketEnvelopeBase {
  type: 'chat.message.created';
  message: AdminSystemChatMessageDTO;
}

export interface InternalChatSocketPresenceSnapshotEvent
  extends InternalChatSocketEnvelopeBase {
  type: 'chat.presence.snapshot';
  presences: InternalChatPresenceDTO[];
}

export interface InternalChatSocketPresenceUpdatedEvent
  extends InternalChatSocketEnvelopeBase {
  type: 'chat.presence.updated';
  presence: InternalChatPresenceDTO;
}

export interface InternalChatSocketErrorEvent
  extends InternalChatSocketEnvelopeBase {
  type: 'chat.error';
  code: string;
  message: string;
}

export type InternalChatSocketEvent =
  | InternalChatSocketConnectedEvent
  | InternalChatSocketMessageCreatedEvent
  | InternalChatSocketPresenceSnapshotEvent
  | InternalChatSocketPresenceUpdatedEvent
  | InternalChatSocketErrorEvent;

export interface ClientDTO {
  id: UUID;
  person_id: UUID;
  company_id: UUID;
  full_name: string;
  short_name: string;
  gender_identity: GenderIdentity;
  marital_status: MaritalStatus;
  birth_date: string;
  cpf: string;
  email: string;
  phone?: string | null;
  cellphone: string;
  has_whatsapp: boolean;
  client_since?: string | null;
  notes?: string | null;
  is_active: boolean;
}

export interface CreateClientInput {
  full_name: string;
  short_name: string;
  gender_identity: GenderIdentity;
  marital_status: MaritalStatus;
  birth_date: string;
  cpf: string;
  email: string;
  phone?: string;
  cellphone: string;
  has_whatsapp: boolean;
  client_since?: string;
  notes?: string;
  image_url?: string;
  upload_object_key?: string;
}

export interface UpdateClientInput {
  full_name?: string;
  short_name?: string;
  gender_identity?: GenderIdentity;
  marital_status?: MaritalStatus;
  birth_date?: string;
  cpf?: string;
  email?: string;
  phone?: string;
  cellphone?: string;
  has_whatsapp?: boolean;
  client_since?: string;
  notes?: string;
  image_url?: string;
  upload_object_key?: string;
}

export interface ClientListApiResponseDTO extends PaginatedResponse<ClientDTO> {}

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
  company_id?: UUID;
  owner_name?: string;
  name: string;
  size: PetSize;
  kind: PetKind;
  temperament: PetTemperament;
  image_url?: string | null;
  birth_date?: string | null;
  is_active: boolean;
  notes?: string | null;
}

export interface CreatePetInput {
  owner_id: UUID;
  name: string;
  size: PetSize;
  kind: PetKind;
  temperament: PetTemperament;
  image_url?: string;
  upload_object_key?: string;
  birth_date?: string;
  notes?: string;
}

export interface UpdatePetInput {
  owner_id?: UUID;
  name?: string;
  size?: PetSize;
  kind?: PetKind;
  temperament?: PetTemperament;
  image_url?: string;
  upload_object_key?: string;
  birth_date?: string;
  notes?: string;
}

export interface PetListApiResponseDTO extends PaginatedResponse<PetDTO> {}

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

export interface CreateServiceInput {
  type_name: string;
  title: string;
  description: string;
  notes?: string;
  price: string;
  discount_rate?: string;
  image_url?: string;
  is_active?: boolean;
}

export interface UpdateServiceInput {
  type_name?: string;
  title?: string;
  description?: string;
  notes?: string;
  price?: string;
  discount_rate?: string;
  image_url?: string;
  is_active?: boolean;
}

export interface ServiceListApiResponseDTO extends PaginatedResponse<ServiceDTO> {}

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

export interface ScheduleListApiResponseDTO extends PaginatedResponse<ScheduleDTO> {}

export interface ScheduleApiResponseDTO {
  data: ScheduleDTO;
}

export interface ScheduleHistoryItemDTO {
  id: UUID;
  schedule_id: UUID;
  status: ScheduleStatus;
  changed_at: string;
  changed_by: UUID;
  notes?: string | null;
}

export interface ScheduleHistoryApiResponseDTO {
  data: ScheduleHistoryItemDTO[];
}

export interface CreateScheduleInput {
  client_id: UUID;
  pet_id: UUID;
  service_ids?: UUID[];
  scheduled_at: string;
  estimated_end?: string;
  notes?: string;
  status?: ScheduleStatus;
  status_notes?: string;
}

export interface UpdateScheduleInput {
  client_id?: UUID;
  pet_id?: UUID;
  service_ids?: UUID[];
  scheduled_at?: string;
  estimated_end?: string;
  notes?: string;
  status?: ScheduleStatus;
  status_notes?: string;
}

export interface CreateUploadIntentInput {
  resource: string;
  field: string;
  file_name: string;
  content_type: string;
  size_bytes: number;
}

export interface UploadIntentDTO {
  upload_url: string;
  method: string;
  headers?: Record<string, string>;
  object_key: string;
  public_url: string;
  expires_at?: string;
}

export interface UploadIntentApiResponseDTO {
  data: UploadIntentDTO;
}

export interface CompleteUploadInput {
  resource: string;
  field: string;
  object_key: string;
}

export interface CompleteUploadDTO {
  object_key: string;
  public_url: string;
}

export interface CompleteUploadApiResponseDTO {
  data: CompleteUploadDTO;
}

export interface HealthDTO {
  status: string;
  timestamp: string;
  version?: string;
}
