import { describe, expect, it } from 'vitest';

import {
  APP_ROUTES,
  COMPANY_ROUTE_PATTERNS,
  buildCompanyRoute,
} from '../src';

describe('APP_ROUTES', () => {
  it('keeps current live app routes unchanged until the router migration starts', () => {
    expect(APP_ROUTES.dashboard).toBe('/dashboard');
    expect(APP_ROUTES.schedules).toBe('/schedules');
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
