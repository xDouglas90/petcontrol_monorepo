import type { ReactNode } from 'react';
import type { LoginSession } from '@petcontrol/shared-types';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { AppLayout } from './_layout';
import { useAuthStore } from '@/lib/auth/auth.store';
import { useUIStore } from '@/stores/ui.store';
import { ApiError } from '@/lib/api/rest-client';

// Mocking TanStack Router
const mockNavigate = vi.fn();
let isDesktopViewport = true;
vi.mock('@tanstack/react-router', () => ({
  Navigate: (props: {
    to: string;
    replace?: boolean;
    search?: Record<string, unknown>;
    hash?: string;
  }) => {
    const args: Array<string | boolean | Record<string, unknown> | undefined> =
      [props.to, props.replace];
    if (props.search !== undefined || props.hash !== undefined) {
      args.push(props.search, props.hash);
    }
    mockNavigate(...args);
    return null;
  },
  Link: ({
    children,
    to,
    onClick,
  }: {
    children: ReactNode;
    to: string;
    onClick?: () => void;
  }) => (
    <a href={to} onClick={onClick}>
      {children}
    </a>
  ),
  Outlet: () => <div data-testid="outlet">Content</div>,
  useParams: vi.fn(() => ({})),
  useLocation: vi.fn(() => ({ pathname: '', search: {}, hash: '' })),
}));

import { useParams, useLocation } from '@tanstack/react-router';

// Mocking queries
const mockUseCurrentCompanyQuery = vi.fn();
const mockUseCurrentUserQuery = vi.fn();
vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useCurrentUserQuery: () => mockUseCurrentUserQuery(),
}));

describe('AppLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    const matchMediaMock = vi.fn().mockImplementation((query: string) => ({
      matches: query === '(min-width: 1024px)' ? isDesktopViewport : false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));
    vi.stubGlobal('matchMedia', matchMediaMock);
    isDesktopViewport = true;
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Correct Company',
        name: 'Correct Company LTDA',
        active_package: 'starter',
        logo_url: null,
      },
      refetch: vi.fn(),
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
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
    vi.unstubAllGlobals();
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
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: true,
      data: undefined,
    });

    render(<AppLayout />);

    expect(
      screen.getByText(
        'Sincronizando o novo shell e carregando o contexto autenticado.',
      ),
    ).toBeTruthy();
  });

  it('redireciona para o slug correto se houver mismatch na URL', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Correct Company',
        name: 'Correct Company LTDA',
        active_package: 'starter',
        logo_url: null,
      },
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'WRONG-SLUG' });
    vi.mocked(useLocation).mockReturnValue({
      pathname: '/WRONG-SLUG/schedules',
      search: {},
      hash: '',
      state: {},
      key: 'test',
      href: '/WRONG-SLUG/schedules',
    } as unknown as ReturnType<typeof useLocation>);

    render(<AppLayout />);

    expect(mockNavigate).toHaveBeenCalledWith(
      '/correct-slug/schedules',
      true,
      {},
      '',
    );
  });

  it('renderiza o layout e o outlet se o slug estiver correto', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Correct Company',
        name: 'Correct Company LTDA',
        active_package: 'starter',
        logo_url: null,
      },
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });

    render(<AppLayout />);

    expect(screen.getByTestId('outlet')).toBeTruthy();
    expect(screen.getAllByText('Correct Company')).not.toHaveLength(0);
    expect(screen.getByRole('link', { name: 'Clientes' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Pets' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Agendamentos' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Configurações' })).toBeTruthy();
    expect(screen.getByText('Upgrade para basic')).toBeTruthy();
  });

  it('usa fallback visual quando tenant não possui logo e usuário não possui imagem', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Acme Vet',
        name: 'Acme Vet LTDA',
        active_package: 'starter',
        logo_url: null,
      },
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Joana Souza',
        short_name: 'Joana',
        image_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });

    render(<AppLayout />);

    expect(screen.getByText('AV')).toBeTruthy();
    expect(screen.getAllByText('J').length).toBeGreaterThan(0);
  });

  it('ajusta o card de upgrade para tenants em premium', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Correct Company',
        name: 'Correct Company LTDA',
        active_package: 'premium',
        logo_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });

    render(<AppLayout />);

    expect(screen.getByText('Plano consolidado')).toBeTruthy();
    expect(screen.getByText('Ver detalhes')).toBeTruthy();
  });

  it('reabre a sidebar automaticamente em telas grandes', () => {
    useUIStore.setState({ sidebarOpen: false });
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Correct Company',
        name: 'Correct Company LTDA',
        active_package: 'starter',
        logo_url: null,
      },
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });

    render(<AppLayout />);

    expect(useUIStore.getState().sidebarOpen).toBe(true);
    expect(screen.getByRole('link', { name: 'Clientes' })).toBeTruthy();
  });

  it('abre o drawer no mobile ao clicar no menu hambúrguer', () => {
    isDesktopViewport = false;
    useUIStore.setState({ sidebarOpen: false });
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'correct-slug',
        fantasy_name: 'Correct Company',
        name: 'Correct Company LTDA',
        active_package: 'starter',
        logo_url: null,
      },
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });

    render(<AppLayout />);

    expect(screen.queryByRole('link', { name: 'Clientes' })).toBeNull();

    fireEvent.click(screen.getByTitle('Alternar sidebar'));

    expect(useUIStore.getState().sidebarOpen).toBe(true);
    expect(screen.getByRole('link', { name: 'Clientes' })).toBeTruthy();
    expect(
      screen.getByRole('button', { name: 'Fechar menu lateral' }),
    ).toBeTruthy();
  });

  it('normaliza o slug para lowercase na navegação canônica', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        slug: 'PETCONTROL-DEV',
        fantasy_name: 'PetControl Dev',
        name: 'PetControl Desenvolvimento LTDA',
        active_package: 'starter',
        logo_url: null,
      },
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
    });

    vi.mocked(useParams).mockReturnValue({ companySlug: 'petcontrol-dev-old' });
    vi.mocked(useLocation).mockReturnValue({
      pathname: '/petcontrol-dev-old/dashboard',
      search: {},
      hash: '',
      state: {},
      key: 'test',
      href: '/petcontrol-dev-old/dashboard',
    } as unknown as ReturnType<typeof useLocation>);

    render(<AppLayout />);

    expect(mockNavigate).toHaveBeenCalledWith(
      '/petcontrol-dev/dashboard',
      true,
      {},
      '',
    );
  });

  it('exibe tela de erro se a query da empresa falhar', () => {
    mockUseCurrentCompanyQuery.mockReturnValue({
      isError: true,
      error: new Error('API Error'),
      refetch: vi.fn(),
    });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
      refetch: vi.fn(),
    });

    render(<AppLayout />);

    expect(screen.getByText('Erro de Contexto')).toBeTruthy();
    expect(screen.getByText('Sair')).toBeTruthy();
  });

  it('faz logout, limpa a sessão e sai da área com slug', async () => {
    vi.mocked(useParams).mockReturnValue({ companySlug: 'correct-slug' });
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
      refetch: vi.fn(),
    });

    render(<AppLayout />);

    fireEvent.click(screen.getByText('Sair'));

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
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria da Silva',
        short_name: 'Maria',
        image_url: null,
      },
      refetch: vi.fn(),
    });

    render(<AppLayout />);

    await waitFor(() => {
      expect(useAuthStore.getState().session).toBeNull();
      expect(mockNavigate).toHaveBeenCalledWith('/login', true);
    });
  });
});
