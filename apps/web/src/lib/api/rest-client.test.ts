import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  createAdminSystemChatMessage,
  completeUpload,
  createSchedule,
  createUploadIntent,
  deleteSchedule,
  getCurrentCompany,
  getCurrentCompanySystemConfig,
  getCurrentUser,
  listAdminSystemChatMessages,
  listClients,
  listPets,
  listSchedules,
  listServices,
  login,
  uploadToGCS,
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

  it('busca usuário corrente com bearer token', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            user_id: 'u-1',
            company_id: 'c-1',
            person_id: 'p-1',
            role: 'admin',
            kind: 'owner',
            full_name: 'Maria Silva',
            short_name: 'Maria',
            image_url: 'https://cdn.example.com/users/maria.png',
          },
        }),
        { status: 200 },
      ),
    );

    const user = await getCurrentUser('token-123');

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/users/me',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
      }),
    );
    expect(user.short_name).toBe('Maria');
  });

  it('busca company-system-configs/current com bearer token', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            company_id: 'c-1',
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
        }),
        { status: 200 },
      ),
    );

    const config = await getCurrentCompanySystemConfig('token-123');

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/company-system-configs/current',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
      }),
    );
    expect(config.schedule_days).toEqual(['monday', 'tuesday', 'wednesday']);
    expect(config.min_schedules_per_day).toBe(4);
  });

  it('busca histórico persistido do chat admin-system com bearer token', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [
            {
              id: 'chat-message-1',
              conversation_id: 'chat-conversation-1',
              company_id: 'c-1',
              sender_user_id: 'u-2',
              sender_name: 'System',
              sender_role: 'system',
              sender_image_url: null,
              body: 'Tudo certo por aqui.',
              created_at: '2026-04-20T09:00:00Z',
            },
          ],
        }),
        { status: 200 },
      ),
    );

    const messages = await listAdminSystemChatMessages('token-123', 'u-2');

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/chat/system/u-2/messages',
      expect.objectContaining({
        method: 'GET',
        headers: expect.objectContaining({ Authorization: 'Bearer token-123' }),
      }),
    );
    expect(messages[0]?.body).toBe('Tudo certo por aqui.');
  });

  it('envia mensagem persistida do chat admin-system com bearer token', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            id: 'chat-message-2',
            conversation_id: 'chat-conversation-1',
            company_id: 'c-1',
            sender_user_id: 'u-1',
            sender_name: 'Maria',
            sender_role: 'admin',
            sender_image_url: null,
            body: 'Precisamos revisar os atendimentos do dia.',
            created_at: '2026-04-20T09:15:00Z',
          },
        }),
        { status: 201 },
      ),
    );

    const message = await createAdminSystemChatMessage('token-123', 'u-2', {
      message: 'Precisamos revisar os atendimentos do dia.',
    });

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/chat/system/u-2/messages',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          Authorization: 'Bearer token-123',
          'Content-Type': 'application/json',
        }),
        body: JSON.stringify({
          message: 'Precisamos revisar os atendimentos do dia.',
        }),
      }),
    );
    expect(message.sender_role).toBe('admin');
  });

  it('usa o contrato completo do upload intent e propaga headers no upload para o GCS', async () => {
    fetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              upload_url: 'https://storage.googleapis.com/upload-signed',
              method: 'PUT',
              headers: {
                'Content-Type': 'image/png',
                'x-goog-meta-origin': 'petcontrol-web',
              },
              object_key: 'uploads/pets/image_url/2026/04/thor.png',
              public_url:
                'https://cdn.example.com/uploads/pets/image_url/2026/04/thor.png',
              expires_at: '2026-04-17T19:00:00Z',
            },
          }),
          { status: 201 },
        ),
      )
      .mockResolvedValueOnce(new Response(null, { status: 200 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              object_key: 'uploads/pets/image_url/2026/04/thor.png',
              public_url:
                'https://cdn.example.com/uploads/pets/image_url/2026/04/thor.png',
            },
          }),
          { status: 200 },
        ),
      );

    const intent = await createUploadIntent('token-123', {
      resource: 'pets',
      field: 'image_url',
      file_name: 'thor.png',
      content_type: 'image/png',
      size_bytes: 1024,
    });

    const file = new File(['binary'], 'thor.png', { type: 'image/png' });
    await uploadToGCS(intent, file);

    const completed = await completeUpload('token-123', {
      resource: 'pets',
      field: 'image_url',
      object_key: intent.object_key,
    });

    expect(intent.method).toBe('PUT');
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      'https://storage.googleapis.com/upload-signed',
      expect.objectContaining({
        method: 'PUT',
        body: file,
        headers: expect.any(Headers),
      }),
    );

    const uploadCall = fetchMock.mock.calls[1];
    const uploadHeaders = uploadCall?.[1] && 'headers' in uploadCall[1]
      ? (uploadCall[1].headers as Headers)
      : null;
    expect(uploadHeaders?.get('Content-Type')).toBe('image/png');
    expect(uploadHeaders?.get('x-goog-meta-origin')).toBe('petcontrol-web');
    expect(completed.object_key).toBe(intent.object_key);
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
            meta: { total: 1, limit: 10, page: 1, total_pages: 1 },
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
            meta: { total: 1, limit: 10, page: 1, total_pages: 1 },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
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
            meta: { total: 1, limit: 10, page: 1, total_pages: 1 },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
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
            meta: { total: 1, limit: 10, page: 1, total_pages: 1 },
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
