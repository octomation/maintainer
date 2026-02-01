import { useEffect, useLayoutEffect, useRef, useState } from 'preact/hooks';
import type { ApiClient } from '../api/client';
import type { Repository } from '../api/types';
import type { Theme } from '../theme';
import { BrandMark } from './Glyphs';
import { RepositoryTree } from './RepositoryTree';
import { Icon, type IconName } from './Icon';

export interface Route {
  path: '/status' | '/inbox' | '/tablo';
  title: string;
  icon: IconName;
}

/** The pages in the order of the sidebar. */
export const ROUTES: Route[] = [
  { path: '/inbox', title: 'Inbox', icon: 'inbox-solid' },
  { path: '/tablo', title: 'Tablo', icon: 'board-solid' },
  { path: '/status', title: 'Status', icon: 'branch-solid' },
];

export interface SidebarProps {
  current: string;
  theme: Theme;
  onTheme: () => void;
  /** A short status at the bottom, such as the index generation. */
  status?: string;
  /** The API client of the Reindex item; without it the item is not shown. */
  client?: ApiClient;
  /** What happens once the index has a new generation; the page reloads by default. */
  onReindexed?: () => void;
  /** How often the catalog is asked for the new generation. */
  pollMs?: number;
  /** The sidebar is an overlay on a narrow screen: open, and how it closes. */
  open?: boolean;
  onClose?: () => void;
  /** Opens the command bar; without it the search button is not shown. */
  onSearch?: () => void;
  /** The repositories of the catalog: the sidebar shows all of them as a tree by owner. */
  repositories?: Repository[];
  /** The repository the inbox is filtered by alone, marked as current. */
  currentRepository?: string | null;
}


/** How long Reindex waits for a new generation. */
export const REINDEX_TIMEOUT_MS = 60_000;

const THEME_TITLE: Record<Theme, string> = { system: 'System', light: 'Light', dark: 'Dark' };
const THEME_ICON: Record<Theme, IconName> = { system: 'desktop', light: 'sun', dark: 'moon' };

/**
 * The sidebar: the maintainer mark, the pages, and at the bottom the index
 * state, Reindex and the theme. 220 px on a desktop; an overlay from the
 * menu button of the top bar on a narrow screen.
 */
export function Sidebar({ current, theme, onTheme, status, client, onReindexed, pollMs = 500, open = false, onClose, onSearch, repositories = [], currentRepository = null }: SidebarProps) {
  const [reindexing, setReindexing] = useState(false);
  const [error, setError] = useState('');
  const alive = useRef(true);
  useLayoutEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
    };
  }, []);

  // POST /refresh, then poll the catalog until the generation changes, then reload.
  const reindex = async () => {
    if (!client || reindexing) return;
    setReindexing(true);
    setError('');
    try {
      const started = await client.refresh();
      const deadline = Date.now() + REINDEX_TIMEOUT_MS;
      while (Date.now() < deadline) {
        await new Promise((resolve) => setTimeout(resolve, pollMs));
        if (!alive.current) return;
        const { envelope } = await client.getCatalog();
        if (envelope.generation !== started.generation) {
          if (onReindexed) onReindexed();
          else location.reload();
          return;
        }
      }
      throw new Error('the index did not change within 60 s');
    } catch (err) {
      if (alive.current) setError(`Reindex failed: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      if (alive.current) setReindexing(false);
    }
  };

  // The overlay takes the focus to the current page; the shell gives it back on close.
  const aside = useRef<HTMLElement>(null);
  useEffect(() => {
    if (open) (aside.current?.querySelector<HTMLElement>('[aria-current="page"]') ?? aside.current?.querySelector<HTMLElement>('a'))?.focus();
  }, [open]);

  const themeLabel = `Theme: ${THEME_TITLE[theme].toLowerCase()}; switch`;
  return (
    <>
      {open && <div class="sidebar__backdrop" onClick={onClose} data-testid="sidebar-backdrop" />}
      <aside class={`sidebar${open ? ' sidebar--open' : ''}`} aria-label="Sidebar" ref={aside}>
        <div class="sidebar__brand">
          <a class="sidebar__brand-link" href="/inbox">
            <BrandMark />
            <span>maintainer</span>
          </a>
          {onClose && (
            <button type="button" class="icon-button sidebar__close" onClick={onClose} aria-label="Close the sidebar" title="Close">
              <Icon name="x" size={14} />
            </button>
          )}
        </div>
        {(client || onSearch) && (
          <div class="sidebar__actions">
            {client && (
              <button type="button" class="sidebar-action" disabled={reindexing} aria-busy={reindexing} onClick={() => void reindex()} title="Rebuild the task index from the mirrors">
                <Icon name="refresh" size={14} />
                {reindexing ? 'Reindexing…' : 'Reindex'}
              </button>
            )}
            {onSearch && (
              <button type="button" class="sidebar-action sidebar-action--square" onClick={onSearch} aria-label="Search" title="Search issues (/ or Ctrl K)">
                <Icon name="search" size={14} />
              </button>
            )}
          </div>
        )}
        <nav class="sidebar__nav" aria-label="Pages">
          <ul class="sidebar__items">
            {ROUTES.map((route) => (
              <li key={route.path}>
                <a class="sidebar-item" href={route.path} aria-current={current === route.path && !currentRepository ? 'page' : undefined}>
                  <Icon name={route.icon} tight />
                  <span>{route.title}</span>
                </a>
              </li>
            ))}
          </ul>
        </nav>
        {repositories.length > 0 && (
          <nav class="sidebar__group" aria-label="Repositories">
            <div class="sidebar__section" aria-hidden="true">
              Repositories
            </div>
            <RepositoryTree repositories={repositories} current={currentRepository} />
          </nav>
        )}
        <div class="sidebar__footer">
          {error ? (
            <p class="sidebar__status sidebar__status--error" role="alert">
              {error}
            </p>
          ) : (
            status && <p class="sidebar__status">{status}</p>
          )}
          <button type="button" class="sidebar-footer-item" onClick={onTheme} aria-label={themeLabel} title={themeLabel}>
            <Icon name={THEME_ICON[theme]} size={14} />
            <span aria-hidden="true">{THEME_TITLE[theme]}</span>
          </button>
        </div>
      </aside>
    </>
  );
}
