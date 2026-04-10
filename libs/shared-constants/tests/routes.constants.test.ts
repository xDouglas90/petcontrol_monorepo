import { describe, expect, it } from 'vitest';

import {
  APP_ROUTES,
  COMPANY_ROUTE_PATTERNS,
  buildCompanyRoute,
} from '../src';

describe('APP_ROUTES', () => {
  it('defines authenticated routes with companySlug in the path', () => {
    expect(APP_ROUTES.dashboard).toBe('/$companySlug/dashboard');
    expect(APP_ROUTES.schedules).toBe('/$companySlug/schedules');
  });

  it('documents the tenant-scoped route convention', () => {
    expect(COMPANY_ROUTE_PATTERNS.dashboard).toBe('/$companySlug/dashboard');
    expect(COMPANY_ROUTE_PATTERNS.schedules).toBe('/$companySlug/schedules');
  });

  it('keeps /login without slug', () => {
    expect(APP_ROUTES.login).toBe('/login');
  });
});

describe('buildCompanyRoute', () => {
  it('builds a dashboard route correctly', () => {
    expect(buildCompanyRoute('my-company', 'dashboard')).toBe(
      '/my-company/dashboard',
    );
  });

  it('builds a schedules route correctly', () => {
    expect(buildCompanyRoute('test-slug', 'schedules')).toBe(
      '/test-slug/schedules',
    );
  });

  it('handles custom slugs and preserves them', () => {
    expect(buildCompanyRoute('PETCONTROL-DEV', 'dashboard')).toBe(
      '/PETCONTROL-DEV/dashboard',
    );
  });
});
