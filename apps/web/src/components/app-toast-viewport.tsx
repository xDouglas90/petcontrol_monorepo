import { cn } from '@petcontrol/ui/web';

import { useToastStore } from '@/stores/toast.store';

export function AppToastViewport() {
  const toasts = useToastStore((state) => state.toasts);
  const dismissToast = useToastStore((state) => state.dismissToast);

  if (toasts.length === 0) {
    return null;
  }

  return (
    <div className="pointer-events-none fixed right-4 top-4 z-[100] flex w-[min(24rem,calc(100vw-2rem))] flex-col gap-3">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={cn(
            'pointer-events-auto rounded-2xl border px-4 py-3 shadow-lg backdrop-blur-sm',
            toast.variant === 'error'
              ? 'border-rose-400/40 bg-rose-500/10 text-rose-200'
              : 'border-emerald-400/30 bg-emerald-500/10 text-emerald-200',
          )}
        >
          <div className="flex items-start justify-between gap-3">
            <p className="text-sm font-medium">{toast.message}</p>
            <button
              type="button"
              onClick={() => dismissToast(toast.id)}
              className="text-xs font-semibold uppercase tracking-[0.16em] text-muted transition hover:text-foreground"
            >
              fechar
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
