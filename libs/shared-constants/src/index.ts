export * from './routes.constants.js';
export * from './pagination.constants.js';

export const API_PATHS = {
  authLogin: '/auth/login',
  currentCompany: '/companies/current',
  currentUser: '/users/me',
  currentCompanySystemConfig: '/company-system-configs/current',
  companyUsers: '/company-users',
  companyUserPermissions: (userId: string) => `/company-users/${userId}/permissions`,
  adminSystemChatMessages: (userId: string) => `/chat/system/${userId}/messages`,
  adminSystemChatSocket: (userId: string) => `/chat/system/${userId}/ws`,
  schedules: '/schedules',
  scheduleHistory: (scheduleId: string) => `/schedules/${scheduleId}/history`,
  people: '/people',
  peopleById: (personId: string) => `/people/${personId}`,
  clients: '/clients',
  pets: '/pets',
  petsById: (petId: string) => `/pets/${petId}`,
  services: '/services',
  uploadsIntent: '/uploads/intent',
  uploadsComplete: '/uploads/complete',
  health: '/health',
} as const;

export const MODULE_CODES = {
  people: 'PPL',
  scheduling: 'SCH',
  crm: 'CRM',
  finance: 'FIN',
} as const;

export const STORAGE_KEYS = {
  auth: 'petcontrol-web-auth',
  ui: 'petcontrol-web-ui',
} as const;

export const AUTH_MODES = {
  api: 'api',
  mock: 'mock',
} as const;

export const INTERNAL_CHAT_SOCKET = {
  subprotocol: 'petcontrol.internal-chat.v1',
  reconnectBackoffMs: [1000, 2000, 5000, 10000] as const,
  heartbeatIntervalMs: 30000,
  writeTimeoutMs: 10000,
  maxMessageBytes: 4096,
} as const;
