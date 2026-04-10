import type { ReactNode } from 'react';
import type { LoginSession } from '@petcontrol/shared-types';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { AppLayout } from './_layout';
import { useAuthStore } from '../../lib/auth/auth.store';
import { useUIStore } from '../../stores/ui.store';
import { ApiError } from '../../lib/api/rest-client';

// Mocking TanStack Router
const mockNavigate = vi.fn();
vi.mock('@tanstack/react-router', () => ({
  Navigate: ({ to, replace }: { to: string; replace: boolean }) => {
    mockNavigate(to, replace);
    return null;
  },
  Link: ({ children, to }: { children: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  Outlet: () => <div data-testid="outlet">Content</div>,
  useParams: vi.fn(() => ({})),
}));

import { useParams } from '@tanstack/react-router';

// Mocking queries
const mockUseCurrentCompanyQuery = vi.fn();
vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
}));

describe('AppLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { slug: 'correct-slug' },
      refetch: vi.fn(),
    });
    const session: LoginSession = {
      accessToken: 'token-123',
      tokenType: 'Bearer',
      userId: 'user-1',
      companyId: 'company-1',
      role: 'admin',
      kind: 'owner',
    };

    useAuthStore.setState({
      session,
      hydrated: true,
    });
    useUIStore.setState({
      sidebarOpen: true,
      theme: 'midnight',
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('redireciona para /login se não houver sessão', () => {
    useAuthStore.setState({ session: null, hydrated: true });
    
    render(<AppLayout />);
    
    expect(mockNavigate).toHaveBeenCalledWith('/login', true);
  });

  it('exibe LoadingScreen enquanto a empresa está carregando', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: true,
      data: undefined,
    });
    
    render(<AppLayout />);
    
    expect(screen.getByText('Carregando painel')).toBeTruthy();
  });

  it('redireciona para o slug correto se houver mismatch na URL', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: { slug: 'correct-slug' },
    });
    
    vi.mocked(useParams).mockReturnValue({ companySlug: 'WRONG-SLUG' });
    
    render(<AppLayout />);
    
    expect(mockNavigate).toHaveBeenCalledWith('/correct-slug/dashboard', true);
  });

  it('renderiza o layout e o outlet se o slug estiver correto', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: { slug: 'correct-slug' },
    });
    
    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });
    
    render(<AppLayout />);
    
    expect(screen.getByTestId('outlet')).toBeTruthy();
    expect(screen.getByText('PetControl')).toBeTruthy();
  });

  it('exibe tela de erro se a query da empresa falhar', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isError: true,
      error: new Error('API Error'),
      refetch: vi.fn(),
    });
    
    render(<AppLayout />);
    
    expect(screen.getByText('Erro de Contexto')).toBeTruthy();
    expect(screen.getByText('Sair')).toBeTruthy();
  });

  it('faz logout, limpa a sessão e sai da área com slug', async () => {
    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });

    render(<AppLayout />);

    fireEvent.click(screen.getByText('Encerrar sessão'));

    await waitFor(() => {
      expect(useAuthStore.getState().session).toBeNull();
      expect(mockNavigate).toHaveBeenCalledWith('/login', true);
    });
  });

  it('limpa a sessão e redireciona para /login quando a empresa corrente responde 401', async () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      isError: true,
      error: new ApiError('unauthorized', 401, { error: 'unauthorized' }),
      refetch: vi.fn(),
    });

    render(<AppLayout />);

    await waitFor(() => {
      expect(useAuthStore.getState().session).toBeNull();
      expect(mockNavigate).toHaveBeenCalledWith('/login', true);
    });
  });
});
