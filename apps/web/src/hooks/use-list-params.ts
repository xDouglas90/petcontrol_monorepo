import { useState, useCallback, useMemo } from 'react';
import { DEFAULT_PAGE, DEFAULT_PAGE_SIZE } from '@petcontrol/shared-constants';
import type { ListQueryParams } from '@petcontrol/shared-types';

export function useListParams(
  initialLimit = DEFAULT_PAGE_SIZE,
  initialSearch = '',
  initialPage = DEFAULT_PAGE,
) {
  const [page, setPage] = useState(Math.max(DEFAULT_PAGE, initialPage));
  const [limit] = useState(initialLimit);
  const [search, setSearchRaw] = useState(initialSearch);

  const setSearch = useCallback((value: string) => {
    setSearchRaw(value);
    setPage(DEFAULT_PAGE);
  }, []);

  const goToPage = useCallback((target: number) => {
    setPage(Math.max(DEFAULT_PAGE, target));
  }, []);

  const params: ListQueryParams = useMemo(
    () => ({
      page,
      limit,
      ...(search ? { search } : {}),
    }),
    [page, limit, search],
  );

  return { page, limit, search, params, setSearch, goToPage } as const;
}
