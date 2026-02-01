import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'preact/hooks';
import type { Task, TaskDetail, TaskPage, TaskQuery, TaskSort } from '../../api/types';
import { COLUMN_TITLES } from '../../api/types';
import { DisplayMenu, DisplaySelect } from '../../components/Display';
import { impactOf, priorityTitle } from '../../components/Chip';
import { RepoAvatar, StatusGlyph } from '../../components/Glyphs';
import { Icon } from '../../components/Icon';
import { IssuePage } from '../../components/IssuePage';
import { Empty, ErrorState, Indexing, Loading, ProgressLine } from '../../components/States';
import { age, date, flagList, TaskDetails } from '../../components/TaskDetails';
import { hasFilters, TaskFilters } from '../../components/TaskFilters';
import { TopBar } from '../../components/TopBar';
import { useCatalog } from '../../lib/catalog';
import { NARROW, matches, useMediaQuery } from '../../lib/media';
import { pushQuery, queryFromSearch, replaceSearch, searchFromQuery } from '../../lib/query';
import { closeTask, openTask, taskFromSearch, useOpenTask } from '../../lib/task';
import type { PageProps } from '../props';
import { groupingFromSearch, GROUPINGS, groupTasks, type TaskGrouping } from './grouping';
import './inbox.css';

/** The rows of one list answer; "Load more" asks for the next ones. */
export const PAGE_SIZE = 100;
/** The service caps a page here, so a requery after a triage keeps at most this many rows. */
const MAX_LIMIT = 500;

const SORTS: { value: TaskSort; label: string }[] = [
  { value: 'score', label: 'Score' },
  { value: 'importance', label: 'Priority' },
  { value: 'updated', label: 'Updated' },
  { value: 'created', label: 'Created' },
];

/**
 * /inbox: the issues of every mirrored repository as a mail app. The list on
 * the left, the issue page on the right, one border between; the filters
 * and the ordering live in the URL, the selection does not.
 */
export function InboxPage({ client, onStatus }: PageProps) {
  const [query, setQuery] = useState<TaskQuery>(() => queryFromSearch(typeof location !== 'undefined' ? location.search : ''));
  const [page, setPage] = useState<TaskPage | null>(null);
  // The query the rows and the count belong to: a failed request leaves them as they were.
  const [shownFor, setShownFor] = useState<TaskQuery>(query);
  const { catalog, failed: catalogFailed, retry: retryCatalog } = useCatalog(client);
  const [tasks, setTasks] = useState<Task[]>([]);
  // The grouping of the list, `group` in the URL.
  const [grouping, setGrouping] = useState<TaskGrouping>(() => groupingFromSearch(typeof location !== 'undefined' ? location.search : ''));
  const changeGrouping = useCallback((next: TaskGrouping) => {
    setGrouping(next);
    const params = new URLSearchParams(location.search);
    if (next === 'none') params.delete('group');
    else params.set('group', next);
    replaceSearch(params);
  }, []);
  const [selected, setSelected] = useState<string | null>(() => (typeof location !== 'undefined' ? taskFromSearch(location.search) : null));
  // The full issue page, addressed in the URL; back from it keeps that task selected.
  const full = useOpenTask();
  const fromFull = useRef<string | null>(null);
  useEffect(() => {
    if (full) {
      fromFull.current = full;
      setSelected(full);
      return;
    }
    if (fromFull.current === null) return;
    const id = fromFull.current;
    fromFull.current = null;
    setTimeout(() => focusRow(id), 0);
  }, [full]);
  // The narrow layout: the pane opens over the list, which then goes inert.
  const narrow = useMediaQuery(NARROW);
  const [indexing, setIndexing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [more, setMore] = useState(false);
  // Any list request in flight: "Load more" waits for it.
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<unknown>(null);
  // A failed refresh of a list that is already shown: the rows stay, the failure is told.
  const [refreshError, setRefreshError] = useState<unknown>(null);
  const [attempt, setAttempt] = useState(0);
  // Every list request takes a number; an answer to an older request is discarded.
  const sequence = useRef(0);
  const loaded = useRef(0);
  // The query the rows were loaded for: rows of another query are never appended to.
  const rowsFor = useRef('');
  const list = useRef<HTMLOListElement>(null);
  const pane = useRef<HTMLElement>(null);
  const focusPane = useRef(false);
  const hasRows = useRef(false);
  const returnTo = useRef<string | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const request = useCallback(
    (paging: { offset: number; limit: number }, mode: 'replace' | 'append' | 'requery') => {
      const seq = ++sequence.current;
      const key = searchFromQuery(query);
      if (timer.current) clearTimeout(timer.current);
      setPending(true);
      if (mode === 'append') setMore(true);
      else if (mode === 'replace') setLoading(true);
      client
        .listTasks({ ...query, ...paging })
        .then(({ status, envelope }) => {
          if (seq !== sequence.current) return;
          if (status === 202 || envelope.data === null) {
            setIndexing(true);
            timer.current = setTimeout(() => request(paging, mode), 1000);
            return;
          }
          const data = envelope.data;
          setPending(false);
          setIndexing(false);
          setLoading(false);
          setMore(false);
          setError(null);
          setRefreshError(null);
          if (mode === 'append' && rowsFor.current !== key) return;
          setPage(data);
          setShownFor(query);
          hasRows.current = true;
          retryCatalog();
          setTasks((previous) => {
            const next = mode === 'append' ? [...previous, ...data.tasks] : data.tasks;
            loaded.current = next.length;
            rowsFor.current = key;
            return next;
          });
          onStatus?.(`generation ${envelope.generation}${envelope.warnings.length ? ` · ${envelope.warnings.length} warnings` : ''}`);
        })
        .catch((err: unknown) => {
          if (seq !== sequence.current) return;
          setPending(false);
          setLoading(false);
          setMore(false);
          setIndexing(false);
          if (hasRows.current) setRefreshError(err);
          else setError(err);
        });
    },
    [client, query, onStatus, retryCatalog],
  );

  // A new query replaces the list from the top.
  useEffect(() => {
    request({ offset: 0, limit: PAGE_SIZE }, 'replace');
    return () => {
      sequence.current++;
      if (timer.current) clearTimeout(timer.current);
    };
  }, [request, attempt]);

  const change = useCallback((next: TaskQuery) => {
    setQuery(next);
    pushQuery(next);
  }, []);

  const loadMore = () => {
    if (pending || rowsFor.current !== searchFromQuery(query)) return;
    request({ offset: loaded.current, limit: PAGE_SIZE }, 'append');
  };

  // After a triage the row shows the new labels at once and the list is asked
  // again with the same filters, sort and extent; the selection stays.
  const changed = useCallback(
    (detail: TaskDetail) => {
      setTasks((previous) => previous.map((task) => (task.id === detail.id ? { ...task, ...withoutDetail(detail) } : task)));
      request({ offset: 0, limit: Math.min(Math.max(loaded.current, PAGE_SIZE), MAX_LIMIT) }, 'requery');
    },
    [request],
  );

  const focusRow = (id: string | null) => {
    const rows = Array.from(list.current?.querySelectorAll<HTMLElement>('[data-id]') ?? []);
    (rows.find((row) => row.dataset.id === id) ?? rows[0])?.focus();
  };

  // Escape and the back button: the focus returns to the row; on a narrow
  // screen the pane closes too, the selection stays on a wide one.
  const close = () => {
    if (!matches(NARROW)) {
      focusRow(selected);
      return;
    }
    // The list is inert until the pane is gone: the focus returns after that render.
    returnTo.current = selected;
    setSelected(null);
  };

  // Opening a task: the pane takes the focus, so Escape reaches it and a
  // narrow screen does not leave the focus behind the overlay; the same row
  // opened again focuses the pane at once, since the selection does not change.
  const open = (task: Task) => {
    if (task.id === selected) {
      pane.current?.focus();
      return;
    }
    focusPane.current = true;
    setSelected(task.id);
  };

  useLayoutEffect(() => {
    if (selected && focusPane.current) pane.current?.focus();
    focusPane.current = false;
    if (!selected && returnTo.current !== null) {
      focusRow(returnTo.current);
      returnTo.current = null;
    }
  }, [selected]);

  const onPaneKey = (event: KeyboardEvent) => {
    if (event.key !== 'Escape') return;
    event.preventDefault();
    close();
  };

  // The arrows follow the rows as shown, across the groups.
  const onRowKey = (event: KeyboardEvent, task: Task) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      open(task);
      return;
    }
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      const rows = Array.from(list.current?.querySelectorAll<HTMLElement>('[data-id]') ?? []);
      const index = rows.findIndex((row) => row.dataset.id === task.id);
      rows[index + (event.key === 'ArrowDown' ? 1 : -1)]?.focus();
    }
  };

  // With the pane over the list, what is behind it takes no focus and no click.
  const behind = narrow && selected !== null;
  const sort = query.sort ?? 'score';
  const groups = groupTasks(tasks, grouping, page?.facets ?? null, page?.total ?? tasks.length, catalog);
  // The caption describes the results that are shown; after a failed request they are the previous ones.
  const showing = page
    ? [page.total, stateWord(shownFor), page.total === 1 ? 'task' : 'tasks'].filter(Boolean).join(' ') +
      (tasks.length < page.total ? ` · showing ${tasks.length}` : '') +
      (refreshError !== null ? ' · showing the previous results' : '')
    : '';
  const filtered = hasFilters(query);
  const retry = () => setAttempt((n) => n + 1);
  // A refresh failure is told where it can be seen: in the pane when it covers the list.
  const refreshNotice = refreshError !== null && (
    <div class="page__notices">
      <div class="notice notice--error" role="alert">
        <p>The list could not be refreshed: {refreshError instanceof Error ? refreshError.message : String(refreshError)}</p>
        <button type="button" class="button" onClick={retry}>
          Try again
        </button>
      </div>
    </div>
  );

  return (
    <section class={`page inbox${selected ? ' inbox--open' : ''}${full ? ' page--issue' : ''}`}>
      <TopBar
        title="Inbox"
        inert={behind}
        actions={
          <DisplayMenu>
            {() => (
              <>
                <DisplaySelect label="Grouping" value={grouping} options={GROUPINGS} onChange={changeGrouping} />
                <DisplaySelect label="Ordering" value={sort} options={SORTS} onChange={(value) => change({ ...query, sort: value })} />
                <p class="display-menu__note">Score weighs impact, ease, type, age and the milestone's due date. A group counts every task of the selection; Load more fills the groups.</p>
              </>
            )}
          </DisplayMenu>
        }
      >
        <TaskFilters value={query} facets={page?.facets ?? null} catalog={catalog} onChange={change} showColumn />
      </TopBar>
      {(indexing || (loading && page)) && <ProgressLine />}
      {catalogFailed && !catalog && (
        <div class="page__notices" inert={behind}>
          <div class="notice notice--error" role="alert">
            <p>The catalog could not be loaded, so milestones show their keys.</p>
            <button type="button" class="button" onClick={retryCatalog}>
              Try again
            </button>
          </div>
        </div>
      )}
      <div class="page__content inbox__panes">
        <div class="inbox__list-pane" inert={behind}>
          {!behind && refreshNotice}
          {error ? (
            <div class="page__notices">
              <ErrorState error={error} onRetry={retry} />
            </div>
          ) : indexing && !page ? (
            <Indexing shape="rows" />
          ) : loading && !page ? (
            <Loading shape="rows" count={8} />
          ) : page && tasks.length === 0 ? (
            <Empty
              action={
                filtered && (
                  <button type="button" class="button" onClick={() => change({ state: query.state, sort: query.sort })}>
                    Clear filters
                  </button>
                )
              }
            >
              {filtered ? 'No tasks match the filters.' : 'No tasks in the collection.'}
            </Empty>
          ) : (
            <>
              <p class="count-strip" role="status" aria-live="polite" aria-busy={loading}>
                {showing}
                {page && (indexing || loading) && <span class="count-strip__state">{indexing ? 'Indexing the mirrors…' : 'Loading…'}</span>}
              </p>
              <ol class="inbox__list" ref={list} role="listbox" aria-label="Tasks" aria-busy={loading}>
                {grouping === 'none'
                  ? tasks.map((task) => <Row key={task.id} task={task} selected={task.id === selected} onOpen={() => open(task)} onKey={(event) => onRowKey(event, task)} />)
                  : groups.map((group) => (
                      <li class="inbox__group" role="group" aria-labelledby={`inbox-${group.key}`} key={group.key}>
                        <div class="group-header inbox__group-header" id={`inbox-${group.key}`}>
                          <span class="group-header__glyph">{group.glyph}</span>
                          <span class="group-header__name">{group.title}</span>
                          <span class="group-header__count" aria-label={`${group.count} tasks`}>
                            {group.count}
                          </span>
                        </div>
                        <ol class="inbox__list">
                          {group.rows.map((task) => (
                            <Row key={task.id} task={task} selected={task.id === selected} onOpen={() => open(task)} onKey={(event) => onRowKey(event, task)} />
                          ))}
                        </ol>
                      </li>
                    ))}
              </ol>
              {page && tasks.length < page.total && (
                <p class="inbox__more">
                  <button type="button" class="button button--ghost" disabled={pending} onClick={loadMore}>
                    {more ? 'Loading…' : `Load more (${page.total - tasks.length} left)`}
                  </button>
                </p>
              )}
            </>
          )}
        </div>
        {selected ? (
          <aside class="inbox__pane" aria-label="Task" ref={pane} tabIndex={-1} onKeyDown={onPaneKey}>
            {behind && refreshNotice}
            <TaskDetails key={selected} client={client} id={selected} onChanged={changed} onBack={narrow ? close : undefined} onExpand={() => openTask(selected)} />
          </aside>
        ) : (
          <aside class="inbox__pane inbox__pane--empty" aria-label="Task">
            <Empty>
              <Icon name="inbox" size={32} class="empty-state__icon" />
              <span>Select a task to read it.</span>
            </Empty>
          </aside>
        )}
      </div>
      {full && <IssuePage client={client} id={full} onChanged={changed} onBack={closeTask} backLabel="Back to the inbox" />}
    </section>
  );
}

interface RowProps {
  task: Task;
  selected: boolean;
  onOpen: () => void;
  onKey: (event: KeyboardEvent) => void;
}

/**
 * A 56 px row: the repository's circle, with a flag as its event badge when
 * the issue is flagged, "CODE title", then one plain line of the repository
 * and the milestone; a 24 px column on the right with the status over the age.
 */
function Row({ task, selected, onOpen, onKey }: RowProps) {
  const flags = flagList(task);
  const reasons = flags.map((flag) => `${flag.title}: ${flag.reason}`).join(' ');
  const repo = task.repository.slice(task.repository.indexOf('/') + 1);
  return (
    <li
      class={`inbox-row${selected ? ' is-selected' : ''}${task.state === 'closed' ? ' is-closed' : ''}`}
      role="option"
      aria-selected={selected}
      tabIndex={0}
      data-id={task.id}
      onClick={onOpen}
      onKeyDown={onKey}
    >
      <RepoAvatar repository={task.repository} badge={flags.length > 0 ? <Icon name="flag" size={12} title={`Flagged: ${reasons}`} /> : undefined} />
      <span class="inbox-row__body">
        <span class="inbox-row__line">
          <span class="inbox-row__code">{task.code}</span>
          <span class="inbox-row__title" title={task.title}>
            {task.state === 'closed' && <span class="sr-only">Closed: </span>}
            {task.title}
          </span>
        </span>
        <span class="inbox-row__sub" title={task.repository}>
          {repo}
          {task.milestone ? ` · ${task.milestone.title}` : ''}
        </span>
        {/* The priority is in the reader; a screen reader hears it with the row. */}
        <span class="sr-only">{priorityTitle(impactOf(task.labels))}</span>
      </span>
      <span class="inbox-row__side">
        <StatusGlyph column={task.column} title={COLUMN_TITLES[task.column]} size={12} />
        <span class="inbox-row__age" title={`Updated ${date(task.updated_at)}`}>
          {age(task.updated_at)}
        </span>
      </span>
    </li>
  );
}

function stateWord(query: TaskQuery): string {
  return query.state === 'all' ? '' : query.state === 'closed' ? 'closed' : 'open';
}

/** The list fields of a detail: the row keeps its shape, the body stays in the pane. */
function withoutDetail(detail: TaskDetail): Task {
  const { body_html: _body, comments: _comments, provenance: _provenance, ...task } = detail;
  return task;
}
