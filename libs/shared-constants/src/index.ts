export * from './routes.constants.js';
export * from './pagination.constants.js';

export const API_PATHS = {
  authLogin: '/auth/login',
  currentCompany: '/companies/current',
  currentUser: '/users/me',
  currentCompanySystemConfig: '/company-system-configs/current',
  companyUsers: '/company-users',
  adminSystemChatMessages: (userId: string) => `/chat/system/${userId}/messages`,
  schedules: '/schedules',
  scheduleHistory: (scheduleId: string) => `/schedules/${scheduleId}/history`,
  clients: '/clients',
  pets: '/pets',
  services: '/services',
  uploadsIntent: '/uploads/intent',
  uploadsComplete: '/uploads/complete',
} as const;

export const MODULE_CODES = {
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
