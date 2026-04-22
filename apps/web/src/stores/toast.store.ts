import { create } from 'zustand';

export type AppToast = {
  id: number;
  message: string;
  variant: 'success' | 'error';
};

type ToastStore = {
  toasts: AppToast[];
  pushToast: (
    message: string,
    variant: AppToast['variant'],
    durationMs?: number,
  ) => void;
  dismissToast: (id: number) => void;
};

let nextToastID = 1;

export const useToastStore = create<ToastStore>((set, get) => ({
  toasts: [],
  pushToast: (message, variant, durationMs = 4000) => {
    const id = nextToastID++;

    set((state) => ({
      toasts: [...state.toasts, { id, message, variant }],
    }));

    if (typeof window !== 'undefined') {
      window.setTimeout(() => {
        get().dismissToast(id);
      }, durationMs);
    }
  },
  dismissToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((toast) => toast.id !== id),
    })),
}));
