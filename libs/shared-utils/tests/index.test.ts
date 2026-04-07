import { describe, expect, it } from 'vitest';

import { isNonEmptyTrimmed, normalizeUrl, safeLowerCase } from '../src';

describe('shared-utils', () => {
  it('normaliza barra final da URL', () => {
    expect(normalizeUrl('http://localhost:8082/api/v1/')).toBe(
      'http://localhost:8082/api/v1',
    );
  });

  it('valida string não vazia com trim', () => {
    expect(isNonEmptyTrimmed('  ok  ')).toBe(true);
    expect(isNonEmptyTrimmed('   ')).toBe(false);
  });

  it('normaliza para lowercase de forma segura', () => {
    expect(safeLowerCase('  ADMIN@PETCONTROL.LOCAL ')).toBe(
      'admin@petcontrol.local',
    );
  });
});
