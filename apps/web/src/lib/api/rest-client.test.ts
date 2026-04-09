import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  createSchedule,
  deleteSchedule,
  getCurrentCompany,
  listSchedules,
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
            data: [
              {
                id: 's-1',
                company_id: 'c-1',
                client_id: 'cli-1',
                pet_id: 'pet-1',
                scheduled_at: '2026-04-08T10:00:00.000Z',
                estimated_end: null,
                notes: 'banho',
                current_status: 'waiting',
              },
            ],
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
    expect(schedules).toHaveLength(1);

    const created = await createSchedule('token-123', {
      client_id: 'cli-2',
      pet_id: 'pet-2',
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
});
