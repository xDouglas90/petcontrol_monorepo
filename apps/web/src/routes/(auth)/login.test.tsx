import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { LoginPage } from './login';
import { useAuthStore } from '@/lib/auth/auth.store';

const navigateMock = vi.fn();
const loginMock = vi.fn();
const checkHealthMock = vi.fn();

vi.mock('@/lib/api/rest-client', () => ({
  login: (...args: unknown[]) => loginMock(...args),
  checkHealth: (...args: unknown[]) => checkHealthMock(...args),
  getAuthMode: () => 'api',
  ApiError: class ApiError extends Error {
    status: number;
    details: unknown;

    constructor(message: string, status = 400, details?: unknown) {
      super(message);
      this.name = 'ApiError';
      this.status = status;
      this.details = details;
    }
  },
}));

vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>();
  return {
    ...actual,
    useNavigate: () => navigateMock,
    Navigate: () => null,
  };
});

describe('LoginPage', () => {
  beforeEach(() => {
    localStorage.clear();
    navigateMock.mockReset();
    loginMock.mockReset();
    checkHealthMock.mockReset();
    checkHealthMock.mockResolvedValue({
      status: 'ok',
      timestamp: new Date().toISOString(),
      version: 'test',
    });
    useAuthStore.setState({
      session: null,
      hydrated: true,
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('salva a sessão e navega para home após login bem-sucedido', async () => {
    loginMock.mockResolvedValue({
      accessToken: 'token-123',
      tokenType: 'Bearer',
      userId: 'user-1',
      companyId: 'company-1',
      role: 'admin',
      kind: 'owner',
    });

    const queryClient = new QueryClient();
    render(
      <QueryClientProvider client={queryClient}>
        <LoginPage />
      </QueryClientProvider>,
    );

    fireEvent.change(screen.getByLabelText('E-mail'), {
      target: { value: 'admin@petcontrol.local' },
    });
    fireEvent.change(screen.getByLabelText('Senha'), {
      target: { value: 'password123' },
    });
    fireEvent.click(screen.getByRole('button', { name: /entrar no sistema/i }));

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalled();
      expect(loginMock.mock.calls[0]?.[0]).toEqual({
        email: 'admin@petcontrol.local',
        password: 'password123',
      });
    });

    await waitFor(() => {
      expect(useAuthStore.getState().session).toMatchObject({
        accessToken: 'token-123',
        companyId: 'company-1',
      });
      expect(navigateMock).toHaveBeenCalledWith({ to: '/' });
    });
  });
});
