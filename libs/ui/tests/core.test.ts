import { describe, expect, it } from 'vitest';

import { defaultSpacingScale, joinClassNames } from '../src/core';

describe('ui/core', () => {
  it('mantém escala de espaçamento padrão', () => {
    expect(defaultSpacingScale).toEqual({
      xs: 4,
      sm: 8,
      md: 12,
      lg: 16,
      xl: 24,
    });
  });

  it('concatena classes ignorando valores falsy', () => {
    const result = joinClassNames('px-4', false, 'rounded', undefined, 'py-2');
    expect(result).toBe('px-4 rounded py-2');
  });
});
