import { describe, expect, it } from 'vitest';

import {
  APP_ROUTE_SEGMENTS,
  APP_ROUTES,
  API_PATHS,
  AUTH_MODES,
  COMPANY_ROUTE_PARAM,
  COMPANY_ROUTE_PATTERNS,
  STORAGE_KEYS,
  buildCompanyRoute,
} from '../src';

describe('shared-constants', () => {
  it('expõe rotas e segmentos esperados', () => {
    expect(APP_ROUTES.home).toBe('/');
    expect(APP_ROUTES.login).toBe('/login');
    expect(APP_ROUTES.dashboard).toBe('/dashboard');
    expect(APP_ROUTES.schedules).toBe('/schedules');

    expect(APP_ROUTE_SEGMENTS.login).toBe('login');
    expect(APP_ROUTE_SEGMENTS.dashboard).toBe('dashboard');
    expect(APP_ROUTE_SEGMENTS.schedules).toBe('schedules');

    expect(COMPANY_ROUTE_PARAM).toBe('companySlug');
    expect(COMPANY_ROUTE_PATTERNS.dashboard).toBe('/:companySlug/dashboard');
    expect(COMPANY_ROUTE_PATTERNS.schedules).toBe('/:companySlug/schedules');
    expect(buildCompanyRoute('company-x', 'dashboard')).toBe(
      '/company-x/dashboard',
    );
  });

  it('expõe paths e chaves de storage estáveis', () => {
    expect(API_PATHS.authLogin).toBe('/auth/login');
    expect(STORAGE_KEYS.auth).toBe('petcontrol-web-auth');
    expect(STORAGE_KEYS.ui).toBe('petcontrol-web-ui');
  });

  it('expõe modos de autenticação suportados', () => {
    expect(AUTH_MODES.api).toBe('api');
    expect(AUTH_MODES.mock).toBe('mock');
  });
});
