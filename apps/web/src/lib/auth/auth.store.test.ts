import { beforeEach, describe, expect, it } from 'vitest';
import type { LoginSession } from '@petcontrol/shared-types';

import {
  selectIsAuthenticated,
  selectSession,
  useAuthStore,
} from './auth.store';

const sessionFixture: LoginSession = {
  accessToken: 'token-abc',
  tokenType: 'Bearer',
  userId: 'u-1',
  companyId: 'c-1',
  role: 'admin',
  kind: 'owner',
};

describe('auth.store', () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({
      session: null,
      hydrated: false,
    });
  });

  it('seta e limpa sessão corretamente', () => {
    useAuthStore.getState().setSession(sessionFixture);

    expect(selectSession(useAuthStore.getState())).toEqual(sessionFixture);
    expect(selectIsAuthenticated(useAuthStore.getState())).toBe(true);

    useAuthStore.getState().clearSession();

    expect(selectSession(useAuthStore.getState())).toBeNull();
    expect(selectIsAuthenticated(useAuthStore.getState())).toBe(false);
  });

  it('marca store como hidratada', () => {
    useAuthStore.getState().markHydrated();

    expect(useAuthStore.getState().hydrated).toBe(true);
  });
});
