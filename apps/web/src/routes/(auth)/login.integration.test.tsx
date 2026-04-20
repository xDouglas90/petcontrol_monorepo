import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';

import { queryClientForWeb, router } from '@/router';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseCurrentCompanyQuery = vi.fn();
const mockUseCurrentCompanySystemConfigQuery = vi.fn();
const mockUseCurrentUserQuery = vi.fn();
const mockUseCompanyUsersQuery = vi.fn();
const mockUseAdminSystemChatMessagesQuery = vi.fn();
const mockUseCreateAdminSystemChatMessageMutation = vi.fn();
const mockUseSchedulesQuery = vi.fn();
const mockUseScheduleHistoriesQuery = vi.fn();

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useCurrentCompanySystemConfigQuery: () =>
    mockUseCurrentCompanySystemConfigQuery(),
  useCurrentUserQuery: () => mockUseCurrentUserQuery(),
  useCompanyUsersQuery: () => mockUseCompanyUsersQuery(),
  useAdminSystemChatMessagesQuery: () => mockUseAdminSystemChatMessagesQuery(),
  useCreateAdminSystemChatMessageMutation: () =>
    mockUseCreateAdminSystemChatMessageMutation(),
  useSchedulesQuery: () => mockUseSchedulesQuery(),
  useScheduleHistoriesQuery: () => mockUseScheduleHistoriesQuery(),
  domainQueryKeys: {
    currentCompany: () => ['domain', 'company', 'current'] as const,
    schedules: () => ['domain', 'schedules'] as const,
  },
}));

describe('Login flow integration', () => {
  beforeEach(async () => {
    localStorage.clear();
    vi.stubGlobal('scrollTo', vi.fn());
    HTMLElement.prototype.scrollTo = vi.fn();
    mockUseCurrentCompanyQuery.mockReset();
    mockUseCurrentCompanySystemConfigQuery.mockReset();
    mockUseCurrentUserQuery.mockReset();
    mockUseCompanyUsersQuery.mockReset();
    mockUseAdminSystemChatMessagesQuery.mockReset();
    mockUseCreateAdminSystemChatMessageMutation.mockReset();
    mockUseSchedulesQuery.mockReset();
    mockUseScheduleHistoriesQuery.mockReset();

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
      data: [],
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

    expect(screen.getByText('Ocupação por horário operacional')).toBeTruthy();
    expect(screen.getAllByText('PetControl Dev').length).toBeGreaterThan(0);
    expect(screen.getAllByRole('heading', { name: 'Olá, Maria' }).length).toBeGreaterThan(0);
  });
});
