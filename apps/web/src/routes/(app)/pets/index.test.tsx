import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';

import { PetsPage } from './index';

const mockUseClientsQuery = vi.fn();
const mockUsePeopleQuery = vi.fn();
const mockUsePetsQuery = vi.fn();
const mockUsePetQuery = vi.fn();
const mockUseCreatePetMutation = vi.fn();
const mockUseUpdatePetMutation = vi.fn();
const mockUseDeletePetMutation = vi.fn();
const mockPushToast = vi.fn();

const idleMutation = {
  mutateAsync: vi.fn(),
  isPending: false,
  error: null,
};

vi.mock('@/lib/api/domain.queries', () => ({
  useClientsQuery: () => mockUseClientsQuery(),
  usePeopleQuery: (params?: unknown) => mockUsePeopleQuery(params),
  usePetsQuery: (params?: unknown) => mockUsePetsQuery(params),
  usePetQuery: (petId?: string) => mockUsePetQuery(petId),
  useCreatePetMutation: () => mockUseCreatePetMutation(),
  useUpdatePetMutation: () => mockUseUpdatePetMutation(),
  useDeletePetMutation: () => mockUseDeletePetMutation(),
}));

vi.mock('@/stores/toast.store', () => ({
  useToastStore: (
    selector: (state: { pushToast: typeof mockPushToast }) => unknown,
  ) => selector({ pushToast: mockPushToast }),
}));

describe('PetsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.history.replaceState({}, '', '/petcontrol-dev/pets');

    mockUseCreatePetMutation.mockReturnValue(idleMutation);
    mockUseUpdatePetMutation.mockReturnValue(idleMutation);
    mockUseDeletePetMutation.mockReturnValue(idleMutation);

    mockUseClientsQuery.mockReturnValue({
      data: {
        data: [{ id: 'client-1', full_name: 'Maria Silva', short_name: 'Maria' }],
      },
      isLoading: false,
      isError: false,
    });

    mockUsePeopleQuery.mockReturnValue({
      data: {
        data: [
          { id: 'guardian-1', full_name: 'Guardiao 1', short_name: 'G1' },
          { id: 'guardian-2', full_name: 'Guardiao 2', short_name: 'G2' },
        ],
      },
      isLoading: false,
      isError: false,
    });

    mockUsePetsQuery.mockReturnValue({
      data: {
        data: [
          {
            id: 'pet-1',
            owner_id: 'client-1',
            owner_name: 'Maria Silva',
            name: 'Thor',
            race: 'Labrador',
            color: 'Caramelo',
            sex: 'M',
            size: 'medium',
            kind: 'dog',
            temperament: 'playful',
            image_url: null,
            birth_date: '2021-08-20',
            is_active: true,
            is_deceased: false,
            is_vaccinated: true,
            is_neutered: false,
            is_microchipped: false,
            microchip_number: null,
            microchip_expiration_date: null,
            notes: 'Observacao',
          },
        ],
        meta: { total: 1, page: 1, limit: 20, total_pages: 1 },
      },
      isLoading: false,
      isError: false,
    });

    mockUsePetQuery.mockReturnValue({
      data: {
        data: {
          id: 'pet-1',
          owner_id: 'client-1',
          owner_name: 'Maria Silva',
          name: 'Thor',
          race: 'Labrador',
          color: 'Caramelo',
          sex: 'M',
          size: 'medium',
          kind: 'dog',
          temperament: 'playful',
          image_url: null,
          birth_date: '2021-08-20',
          is_active: true,
          is_deceased: false,
          is_vaccinated: true,
          is_neutered: false,
          is_microchipped: false,
          microchip_number: null,
          microchip_expiration_date: null,
          notes: 'Observacao',
          guardians: [
            {
              guardian_id: 'guardian-1',
              full_name: 'Guardiao 1',
              short_name: 'G1',
              image_url: null,
              email: 'guardian1@petcontrol.local',
              cellphone: '+5511999990001',
              has_whatsapp: true,
            },
          ],
        },
      },
      isLoading: false,
      isError: false,
    });
  });

  afterEach(() => {
    cleanup();
  });

  it('renderiza guardioes no detalhe lateral', () => {
    render(<PetsPage />);
    expect(screen.getByText('Guardiao 1')).toBeTruthy();
  });

  it('exibe seletor de guardiao ao ativar o toggle no formulario', () => {
    render(<PetsPage />);

    fireEvent.click(screen.getByRole('button', { name: 'Inserir pet' }));

    const toggle = screen.getByLabelText('Inserir guardião');
    fireEvent.click(toggle);

    expect(screen.getByLabelText('Guardião')).toBeTruthy();
    expect(screen.getByRole('option', { name: 'Guardiao 1' })).toBeTruthy();
  });
});
