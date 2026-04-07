import { beforeEach, describe, expect, it, vi } from 'vitest';

import { login } from './rest-client';

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
      'http://localhost:8082/api/v1/auth/login',
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
});
