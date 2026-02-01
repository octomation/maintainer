import { useCallback, useEffect, useRef, useState } from 'preact/hooks';
import type { Column, Envelope, Facets, Task, TaskDetail, TaskQuery, TaskSort } from '../../api/types';
import { COLUMNS, COLUMN_TITLES } from '../../api/types';
import { Estimate, otherLabels, Priority } from '../../components/Chip';
import { IssuePage } from '../../components/IssuePage';
import { decorationOf, RepoAvatar, StatusGlyph } from '../../components/Glyphs';
import { Icon } from '../../components/Icon';
import { Empty, ErrorState, Indexing, ProgressLine, Skeleton } from '../../components/States';
import { DisplayMenu, DisplayRow, DisplaySelect } from '../../components/Display';
import { flagList } from '../../components/TaskDetails';
import { hasFilters, TaskFilters } from '../../components/TaskFilters';
import { TopBar, ViewSwitch } from '../../components/TopBar';
import { useCatalog } from '../../lib/catalog';
import { pushQuery, queryFromSearch, searchFromQuery } from '../../lib/query';
import { closeTask, openTask, useOpenTask } from '../../lib/task';
import type { PageProps } from '../props';
import './tablo.css';

const ORDERINGS: { value: TaskSort | 'default'; label: string }[] = [
  { value: 'default', label: 'Default' },
  { value: 'score', label: 'Score' },
  { value: 'importance', label: 'Priority' },
  { value: 'updated', label: 'Updated' },
  { value: 'created', label: 'Created' },
];

type CardProperty = 'priority' | 'estimate' | 'code' | 'labels' | 'repository';
type CardProperties = Record<CardProperty, boolean>;
const CARD_PROPERTIES: { key: CardProperty; title: string }[] = [
  { key: 'priority', title: 'Priority' },
  { key: 'estimate', title: 'Estimate' },
  { key: 'code', title: 'Code' },
  { key: 'labels', title: 'Labels' },
  { key: 'repository', title: 'Repository' },
];

interface BoardDisplay {
  empty: boolean;
  properties: CardProperties;
}

const DISPLAY_KEY = 'maintainer.tablo.display';
const DEFAULT_DISPLAY: BoardDisplay = { empty: true, properties: { priority: true, estimate: true, code: true, labels: true, repository: true } };

/** The display options of the board; the defaults when none are stored or storage fails. */
function readDisplay(): BoardDisplay {
  try {
    const stored = JSON.parse(localStorage.getItem(DISPLAY_KEY) ?? 'null') as Partial<BoardDisplay> | null;
    if (!stored) return DEFAULT_DISPLAY;
    const properties = { ...DEFAULT_DISPLAY.properties };
    for (const { key } of CARD_PROPERTIES) if (typeof stored.properties?.[key] === 'boolean') properties[key] = stored.properties[key];
    return { empty: stored.empty ?? true, properties };
  } catch {
    return DEFAULT_DISPLAY;
  }
}

function writeDisplay(display: BoardDisplay): void {
  try {
    localStorage.setItem(DISPLAY_KEY, JSON.stringify(display));
  } catch {
    // storage is a convenience
  }
}

/** The cards a column loads per request. */
export const PAGE = 50;
/** The most the service pages at once; a column reload after a triage keeps its pages up to this. */
const MAX_LIMIT = 500;

/** The open columns are ordered by score, the closed ones by the last update. */
export function sortFor(column: Column): TaskSort {
  return column === 'done' || column === 'wont_do' ? 'updated' : 'score';
}

/** The filters and the ordering the board shares with the URL: the column and the paging are the board's own. */
function boardQuery(query: TaskQuery): TaskQuery {
  const next = { ...query };
  delete next.column;
  delete next.limit;
  delete next.offset;
  return next;
}

interface ColumnRequest {
  offset: number;
  limit: number;
  replace: boolean;
}

interface ColumnState {
  tasks: Task[];
  total: number | null;
  /** The query the cards and the total were loaded for; another query makes them stale. */
  query: string;
  loading: boolean;
  /** The request in flight replaces the cards, as after a reset or a triage. */
  replacing: boolean;
  error: unknown;
  /** The request that failed, so Try again repeats it as it was. */
  failed: ColumnRequest | null;
}

type Board = Record<Column, ColumnState>;

function emptyBoard(): Board {
  const board = {} as Board;
  for (const column of COLUMNS) board[column] = { tasks: [], total: null, query: '', loading: true, replacing: true, error: null, failed: null };
  return board;
}

function status(envelope: Envelope<unknown>): string {
  return `generation ${envelope.generation}${envelope.warnings.length ? ` · ${envelope.warnings.length} warnings` : ''}`;
}

/**
 * /tablo: the issues of the collection on one basic board of five columns.
 * One request without a column gives the facets and the column counts; each
 * column then pages its own cards. The pages reset when the filters or the
 * index generation change; a triage from the drawer reloads its column.
 */
export function TabloPage({ client, onStatus }: PageProps) {
  const [query, setQuery] = useState<TaskQuery>(() => ({ state: 'all', ...boardQuery(queryFromSearch(typeof location !== 'undefined' ? location.search : '')) }));
  const [epoch, setEpoch] = useState(0);
  const [facets, setFacets] = useState<Facets | null>(null);
  const { catalog, failed: catalogFailed, retry: retryCatalog } = useCatalog(client);
  const [board, setBoard] = useState<Board>(emptyBoard);
  const [indexing, setIndexing] = useState(false);
  // The summary request is in flight: the board is busy from its start.
  const [summarizing, setSummarizing] = useState(true);
  const [error, setError] = useState<unknown>(null);
  // The counts could not be refreshed after a triage; the board stays as it is.
  const [facetError, setFacetError] = useState<unknown>(null);
  // What the cards show and whether empty columns are, remembered in the browser.
  const [display, setDisplay] = useState<BoardDisplay>(readDisplay);
  const changeDisplay = (next: BoardDisplay) => {
    setDisplay(next);
    writeDisplay(next);
  };
  // The full issue page of a card, addressed in the URL; the board stays mounted under it.
  const task = useOpenTask();
  const opener = useRef<string | null>(null);
  // The content container stays mounted whether it holds the board or the empty view: the drawer's focus returns to it.
  const content = useRef<HTMLDivElement>(null);
  // The token of the current board: an answer of an older filter set or epoch is discarded.
  const token = useRef(0);
  // The generation the board was loaded from; null accepts the next answer's.
  const generation = useRef<number | null>(null);
  const boardRef = useRef(board);
  boardRef.current = board;

  const change = useCallback((next: TaskQuery) => {
    const clean = boardQuery(next);
    setQuery(clean);
    pushQuery(clean);
  }, []);

  const reset = useCallback(() => setEpoch((n) => n + 1), []);

  /** True when the answer comes from another generation than the board: the board reloads. */
  const moved = useCallback(
    (envelope: Envelope<unknown>): boolean => {
      if (generation.current === null) {
        generation.current = envelope.generation;
        return false;
      }
      if (envelope.generation === generation.current) return false;
      generation.current = null;
      reset();
      return true;
    },
    [reset],
  );

  const loadColumn = useCallback(
    (selection: TaskQuery, column: Column, offset: number, limit: number, replace: boolean) => {
      const mine = token.current;
      setBoard((prev) => ({ ...prev, [column]: { ...prev[column], loading: true, replacing: replace, error: null, failed: null } }));
      client
        .listTasks({ ...selection, column: [column], sort: selection.sort ?? sortFor(column), limit, offset })
        .then(({ envelope }) => {
          if (mine !== token.current) return;
          const page = envelope.data;
          if (!page) {
            setBoard((prev) => ({ ...prev, [column]: { ...prev[column], loading: false } }));
            return;
          }
          if (moved(envelope)) return;
          setBoard((prev) => {
            const tasks = replace ? page.tasks : [...prev[column].tasks, ...page.tasks];
            return { ...prev, [column]: { tasks, total: page.total, query: searchFromQuery(selection), loading: false, replacing: false, error: null, failed: null } };
          });
        })
        .catch((err: unknown) => {
          if (mine === token.current) setBoard((prev) => ({ ...prev, [column]: { ...prev[column], loading: false, error: err, failed: { offset, limit, replace } } }));
        });
    },
    [client, moved],
  );

  useEffect(() => {
    token.current += 1;
    const mine = token.current;
    generation.current = null;
    // The board stays on screen until the new one arrives, as the inbox keeps its rows.
    setError(null);
    setFacetError(null);
    setSummarizing(true);
    let timer: ReturnType<typeof setTimeout> | undefined;
    const load = () => {
      client
        .listTasks({ ...query, limit: 1 })
        .then(({ status: code, envelope }) => {
          if (mine !== token.current) return;
          if (code === 202 || envelope.data === null) {
            setIndexing(true);
            timer = setTimeout(load, 1000);
            return;
          }
          setIndexing(false);
          setSummarizing(false);
          generation.current = envelope.generation;
          setFacets(envelope.data.facets);
          onStatus?.(status(envelope));
          retryCatalog();
          for (const column of COLUMNS) loadColumn(query, column, 0, PAGE, true);
        })
        .catch((err: unknown) => {
          if (mine !== token.current) return;
          setIndexing(false);
          setSummarizing(false);
          setError(err);
        });
    };
    load();
    return () => {
      if (timer) clearTimeout(timer);
    };
  }, [client, query, epoch, onStatus, loadColumn, retryCatalog]);

  const key = searchFromQuery(query);
  /** The cards of a column belong to another query until their replacement succeeds. */
  const stale = (column: Column) => board[column].query !== key;

  const loadMore = (column: Column) => {
    const state = board[column];
    if (state.loading || stale(column)) return;
    loadColumn(query, column, state.tasks.length, PAGE, false);
  };

  /** Try again repeats the request that failed, a replacement from the top included. */
  const retryColumn = (column: Column) => {
    const failed = board[column].failed ?? { offset: board[column].tasks.length, limit: PAGE, replace: false };
    loadColumn(query, column, failed.offset, failed.limit, failed.replace);
  };

  const open = useCallback((t: Task) => {
    opener.current = t.id;
    openTask(t.id);
  }, []);
  // Back on the board: the focus returns to the card that was opened, or to the board when it is gone.
  useEffect(() => {
    if (task !== null || opener.current === null) return;
    const id = opener.current;
    opener.current = null;
    const card = Array.from(content.current?.querySelectorAll<HTMLElement>('[data-id]') ?? []).find((node) => node.dataset.id === id);
    (card ?? content.current)?.focus();
  }, [task]);

  /** The counts after a triage; a failure is told, not swallowed. */
  const refreshFacets = useCallback(() => {
    generation.current = null;
    const mine = token.current;
    setFacetError(null);
    client
      .listTasks({ ...query, limit: 1 })
      .then(({ envelope }) => {
        if (mine !== token.current || !envelope.data || moved(envelope)) return;
        setFacets(envelope.data.facets);
        onStatus?.(status(envelope));
      })
      .catch((err: unknown) => {
        if (mine === token.current) setFacetError(err);
      });
  }, [client, query, onStatus, moved]);

  /** After a triage the index has a new generation: the counts and the task's column are requested again, with the pages it had. */
  const changed = useCallback(
    (detail: TaskDetail) => {
      refreshFacets();
      const loaded = boardRef.current[detail.column].tasks.length;
      loadColumn(query, detail.column, 0, Math.min(MAX_LIMIT, Math.max(PAGE, loaded)), true);
    },
    [query, refreshFacets, loadColumn],
  );

  const count = (column: Column) => facets?.columns.find((f) => f.value === column)?.count ?? board[column].total ?? 0;
  // Busy: a column is being filled or replaced, as on a reset that kept the board on screen.
  const busy = facets !== null && (summarizing || COLUMNS.some((column) => board[column].loading && (board[column].tasks.length === 0 || board[column].replacing)));
  const settled = facets !== null && !summarizing && COLUMNS.every((column) => !board[column].loading && board[column].total !== null && !stale(column));
  const nothing = settled && COLUMNS.every((column) => board[column].total === 0);

  return (
    <section class={`page tablo${task ? ' page--issue' : ''}`}>
      <TopBar
        title="Tablo"
        actions={
          <>
            <ViewSwitch current="board" search={searchFromQuery({ ...query, sort: undefined })} />
            <DisplayMenu>
              {() => (
                <>
                  <DisplaySelect label="Ordering" value={query.sort ?? 'default'} options={ORDERINGS} onChange={(value) => change({ ...query, sort: value === 'default' ? undefined : value })} />
                  <p class="display-menu__note">The default orders the open columns by score and Done and Won't do by the last update.</p>
                  <DisplayRow label="Empty columns">
                    <input type="checkbox" role="switch" class="switch" aria-label="Show empty columns" checked={display.empty} onChange={() => changeDisplay({ ...display, empty: !display.empty })} />
                  </DisplayRow>
                  <div class="display-menu__section">
                    <DisplayRow label="Display properties" id="board-properties">
                      <span />
                    </DisplayRow>
                    <div class="ordering" role="group" aria-labelledby="board-properties">
                      {CARD_PROPERTIES.map((property) => (
                        <button
                          type="button"
                          key={property.key}
                          class={`ordering__key${display.properties[property.key] ? ' is-active' : ''}`}
                          aria-pressed={display.properties[property.key]}
                          onClick={() => changeDisplay({ ...display, properties: { ...display.properties, [property.key]: !display.properties[property.key] } })}
                        >
                          {property.title}
                        </button>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </DisplayMenu>
          </>
        }
      >
        <TaskFilters value={query} facets={facets} catalog={catalog} onChange={change} defaultState="all" />
      </TopBar>
      {(indexing || busy) && <ProgressLine />}
      {facets && (indexing || busy) && (
        <p class="sr-only" role="status">
          {indexing ? 'Indexing the mirrors…' : 'Loading…'}
        </p>
      )}
      {catalogFailed && !catalog && (
        <div class="page__notices">
          <div class="notice notice--error" role="alert">
            <p>The catalog could not be loaded, so milestones show their keys.</p>
            <button type="button" class="button" onClick={retryCatalog}>
              Try again
            </button>
          </div>
        </div>
      )}
      {error && facets && (
        <div class="page__notices">
          <div class="notice notice--error" role="alert">
            <p>The board could not be refreshed: {error instanceof Error ? error.message : String(error)}</p>
            <button type="button" class="button" onClick={reset}>
              Try again
            </button>
          </div>
        </div>
      )}
      {facetError !== null && (
        <div class="page__notices">
          <div class="notice notice--error" role="alert">
            <p>The counts could not be refreshed: {facetError instanceof Error ? facetError.message : String(facetError)}</p>
            <button type="button" class="button" onClick={refreshFacets}>
              Try again
            </button>
          </div>
        </div>
      )}
      {error && !facets ? (
        <div class="page__notices">
          <ErrorState error={error} onRetry={reset} />
        </div>
      ) : indexing && !facets ? (
        <Indexing shape="cards" />
      ) : !facets ? (
        <div class="page__content board" role="status" aria-live="polite">
          <span class="sr-only">Loading the board…</span>
          {COLUMNS.map((column) => (
            <div class="column" key={column} aria-hidden="true">
              <div class="column__head">
                <StatusGlyph column={column} />
                <span class="skeleton column__skeleton-title" />
              </div>
              <div class="column__cards">
                <Skeleton shape="cards" count={3} />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div class={`page__content${nothing && hasFilters(query) ? '' : ' board'}`} ref={content} tabIndex={-1}>
          {nothing && hasFilters(query) ? (
            <Empty
              action={
                <button type="button" class="button" onClick={() => change({ state: query.state })}>
                  Clear filters
                </button>
              }
            >
              No tasks match the filters.
            </Empty>
          ) : (
            COLUMNS.filter((column) => display.empty || count(column) > 0 || board[column].loading || board[column].error).map((column) => (
              <BoardColumn key={column} column={column} count={count(column)} state={board[column]} stale={stale(column)} onOpen={open} onMore={() => loadMore(column)} onRetry={() => retryColumn(column)} search={key} properties={display.properties} />
            ))
          )}
        </div>
      )}
      {task && <IssuePage client={client} id={task} onChanged={changed} onBack={closeTask} backLabel="Back to the board" />}
    </section>
  );
}

interface BoardColumnProps {
  column: Column;
  count: number;
  state: ColumnState;
  /** The cards belong to the previous query: shown dimmed, with no Load more, until replaced. */
  stale: boolean;
  onOpen: (task: Task) => void;
  onMore: () => void;
  onRetry: () => void;
  /** The board's filters, for the column's link to the inbox. */
  search: string;
  properties: CardProperties;
}

function BoardColumn({ column, count, state, stale, onOpen, onMore, onRetry, search, properties }: BoardColumnProps) {
  const title = COLUMN_TITLES[column];
  const inbox = new URLSearchParams(search);
  inbox.set('column', column);
  const more = state.total !== null && state.tasks.length < state.total;
  return (
    <section class={`column column--${column}${stale ? ' column--stale' : ''}`} aria-label={title} aria-busy={state.loading}>
      <header class="column__head">
        <StatusGlyph column={column} />
        <h2 class="column__title">{title}</h2>
        <span class="column__count" aria-label={`${count} tasks`}>
          {count}
        </span>
        <a class="icon-button column__action" href={`/inbox?${inbox.toString()}`} aria-label={`Open ${title} in the inbox`} title={`Open ${title} in the inbox`}>
          <Icon name="list" />
        </a>
      </header>
      {stale && state.tasks.length > 0 && <p class="column__stale">Previous results</p>}
      <ul class="column__cards">
        {state.tasks.map((task) => (
          <Card key={task.key} task={task} onOpen={onOpen} properties={properties} />
        ))}
      </ul>
      {state.error ? (
        <div class="column__state">
          <div class="notice notice--error" role="alert">
            <p>{state.error instanceof Error ? state.error.message : String(state.error)}</p>
            <button type="button" class="button" onClick={onRetry}>
              Try again
            </button>
          </div>
        </div>
      ) : state.loading ? (
        <div class="column__state" role="status" aria-live="polite">
          <span class="sr-only">{state.tasks.length ? 'Loading more…' : 'Loading…'}</span>
          <Skeleton shape="cards" count={state.tasks.length ? 1 : 3} />
        </div>
      ) : state.tasks.length === 0 ? (
        <div class="column__state column__state--empty">No tasks</div>
      ) : (
        more && (
          <div class="column__state">
            <button type="button" class="button button--ghost column__more" onClick={onMore} disabled={stale} aria-label={`Load more in ${title}`} title={stale ? 'Waiting for the results of the new filters' : undefined}>
              Load more ({state.total! - state.tasks.length} left)
            </button>
          </div>
        )
      )}
    </section>
  );
}

/** The label chips a card shows before it counts the rest: one fits beside the code at 320 px. */
const CARD_LABELS = 1;

/**
 * A card of about 100 px: the title on up to two lines with the
 * repository's circle at the top right, then one footer that never wraps:
 * the priority, the estimate, the code, up to two labels and "+N". The
 * milestone and the repository's name are on the issue page.
 */
function Card({ task, onOpen, properties }: { task: Task; onOpen: (task: Task) => void; properties: CardProperties }) {
  const flags = flagList(task);
  const reasons = flags.map((f) => `${f.title}: ${f.reason}`).join(' ');
  const labels = properties.labels ? otherLabels(task.labels) : [];
  const shown = labels.slice(0, CARD_LABELS);
  const hidden = labels.slice(CARD_LABELS);
  return (
    <li class="card-item">
      <button type="button" class="card" data-id={task.id} onClick={() => onOpen(task)}>
        <span class="card__head">
          <span class="card__title">{task.title}</span>
          {properties.repository && (
            <span class="card__avatar" title={task.repository}>
              <RepoAvatar repository={task.repository} size={16} />
              <span class="sr-only">{task.repository}</span>
            </span>
          )}
        </span>
        <span class="card__footer">
          {properties.priority && <Priority labels={task.labels} />}
          {properties.estimate && <Estimate labels={task.labels} />}
          {properties.code && <span class="card__code">{task.code}</span>}
          {flags.length > 0 && (
            <span class="card__flag" title={reasons} aria-label={`Flagged, ${reasons}`}>
              <Icon name="flag" size={12} />
              <span class="sr-only">{flags.map((f) => f.title).join(', ')}</span>
            </span>
          )}
          {shown.map((label) => (
            <span class="board-chip" key={label} title={label}>
              <span class="label-chip__dot" style={{ '--dot': `var(--${decorationOf(label)})` }} aria-hidden="true" />
              <span class="board-chip__text">{label}</span>
            </span>
          ))}
          {hidden.length > 0 && (
            <span class="board-chip board-chip--more" title={hidden.join(', ')} aria-label={`${hidden.length} more labels: ${hidden.join(', ')}`}>
              +{hidden.length}
            </span>
          )}
        </span>
      </button>
    </li>
  );
}
