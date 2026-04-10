import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';

import { queryClientForWeb, router } from '@/router';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseCurrentCompanyQuery = vi.fn();
const mockUseSchedulesQuery = vi.fn();

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useSchedulesQuery: () => mockUseSchedulesQuery(),
  domainQueryKeys: {
    currentCompany: () => ['domain', 'company', 'current'] as const,
    schedules: () => ['domain', 'schedules'] as const,
  },
}));

describe('Login flow integration', () => {
  beforeEach(async () => {
    localStorage.clear();
    vi.stubGlobal('scrollTo', vi.fn());
    mockUseCurrentCompanyQuery.mockReset();
    mockUseSchedulesQuery.mockReset();

    useAuthStore.setState({
      session: {
        accessToken: 'token-123',
        tokenType: 'Bearer',
        userId: 'user-1',
        companyId: 'company-1',
        role: 'admin',
        kind: 'owner',
      },
      hydrated: true,
    });
    useUIStore.setState({
      sidebarOpen: true,
      theme: 'midnight',
    });

    mockUseCurrentCompanyQuery.mockReturnValue({
      data: {
        id: 'company-1',
        slug: 'petcontrol-dev',
        name: 'PetControl Desenvolvimento LTDA',
        fantasy_name: 'PetControl Dev',
        cnpj: '12345678000195',
        active_package: 'starter',
        is_active: true,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });

    mockUseSchedulesQuery.mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    });

    await router.navigate({ to: '/' });
  });

  afterEach(() => {
    cleanup();
    queryClientForWeb().clear();
    vi.unstubAllGlobals();
  });

  it('com sessão persistida resolve a empresa corrente e vai para /:companySlug/dashboard', async () => {
    render(
      <QueryClientProvider client={queryClientForWeb()}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/petcontrol-dev/dashboard');
    });

    expect(
      screen.getByText('Dashboard conectado ao backend com dados reais de tenant, empresa e agendamentos.'),
    ).toBeTruthy();
    expect(screen.getByText('PetControl Dev')).toBeTruthy();
  });
});
