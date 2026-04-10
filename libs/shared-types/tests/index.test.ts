import { describe, expect, expectTypeOf, it } from 'vitest';

import {
  GENDER_IDENTITIES,
  MARITAL_STATUSES,
  MODULE_CODES,
  PET_KINDS,
  PET_SIZES,
  PET_TEMPERAMENTS,
  SCHEDULE_STATUSES,
  type ClientDTO,
  type PetDTO,
  type ServiceDTO,
  type ScheduleDTO,
} from '../src';

describe('shared-types domain contracts', () => {
  it('exports module codes for the current seeded domain set', () => {
    expect(MODULE_CODES).toEqual(['SCH', 'CRM', 'FIN']);
  });

  it('exports pet enums aligned with the database schema', () => {
    expect(PET_SIZES).toEqual(['small', 'medium', 'large', 'giant']);
    expect(PET_KINDS).toContain('dog');
    expect(PET_KINDS).toContain('other');
    expect(PET_TEMPERAMENTS).toContain('playful');
  });

  it('exports person enums required by the clients contract', () => {
    expect(GENDER_IDENTITIES).toContain('woman_cisgender');
    expect(MARITAL_STATUSES).toContain('single');
  });

  it('keeps schedule statuses stable for shared consumers', () => {
    expect(SCHEDULE_STATUSES).toContain('confirmed');
    expect(SCHEDULE_STATUSES).toContain('delivered');
  });
});

describe('shared-types DTO compatibility', () => {
  it('accepts a client payload with operational fields', () => {
    const client: ClientDTO = {
      id: '11111111-1111-1111-1111-111111111111',
      person_id: '22222222-2222-2222-2222-222222222222',
      company_id: '33333333-3333-3333-3333-333333333333',
      full_name: 'Maria Silva',
      short_name: 'Maria',
      gender_identity: 'woman_cisgender',
      marital_status: 'single',
      birth_date: '1992-06-15',
      cpf: '12345678901',
      email: 'maria@example.com',
      phone: '+551130000000',
      cellphone: '+5511999999999',
      has_whatsapp: true,
      client_since: '2026-04-10',
      notes: 'Cliente recorrente',
      is_active: true,
    };

    expect(client.full_name).toBe('Maria Silva');
    expect(client.is_active).toBe(true);
  });

  it('accepts a pet payload associated to a client', () => {
    const pet: PetDTO = {
      id: '44444444-4444-4444-4444-444444444444',
      owner_id: '11111111-1111-1111-1111-111111111111',
      company_id: '33333333-3333-3333-3333-333333333333',
      owner_name: 'Maria Silva',
      name: 'Thor',
      size: 'medium',
      kind: 'dog',
      temperament: 'playful',
      birth_date: '2023-01-15',
      is_active: true,
      notes: 'Gosta de brincar',
    };

    expect(pet.kind).toBe('dog');
    expect(pet.temperament).toBe('playful');
    expect(pet.owner_name).toBe('Maria Silva');
  });

  it('accepts a service payload with catalog pricing fields', () => {
    const service: ServiceDTO = {
      id: '55555555-5555-5555-5555-555555555555',
      type_id: '66666666-6666-6666-6666-666666666666',
      type_name: 'Banho',
      title: 'Banho completo',
      description: 'Banho com secagem e perfume',
      price: '89.90',
      discount_rate: '0.00',
      is_active: true,
    };

    expect(service.type_name).toBe('Banho');
    expect(service.price).toBe('89.90');
  });

  it('allows enriched schedule payloads without breaking the current contract', () => {
    const schedule: ScheduleDTO = {
      id: '77777777-7777-7777-7777-777777777777',
      company_id: '33333333-3333-3333-3333-333333333333',
      client_id: '11111111-1111-1111-1111-111111111111',
      pet_id: '44444444-4444-4444-4444-444444444444',
      client_name: 'Maria Silva',
      pet_name: 'Thor',
      service_ids: ['55555555-5555-5555-5555-555555555555'],
      service_titles: ['Banho completo'],
      scheduled_at: '2026-04-10T14:00:00Z',
      estimated_end: '2026-04-10T15:00:00Z',
      notes: 'Levar coleira vermelha',
      current_status: 'confirmed',
    };

    expect(schedule.client_name).toBe('Maria Silva');
    expect(schedule.service_titles).toContain('Banho completo');
    expectTypeOf(schedule.current_status).toEqualTypeOf<
      'waiting' | 'confirmed' | 'canceled' | 'in_progress' | 'finished' | 'delivered'
    >();
  });
});
