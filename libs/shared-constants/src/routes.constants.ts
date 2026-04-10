export const APP_ROUTES = {
  home: '/',
  login: '/login',
  dashboard: '/$companySlug/dashboard',
  schedules: '/$companySlug/schedules',
} as const;

export const APP_ROUTE_SEGMENTS = {
  login: 'login',
  dashboard: 'dashboard',
  schedules: 'schedules',
} as const;

export const COMPANY_ROUTE_PARAM = 'companySlug' as const;

export const COMPANY_ROUTE_PATTERNS = {
  dashboard: '/$companySlug/dashboard',
  schedules: '/$companySlug/schedules',
} as const;

// `companySlug` is URL/UX context only. Authorization remains JWT + company_id.
// See docs/conventions/company-slug-routing.md for the routing convention.
export function normalizeCompanySlug(companySlug: string) {
  return companySlug.trim().toLowerCase();
}

export function buildCompanyRoute(
  companySlug: string,
  route: keyof typeof COMPANY_ROUTE_PATTERNS,
) {
  return `/${normalizeCompanySlug(companySlug)}/${APP_ROUTE_SEGMENTS[route]}`;
}
