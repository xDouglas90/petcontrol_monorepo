import type { PaginationMeta } from '@petcontrol/shared-types';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationBarProps {
  meta: PaginationMeta | undefined;
  onPageChange: (page: number) => void;
}

export function PaginationBar({ meta, onPageChange }: PaginationBarProps) {
  if (!meta || meta.total_pages <= 1) return null;

  const { page, total_pages, total } = meta;

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 pt-4 text-sm text-foreground">
      <span className="text-xs text-muted">
        Página {page} de {total_pages} · {total} registro{total !== 1 ? 's' : ''}
      </span>

      <div className="flex items-center gap-1">
        <button
          type="button"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
          title="Página anterior"
          className="rounded-xl border border-border/50 bg-surface/50 p-1.5 text-muted transition hover:bg-surface disabled:cursor-not-allowed disabled:opacity-40"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>

        {generatePageNumbers(page, total_pages).map((num, idx) =>
          num === null ? (
            <span key={`ellipsis-${idx}`} className="px-1 text-muted">
              …
            </span>
          ) : (
            <button
              key={num}
              type="button"
              onClick={() => onPageChange(num)}
              className={`min-w-[2rem] rounded-xl border px-2 py-1 text-xs font-medium transition ${
                num === page
                  ? 'border-primary/40 bg-primary/20 text-primary'
                  : 'border-border/50 bg-surface/50 text-muted hover:bg-surface'
              }`}
            >
              {num}
            </button>
          ),
        )}

        <button
          type="button"
          disabled={page >= total_pages}
          onClick={() => onPageChange(page + 1)}
          title="Próxima página"
          className="rounded-xl border border-border/50 bg-surface/50 p-1.5 text-muted transition hover:bg-surface disabled:cursor-not-allowed disabled:opacity-40"
        >
          <ChevronRight className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}

function generatePageNumbers(
  current: number,
  total: number,
): (number | null)[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: (number | null)[] = [1];

  if (current > 3) pages.push(null);

  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  if (current < total - 2) pages.push(null);

  pages.push(total);

  return pages;
}
