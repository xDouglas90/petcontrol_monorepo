export const APP_ROUTES = {
  home: '/',
  login: '/login',
  dashboard: '/dashboard',
} as const;

export const APP_ROUTE_SEGMENTS = {
  login: 'login',
  dashboard: 'dashboard',
} as const;

export const API_PATHS = {
  authLogin: '/auth/login',
} as const;

export const STORAGE_KEYS = {
  auth: 'petcontrol-web-auth',
  ui: 'petcontrol-web-ui',
} as const;

export const AUTH_MODES = {
  api: 'api',
  mock: 'mock',
} as const;
