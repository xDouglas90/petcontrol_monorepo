import { describe, expect, it } from 'vitest';

import { domainQueryKeys } from './domain.queries';

describe('domain query keys', () => {
  it('mantém chaves estáveis para empresa corrente e domínios operacionais', () => {
    expect(domainQueryKeys.currentCompany()).toEqual([
      'domain',
      'company',
      'current',
    ]);
    expect(domainQueryKeys.currentCompanySystemConfig()).toEqual([
      'domain',
      'company',
      'system-config',
      'current',
    ]);
    expect(domainQueryKeys.companyUsers()).toEqual([
      'domain',
      'company-users',
    ]);
    expect(domainQueryKeys.adminSystemChatMessages('user-system-1')).toEqual([
      'domain',
      'chat',
      'admin-system',
      'user-system-1',
      'messages',
    ]);
    expect(domainQueryKeys.clients()).toEqual(['domain', 'clients', {}]);
    expect(domainQueryKeys.pets()).toEqual(['domain', 'pets', {}]);
    expect(domainQueryKeys.services()).toEqual(['domain', 'services', {}]);
    expect(domainQueryKeys.schedules()).toEqual(['domain', 'schedules', {}]);
    expect(domainQueryKeys.scheduleHistory('schedule-1')).toEqual([
      'domain',
      'schedules',
      'schedule-1',
      'history',
    ]);
  });

  it('inclui params na queryKey quando fornecidos', () => {
    const params = { page: 2, limit: 10, search: 'Thor' };
    expect(domainQueryKeys.clients(params)).toEqual([
      'domain',
      'clients',
      params,
    ]);
  });

  it('mantém chaves de prefixo para invalidação em lote', () => {
    expect(domainQueryKeys.allClients()).toEqual(['domain', 'clients']);
    expect(domainQueryKeys.allPets()).toEqual(['domain', 'pets']);
    expect(domainQueryKeys.allServices()).toEqual(['domain', 'services']);
    expect(domainQueryKeys.allSchedules()).toEqual(['domain', 'schedules']);
  });
});
