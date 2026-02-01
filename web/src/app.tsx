import { useCallback, useEffect, useMemo, useRef, useState } from 'preact/hooks';
import { ApiClient } from './api/client';
import { CommandBar, type IssueCommands, type Mode } from './components/CommandBar';
import { ROUTES, Sidebar } from './components/Sidebar';
import { useCatalog } from './lib/catalog';
import { QUERY_EVENT } from './lib/query';
import { openTask } from './lib/task';
import { ShellContext } from './components/TopBar';
import { NARROW, useMediaQuery } from './lib/media';
import { applyTheme, nextTheme, readTheme, type Theme } from './theme';
import { StatusPage } from './pages/status';
import { InboxPage } from './pages/inbox';
import { TabloPage } from './pages/tablo';

export interface AppProps {
  /** The path to render; defaults to location.pathname. */
  path?: string;
  client?: ApiClient;
}

/**
 * The shell: the sidebar and the page of the pathname in a 100dvh grid; the
 * page scrolls inside. The service serves the document at /status, /inbox
 * and /tablo only, so there is no client-side router: a link is a
 * navigation. On a narrow screen the sidebar is an overlay that the top bar
 * opens.
 */
export function App({ path, client }: AppProps) {
  const current = path ?? (typeof location !== 'undefined' ? location.pathname : '/status');
  const api = useMemo(() => client ?? new ApiClient(), [client]);
  const [theme, setTheme] = useState<Theme>(() => readTheme());
  useEffect(() => applyTheme(theme), [theme]);
  const route = ROUTES.find((r) => r.path === current) ?? ROUTES[ROUTES.length - 1];
  const [status, setStatus] = useState('');
  const narrow = useMediaQuery(NARROW);
  const [open, setOpen] = useState(false);
  useEffect(() => {
    if (!narrow) setOpen(false);
  }, [narrow]);
  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open]);
  // The element that opened the overlay, such as the menu button, gets the focus back.
  const opener = useRef<HTMLElement | null>(null);
  // The open issues register their commands; the last one is in front.
  const issues = useRef<(() => IssueCommands | null)[]>([]);
  const [issue, setIssue] = useState<IssueCommands | null>(null);
  const [mode, setMode] = useState<Mode>('root');
  const shell = useMemo(
    () => ({
      openSidebar: () => {
        opener.current = document.activeElement as HTMLElement | null;
        setOpen(true);
      },
      registerIssue: (get: () => IssueCommands | null) => {
        issues.current = [...issues.current, get];
        return () => {
          issues.current = issues.current.filter((g) => g !== get);
        };
      },
    }),
    [],
  );
  /** Opens the command bar, issue-scoped when an issue is open. */
  const search = useCallback((at: Mode = 'root') => {
    const top = issues.current[issues.current.length - 1];
    setIssue(top?.() ?? null);
    setMode(at);
    setOpen(false);
    setSearching(true);
  }, []);
  // P and E on an open issue: the command bar at its priority or estimate levels.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const key = event.key.toLowerCase();
      if ((key !== 'p' && key !== 'e') || event.metaKey || event.ctrlKey || event.altKey || event.defaultPrevented) return;
      const target = event.target as HTMLElement | null;
      if (target?.closest('input, textarea, select, [contenteditable="true"]')) return;
      if (document.querySelector('[aria-modal="true"], .popover')) return;
      const top = issues.current[issues.current.length - 1]?.();
      if (!top) return;
      event.preventDefault();
      search(key === 'p' ? 'priority' : 'estimate');
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [search]);
  const close = useCallback(() => setOpen(false), []);
  const { catalog } = useCatalog(api);
  const [searching, setSearching] = useState(false);
  // / and Ctrl or Cmd K open the command bar, unless a text field has the focus.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const k = event.key.toLowerCase();
      const combo = k === 'k' && (event.metaKey || event.ctrlKey) && !event.altKey;
      const slash = event.key === '/' && !event.metaKey && !event.ctrlKey && !event.altKey;
      if (!combo && !slash) return;
      const target = event.target as HTMLElement | null;
      if (slash && target?.closest('input, textarea, select, [contenteditable="true"]')) return;
      if (target?.closest('[aria-modal="true"]')) return;
      event.preventDefault();
      search('root');
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);
  // An issue opens as the full page on the inbox or the board, from the status page in the inbox.
  const openIssue = useCallback(
    (id: string) => {
      if (route.path === '/inbox' || route.path === '/tablo') openTask(id);
      else location.assign(`/inbox?task=${encodeURIComponent(id)}`);
    },
    [route.path],
  );
  // The repository the inbox is filtered by alone: the current one of the sidebar, kept in step with the filters.
  const readRepo = useCallback(() => {
    if (route.path !== '/inbox' || typeof location === 'undefined') return null;
    const repos = new URLSearchParams(location.search).getAll('repo');
    return repos.length === 1 ? repos[0] : null;
  }, [route.path]);
  const [filteredRepo, setFilteredRepo] = useState<string | null>(readRepo);
  useEffect(() => {
    const update = () => setFilteredRepo(readRepo());
    addEventListener(QUERY_EVENT, update);
    addEventListener('popstate', update);
    return () => {
      removeEventListener(QUERY_EVENT, update);
      removeEventListener('popstate', update);
    };
  }, [readRepo]);
  useEffect(() => {
    if (open || !opener.current) return;
    opener.current.focus();
    opener.current = null;
  }, [open]);
  return (
    <ShellContext.Provider value={shell}>
      <div class="app">
        <Sidebar
          current={route.path}
          theme={theme}
          onTheme={() => setTheme((t) => nextTheme(t))}
          status={status}
          client={api}
          open={narrow && open}
          onClose={narrow ? close : undefined}
          onSearch={() => search('root')}
          repositories={catalog?.repositories ?? []}
          currentRepository={filteredRepo}
        />
        <main class="main" id="main" inert={narrow && open}>
          {route.path === '/status' && <StatusPage client={api} onStatus={setStatus} />}
          {route.path === '/inbox' && <InboxPage client={api} onStatus={setStatus} />}
          {route.path === '/tablo' && <TabloPage client={api} onStatus={setStatus} />}
        </main>
        <CommandBar client={api} open={searching} onClose={() => setSearching(false)} onOpenTask={openIssue} issue={issue} initialMode={mode} />
      </div>
    </ShellContext.Provider>
  );
}
