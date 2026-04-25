import { useEffect, useRef, useState } from 'react';
import { Search, X } from 'lucide-react';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
  id?: string;
}

export function SearchBar({
  value,
  onChange,
  placeholder = 'Buscar...',
  debounceMs = 350,
  id = 'search-bar',
}: SearchBarProps) {
  const [local, setLocal] = useState(value);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    setLocal(value);
  }, [value]);

  function handleChange(next: string) {
    setLocal(next);
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => onChange(next), debounceMs);
  }

  function clear() {
    setLocal('');
    onChange('');
  }

  return (
    <div className="relative">
      <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
      <input
        id={id}
        title={placeholder}
        type="search"
        value={local}
        onChange={(event) => handleChange(event.target.value)}
        placeholder={placeholder}
        className="w-full rounded-2xl border border-border/50 bg-surface/50 py-2 pl-9 pr-9 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-primary/50 focus:ring-2 focus:ring-primary/20"
      />
      {local ? (
        <button
          type="button"
          onClick={clear}
          title="Limpar busca"
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full p-1 text-muted transition hover:bg-surface hover:text-foreground"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      ) : null}
    </div>
  );
}
