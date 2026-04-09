import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

export type AsyncViewState = 'loading' | 'error' | 'empty' | 'ready';

export function resolveAsyncViewState(params: {
  isLoading: boolean;
  isError: boolean;
  itemCount: number;
}): AsyncViewState {
  if (params.isLoading) {
    return 'loading';
  }
  if (params.isError) {
    return 'error';
  }
  if (params.itemCount === 0) {
    return 'empty';
  }
  return 'ready';
}

export function formatScheduleStatus(status: string): string {
  switch (status) {
    case 'waiting':
      return 'Aguardando';
    case 'confirmed':
      return 'Confirmado';
    case 'canceled':
      return 'Cancelado';
    case 'in_progress':
      return 'Em andamento';
    case 'finished':
      return 'Finalizado';
    case 'delivered':
      return 'Entregue';
    default:
      return status;
  }
}

export function scheduleStatusColorClass(status: string): string {
  switch (status) {
    case 'waiting':
      return 'bg-amber-400/20 text-amber-100 border-amber-400/40';
    case 'confirmed':
      return 'bg-sky-400/20 text-sky-100 border-sky-400/40';
    case 'in_progress':
      return 'bg-violet-400/20 text-violet-100 border-violet-400/40';
    case 'finished':
      return 'bg-emerald-400/20 text-emerald-100 border-emerald-400/40';
    case 'delivered':
      return 'bg-teal-400/20 text-teal-100 border-teal-400/40';
    case 'canceled':
      return 'bg-rose-400/20 text-rose-100 border-rose-400/40';
    default:
      return 'bg-slate-400/20 text-slate-100 border-slate-400/40';
  }
}
