import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';

import { SchedulesPage } from './index';

vi.mock('@/lib/api/domain.queries', () => ({
  useClientsQuery: () => ({
    data: [
      {
        id: 'client-1',
        full_name: 'Maria Silva',
      },
    ],
  }),
  usePetsQuery: () => ({
    data: [
      {
        id: 'pet-1',
        owner_id: 'client-1',
        name: 'Thor',
      },
    ],
  }),
  useServicesQuery: () => ({
    data: [
      {
        id: 'service-1',
        title: 'Banho completo',
      },
    ],
  }),
  useSchedulesQuery: () => ({
    data: [
      {
        id: 'schedule-1',
        company_id: 'company-1',
        client_id: 'client-1',
        pet_id: 'pet-1',
        client_name: 'Maria Silva',
        pet_name: 'Thor',
        service_ids: ['service-1'],
        service_titles: ['Banho completo'],
        scheduled_at: '2026-04-10T14:00:00Z',
        estimated_end: null,
        notes: 'Observação',
        current_status: 'confirmed',
      },
    ],
    isLoading: false,
    isError: false,
  }),
  useCreateScheduleMutation: () => ({
    isPending: false,
    mutateAsync: vi.fn(),
    error: null,
  }),
  useUpdateScheduleMutation: () => ({
    isPending: false,
    mutateAsync: vi.fn(),
    error: null,
  }),
  useDeleteScheduleMutation: () => ({
    isPending: false,
    mutateAsync: vi.fn(),
    error: null,
  }),
}));

describe('SchedulesPage', () => {
  afterEach(() => {
    cleanup();
  });

  it('renderiza contexto operacional legível e seletores reais', () => {
    render(<SchedulesPage />);

    expect(screen.getByRole('option', { name: 'Maria Silva' })).toBeTruthy();
    expect(screen.getByRole('option', { name: 'Thor' })).toBeTruthy();
    expect(screen.getByRole('option', { name: 'Banho completo' })).toBeTruthy();

    expect(screen.getAllByText('Maria Silva')).not.toHaveLength(0);
    expect(screen.getAllByText('Thor')).not.toHaveLength(0);
    expect(screen.getAllByText('Banho completo')).not.toHaveLength(0);
    expect(screen.queryByPlaceholderText('UUID do client')).toBeNull();
    expect(screen.queryByPlaceholderText('UUID do pet')).toBeNull();
  });
});
