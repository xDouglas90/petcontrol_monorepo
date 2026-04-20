export const APP_ROUTES = {
  home: "/",
  login: "/login",
  dashboard: "/$companySlug/dashboard",
  schedules: "/$companySlug/schedules",
  clients: "/$companySlug/clients",
  pets: "/$companySlug/pets",
  services: "/$companySlug/services",
  settings: "/$companySlug/settings",
} as const;

export const APP_ROUTE_SEGMENTS = {
  login: "login",
  dashboard: "dashboard",
  schedules: "schedules",
  clients: "clients",
  pets: "pets",
  services: "services",
  settings: "settings",
} as const;

export const COMPANY_ROUTE_PARAM = "companySlug" as const;

export const COMPANY_ROUTE_PATTERNS = {
  dashboard: "/$companySlug/dashboard",
  schedules: "/$companySlug/schedules",
  clients: "/$companySlug/clients",
  pets: "/$companySlug/pets",
  services: "/$companySlug/services",
  settings: "/$companySlug/settings",
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
