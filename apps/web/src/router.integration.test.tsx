import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';
import type { CompanyDTO, ScheduleDTO } from '@petcontrol/shared-types';

import { queryClientForWeb, router } from '@/router';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseCurrentCompanyQuery = vi.fn();
const mockUseClientsQuery = vi.fn();
const mockUsePetsQuery = vi.fn();
const mockUseServicesQuery = vi.fn();
const mockUseSchedulesQuery = vi.fn();
const mockCreateScheduleMutation = vi.fn();
const mockUpdateScheduleMutation = vi.fn();
const mockDeleteScheduleMutation = vi.fn();

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useClientsQuery: () => mockUseClientsQuery(),
  usePetsQuery: () => mockUsePetsQuery(),
  useServicesQuery: () => mockUseServicesQuery(),
  useSchedulesQuery: () => mockUseSchedulesQuery(),
  useCreateScheduleMutation: () => mockCreateScheduleMutation(),
  useUpdateScheduleMutation: () => mockUpdateScheduleMutation(),
  useDeleteScheduleMutation: () => mockDeleteScheduleMutation(),
  domainQueryKeys: {
    currentCompany: () => ['domain', 'company', 'current'] as const,
    clients: () => ['domain', 'clients'] as const,
    pets: () => ['domain', 'pets'] as const,
    services: () => ['domain', 'services'] as const,
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
        client_name: 'Maria Silva',
        pet_name: 'Thor',
        service_ids: ['service-1'],
        service_titles: ['Banho completo'],
        scheduled_at: new Date('2026-04-10T13:00:00Z').toISOString(),
        estimated_end: null,
        notes: 'Banho e tosa',
        current_status: 'confirmed',
      },
    ];

    mockUseCurrentCompanyQuery.mockReset();
    mockUseClientsQuery.mockReset();
    mockUsePetsQuery.mockReset();
    mockUseServicesQuery.mockReset();
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
      data: {
        data: schedules,
        meta: { total: schedules.length, page: 1, limit: 10, total_pages: 1 }
      },
      isLoading: false,
      isError: false,
    });
    mockUseClientsQuery.mockReturnValue({
      data: {
        data: [
          {
            id: '11111111-1111-1111-1111-111111111111',
            full_name: 'Maria Silva',
          },
        ],
        meta: { total: 1, page: 1, limit: 10, total_pages: 1 }
      },
      isLoading: false,
      isError: false,
    });
    mockUsePetsQuery.mockReturnValue({
      data: {
        data: [
          {
            id: '22222222-2222-2222-2222-222222222222',
            owner_id: '11111111-1111-1111-1111-111111111111',
            name: 'Thor',
          },
        ],
        meta: { total: 1, page: 1, limit: 10, total_pages: 1 }
      },
      isLoading: false,
      isError: false,
    });
    mockUseServicesQuery.mockReturnValue({
      data: {
        data: [{ id: 'service-1', title: 'Banho completo' }],
        meta: { total: 1, page: 1, limit: 10, total_pages: 1 }
      },
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
    // @ts-expect-error - Navigate com URL raw interage melhor com os testes de redirect
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
    expect(screen.getAllByText('Maria Silva')).not.toHaveLength(0);
    expect(screen.getAllByText('Thor')).not.toHaveLength(0);
    expect(screen.getAllByText('Banho completo')).not.toHaveLength(0);

    const dashboardLink = screen.getByRole('link', { name: 'Dashboard' });
    const schedulesLink = screen.getByRole('link', { name: 'Schedules' });
    const clientsLink = screen.getByRole('link', { name: 'Clients' });
    const petsLink = screen.getByRole('link', { name: 'Pets' });
    const servicesLink = screen.getByRole('link', { name: 'Services' });

    expect(dashboardLink.getAttribute('href')).toBe(
      '/petcontrol-dev/dashboard',
    );
    expect(schedulesLink.getAttribute('href')).toBe(
      '/petcontrol-dev/schedules',
    );
    expect(clientsLink.getAttribute('href')).toBe('/petcontrol-dev/clients');
    expect(petsLink.getAttribute('href')).toBe('/petcontrol-dev/pets');
    expect(servicesLink.getAttribute('href')).toBe('/petcontrol-dev/services');
  });

  it('redireciona para o slug canônico quando o slug na URL é inválido ou diferente', async () => {
    // @ts-expect-error - Simula navegação manual para um slug incorreto
    await router.navigate({ to: '/wrong-slug/schedules' });

    render(
      <QueryClientProvider client={queryClientForWeb()}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/petcontrol-dev/schedules');
    }, { timeout: 5000 });

    // Verifica se o layout administrativo foi carregado (e consequentemente o redirect funcionou)
    expect(await screen.findByText(/Painel administrativo/i, {}, { timeout: 5000 })).toBeTruthy();
    expect(screen.getByText(/Tenant atual/i)).toBeTruthy();
    expect(screen.getByText(/@petcontrol-dev/i)).toBeTruthy();
  });
});
