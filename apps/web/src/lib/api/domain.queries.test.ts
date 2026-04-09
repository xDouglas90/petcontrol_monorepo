import { describe, expect, it } from 'vitest';

import { domainQueryKeys } from './domain.queries';

describe('domain query keys', () => {
  it('mantém chaves estáveis para empresa corrente e schedules', () => {
    expect(domainQueryKeys.currentCompany()).toEqual([
      'domain',
      'company',
      'current',
    ]);
    expect(domainQueryKeys.schedules()).toEqual(['domain', 'schedules']);
  });
});
