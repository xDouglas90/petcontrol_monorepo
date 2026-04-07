import { createJSONStorage, persist } from 'zustand/middleware';
import { create } from 'zustand';

export type AppTheme = 'midnight' | 'ember';

interface UIState {
  sidebarOpen: boolean;
  theme: AppTheme;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  setTheme: (theme: AppTheme) => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarOpen: true,
      theme: 'midnight',
      toggleSidebar: () =>
        set((state) => ({ sidebarOpen: !state.sidebarOpen })),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'petcontrol-web-ui',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        sidebarOpen: state.sidebarOpen,
        theme: state.theme,
      }),
    },
  ),
);
