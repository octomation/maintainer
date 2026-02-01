import type { ApiClient } from '../api/client';

/** What the shell hands every page. */
export interface PageProps {
  client: ApiClient;
  /** Reports a short status for the bottom of the sidebar, such as "generation 3 · 2 warnings". */
  onStatus?: (status: string) => void;
}
