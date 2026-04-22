import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';

import { PeoplePage } from './index';

const mockUseCurrentUserQuery = vi.fn();
const mockUsePeopleQuery = vi.fn();
const mockUsePersonQuery = vi.fn();
const mockUsePetsQuery = vi.fn();
const mockUseCreatePersonMutation = vi.fn();
const mockUseUpdatePersonMutation = vi.fn();
const mockPushToast = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useParams: vi.fn(() => ({ companySlug: 'petcontrol-dev' })),
  Navigate: () => null,
}));

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentUserQuery: () => mockUseCurrentUserQuery(),
  usePeopleQuery: (params?: unknown) => mockUsePeopleQuery(params),
  usePersonQuery: (personId?: string) => mockUsePersonQuery(personId),
  usePetsQuery: () => mockUsePetsQuery(),
  useCreatePersonMutation: () => mockUseCreatePersonMutation(),
  useUpdatePersonMutation: () => mockUseUpdatePersonMutation(),
}));

vi.mock('@/stores/toast.store', () => ({
  useToastStore: (selector: (state: { pushToast: typeof mockPushToast }) => unknown) =>
    selector({ pushToast: mockPushToast }),
}));

describe('PeoplePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.history.replaceState({}, '', '/petcontrol-dev/people');

    mockUseCreatePersonMutation.mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    });
    mockUseUpdatePersonMutation.mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    });
    mockUsePetsQuery.mockReturnValue({
      data: { data: [], meta: { total: 0, page: 1, limit: 200, total_pages: 1 } },
      isLoading: false,
      isError: false,
    });
    mockUsePeopleQuery.mockReturnValue({
      data: {
        data: [
          {
            id: 'person-1',
            company_id: 'company-1',
            company_person_id: 'cp-1',
            kind: 'client',
            full_name: 'Maria Silva',
            short_name: 'Maria',
            image_url: null,
            cpf: '12345678901',
            has_system_user: false,
            is_active: true,
            created_at: '2026-04-10T10:00:00Z',
            updated_at: null,
          },
          {
            id: 'person-2',
            company_id: 'company-1',
            company_person_id: 'cp-2',
            kind: 'supplier',
            full_name: 'Fornecedor XPTO',
            short_name: 'XPTO',
            image_url: null,
            cpf: null,
            has_system_user: false,
            is_active: true,
            created_at: '2026-04-11T10:00:00Z',
            updated_at: null,
          },
        ],
        meta: { total: 2, page: 1, limit: 100, total_pages: 1 },
      },
      isLoading: false,
      isError: false,
    });
    mockUsePersonQuery.mockImplementation((personId?: string) => {
      if (personId === 'person-2') {
        return {
          data: {
            id: 'person-2',
            company_id: 'company-1',
            company_person_id: 'cp-2',
            kind: 'supplier',
            full_name: 'Fornecedor XPTO',
            short_name: 'XPTO',
            image_url: null,
            cpf: null,
            has_system_user: false,
            is_active: true,
            created_at: '2026-04-11T10:00:00Z',
            updated_at: null,
            gender_identity: 'not_to_expose',
            marital_status: 'single',
            birth_date: '1990-10-10',
            contact: {
              email: 'fornecedor@xpto.local',
              phone: null,
              cellphone: '+5511999999999',
              has_whatsapp: false,
              instagram_user: null,
              emergency_contact: null,
              emergency_phone: null,
            },
            address: null,
            client_details: null,
            employee_details: null,
            employee_documents: null,
            employee_benefits: null,
            linked_user: null,
            guardian_pets: [],
          },
          isLoading: false,
          isError: false,
        };
      }

      return {
        data: {
          id: 'person-1',
          company_id: 'company-1',
          company_person_id: 'cp-1',
          kind: 'client',
          full_name: 'Maria Silva',
          short_name: 'Maria',
          image_url: null,
          cpf: '12345678901',
          has_system_user: false,
          is_active: true,
          created_at: '2026-04-10T10:00:00Z',
          updated_at: null,
          gender_identity: 'woman_cisgender',
          marital_status: 'single',
          birth_date: '1992-06-15',
          contact: {
            email: 'maria@petcontrol.local',
            phone: null,
            cellphone: '+5511988887777',
            has_whatsapp: true,
            instagram_user: null,
            emergency_contact: null,
            emergency_phone: null,
          },
          address: null,
          client_details: null,
          employee_details: null,
          employee_documents: null,
          employee_benefits: null,
          linked_user: null,
          guardian_pets: [],
        },
        isLoading: false,
        isError: false,
      };
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('renderiza lista e abre o painel de detalhe da pessoa selecionada', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    await waitFor(() => {
      expect(screen.getByText('Seleção atual')).toBeTruthy();
    });

    expect(screen.getAllByText('Maria Silva').length).toBeGreaterThan(0);
    expect(screen.getByText('maria@petcontrol.local')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Editar' })).toBeTruthy();
  });

  it('aplica filtro por tipo de pessoa na listagem', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    expect(screen.getAllByText('Maria Silva').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Fornecedor XPTO').length).toBeGreaterThan(0);

    fireEvent.change(screen.getByLabelText('Filtrar por tipo de pessoa'), {
      target: { value: 'supplier' },
    });

    await waitFor(() => {
      expect(mockUsePeopleQuery).toHaveBeenLastCalledWith({
        page: 1,
        limit: 100,
        kind: 'supplier',
      });
    });
    expect(screen.getByText('XPTO · Fornecedor')).toBeTruthy();
    expect(window.location.search).toContain('kind=supplier');
  });

  it('inicializa o filtro por tipo a partir da URL', async () => {
    window.history.replaceState({}, '', '/petcontrol-dev/people?kind=supplier');

    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    await waitFor(() => {
      expect(mockUsePeopleQuery).toHaveBeenCalledWith({
        page: 1,
        limit: 100,
        kind: 'supplier',
      });
    });

    expect(
      (screen.getByLabelText('Filtrar por tipo de pessoa') as HTMLSelectElement)
        .value,
    ).toBe('supplier');
  });

  it('inicializa a busca a partir da URL e propaga search para a query', async () => {
    window.history.replaceState(
      {},
      '',
      '/petcontrol-dev/people?search=fornecedor',
    );

    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    await waitFor(() => {
      expect(mockUsePeopleQuery).toHaveBeenCalledWith({
        page: 1,
        limit: 100,
        search: 'fornecedor',
      });
    });

    expect(
      (screen.getByPlaceholderText(
        'Buscar por nome, CPF ou tipo...',
      ) as HTMLInputElement).value,
    ).toBe('fornecedor');
  });

  it('atualiza a URL e a query quando a busca muda', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    fireEvent.change(
      screen.getByPlaceholderText('Buscar por nome, CPF ou tipo...'),
      {
        target: { value: 'xpto' },
      },
    );

    await waitFor(() => {
      expect(mockUsePeopleQuery).toHaveBeenLastCalledWith({
        page: 1,
        limit: 100,
        search: 'xpto',
      });
    });

    expect(window.location.search).toContain('search=xpto');
  });

  it('renderiza paginação e troca de página na query', async () => {
    mockUsePeopleQuery.mockReturnValue({
      data: {
        data: [
          {
            id: 'person-1',
            company_id: 'company-1',
            company_person_id: 'cp-1',
            kind: 'client',
            full_name: 'Maria Silva',
            short_name: 'Maria',
            image_url: null,
            cpf: '12345678901',
            has_system_user: false,
            is_active: true,
            created_at: '2026-04-10T10:00:00Z',
            updated_at: null,
          },
        ],
        meta: { total: 201, page: 1, limit: 100, total_pages: 3 },
      },
      isLoading: false,
      isError: false,
    });

    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    expect(screen.getByText('Página 1 de 3 · 201 registros')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: '2' }));

    await waitFor(() => {
      expect(mockUsePeopleQuery).toHaveBeenLastCalledWith({
        page: 2,
        limit: 100,
      });
    });

    expect(window.location.search).toContain('page=2');
  });

  it('inicializa a página a partir da URL', async () => {
    window.history.replaceState({}, '', '/petcontrol-dev/people?page=3');

    mockUsePeopleQuery.mockReturnValue({
      data: {
        data: [],
        meta: { total: 201, page: 3, limit: 100, total_pages: 3 },
      },
      isLoading: false,
      isError: false,
    });

    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-1',
        company_id: 'company-1',
        person_id: 'person-admin',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    await waitFor(() => {
      expect(mockUsePeopleQuery).toHaveBeenCalledWith({
        page: 3,
        limit: 100,
      });
    });
  });

  it('limita tipos e alterna criação de usuário para perfil system', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      data: {
        user_id: 'user-system-1',
        company_id: 'company-1',
        person_id: 'person-system',
        role: 'system',
        kind: 'employee',
        full_name: 'Operação',
        short_name: 'Ops',
        image_url: null,
      },
      isSuccess: true,
      isLoading: false,
      isError: false,
    });

    render(<PeoplePage />);

    fireEvent.click(screen.getByRole('button', { name: 'Inserir pessoa' }));

    const kindSelect = screen.getByLabelText('Tipo de pessoa');
    const kindOptions = Array.from(kindSelect.querySelectorAll('option')).map(
      (option) => option.textContent?.trim(),
    );
    expect(kindOptions).toEqual(['Cliente', 'Fornecedor']);

    expect(screen.getByText('Criar usuário de sistema')).toBeTruthy();

    fireEvent.change(kindSelect, { target: { value: 'supplier' } });

    await waitFor(() => {
      expect(screen.queryByText('Criar usuário de sistema')).toBeNull();
    });
  });
});
