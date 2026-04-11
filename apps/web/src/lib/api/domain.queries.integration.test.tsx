import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act } from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { CreateScheduleInput, ScheduleDTO } from '@petcontrol/shared-types';

import {
  useCreateScheduleMutation,
  useSchedulesQuery,
} from './domain.queries';
import { useAuthStore } from '@/lib/auth/auth.store';

let schedulesFixture: ScheduleDTO[] = [];
const listSchedulesMock = vi.fn();
const createScheduleMock = vi.fn();

vi.mock('@/lib/api/rest-client', async () => {
  const actual = await vi.importActual<typeof import('./rest-client')>('./rest-client');

  return {
    ...actual,
    getCurrentCompany: vi.fn(),
    listSchedules: (...args: Parameters<typeof listSchedulesMock>) =>
      listSchedulesMock(...args),
    createSchedule: (...args: Parameters<typeof createScheduleMock>) =>
      createScheduleMock(...args),
    updateSchedule: vi.fn(),
    deleteSchedule: vi.fn(),
  };
});

function QueryHarness() {
  const schedulesQuery = useSchedulesQuery();
  const createMutation = useCreateScheduleMutation();

  return (
    <div>
      <span data-testid="schedule-count">
        {String(schedulesQuery.data?.data?.length ?? 0)}
      </span>
      <button
        type="button"
        onClick={() =>
          createMutation.mutate({
            client_id: 'client-2',
            pet_id: 'pet-2',
            scheduled_at: '2026-04-08T12:00:00.000Z',
            notes: 'Consulta',
          })
        }
      >
        criar
      </button>
    </div>
  );
}

describe('domain queries integration', () => {
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

    schedulesFixture = [
      {
        id: 'schedule-1',
        company_id: 'company-1',
        client_id: 'client-1',
        pet_id: 'pet-1',
        scheduled_at: '2026-04-08T10:00:00.000Z',
        estimated_end: null,
        notes: 'Banho',
        current_status: 'waiting',
      },
    ];

    listSchedulesMock.mockReset();
    createScheduleMock.mockReset();

    listSchedulesMock.mockImplementation(async () => ({
      data: [...schedulesFixture],
      meta: { total: schedulesFixture.length, limit: 10, page: 1, total_pages: 1 }
    }));
    createScheduleMock.mockImplementation(
      async (_token: string, input: CreateScheduleInput) => {
        const created: ScheduleDTO = {
          id: 'schedule-2',
          company_id: 'company-1',
          client_id: input.client_id,
          pet_id: input.pet_id,
          scheduled_at: input.scheduled_at,
          estimated_end: input.estimated_end ?? null,
          notes: input.notes ?? null,
          current_status: input.status ?? 'waiting',
        };
        schedulesFixture = [...schedulesFixture, created];
        return created;
      },
    );
  });

  afterEach(() => {
    cleanup();
  });

  it('invalida e recarrega schedules após create', async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <QueryHarness />
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId('schedule-count').textContent).toBe('1');
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'criar' }));
    });

    await waitFor(() => {
      expect(screen.getByTestId('schedule-count').textContent).toBe('2');
    });

    expect(createScheduleMock).toHaveBeenCalledTimes(1);
    expect(listSchedulesMock).toHaveBeenCalledTimes(2);
  });
});
