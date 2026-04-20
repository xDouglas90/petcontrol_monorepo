import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';
import type { CompanyDTO, ScheduleDTO } from '@petcontrol/shared-types';

import { queryClientForWeb, router } from '@/router';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseCurrentCompanyQuery = vi.fn();
const mockUseCurrentCompanySystemConfigQuery = vi.fn();
const mockUseCurrentUserQuery = vi.fn();
const mockUseCompanyUsersQuery = vi.fn();
const mockUseAdminSystemChatMessagesQuery = vi.fn();
const mockUseCreateAdminSystemChatMessageMutation = vi.fn();
const mockUseClientsQuery = vi.fn();
const mockUsePetsQuery = vi.fn();
const mockUseServicesQuery = vi.fn();
const mockUseSchedulesQuery = vi.fn();
const mockUseScheduleHistoriesQuery = vi.fn();
const mockCreateScheduleMutation = vi.fn();
const mockUpdateScheduleMutation = vi.fn();
const mockDeleteScheduleMutation = vi.fn();

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useCurrentCompanySystemConfigQuery: () =>
    mockUseCurrentCompanySystemConfigQuery(),
  useCurrentUserQuery: () => mockUseCurrentUserQuery(),
  useCompanyUsersQuery: () => mockUseCompanyUsersQuery(),
  useAdminSystemChatMessagesQuery: () => mockUseAdminSystemChatMessagesQuery(),
  useCreateAdminSystemChatMessageMutation: () =>
    mockUseCreateAdminSystemChatMessageMutation(),
  useClientsQuery: () => mockUseClientsQuery(),
  usePetsQuery: () => mockUsePetsQuery(),
  useServicesQuery: () => mockUseServicesQuery(),
  useSchedulesQuery: () => mockUseSchedulesQuery(),
  useScheduleHistoriesQuery: () => mockUseScheduleHistoriesQuery(),
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
    mockUseCurrentCompanySystemConfigQuery.mockReset();
    mockUseCurrentUserQuery.mockReset();
    mockUseCompanyUsersQuery.mockReset();
    mockUseAdminSystemChatMessagesQuery.mockReset();
    mockUseCreateAdminSystemChatMessageMutation.mockReset();
    mockUseClientsQuery.mockReset();
    mockUsePetsQuery.mockReset();
    mockUseServicesQuery.mockReset();
    mockUseSchedulesQuery.mockReset();
    mockUseScheduleHistoriesQuery.mockReset();
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
    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria Silva',
        short_name: 'Maria',
        image_url: null,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mockUseCurrentCompanySystemConfigQuery.mockReturnValue({
      data: {
        company_id: 'company-1',
        schedule_init_time: '08:00',
        schedule_pause_init_time: '12:00',
        schedule_pause_end_time: '13:00',
        schedule_end_time: '18:00',
        min_schedules_per_day: 4,
        max_schedules_per_day: 18,
        schedule_days: [
          'monday',
          'tuesday',
          'wednesday',
          'thursday',
          'friday',
          'saturday',
        ],
        dynamic_cages: false,
        total_small_cages: 8,
        total_medium_cages: 6,
        total_large_cages: 4,
        total_giant_cages: 2,
        whatsapp_notifications: true,
        whatsapp_conversation: true,
        whatsapp_business_phone: '+5511999990001',
      },
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
    mockUseCompanyUsersQuery.mockReturnValue({
      data: [
        {
          id: 'company-user-system-1',
          company_id: 'company-1',
          user_id: 'user-system-1',
          kind: 'employee',
          role: 'system',
          is_owner: false,
          is_active: true,
          full_name: 'System PetControl',
          short_name: 'System',
          image_url: null,
          joined_at: '2026-04-10T10:00:00Z',
          left_at: null,
        },
      ],
      isLoading: false,
      isError: false,
    });
    mockUseAdminSystemChatMessagesQuery.mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    });
    mockUseCreateAdminSystemChatMessageMutation.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
      isError: false,
    });
    mockUseScheduleHistoriesQuery.mockReturnValue([]);
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
      screen.getByText('Agendamentos em andamento'),
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
    const schedulesLink = screen.getByRole('link', { name: 'Agendamentos' });
    const clientsLink = screen.getByRole('link', { name: 'Clientes' });
    const petsLink = screen.getByRole('link', { name: 'Pets' });
    const settingsLink = screen.getByRole('link', { name: 'Configurações' });

    expect(dashboardLink.getAttribute('href')).toBe(
      '/petcontrol-dev/dashboard',
    );
    expect(schedulesLink.getAttribute('href')).toBe(
      '/petcontrol-dev/schedules',
    );
    expect(clientsLink.getAttribute('href')).toBe('/petcontrol-dev/clients');
    expect(petsLink.getAttribute('href')).toBe('/petcontrol-dev/pets');
    expect(settingsLink.getAttribute('href')).toBe('/petcontrol-dev/settings');
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

    await waitFor(() => {
      expect(screen.getAllByText(/Dashboard/i).length).toBeGreaterThan(0);
    }, { timeout: 5000 });
    expect(screen.getAllByText(/PetControl Dev/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/Agendamentos do tenant/i)).toBeTruthy();
  }, 10000);
});
