import { useCallback, useEffect, useMemo, useRef, useState } from 'preact/hooks';
import type { CheckoutRow, CheckoutsData, SortSpec } from '../../api/types';
import { CHECKOUT_COLUMNS, CHECKOUT_COLUMN_TITLES, type CheckoutColumn } from '../../api/types';
import { DisplayMenu, DisplayRow, DisplaySelect } from '../../components/Display';
import { CheckoutGlyph, RepoAvatar } from '../../components/Glyphs';
import { Icon } from '../../components/Icon';
import { Drawer } from '../../components/Drawer';
import { Empty, ErrorState, Loading, ProgressLine } from '../../components/States';
import { Popover } from '../../components/Popover';
import { TopBar, ViewSwitch } from '../../components/TopBar';
import type { PageProps } from '../props';
import { CHECKOUT_GROUP_TITLES, columnTemplate, cycleSort, formatTime, groupOf, groupRows, type Grouping, type Layout, layoutFromSearch, readColumns, sortParam, sortSummary, STATUS_COLUMNS, type StatusColumns, summaryLine, writeColumns } from './model';
import './status.css';

/** How often the page asks again while the service collects. */
const POLL_MS = 1000;
/** How long the search box waits before it sends `q`. */
const SEARCH_DEBOUNCE_MS = 120;

const GROUPINGS: { value: Grouping; label: string }[] = [
  { value: 'state', label: 'State' },
  { value: 'owner', label: 'Owner' },
  { value: 'none', label: 'No grouping' },
];

/**
 * /status: the checkouts of `maintainer status` as a list grouped by their
 * state. The page sends the ordering chain and the query; the order, the
 * cells and the summary come from the server, which answers from its cached
 * snapshot without a new Git collection; the groups keep the server's order
 * inside. Refresh asks for a collection and the page polls until it is done.
 */
export function StatusPage({ client, onStatus }: PageProps) {
  const [sort, setSort] = useState<SortSpec[]>([]);
  const [input, setInput] = useState('');
  const [q, setQ] = useState('');
  const [grouping, setGrouping] = useState<Grouping>('state');
  // The list or the board of the checkouts, in the URL as layout=board.
  const [layout, setLayout] = useState<Layout>(() => layoutFromSearch(typeof location !== 'undefined' ? location.search : ''));
  // A board column's action: the list again, with the focus on that group's first row.
  const focusGroup = useRef<string | null>(null);
  const showInList = (key: string) => {
    focusGroup.current = key;
    changeLayout('list');
  };
  useEffect(() => {
    if (layout !== 'list' || focusGroup.current === null) return;
    const header = document.getElementById(`group-${focusGroup.current}`);
    focusGroup.current = null;
    header?.scrollIntoView?.({ block: 'start' });
    (header?.parentElement ?? document.querySelector('.checkouts'))?.querySelector<HTMLElement>('[role="option"]')?.focus();
  }, [layout]);
  const changeLayout = useCallback((next: Layout) => {
    setLayout(next);
    const params = new URLSearchParams(location.search);
    if (next === 'board') params.set('layout', 'board');
    else params.delete('layout');
    const search = params.toString();
    history.replaceState(null, '', `${location.pathname}${search ? `?${search}` : ''}`);
  }, []);
  const [data, setData] = useState<CheckoutsData | null>(null);
  const [collectedAt, setCollectedAt] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [error, setError] = useState<unknown>(null);
  // The failed request: Try again repeats a collection when that is what failed.
  const [failed, setFailed] = useState<'load' | 'refresh'>('load');
  const [attempt, setAttempt] = useState(0);
  const [selected, setSelected] = useState<CheckoutRow | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => setQ(input.trim()), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [input]);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const load = () => {
      client
        .getCheckouts({ q, sort: sortParam(sort) })
        .then(({ status, envelope }) => {
          if (cancelled) return;
          setWarnings(envelope.warnings);
          setCollectedAt(envelope.collected_at);
          setError(null);
          if (status === 202 || envelope.data === null) {
            setRefreshing(true);
            timer = setTimeout(load, POLL_MS);
            return;
          }
          setData(envelope.data);
          setRefreshing(envelope.refreshing);
          if (envelope.refreshing) timer = setTimeout(load, POLL_MS);
        })
        .catch((err: unknown) => {
          if (cancelled) return;
          setError(err);
          setFailed('load');
          // A failed read ends the polling: Refresh is usable again.
          setRefreshing(false);
        });
    };
    load();
    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, [client, q, sort, attempt]);

  useEffect(() => {
    if (!onStatus) return;
    const at = collectedAt ? `collected ${formatTime(collectedAt)}` : '';
    onStatus(refreshing ? [at, 'collecting…'].filter(Boolean).join(' · ') : at);
  }, [onStatus, collectedAt, refreshing]);

  const refresh = useCallback(() => {
    setRefreshing(true);
    setError(null);
    client
      .refreshCheckouts()
      .then(() => setAttempt((n) => n + 1))
      .catch((err: unknown) => {
        setRefreshing(false);
        setFailed('refresh');
        setError(err);
      });
  }, [client]);
  const retry = useCallback(() => {
    if (failed === 'refresh') refresh();
    else setAttempt((n) => n + 1);
  }, [failed, refresh]);

  const onSort = useCallback((key: string, additive: boolean) => {
    setSort((keys) => cycleSort(keys, key, additive));
  }, []);

  const chain = data?.sort ?? sort;

  const notices = useMemo(() => {
    const seen = new Set<string>();
    const out: string[] = [];
    for (const text of [data?.error ?? '', ...warnings]) {
      if (text && !seen.has(text)) {
        seen.add(text);
        out.push(text);
      }
    }
    return out;
  }, [data, warnings]);

  const groups = useMemo(() => (data ? groupRows(data.rows, grouping) : []), [data, grouping]);
  // The properties of a row in columns of fixed width, measured from the values shown, the same for every group.
  const [columns, setColumns] = useState<StatusColumns>(readColumns);
  const changeColumns = (next: StatusColumns) => {
    setColumns(next);
    writeColumns(next);
  };
  // Measured in Inter once the typeface is in; until then the estimate holds.
  const [fonts, setFonts] = useState(0);
  useEffect(() => {
    let alive = true;
    void document.fonts?.ready?.then(() => alive && setFonts((n) => n + 1));
    return () => {
      alive = false;
    };
  }, []);
  const grid = useMemo(() => columnTemplate(data?.rows ?? [], columns), [data, columns, fonts]);
  const message = error instanceof Error ? error.message : error ? String(error) : '';
  const close = useCallback(() => setSelected(null), []);

  return (
    <section class="page status">
      <TopBar
        title="Status"
        actions={
          <>
            <ViewSwitch current={layout} onChange={changeLayout} />
            <DisplayMenu>
              {() => (
                <>
                  <DisplaySelect label="Grouping" value={grouping} options={GROUPINGS} onChange={setGrouping} />
                  <DisplayRow label="Ordering" id="status-ordering">
                    <span class="display-menu__muted" data-testid="sort-summary">
                      {data ? sortSummary(data.sort, data.query) : 'Sort: default (repository / path)'}
                    </span>
                  </DisplayRow>
                  <div class="display-menu__section">
                  <DisplayRow label="Display properties" id="status-properties">
                    <span />
                  </DisplayRow>
                  <div class="ordering" role="group" aria-labelledby="status-properties">
                    {STATUS_COLUMNS.map((column) => (
                      <button
                        type="button"
                        key={column.key}
                        class={`ordering__key${columns[column.key] ? ' is-active' : ''}`}
                        aria-pressed={columns[column.key]}
                        onClick={() => changeColumns({ ...columns, [column.key]: !columns[column.key] })}
                      >
                        {column.title}
                      </button>
                    ))}
                  </div>
                </div>
                <div class="ordering" role="group" aria-labelledby="status-ordering">
                    {CHECKOUT_COLUMNS.map((key) => {
                      const index = chain.findIndex((spec) => spec.column === key);
                      const spec = index >= 0 ? chain[index] : null;
                      const title = CHECKOUT_COLUMN_TITLES[key];
                      return (
                        <button
                          type="button"
                          key={key}
                          class={`ordering__key${spec ? ' is-active' : ''}`}
                          aria-pressed={spec !== null}
                          aria-label={spec ? `${title}, sorted ${spec.desc ? 'descending' : 'ascending'}, priority ${index + 1}` : title}
                          title={`Order by ${title.toLowerCase()}: a click or Enter cycles ascending, descending, off; with Shift it joins the chain after the current keys`}
                          onClick={(event) => onSort(key, event.shiftKey)}
                        >
                          <span>{title}</span>
                          {spec && (
                            <span class="ordering__mark" aria-hidden="true">
                              <Icon name={spec.desc ? 'arrow-down' : 'arrow-up'} size={12} />
                              {index + 1}
                            </span>
                          )}
                        </button>
                      );
                    })}
                  </div>
                  <div class="display-menu__section">
                    {data && (
                      <p class="display-menu__note" data-testid="summary">
                        {data.query.trim() !== '' && <span>shown {data.rows.length} / {data.total} · </span>}
                        {summaryLine(data.summary)}
                      </p>
                    )}
                    <DisplayRow label="Snapshot">
                      <span class="display-menu__muted" role="status" aria-live="polite">
                        {collectedAt ? `snapshot ${formatTime(collectedAt)}` : refreshing ? 'no snapshot yet' : ''}
                        {collectedAt && refreshing ? ' · collecting…' : ''}
                      </span>
                    </DisplayRow>
                  </div>
                  <div class="display-menu__footer">
                    <button type="button" class="button button--ghost" onClick={() => setSort([])} disabled={chain.length === 0}>
                      Reset to default
                    </button>
                    <span class="display-menu__spacer" />
                    <button type="button" class="button" onClick={refresh} disabled={refreshing} aria-busy={refreshing}>
                      {refreshing ? 'Collecting…' : 'Refresh'}
                    </button>
                  </div>
                </>
              )}
            </DisplayMenu>
          </>
        }
      >
        <form class="filters" role="search" aria-label="Checkout search" onSubmit={(event) => event.preventDefault()}>
          <Popover
            role="group"
            label="Search the checkouts"
            class="filters__menu"
            trigger={(props) => (
              <button type="button" {...props} class="filter-button" title="Search the checkouts">
                <Icon name="plus" size={12} />
                <span class="filter-button__text">Filter</span>
              </button>
            )}
          >
            {() => (
              <div class="filter-menu">
                <div class="menu-search">
                  <input type="search" data-autofocus aria-label="Search" placeholder="Repository, branch, path…" value={input} onInput={(event) => setInput((event.target as HTMLInputElement).value)} />
                </div>
                <div class="menu-heading">A fuzzy match, as the search of maintainer status.</div>
              </div>
            )}
          </Popover>
          {input.trim() !== '' && (
            <div class="filters__chips" role="group" aria-label="Active filters">
              <span class="filter-chip">
                <span class="filter-chip__dim">
                  <Icon name="search" size={12} />
                  Search
                </span>
                <span class="filter-chip__op">matches</span>
                <span class="filter-chip__value">“{input.trim()}”</span>
                <button type="button" class="filter-chip__remove" aria-label="Remove search filter" title="Remove search filter" onClick={() => setInput('')}>
                  <Icon name="x" size={12} />
                </button>
              </span>
            </div>
          )}
          <span class="filters__spacer" />
        </form>
      </TopBar>
      {refreshing && <ProgressLine />}
      <div class="page__notices">
        {error && data && (
          <div class="notice notice--error" role="alert">
            <p>
              {failed === 'refresh' ? 'The collection could not be started: ' : ''}
              {message}
            </p>
            <button type="button" class="button" onClick={retry}>
              Try again
            </button>
          </div>
        )}
        {notices.map((text) => (
          <div key={text} class="notice" role="alert">
            <p>The last collection failed: {text}</p>
          </div>
        ))}
      </div>
      {error && !data ? (
        <div class="page__notices">
          <ErrorState error={error} onRetry={retry} />
        </div>
      ) : !data ? (
        <div class="page__content">
          <Loading shape="table" count={8}>
            Reading local repository status…
          </Loading>
        </div>
      ) : (
        <div class="page__content status__content">
          {data.rows.length === 0 ? (
            <Empty
              action={
                data.total > 0 && (
                  <button type="button" class="button" onClick={() => setInput('')}>
                    Clear search
                  </button>
                )
              }
            >
              {data.total === 0 ? 'No checkouts in the collection.' : 'No checkout matches the search.'}
            </Empty>
          ) : layout === 'board' ? (
            <div class="board status__board" aria-label="Checkouts">
              {groups.map((group) => (
                <section class="column" key={group.key} aria-label={group.title || 'Checkouts'}>
                  <header class="column__head">
                    {group.glyph && <CheckoutGlyph group={group.glyph} />}
                    <h2 class="column__title">{group.title || 'Checkouts'}</h2>
                    <span class="column__count" aria-label={`${group.rows.length} checkouts`}>
                      {group.rows.length}
                    </span>
                    <button type="button" class="icon-button column__action" onClick={() => showInList(group.key)} aria-label={`Show ${group.title || 'the checkouts'} in the list`} title={`Show ${group.title || 'the checkouts'} in the list`}>
                      <Icon name="list" size={14} />
                    </button>
                  </header>
                  <ul class="column__cards">
                    {group.rows.map((row) => (
                      <CheckoutCard key={rowKey(row)} row={row} onOpen={() => setSelected(row)} />
                    ))}
                  </ul>
                </section>
              ))}
            </div>
          ) : (
            <div class="checkouts" role="listbox" aria-label="Checkouts" style={{ '--checkout-columns': grid }}>
              {groups.map((group) => (
                <div class="checkouts__group" role="group" aria-labelledby={group.title ? `group-${group.key}` : undefined} aria-label={group.title ? undefined : 'Checkouts'} key={group.key}>
                  {group.title && (
                    <div class="group-header" id={`group-${group.key}`}>
                      {group.glyph && <CheckoutGlyph group={group.glyph} />}
                      <span class="group-header__name">{group.title}</span>
                      <span class="group-header__count">{group.rows.length}</span>
                    </div>
                  )}
                  {group.rows.map((row) => (
                    <CheckoutItem key={rowKey(row)} row={row} selected={selected !== null && rowKey(selected) === rowKey(row)} onOpen={() => setSelected(row)} columns={columns} />
                  ))}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
      <Drawer open={selected !== null} title={selected?.repository ?? ''} onClose={close} size="narrow">
        {selected && <CheckoutDetails row={selected} />}
      </Drawer>
    </section>
  );
}

function rowKey(row: CheckoutRow): string {
  return `${row.repository}\u0000${row.path}\u0000${row.branch}`;
}

function rowClass(row: CheckoutRow): string {
  if (row.error) return 'is-error';
  if (row.orphan_reason) return 'is-orphan';
  if (row.placeholder) return 'is-placeholder';
  return '';
}

/** The tone of a status word: the default for synced, warn for ahead, behind and orphan, danger for an error. */
function statusTone(row: CheckoutRow): string {
  if (row.error) return 'status-word status-word--danger';
  if (row.orphan_reason || row.ahead > 0 || row.behind > 0 || /ahead|behind|diverged|unborn/.test(row.status)) return 'status-word status-word--warn';
  return 'status-word';
}

/**
 * A 44 px row with the fixed slots of the list: the push lock, the
 * divergence, the state glyph beside the title, the repository as the title and the
 * branch, the uncommitted counts, the status word and the path. Every text
 * is the server's cell for that column.
 */
function CheckoutItem({ row, selected, onOpen, columns }: { row: CheckoutRow; selected: boolean; onOpen: () => void; columns: StatusColumns }) {
  const cell = (column: CheckoutColumn) => row.cells[CHECKOUT_COLUMNS.indexOf(column)] ?? '';
  const group = groupOf(row);
  const divergence = [row.ahead > 0 ? `↑${row.ahead}` : '', row.behind > 0 ? `↓${row.behind}` : ''].filter(Boolean).join(' ');
  return (
    <div
      class={`list-row checkout ${rowClass(row)}`.trim()}
      role="option"
      aria-selected={selected}
      tabIndex={0}
      onClick={onOpen}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault();
          onOpen();
        }
      }}
    >
      <span class="checkout__lock" data-cell="lock">
        {row.push_locked ? <Icon name="lock" size={16} title="Push locked" /> : cell('lock') ? <span class="sr-only">{cell('lock')}</span> : null}
      </span>
      <span class="checkout__divergence" title={divergence ? `ahead ${row.ahead}, behind ${row.behind}` : undefined}>
        {divergence}
      </span>
      <CheckoutGlyph group={group} title={CHECKOUT_GROUP_TITLES[group]} size={12} />
      <span class="checkout__cell checkout__repo" data-cell="repository" title={cell('repository')}>
        {cell('repository')}
      </span>
      {columns.branch && (
        <span class="checkout__cell checkout__branch" data-cell="branch" title={cell('branch')}>
          {cell('branch')}
        </span>
      )}
      {columns.changes && (
        <span class="checkout__cell checkout__changes" data-cell="uncommitted" title={cell('uncommitted')}>
          {cell('uncommitted')}
        </span>
      )}
      {columns.status && (
        <span class={`checkout__cell checkout__status ${statusTone(row)}`} data-cell="status" title={cell('status')}>
          {cell('status')}
        </span>
      )}
      <span class="checkout__cell checkout__path" data-cell="path" title={columns.path ? cell('path') : undefined}>
        {columns.path ? cell('path') : ''}
      </span>
    </div>
  );
}

/**
 * A card of the status board: the repository as the title with its circle,
 * then one footer of the branch, the divergence (or the status word when the
 * state has no counts), the uncommitted counts where there are any, and the
 * lock; the server's cells, as in the list.
 */
function CheckoutCard({ row, onOpen }: { row: CheckoutRow; onOpen: () => void }) {
  const cell = (column: CheckoutColumn) => row.cells[CHECKOUT_COLUMNS.indexOf(column)] ?? '';
  const divergence = [row.ahead > 0 ? `↑${row.ahead}` : '', row.behind > 0 ? `↓${row.behind}` : ''].filter(Boolean).join(' ');
  // The counts only where something is uncommitted, or where they are unknown after an error.
  const changed = row.error !== undefined || row.added + row.deleted + row.changed_files + row.untracked + row.binary + row.conflicts > 0;
  return (
    <li class="card-item">
      <button type="button" class={`card checkout-card ${rowClass(row)}`.trim()} onClick={onOpen}>
        <span class="card__head">
          <span class="card__title" data-cell="repository">
            {cell('repository')}
          </span>
          <span class="card__avatar" aria-hidden="true">
            <RepoAvatar repository={row.repository} size={16} />
          </span>
        </span>
        <span class="card__footer">
          {cell('branch') && (
            <span class="board-chip" data-cell="branch" title={cell('branch')}>
              <Icon name="git-branch" size={12} />
              <span class="board-chip__text">{cell('branch')}</span>
            </span>
          )}
          {divergence && (
            <span class="card__code" title={`ahead ${row.ahead}, behind ${row.behind}`}>
              {divergence}
            </span>
          )}
          {changed && cell('uncommitted') && (
            <span class="card__code" data-cell="uncommitted" title={cell('uncommitted')}>
              {cell('uncommitted')}
            </span>
          )}
          {row.push_locked && <Icon name="lock" size={14} title="Push locked" />}
          {(row.error || (!divergence && groupOf(row) === 'diverged')) && (
            <span class={`card__code ${statusTone(row)}`} data-cell="status" title={cell('status')}>
              {cell('status')}
            </span>
          )}
        </span>
      </button>
    </li>
  );
}

/** The peek panel: everything the row carries, as the TUI footer shows it. */
function CheckoutDetails({ row }: { row: CheckoutRow }) {
  const branch = row.branch || '—';
  const uncommitted = [
    `+${row.added}/-${row.deleted} tracked lines`,
    row.changed_files > 0 ? `${row.changed_files} changed files` : '',
    row.untracked > 0 ? `${row.untracked} untracked` : '',
    row.binary > 0 ? `${row.binary} binary` : '',
    row.conflicts > 0 ? `${row.conflicts} conflicts` : '',
  ]
    .filter(Boolean)
    .join(' · ');
  const divergence = [row.ahead > 0 ? `ahead ${row.ahead}` : '', row.behind > 0 ? `behind ${row.behind}` : ''].filter(Boolean).join(' · ');
  const group = groupOf(row);
  return (
    <dl class="props__list checkout-props">
      <Prop name="State">
        <span class="prop__value">
          <CheckoutGlyph group={group} />
          {CHECKOUT_GROUP_TITLES[group]}
        </span>
      </Prop>
      <Prop name="Repository">{row.repository}</Prop>
      <Prop name="Path" mono>
        {row.path}
      </Prop>
      {row.display_path && row.display_path !== row.path && (
        <Prop name="Shown as" mono>
          {row.display_path}
        </Prop>
      )}
      <Prop name="Branch" mono>
        {branch}
        {row.branch && row.branch === row.default_branch ? ' (default)' : ''}
      </Prop>
      {row.default_branch && row.branch !== row.default_branch && (
        <Prop name="Default branch" mono>
          {row.default_branch}
        </Prop>
      )}
      <Prop name="Upstream" mono>
        {row.upstream || '—'}
      </Prop>
      <Prop name="Commit" mono>
        {row.commit || '—'}
      </Prop>
      <Prop name="Status">
        <span class={statusTone(row)}>{row.error ? `error: ${row.error}` : divergence ? `${row.status} · ${divergence}` : row.status}</span>
      </Prop>
      {!row.error && !row.placeholder && (
        <Prop name="Uncommitted" mono>
          {uncommitted}
        </Prop>
      )}
      <Prop name="Push lock">
        {row.push_locked ? (
          <span class="prop__value">
            <Icon name="lock" size={14} /> locked
          </span>
        ) : (
          '—'
        )}
      </Prop>
      <Prop name="Pinned">{row.pinned ? 'yes' : 'no'}</Prop>
      {row.orphan_reason && <Prop name="Orphan">{row.orphan_reason}</Prop>}
      {row.active_path && (
        <Prop name="Active path" mono>
          {row.active_path}
        </Prop>
      )}
      {row.remote_checked_at && (
        <Prop name="GitHub check (cached)" mono>
          {formatTime(row.remote_checked_at)}
        </Prop>
      )}
      {row.error && (
        <Prop name="Error" tone="error">
          {row.error}
        </Prop>
      )}
    </dl>
  );
}

function Prop({ name, mono, tone, children }: { name: string; mono?: boolean; tone?: 'error'; children: preact.ComponentChildren }) {
  return (
    <div class="prop prop--stacked">
      <dt class="prop__name">{name}</dt>
      <dd class={['prop__body', mono ? 'mono' : '', tone === 'error' ? 'is-error' : ''].filter(Boolean).join(' ')}>{children}</dd>
    </div>
  );
}
