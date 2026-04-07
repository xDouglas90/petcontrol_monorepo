import { describe, expect, it } from 'vitest';

import {
  APP_ROUTE_SEGMENTS,
  APP_ROUTES,
  API_PATHS,
  AUTH_MODES,
  STORAGE_KEYS,
} from '../src';

describe('shared-constants', () => {
  it('expõe rotas e segmentos esperados', () => {
    expect(APP_ROUTES.home).toBe('/');
    expect(APP_ROUTES.login).toBe('/login');
    expect(APP_ROUTES.dashboard).toBe('/dashboard');

    expect(APP_ROUTE_SEGMENTS.login).toBe('login');
    expect(APP_ROUTE_SEGMENTS.dashboard).toBe('dashboard');
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
