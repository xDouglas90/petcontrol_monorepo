import { describe, expect, it } from "vitest";

import {
  API_PATHS,
  APP_ROUTES,
  COMPANY_ROUTE_PATTERNS,
  MODULE_CODES,
  buildCompanyRoute,
  normalizeCompanySlug,
} from "../src";

describe("APP_ROUTES", () => {
  it("defines authenticated routes with companySlug in the path", () => {
    expect(APP_ROUTES.dashboard).toBe("/$companySlug/dashboard");
    expect(APP_ROUTES.schedules).toBe("/$companySlug/schedules");
    expect(APP_ROUTES.people).toBe("/$companySlug/people");
  });

  it("documents the tenant-scoped route convention", () => {
    expect(COMPANY_ROUTE_PATTERNS.dashboard).toBe("/$companySlug/dashboard");
    expect(COMPANY_ROUTE_PATTERNS.schedules).toBe("/$companySlug/schedules");
    expect(COMPANY_ROUTE_PATTERNS.people).toBe("/$companySlug/people");
  });

  it("keeps operational domain routes explicit as active patterns", () => {
    expect(COMPANY_ROUTE_PATTERNS.clients).toBe("/$companySlug/clients");
    expect(COMPANY_ROUTE_PATTERNS.people).toBe("/$companySlug/people");
    expect(COMPANY_ROUTE_PATTERNS.pets).toBe("/$companySlug/pets");
    expect(COMPANY_ROUTE_PATTERNS.services).toBe("/$companySlug/services");
  });

  it("keeps /login without slug", () => {
    expect(APP_ROUTES.login).toBe("/login");
  });
});

describe("buildCompanyRoute", () => {
  it("builds a dashboard route correctly", () => {
    expect(buildCompanyRoute("my-company", "dashboard")).toBe(
      "/my-company/dashboard",
    );
  });

  it("builds a schedules route correctly", () => {
    expect(buildCompanyRoute("test-slug", "schedules")).toBe(
      "/test-slug/schedules",
    );
  });

  it("builds operational domain routes correctly", () => {
    expect(buildCompanyRoute("test-slug", "clients")).toBe(
      "/test-slug/clients",
    );
    expect(buildCompanyRoute("test-slug", "people")).toBe(
      "/test-slug/people",
    );
    expect(buildCompanyRoute("test-slug", "pets")).toBe("/test-slug/pets");
    expect(buildCompanyRoute("test-slug", "services")).toBe(
      "/test-slug/services",
    );
  });

  it("handles custom slugs and preserves them", () => {
    expect(buildCompanyRoute("PETCONTROL-DEV", "dashboard")).toBe(
      "/petcontrol-dev/dashboard",
    );
  });

  it("normalizes company slugs to canonical lowercase routes", () => {
    expect(buildCompanyRoute("  Company-X  ", "schedules")).toBe(
      "/company-x/schedules",
    );
    expect(normalizeCompanySlug("  PETCONTROL-DEV  ")).toBe("petcontrol-dev");
  });
});

describe("domain constants", () => {
  it("exports API paths for the next operational modules", () => {
    expect(API_PATHS.clients).toBe("/clients");
    expect(API_PATHS.people).toBe("/people");
    expect(API_PATHS.pets).toBe("/pets");
    expect(API_PATHS.services).toBe("/services");
  });

  it("exports stable module codes used across the monorepo", () => {
    expect(MODULE_CODES.people).toBe("PPL");
    expect(MODULE_CODES.scheduling).toBe("SCH");
    expect(MODULE_CODES.crm).toBe("CRM");
    expect(MODULE_CODES.finance).toBe("FIN");
  });
});
