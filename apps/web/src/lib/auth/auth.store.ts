import { createJSONStorage, persist } from 'zustand/middleware';
import { create } from 'zustand';

import type { LoginSession } from '@/lib/api/rest-client';

interface AuthState {
  session: LoginSession | null;
  hydrated: boolean;
  setSession: (session: LoginSession) => void;
  clearSession: () => void;
  markHydrated: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      session: null,
      hydrated: false,
      setSession: (session) => set({ session }),
      clearSession: () => set({ session: null }),
      markHydrated: () => set({ hydrated: true }),
    }),
    {
      name: 'petcontrol-web-auth',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ session: state.session }),
      onRehydrateStorage: () => (state) => {
        state?.markHydrated();
      },
    },
  ),
);

export const selectSession = (state: AuthState) => state.session;
export const selectIsAuthenticated = (state: AuthState) =>
  Boolean(state.session?.accessToken);
