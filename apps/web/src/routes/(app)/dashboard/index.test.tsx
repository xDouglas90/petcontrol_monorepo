import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import type { CompanyDTO, ScheduleDTO } from '@petcontrol/shared-types';

import { DashboardPage } from './index';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseInternalChatSocket = vi.fn();
const mockUseCurrentCompanyQuery = vi.fn();
const mockUseCurrentCompanySystemConfigQuery = vi.fn();
const mockUseCurrentUserQuery = vi.fn();
const mockUseCompanyUsersQuery = vi.fn();
const mockUseAdminSystemChatMessagesQuery = vi.fn();
const mockUseCreateAdminSystemChatMessageMutation = vi.fn();
const mockUseSchedulesQuery = vi.fn();
const mockUseScheduleHistoriesQuery = vi.fn();

vi.mock('@/hooks/use-internal-chat-socket', () => ({
  useInternalChatSocket: () => mockUseInternalChatSocket(),
}));

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
}));

describe('DashboardPage', () => {
  beforeEach(() => {
    localStorage.clear();
    HTMLElement.prototype.scrollTo = vi.fn();
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
  });

  afterEach(() => {
    cleanup();
    mockUseCurrentCompanyQuery.mockReset();
    mockUseCurrentCompanySystemConfigQuery.mockReset();
    mockUseCurrentUserQuery.mockReset();
    mockUseCompanyUsersQuery.mockReset();
    mockUseAdminSystemChatMessagesQuery.mockReset();
    mockUseCreateAdminSystemChatMessageMutation.mockReset();
    mockUseSchedulesQuery.mockReset();
    mockUseScheduleHistoriesQuery.mockReset();
    mockUseInternalChatSocket.mockReset();
    vi.useRealTimers();
  });

  it('renderiza dados reais da empresa corrente e dos schedules', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 3, 19, 10, 30, 0));

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
        client_id: 'client-1',
        pet_id: 'pet-1',
        scheduled_at: new Date().toISOString(),
        estimated_end: null,
        notes: 'Banho e hidratação',
        current_status: 'waiting',
      },
      {
        id: 'schedule-2',
        company_id: 'company-1',
        client_id: 'client-2',
        pet_id: 'pet-2',
        scheduled_at: new Date(2026, 3, 19, 9, 15, 0).toISOString(),
        estimated_end: null,
        notes: 'Consulta',
        current_status: 'finished',
      },
    ];

    mockUseCurrentCompanyQuery.mockReturnValue({
      data: company,
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
    mockUseSchedulesQuery.mockReturnValue({
      data: {
        data: schedules,
        meta: { total: schedules.length, page: 1, limit: 10, total_pages: 1 }
      },
      isLoading: false,
      isError: false,
    });
    mockUseAdminSystemChatMessagesQuery.mockReturnValue({
      data: [
        {
          id: 'chat-message-1',
          conversation_id: 'chat-conversation-1',
          company_id: 'company-1',
          sender_user_id: 'user-system-1',
          sender_name: 'System',
          sender_role: 'system',
          sender_image_url: null,
          body: 'Tudo certo. O monitoramento do tenant já está ativo.',
          created_at: '2026-04-19T10:20:00-03:00',
        },
      ],
      isLoading: false,
      isError: false,
    });
    mockUseCreateAdminSystemChatMessageMutation.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
      isError: false,
    });
    mockUseInternalChatSocket.mockReturnValue({
      presenceMap: {
        'user-system-1': {
          user_id: 'user-system-1',
          status: 'online',
          last_changed_at: '2026-04-19T10:20:00-03:00',
        },
      },
      updatePresenceStatus: vi.fn(),
    });
    mockUseScheduleHistoriesQuery.mockReturnValue([
      {
        data: [
          {
            id: 'history-1',
            schedule_id: 'schedule-2',
            status: 'finished',
            changed_at: new Date(2026, 3, 19, 10, 30, 0).toISOString(),
            changed_by: 'user-1',
            notes: null,
          },
        ],
      },
    ]);

    render(<DashboardPage />);

    expect(screen.getByText('Olá, Maria')).toBeTruthy();
    expect(screen.getByText('Agendamentos/dia')).toBeTruthy();
    expect(screen.getByText('Agendamentos/mês')).toBeTruthy();
    expect(screen.getByText('Eficiência (meta mensal)')).toBeTruthy();
    expect(screen.getByText('Performance')).toBeTruthy();
    expect(screen.getByText('Ocupação por horário operacional')).toBeTruthy();
    expect(screen.getByText('Agendamentos em andamento')).toBeTruthy();
    expect(screen.getByText('Turno da manhã')).toBeTruthy();
    expect(screen.getByText('Finalizado')).toBeTruthy();
    expect(screen.getByText('1h 15min')).toBeTruthy();
    expect(screen.getByText('Chat do sistema')).toBeTruthy();
    expect(screen.getByText('Suporte ao administrador')).toBeTruthy();
    expect(screen.getByRole('combobox', { name: 'Selecionar usuário system' })).toBeTruthy();
    expect(screen.getAllByText('System').length).toBeGreaterThan(0);
    expect(
      screen.getByText('Tudo certo. O monitoramento do tenant já está ativo.'),
    ).toBeTruthy();
    expect(
      screen.getByRole('textbox', {
        name: 'Escrever mensagem para usuário system',
      }),
    ).toBeTruthy();
    const weekSelect = screen.getByRole('combobox', {
      name: 'Selecionar semana de performance',
    }) as HTMLSelectElement;

    expect(weekSelect.value).toBe('15-21');

    fireEvent.change(
      weekSelect,
      {
        target: { value: '1-7' },
      },
    );

    expect(
      screen.getByRole('option', {
        name: '01-07 abr',
      }),
    ).toBeTruthy();
    expect(weekSelect.value).toBe('1-7');
  });

  it('mantém o dashboard estável com dados operacionais mínimos', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 3, 19, 14, 30, 0));

    const company: CompanyDTO = {
      id: 'company-1',
      slug: 'petcontrol-dev',
      name: 'PetControl Desenvolvimento LTDA',
      fantasy_name: 'PetControl Premium',
      cnpj: '12345678000195',
      active_package: 'premium',
      is_active: true,
      logo_url: null,
    };

    mockUseCurrentCompanyQuery.mockReturnValue({
      data: company,
      isLoading: false,
      isError: false,
    });
    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Administrador Premium',
        short_name: 'Admin',
        image_url: null,
      },
      isLoading: false,
      isError: false,
    });
    mockUseCurrentCompanySystemConfigQuery.mockReturnValue({
      data: {
        company_id: 'company-1',
        schedule_init_time: '08:00',
        schedule_pause_init_time: '12:00',
        schedule_pause_end_time: '13:00',
        schedule_end_time: '18:00',
        min_schedules_per_day: 0,
        max_schedules_per_day: 18,
        schedule_days: [],
        dynamic_cages: false,
        total_small_cages: 0,
        total_medium_cages: 0,
        total_large_cages: 0,
        total_giant_cages: 0,
        whatsapp_notifications: false,
        whatsapp_conversation: false,
        whatsapp_business_phone: null,
      },
      isLoading: false,
      isError: false,
    });
    mockUseCompanyUsersQuery.mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    });
    mockUseSchedulesQuery.mockReturnValue({
      data: {
        data: [],
        meta: { total: 0, page: 1, limit: 10, total_pages: 0 },
      },
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
    mockUseInternalChatSocket.mockReturnValue({
      presenceMap: {},
      updatePresenceStatus: vi.fn(),
    });
    mockUseScheduleHistoriesQuery.mockReturnValue([]);

    render(<DashboardPage />);

    expect(screen.getByText('Olá, Admin')).toBeTruthy();
    expect(screen.getByText('0%')).toBeTruthy();
    expect(
      screen.getByText('Nenhum atendimento registrado para o turno atual.'),
    ).toBeTruthy();
    expect(
      screen.getByRole('option', {
        name: 'Nenhum usuário vinculado',
      }),
    ).toBeTruthy();
    expect(screen.getByText('Lista de usuários')).toBeTruthy();
  });
});
