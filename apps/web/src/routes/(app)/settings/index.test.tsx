import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { SettingsPage } from './index';

const mockUseParams = vi.fn();
const mockUseCurrentUserQuery = vi.fn();
const mockUseCurrentCompanyQuery = vi.fn();
const mockUseCurrentCompanySystemConfigQuery = vi.fn();
const mockUseCompanyUsersQuery = vi.fn();
const mockUseCompanyUserPermissionsQuery = vi.fn();
const mockUseUpdateCurrentCompanyMutation = vi.fn();
const mockUseUpdateCurrentCompanySystemConfigMutation = vi.fn();
const mockUseUpdateCompanyUserPermissionsMutation = vi.fn();
const mockUpdateCurrentCompanyMutateAsync = vi.fn();
const mockUpdateCurrentCompanySystemConfigMutateAsync = vi.fn();
const mockUpdateCompanyUserPermissionsMutateAsync = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useParams: () => mockUseParams(),
  Navigate: ({ to }: { to: string }) => <div data-testid="navigate">{to}</div>,
}));

vi.mock('@/lib/api/domain.queries', () => ({
  useCurrentUserQuery: () => mockUseCurrentUserQuery(),
  useCurrentCompanyQuery: () => mockUseCurrentCompanyQuery(),
  useCurrentCompanySystemConfigQuery: () =>
    mockUseCurrentCompanySystemConfigQuery(),
  useCompanyUsersQuery: () => mockUseCompanyUsersQuery(),
  useCompanyUserPermissionsQuery: () => mockUseCompanyUserPermissionsQuery(),
  useUpdateCurrentCompanyMutation: () => mockUseUpdateCurrentCompanyMutation(),
  useUpdateCurrentCompanySystemConfigMutation: () =>
    mockUseUpdateCurrentCompanySystemConfigMutation(),
  useUpdateCompanyUserPermissionsMutation: () =>
    mockUseUpdateCompanyUserPermissionsMutation(),
}));

describe('SettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    mockUseParams.mockReturnValue({ companySlug: 'petcontrol-dev' });
    mockUseCurrentCompanyQuery.mockReturnValue({
      isLoading: false,
      data: {
        id: 'company-1',
        slug: 'petcontrol-dev',
        name: 'PetControl LTDA',
        fantasy_name: 'PetControl',
        cnpj: '12345678000195',
        active_package: 'starter',
        is_active: true,
        logo_url: 'https://cdn.example.com/logo.png',
      },
    });
    mockUseCurrentCompanySystemConfigQuery.mockReturnValue({
      isLoading: false,
      data: {
        company_id: 'company-1',
        schedule_init_time: '08:00',
        schedule_pause_init_time: '12:00',
        schedule_pause_end_time: '13:00',
        schedule_end_time: '18:00',
        min_schedules_per_day: 4,
        max_schedules_per_day: 18,
        schedule_days: ['monday', 'tuesday', 'wednesday'],
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
          id: 'company-user-1',
          company_id: 'company-1',
          user_id: 'system-user-1',
          kind: 'employee',
          role: 'system',
          is_owner: false,
          is_active: true,
          full_name: 'System PetControl',
          short_name: 'System',
          image_url: null,
          joined_at: new Date().toISOString(),
          left_at: null,
        },
      ],
    });
    mockUseCompanyUserPermissionsQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'system-user-1',
        company_id: 'company-1',
        role: 'system',
        kind: 'employee',
        is_owner: false,
        is_active: true,
        managed_by: 'admin-user-1',
        scope: 'tenant_settings',
        permissions: [
          {
            id: 'permission-1',
            code: 'company_settings:edit',
            description: 'Editar configurações gerais',
            default_roles: ['root', 'admin'],
            is_active: false,
            is_default_for_role: false,
            granted_by: null,
            granted_at: null,
          },
        ],
      },
    });
    mockUseUpdateCurrentCompanyMutation.mockReturnValue({
      mutateAsync: mockUpdateCurrentCompanyMutateAsync,
      isPending: false,
      isSuccess: false,
      isError: false,
    });
    mockUseUpdateCurrentCompanySystemConfigMutation.mockReturnValue({
      mutateAsync: mockUpdateCurrentCompanySystemConfigMutateAsync,
      isPending: false,
      isSuccess: false,
      isError: false,
    });
    mockUseUpdateCompanyUserPermissionsMutation.mockReturnValue({
      mutateAsync: mockUpdateCompanyUserPermissionsMutateAsync,
      isPending: false,
      isSuccess: false,
      isError: false,
    });

    mockUpdateCurrentCompanyMutateAsync.mockResolvedValue({});
    mockUpdateCurrentCompanySystemConfigMutateAsync.mockResolvedValue({});
    mockUpdateCompanyUserPermissionsMutateAsync.mockResolvedValue({});
  });

  afterEach(() => {
    cleanup();
  });

  it('renderiza as três seções para admin', () => {
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'admin-user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
        settings_access: {
          can_view: true,
          can_manage_permissions: true,
          active_permission_codes: ['company_settings:edit'],
          editable_permission_codes: ['company_settings:edit'],
        },
      },
    });

    render(<SettingsPage />);

    expect(screen.getByText('Configurações da empresa')).toBeTruthy();
    expect(screen.getByText('Configurações de negócios')).toBeTruthy();
    expect(screen.getByText('Permissões por usuário')).toBeTruthy();
  });

  it('não renderiza a seção de permissões para system autorizado', () => {
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'system-user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'system',
        kind: 'employee',
        full_name: 'System',
        short_name: 'System',
        image_url: null,
        settings_access: {
          can_view: true,
          can_manage_permissions: false,
          active_permission_codes: ['company_settings:edit'],
          editable_permission_codes: ['company_settings:edit'],
        },
      },
    });

    render(<SettingsPage />);

    expect(
      screen.getAllByText('Configurações da empresa').length,
    ).toBeGreaterThan(0);
    expect(screen.queryByText('Permissões por usuário')).toBeNull();
  });

  it('redireciona quando o usuário não possui acesso à tela', () => {
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'common-user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'common',
        kind: 'employee',
        full_name: 'Common',
        short_name: 'Common',
        image_url: null,
        settings_access: {
          can_view: false,
          can_manage_permissions: false,
          active_permission_codes: [],
          editable_permission_codes: [],
        },
      },
    });

    render(<SettingsPage />);

    expect(screen.getByTestId('navigate').textContent).toBe(
      '/petcontrol-dev/dashboard',
    );
  });

  it('salva as configurações da empresa com payload normalizado', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'admin-user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
        settings_access: {
          can_view: true,
          can_manage_permissions: true,
          active_permission_codes: ['company_settings:edit'],
          editable_permission_codes: ['company_settings:edit'],
        },
      },
    });

    render(<SettingsPage />);

    fireEvent.change(screen.getByLabelText('Nome jurídico'), {
      target: { value: '  PetControl Holdings  ' },
    });
    fireEvent.change(screen.getByLabelText('Nome fantasia'), {
      target: { value: '  PetControl Pro  ' },
    });
    fireEvent.change(screen.getByLabelText('Logo URL'), {
      target: { value: '   ' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Salvar empresa' }));

    await waitFor(() => {
      expect(mockUpdateCurrentCompanyMutateAsync).toHaveBeenCalledWith({
        name: 'PetControl Holdings',
        fantasy_name: 'PetControl Pro',
        logo_url: null,
      });
    });
  });

  it('salva as configurações de negócios convertendo números e campos opcionais', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'admin-user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
        settings_access: {
          can_view: true,
          can_manage_permissions: true,
          active_permission_codes: ['company_settings:edit'],
          editable_permission_codes: ['company_settings:edit'],
        },
      },
    });

    render(<SettingsPage />);

    fireEvent.change(screen.getByLabelText('Início da pausa'), {
      target: { value: '' },
    });
    fireEvent.change(screen.getByLabelText('Fim da pausa'), {
      target: { value: '' },
    });
    fireEvent.change(screen.getByLabelText('Mínimo de agendamentos por dia'), {
      target: { value: '6' },
    });
    fireEvent.change(screen.getByLabelText('Máximo de agendamentos por dia'), {
      target: { value: '22' },
    });
    fireEvent.change(screen.getByLabelText('WhatsApp business'), {
      target: { value: '' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Salvar negócio' }));

    await waitFor(() => {
      expect(
        mockUpdateCurrentCompanySystemConfigMutateAsync,
      ).toHaveBeenCalledWith({
        schedule_init_time: '08:00',
        schedule_pause_init_time: null,
        schedule_pause_end_time: null,
        schedule_end_time: '18:00',
        min_schedules_per_day: 6,
        max_schedules_per_day: 22,
        schedule_days: ['monday', 'tuesday', 'wednesday'],
        dynamic_cages: false,
        total_small_cages: 8,
        total_medium_cages: 6,
        total_large_cages: 4,
        total_giant_cages: 2,
        whatsapp_notifications: true,
        whatsapp_conversation: true,
        whatsapp_business_phone: null,
      });
    });
  });

  it('salva as permissões do usuário selecionado', async () => {
    mockUseCurrentUserQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'admin-user-1',
        company_id: 'company-1',
        person_id: 'person-1',
        role: 'admin',
        kind: 'owner',
        full_name: 'Maria',
        short_name: 'Maria',
        image_url: null,
        settings_access: {
          can_view: true,
          can_manage_permissions: true,
          active_permission_codes: ['company_settings:edit'],
          editable_permission_codes: ['company_settings:edit'],
        },
      },
    });
    mockUseCompanyUserPermissionsQuery.mockReturnValue({
      isLoading: false,
      data: {
        user_id: 'system-user-1',
        company_id: 'company-1',
        role: 'system',
        kind: 'employee',
        is_owner: false,
        is_active: true,
        managed_by: 'admin-user-1',
        scope: 'tenant_settings',
        permissions: [
          {
            id: 'permission-1',
            code: 'company_settings:edit',
            description: 'Editar configurações gerais',
            default_roles: ['root', 'admin'],
            is_active: false,
            is_default_for_role: false,
            granted_by: null,
            granted_at: null,
          },
          {
            id: 'permission-2',
            code: 'plan_settings:edit',
            description: 'Editar configurações de plano',
            default_roles: ['root', 'admin', 'system'],
            is_active: true,
            is_default_for_role: true,
            granted_by: 'admin-user-1',
            granted_at: new Date().toISOString(),
          },
        ],
      },
    });

    render(<SettingsPage />);

    fireEvent.click(
      screen.getByRole('checkbox', {
        name: /configurações gerais da empresa/i,
      }),
    );

    fireEvent.click(screen.getByRole('button', { name: 'Salvar permissões' }));

    await waitFor(() => {
      expect(mockUpdateCompanyUserPermissionsMutateAsync).toHaveBeenCalledWith({
        permission_codes: ['plan_settings:edit', 'company_settings:edit'],
      });
    });
  });
});
