import { useCallback, useEffect, useRef, useState } from 'preact/hooks';
import type { ApiClient } from '../api/client';
import type { Catalog } from '../api/types';

/** How often the catalog is asked again while the index is being built. */
const POLL_MS = 1000;

export interface CatalogState {
  catalog: Catalog | null;
  /** The last request failed; `retry` asks again. */
  failed: boolean;
  /** Asks again after a failure; a no-op while a request is in flight or once the catalog is in. */
  retry: () => void;
}

/**
 * The catalog of a page: milestone titles and more. A 202 before the first
 * index is polled like the pages poll their lists; a failure is kept until
 * `retry`, which the pages call on Try again and after their next successful
 * list request, so milestones never stay as keys for long.
 */
export function useCatalog(client: ApiClient): CatalogState {
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const busy = useRef(false);
  const done = useRef(false);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const load = () => {
      busy.current = true;
      client
        .getCatalog()
        .then(({ status, envelope }) => {
          if (cancelled) return;
          if (status === 202 || envelope.data === null) {
            timer = setTimeout(load, POLL_MS);
            return;
          }
          busy.current = false;
          done.current = true;
          setCatalog(envelope.data);
          setFailed(false);
        })
        .catch(() => {
          if (cancelled) return;
          busy.current = false;
          setFailed(true);
        });
    };
    load();
    return () => {
      cancelled = true;
      busy.current = false;
      if (timer) clearTimeout(timer);
    };
  }, [client, attempt]);

  const retry = useCallback(() => {
    if (busy.current || done.current) return;
    setAttempt((n) => n + 1);
  }, []);

  return { catalog, failed, retry };
}
