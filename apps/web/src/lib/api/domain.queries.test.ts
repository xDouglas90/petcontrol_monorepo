import { describe, expect, it } from 'vitest';

import { domainQueryKeys } from './domain.queries';

describe('domain query keys', () => {
  it('mantém chaves estáveis para empresa corrente e domínios operacionais', () => {
    expect(domainQueryKeys.currentCompany()).toEqual([
      'domain',
      'company',
      'current',
    ]);
    expect(domainQueryKeys.clients()).toEqual(['domain', 'clients']);
    expect(domainQueryKeys.pets()).toEqual(['domain', 'pets']);
    expect(domainQueryKeys.services()).toEqual(['domain', 'services']);
    expect(domainQueryKeys.schedules()).toEqual(['domain', 'schedules']);
  });
});
