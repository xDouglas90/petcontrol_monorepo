import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';
import type { CompanyDTO, ScheduleDTO } from '@petcontrol/shared-types';

import { queryClientForWeb, router } from '@/router';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseCurrentCompanyQuery = vi.fn();
const mockUseSchedulesQuery = vi.fn();
const mockCreateScheduleMutation = vi.fn();
const mockUpdateScheduleMutation = vi.fn();
const mockDeleteScheduleMutation = vi.fn();

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useSchedulesQuery: () => mockUseSchedulesQuery(),
  useCreateScheduleMutation: () => mockCreateScheduleMutation(),
  useUpdateScheduleMutation: () => mockUpdateScheduleMutation(),
  useDeleteScheduleMutation: () => mockDeleteScheduleMutation(),
  domainQueryKeys: {
    currentCompany: () => ['domain', 'company', 'current'] as const,
    schedules: () => ['domain', 'schedules'] as const,
  },
}));

describe('Router integration', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.stubGlobal('scrollTo', vi.fn());

    const company: CompanyDTO = {
      id: 'company-1',
      slug: 'petcontrol-dev',
      name: 'PetControl Desenvolvimento LTDA',
      fantasy_name: 'PetControl Dev',
      cnpj: '12345678000195',
      active_package: 'starter',
      is_active: true,
    };

    const schedules: ScheduleDTO[] = [
      {
        id: 'schedule-1',
        company_id: 'company-1',
        client_id: '11111111-1111-1111-1111-111111111111',
        pet_id: '22222222-2222-2222-2222-222222222222',
        scheduled_at: new Date('2026-04-10T13:00:00Z').toISOString(),
        estimated_end: null,
        notes: 'Banho e tosa',
        current_status: 'confirmed',
      },
    ];

    mockUseCurrentCompanyQuery.mockReset();
    mockUseSchedulesQuery.mockReset();
    mockCreateScheduleMutation.mockReset();
    mockUpdateScheduleMutation.mockReset();
    mockDeleteScheduleMutation.mockReset();

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
      data: company,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });

    mockUseSchedulesQuery.mockReturnValue({
      data: schedules,
      isLoading: false,
      isError: false,
    });

    const idleMutation = {
      mutateAsync: vi.fn(),
      isPending: false,
      error: null,
    };
    mockCreateScheduleMutation.mockReturnValue(idleMutation);
    mockUpdateScheduleMutation.mockReturnValue(idleMutation);
    mockDeleteScheduleMutation.mockReturnValue(idleMutation);
  });

  afterEach(() => {
    cleanup();
    queryClientForWeb().clear();
    vi.unstubAllGlobals();
  });

  it('redireciona da home para /:companySlug/dashboard quando há sessão persistida', async () => {
    await router.navigate({ to: '/' });

    render(
      <QueryClientProvider client={queryClientForWeb()}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/petcontrol-dev/dashboard');
    });

    expect(
      screen.getByText(
        'Dashboard conectado ao backend com dados reais de tenant, empresa e agendamentos.',
      ),
    ).toBeTruthy();
  });

  it('mantém o slug atual nos links internos e navega corretamente em /:companySlug/schedules', async () => {
    await router.navigate({ to: '/petcontrol-dev/schedules' });

    render(
      <QueryClientProvider client={queryClientForWeb()}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/petcontrol-dev/schedules');
    });

    expect(screen.getByText('Agendamentos do tenant')).toBeTruthy();
    expect(screen.getByText('Criar agendamento')).toBeTruthy();
    expect(screen.getByText('11111111...')).toBeTruthy();

    const dashboardLink = screen.getByRole('link', { name: 'Dashboard' });
    const schedulesLink = screen.getByRole('link', { name: 'Schedules' });

    expect(dashboardLink.getAttribute('href')).toBe('/petcontrol-dev/dashboard');
    expect(schedulesLink.getAttribute('href')).toBe('/petcontrol-dev/schedules');
  });
});
