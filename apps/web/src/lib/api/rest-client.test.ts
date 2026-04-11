import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  createSchedule,
  deleteSchedule,
  getCurrentCompany,
  listClients,
  listPets,
  listSchedules,
  listServices,
  login,
  updateSchedule,
} from './rest-client';

const fetchMock = vi.fn<typeof fetch>();

describe('rest-client login', () => {
  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
  });

  it('mapeia resposta de login da API para sessão', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            access_token: 'token-123',
            token_type: 'Bearer',
            user_id: 'u-1',
            company_id: 'c-1',
            role: 'admin',
            kind: 'owner',
          },
        }),
        { status: 200 },
      ),
    );

    const session = await login({
      email: 'admin@petcontrol.local',
      password: 'password123',
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/auth/login',
      expect.objectContaining({ method: 'POST' }),
    );
    expect(session).toEqual({
      accessToken: 'token-123',
      tokenType: 'Bearer',
      userId: 'u-1',
      companyId: 'c-1',
      role: 'admin',
      kind: 'owner',
    });
  });

  it('lança ApiError com mensagem da API em falha de autenticação', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ error: 'credenciais inválidas' }), {
        status: 401,
      }),
    );

    await expect(
      login({ email: 'admin@petcontrol.local', password: 'senha-invalida' }),
    ).rejects.toMatchObject({
      name: 'ApiError',
      status: 401,
      message: 'credenciais inválidas',
    });
  });

  it('busca empresa corrente com bearer token', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            id: 'c-1',
            slug: 'petcontrol-dev',
            name: 'PetControl',
            fantasy_name: 'PetControl',
            cnpj: '12345678000195',
            active_package: 'starter',
            is_active: true,
          },
        }),
        { status: 200 },
      ),
    );

    const company = await getCurrentCompany('token-123');

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/companies/current',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
      }),
    );
    expect(company.slug).toBe('petcontrol-dev');
  });

  it('executa ciclo de list/create/update/delete de schedules', async () => {
    fetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              data: [
                {
                  id: 's-1',
                  company_id: 'c-1',
                  client_id: 'cli-1',
                  pet_id: 'pet-1',
                  client_name: 'Maria',
                  pet_name: 'Thor',
                  service_ids: ['svc-1'],
                  service_titles: ['Banho'],
                  scheduled_at: '2026-04-08T10:00:00.000Z',
                  estimated_end: null,
                  notes: 'banho',
                  current_status: 'waiting',
                },
              ],
              meta: { total: 1, limit: 10, page: 1, total_pages: 1 }
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              id: 's-2',
              company_id: 'c-1',
              client_id: 'cli-2',
              pet_id: 'pet-2',
              client_name: 'João',
              pet_name: 'Luna',
              service_ids: ['svc-2'],
              service_titles: ['Tosa'],
              scheduled_at: '2026-04-08T12:00:00.000Z',
              estimated_end: null,
              notes: 'consulta',
              current_status: 'waiting',
            },
          }),
          { status: 201 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              id: 's-2',
              company_id: 'c-1',
              client_id: 'cli-2',
              pet_id: 'pet-2',
              client_name: 'João',
              pet_name: 'Luna',
              service_ids: ['svc-2'],
              service_titles: ['Tosa'],
              scheduled_at: '2026-04-08T12:00:00.000Z',
              estimated_end: null,
              notes: 'consulta atualizada',
              current_status: 'confirmed',
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    const schedules = await listSchedules('token-123');
    expect(schedules.data).toHaveLength(1);

    const created = await createSchedule('token-123', {
      client_id: 'cli-2',
      pet_id: 'pet-2',
      service_ids: ['svc-2'],
      scheduled_at: '2026-04-08T12:00:00.000Z',
      notes: 'consulta',
    });
    expect(created.id).toBe('s-2');

    const updated = await updateSchedule('token-123', 's-2', {
      notes: 'consulta atualizada',
      status: 'confirmed',
    });
    expect(updated.current_status).toBe('confirmed');

    await expect(deleteSchedule('token-123', 's-2')).resolves.toBeUndefined();

    expect(fetchMock).toHaveBeenCalledTimes(4);
  });

  it('busca catálogos operacionais com bearer token', async () => {
    fetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              data: [
                {
                  id: 'cli-1',
                  person_id: 'p-1',
                  company_id: 'c-1',
                  full_name: 'Maria',
                  short_name: 'Maria',
                  gender_identity: 'woman_cisgender',
                  marital_status: 'single',
                  birth_date: '1992-06-15',
                  cpf: '12345678901',
                  email: 'maria@example.com',
                  cellphone: '+5511999999999',
                  has_whatsapp: true,
                  is_active: true,
                },
              ],
              meta: { total: 1, limit: 10, page: 1, total_pages: 1 }
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              data: [
                {
                  id: 'pet-1',
                  owner_id: 'cli-1',
                  owner_name: 'Maria',
                  name: 'Thor',
                  size: 'medium',
                  kind: 'dog',
                  temperament: 'playful',
                  is_active: true,
                },
              ],
              meta: { total: 1, limit: 10, page: 1, total_pages: 1 }
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              data: [
                {
                  id: 'svc-1',
                  type_id: 'type-1',
                  type_name: 'Banho',
                  title: 'Banho completo',
                  description: 'Banho com secagem',
                  price: '89.90',
                  discount_rate: '0.00',
                  is_active: true,
                },
              ],
              meta: { total: 1, limit: 10, page: 1, total_pages: 1 }
            },
          }),
          { status: 200 },
        ),
      );

    await expect(listClients('token-123').then(res => res.data)).resolves.toHaveLength(1);
    await expect(listPets('token-123').then(res => res.data)).resolves.toHaveLength(1);
    await expect(listServices('token-123').then(res => res.data)).resolves.toHaveLength(1);

    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});
