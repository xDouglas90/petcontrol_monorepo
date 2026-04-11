import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';

import { ClientsPage } from './clients';
import { PetsPage } from './pets';
import { ServicesPage } from './services';

const idleMutation = {
  isPending: false,
  mutateAsync: vi.fn(),
  error: null,
};

vi.mock('@/lib/api/domain.queries', () => ({
  useClientsQuery: () => ({
    data: {
      data: [
        {
          id: 'client-1',
          person_id: 'person-1',
          company_id: 'company-1',
          full_name: 'Maria Silva',
          short_name: 'Maria',
          gender_identity: 'woman_cisgender',
          marital_status: 'single',
          birth_date: '1992-06-15',
          cpf: '12345678901',
          email: 'maria@example.com',
          cellphone: '+5511999990001',
          has_whatsapp: true,
          client_since: '2026-04-01',
          notes: 'Cliente recorrente',
          is_active: true,
        },
      ],
      meta: { total: 1, page: 1, limit: 10, total_pages: 1 }
    },
    isLoading: false,
    isError: false,
  }),
  usePetsQuery: () => ({
    data: {
      data: [
        {
          id: 'pet-1',
          owner_id: 'client-1',
          owner_name: 'Maria Silva',
          name: 'Thor',
          size: 'medium',
          kind: 'dog',
          temperament: 'playful',
          is_active: true,
        },
      ],
      meta: { total: 1, page: 1, limit: 10, total_pages: 1 }
    },
    isLoading: false,
    isError: false,
  }),
  useServicesQuery: () => ({
    data: {
      data: [
        {
          id: 'service-1',
          type_id: 'type-1',
          type_name: 'Banho',
          title: 'Banho completo',
          description: 'Banho com secagem e perfume',
          price: '89.90',
          discount_rate: '0.00',
          is_active: true,
        },
      ],
      meta: { total: 1, page: 1, limit: 10, total_pages: 1 }
    },
    isLoading: false,
    isError: false,
  }),
  useCreateClientMutation: () => idleMutation,
  useUpdateClientMutation: () => idleMutation,
  useDeleteClientMutation: () => idleMutation,
  useCreatePetMutation: () => idleMutation,
  useUpdatePetMutation: () => idleMutation,
  useDeletePetMutation: () => idleMutation,
  useCreateServiceMutation: () => idleMutation,
  useUpdateServiceMutation: () => idleMutation,
  useDeleteServiceMutation: () => idleMutation,
}));

describe('operational domain pages', () => {
  afterEach(() => {
    cleanup();
  });

  it('renderiza a tela operacional de clientes', () => {
    render(<ClientsPage />);

    expect(screen.getByRole('heading', { name: 'Clientes' })).toBeTruthy();
    expect(screen.getByText('Maria Silva')).toBeTruthy();
    expect(screen.getByText(/maria@example.com/)).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Criar' })).toBeTruthy();
  });

  it('renderiza a tela operacional de pets com tutor selecionável', () => {
    render(<PetsPage />);

    expect(screen.getByRole('heading', { name: 'Pets' })).toBeTruthy();
    expect(screen.getByText('Thor')).toBeTruthy();
    expect(screen.getByRole('option', { name: 'Maria Silva' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Criar' })).toBeTruthy();
  });

  it('renderiza a tela operacional de serviços', () => {
    render(<ServicesPage />);

    expect(screen.getByRole('heading', { name: 'Serviços' })).toBeTruthy();
    expect(screen.getByText('Banho completo')).toBeTruthy();
    expect(screen.getByText(/R\$ 89.90/)).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Criar' })).toBeTruthy();
  });
});
