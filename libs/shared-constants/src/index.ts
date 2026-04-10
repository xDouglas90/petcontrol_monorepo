export * from './routes.constants.js';

export const API_PATHS = {
  authLogin: '/auth/login',
  currentCompany: '/companies/current',
  schedules: '/schedules',
  clients: '/clients',
  pets: '/pets',
  services: '/services',
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
