import { beforeEach, describe, expect, it } from 'vitest';

import { useUIStore } from './ui.store';

describe('ui.store', () => {
  beforeEach(() => {
    localStorage.clear();
    useUIStore.setState({
      sidebarOpen: true,
      theme: 'midnight',
    });
  });

  it('alterna sidebar com toggleSidebar', () => {
    useUIStore.getState().toggleSidebar();
    expect(useUIStore.getState().sidebarOpen).toBe(false);

    useUIStore.getState().toggleSidebar();
    expect(useUIStore.getState().sidebarOpen).toBe(true);
  });

  it('aplica tema e abertura de sidebar explicitamente', () => {
    useUIStore.getState().setTheme('ember');
    useUIStore.getState().setSidebarOpen(false);

    expect(useUIStore.getState().theme).toBe('ember');
    expect(useUIStore.getState().sidebarOpen).toBe(false);
  });
});
