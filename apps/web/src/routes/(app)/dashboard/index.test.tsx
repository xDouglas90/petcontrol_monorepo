import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import type { CompanyDTO, ScheduleDTO } from '@petcontrol/shared-types';

import { DashboardPage } from './index';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';

const mockUseCurrentCompanyQuery = vi.fn();
const mockUseSchedulesQuery = vi.fn();

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useSchedulesQuery: () => mockUseSchedulesQuery(),
}));

describe('DashboardPage', () => {
  beforeEach(() => {
    localStorage.clear();
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
    mockUseSchedulesQuery.mockReset();
  });

  it('renderiza dados reais da empresa corrente e dos schedules', () => {
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
        scheduled_at: new Date().toISOString(),
        estimated_end: null,
        notes: 'Consulta',
        current_status: 'confirmed',
      },
    ];

    mockUseCurrentCompanyQuery.mockReturnValue({
      data: company,
    });
    mockUseSchedulesQuery.mockReturnValue({
      data: {
        data: schedules,
        meta: { total: schedules.length, page: 1, limit: 10, total_pages: 1 }
      },
      isLoading: false,
      isError: false,
    });

    render(<DashboardPage />);

    expect(screen.getByText('PetControl Dev')).toBeTruthy();
    expect(screen.getByText('Dashboard conectado ao backend com dados reais de tenant, empresa e agendamentos.')).toBeTruthy();
    expect(screen.getAllByText('2')).toHaveLength(2);
    expect(screen.getByText('starter')).toBeTruthy();
    expect(screen.getByText('Banho e hidratação')).toBeTruthy();
  });
});
