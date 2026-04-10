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

  it('documents the tenant-scoped route convention for the upcoming migration', () => {
    expect(COMPANY_ROUTE_PATTERNS.dashboard).toBe('/:companySlug/dashboard');
    expect(COMPANY_ROUTE_PATTERNS.schedules).toBe('/:companySlug/schedules');
    expect(buildCompanyRoute('company-x', 'schedules')).toBe(
      '/company-x/schedules',
    );
  });

  it('keeps /login without slug', () => {
    expect(APP_ROUTES.login).toBe('/login');
  });
});
