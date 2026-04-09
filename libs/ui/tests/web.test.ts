import { describe, expect, it } from 'vitest';

import {
  formatScheduleStatus,
  resolveAsyncViewState,
  scheduleStatusColorClass,
} from '../src/web';

describe('ui/web helpers', () => {
  it('resolve estado assíncrono corretamente', () => {
    expect(
      resolveAsyncViewState({ isLoading: true, isError: false, itemCount: 0 }),
    ).toBe('loading');

    expect(
      resolveAsyncViewState({ isLoading: false, isError: true, itemCount: 0 }),
    ).toBe('error');

    expect(
      resolveAsyncViewState({
        isLoading: false,
        isError: false,
        itemCount: 0,
      }),
    ).toBe('empty');

    expect(
      resolveAsyncViewState({
        isLoading: false,
        isError: false,
        itemCount: 3,
      }),
    ).toBe('ready');
  });

  it('formata status de schedules para rótulos amigáveis', () => {
    expect(formatScheduleStatus('in_progress')).toBe('Em andamento');
    expect(formatScheduleStatus('delivered')).toBe('Entregue');
  });

  it('retorna classes de estilo para status', () => {
    expect(scheduleStatusColorClass('waiting')).toContain('amber');
    expect(scheduleStatusColorClass('canceled')).toContain('rose');
  });
});
