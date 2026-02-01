import { fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { beforeEach, describe, expect, it } from 'vitest';
import { ApiClient } from '../../api/client';
import { fixtures, json, mockApi, type MockCall } from '../../api/mock';
import type { Column, Task, TaskPage } from '../../api/types';
import { COLUMNS, COLUMN_TITLES } from '../../api/types';
import { PAGE, TabloPage, sortFor } from './index';

const page = fixtures.tasks;

/** A fetch like the service: the column selector, sort, limit and offset page the given tasks; the facets stay the fixture's. */
function boardApi(tasks: Task[] = page.data.tasks, options: { generation?: () => number; route?: (call: MockCall) => Response | undefined } = {}) {
  return mockApi({
    route: (call) => {
      const custom = options.route?.(call);
      if (custom) return custom;
      if (call.method !== 'GET' || call.url.pathname !== '/api/v0/tasks') return undefined;
      const params = call.url.searchParams;
      const columns = params.getAll('column');
      const repo = params.getAll('repo');
      let selected = tasks.filter((t) => repo.length === 0 || repo.includes(t.repository));
      const total = selected.length;
      selected = selected.filter((t) => columns.length === 0 || columns.includes(t.column));
      const limit = Number(params.get('limit') ?? 100);
      const offset = Number(params.get('offset') ?? 0);
      const data: TaskPage = { ...page.data, total: columns.length ? selected.length : total, limit, offset, tasks: selected.slice(offset, offset + limit) };
      return json({ ...page, generation: options.generation?.() ?? page.generation, data });
    },
  });
}

function manyBacklog(n: number): Task[] {
  const template = page.data.tasks[0];
  return Array.from({ length: n }, (_, i) => ({ ...template, key: `github:octomation/maintainer#${1000 + i}`, id: `id-${i}`, code: `MAIN-${1000 + i}`, title: `generated ${i}`, column: 'backlog' as Column }));
}

function column(name: Column) {
  return screen.getByRole('region', { name: COLUMN_TITLES[name] });
}

function cards(name: Column) {
  return within(column(name)).queryAllByRole('listitem');
}

/** Filters the board by one repository through the filter menu, then closes the menu. */
function chooseRepository(value: string) {
  fireEvent.click(screen.getByRole('button', { name: /^Filter/ }));
  fireEvent.click(within(screen.getByRole('group', { name: 'Filters' })).getByRole('button', { name: /^Repository/ }));
  const [owner, name] = value.split('/');
  fireEvent.click(within(screen.getByRole('group', { name: 'Repository' })).getByRole('checkbox', { name: `${name} · ${owner}` }));
  fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
}

/** Picks a priority in the issue page of the peek panel. */
function pickPriority(dialog: HTMLElement, value: string) {
  fireEvent.click(within(dialog).getByRole('button', { name: /^Priority:/ }));
  fireEvent.click(within(screen.getByRole('menu', { name: 'Priority' })).getByRole('menuitemradio', { name: value }));
}

function listCalls(api: ReturnType<typeof mockApi>) {
  return api.calls.filter((c) => c.method === 'GET' && c.url.pathname === '/api/v0/tasks');
}

describe('TabloPage', () => {
  beforeEach(() => {
    history.replaceState(null, '', '/tablo');
  });

  it('lays the tasks out in five columns with their counts', async () => {
    const api = boardApi();
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    const regions = screen.getAllByRole('region').map((r) => r.getAttribute('aria-label'));
    expect(regions).toEqual(COLUMNS.map((c) => COLUMN_TITLES[c]));
    for (const c of COLUMNS) {
      const expected = page.data.facets.columns.find((f) => f.value === c)?.count ?? 0;
      expect(within(column(c)).getByLabelText(`${expected} tasks`).textContent).toBe(String(expected));
      await waitFor(() => expect(cards(c).length).toBe(page.data.tasks.filter((t) => t.column === c).length));
    }
    const first = cards('backlog')[0];
    expect(first.textContent).toContain('MAIN-1');
    expect(first.textContent).toContain('fetch: adopt existing checkouts');
    expect(within(first).getByTitle('octomation/maintainer')).toBeTruthy();
    // impact and effort are the priority and estimate glyphs; up to two other labels are chips.
    expect(within(first).getByLabelText('Priority: High')).toBeTruthy();
    expect(within(first).getByLabelText('Estimate: S (easy)')).toBeTruthy();
    expect(first.textContent).not.toContain('impact: high');
    expect(first.textContent).toContain('type: feature');
    // One label chip fits beside the code; the rest are counted.
    expect(within(first).getByLabelText('1 more labels: scope: code').textContent).toBe('+1');
    // One footer: the code is there, the milestone is on the issue page.
    expect(first.querySelector('.card__footer')?.textContent).toContain('MAIN-1');
    expect(first.textContent).not.toContain('Workspace, v0.2');

    const calls = listCalls(api);
    expect(calls[0].url.searchParams.get('state')).toBe('all');
    expect(calls[0].url.searchParams.get('column')).toBeNull();
    expect(calls[0].url.searchParams.get('limit')).toBe('1');
    for (const c of COLUMNS) {
      const call = calls.find((call) => call.url.searchParams.get('column') === c);
      expect(call, c).toBeTruthy();
      expect(call?.url.searchParams.get('limit')).toBe(String(PAGE));
      expect(call?.url.searchParams.get('offset')).toBe('0');
      expect(call?.url.searchParams.get('sort')).toBe(sortFor(c));
      expect(call?.url.searchParams.get('state')).toBe('all');
    }
    expect(sortFor('done')).toBe('updated');
    expect(sortFor('wont_do')).toBe('updated');
    expect(sortFor('todo')).toBe('score');
  });

  it('marks a flagged card with the reason in its tooltip', async () => {
    const api = boardApi();
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    const ambiguous = cards('backlog').find((c) => c.textContent?.includes('MAIN-6'));
    const marker = within(ambiguous!).getByTitle(/boards disagree/);
    expect(marker.textContent).toContain('ambiguous');
    const contradiction = cards('backlog').find((c) => c.textContent?.includes('MAIN-5'));
    expect(within(contradiction!).getByTitle(/listed under Done/).textContent).toContain('contradiction');
    const plain = cards('backlog').find((c) => c.textContent?.includes('MAIN-1'));
    expect(within(plain!).queryByLabelText(/Flagged/)).toBeNull();
    expect(within(ambiguous!).getByLabelText(/Flagged/).querySelector('svg')).toBeTruthy();
  });

  it('pages each column by 50 and loads more on request', async () => {
    const api = boardApi(manyBacklog(120));
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    const more = within(column('backlog')).getByRole('button', { name: 'Load more in Backlog' });
    expect(more.textContent).toContain('70 left');
    fireEvent.click(more);
    await waitFor(() => expect(cards('backlog').length).toBe(100));
    expect(cards('backlog')[50].textContent).toContain('MAIN-1050');
    const second = listCalls(api).filter((c) => c.url.searchParams.get('column') === 'backlog')[1];
    expect(second.url.searchParams.get('offset')).toBe('50');
    expect(second.url.searchParams.get('limit')).toBe('50');
    expect(second.url.searchParams.get('sort')).toBe('score');
    fireEvent.click(within(column('backlog')).getByRole('button', { name: 'Load more in Backlog' }));
    await waitFor(() => expect(cards('backlog').length).toBe(120));
    expect(within(column('backlog')).queryByRole('button', { name: /Load more/ })).toBeNull();
    expect(within(column('todo')).getByText('No tasks')).toBeTruthy();
  });

  it('resets the pages on a filter change and carries the filters in the URL', async () => {
    const api = boardApi(manyBacklog(120));
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    fireEvent.click(within(column('backlog')).getByRole('button', { name: /Load more/ }));
    await waitFor(() => expect(cards('backlog').length).toBe(100));
    chooseRepository('octomation/maintainer');
    await waitFor(() => expect(location.search).toBe('?repo=octomation%2Fmaintainer&state=all'));
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    const after = listCalls(api).filter((c) => c.url.searchParams.get('repo') === 'octomation/maintainer');
    expect(after.length).toBe(1 + COLUMNS.length);
    expect(after.filter((c) => c.url.searchParams.get('column') === 'backlog').map((c) => c.url.searchParams.get('offset'))).toEqual(['0']);
    expect(within(column('backlog')).getByRole('button', { name: /Load more/ }).textContent).toContain('70 left');
  });

  it('restores the filters and the ordering from the URL and never requests a column from it', async () => {
    history.replaceState(null, '', '/tablo?repo=octomation%2Fmaintainer&state=open&column=todo&sort=created');
    const api = boardApi();
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    const chips = screen.getByRole('group', { name: 'Active filters' }).textContent;
    expect(chips).toContain('Repositoryismaintainer · octomation');
    expect(chips).toContain('StateisOpen');
    const summary = listCalls(api)[0].url.searchParams;
    expect(summary.get('state')).toBe('open');
    expect(summary.getAll('column')).toEqual([]);
    // The ordering of the URL applies to every column, the closed ones included.
    expect(listCalls(api).find((c) => c.url.searchParams.get('column') === 'done')?.url.searchParams.get('sort')).toBe('created');
    expect(listCalls(api).find((c) => c.url.searchParams.get('column') === 'backlog')?.url.searchParams.get('sort')).toBe('created');
  });

  it('reloads the board when the index generation changes', async () => {
    let generation = 1;
    const api = boardApi(manyBacklog(120), { generation: () => generation });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    const before = listCalls(api).length;
    generation = 2;
    fireEvent.click(within(column('backlog')).getByRole('button', { name: /Load more/ }));
    await waitFor(() => expect(listCalls(api).length).toBeGreaterThanOrEqual(before + 1 + 1 + COLUMNS.length));
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    const last = listCalls(api).slice(-COLUMNS.length);
    expect(last.map((c) => c.url.searchParams.get('offset'))).toEqual(COLUMNS.map(() => '0'));
  });

  it('opens a card as the full issue page and requests its column again after a triage', async () => {
    const api = boardApi(manyBacklog(60));
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    const opener = within(cards('backlog')[1]).getByRole('button');
    fireEvent.click(opener);
    const dialog = await screen.findByRole('region', { name: 'Issue' });
    // The page is addressed in the URL beside the filters; the board stays mounted under it.
    expect(new URLSearchParams(location.search).get('task')).toBe('id-1');
    expect(document.querySelector('.page--issue')).toBeTruthy();
    await within(dialog).findByRole('heading', { level: 2 });
    expect(within(dialog).getByRole('link', { name: 'Open on GitHub' })).toBeTruthy();
    expect(within(dialog).getAllByText(fixtures.task.data.code)).toHaveLength(2);
    expect(within(dialog).getByRole('button', { name: /^Priority:/ })).toBeTruthy();
    const before = listCalls(api).length;
    pickPriority(dialog, 'Low');
    await waitFor(() => expect(api.calls.filter((c) => c.method === 'PATCH').length).toBe(1));
    await waitFor(() => expect(listCalls(api).length).toBe(before + 2));
    const [summary, reload] = listCalls(api).slice(before);
    expect(summary.url.searchParams.get('column')).toBeNull();
    expect(reload.url.searchParams.get('column')).toBe(fixtures.task.data.column);
    expect(reload.url.searchParams.get('offset')).toBe('0');
    expect(reload.url.searchParams.get('limit')).toBe(String(PAGE));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Back to the board' }));
    await waitFor(() => expect(screen.queryByRole('region', { name: 'Issue' })).toBeNull());
    expect(new URLSearchParams(location.search).get('task')).toBeNull();
    await waitFor(() => expect(document.activeElement).toBe(within(cards('backlog')[1]).getByRole('button')));
  });

  it('opens the full issue page from the URL', async () => {
    history.replaceState(null, '', '/tablo?task=id-2');
    const api = boardApi(manyBacklog(5));
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    const issue = await screen.findByRole('region', { name: 'Issue' });
    await within(issue).findByRole('heading', { level: 2 });
    fireEvent.keyDown(issue, { key: 'Escape' });
    await waitFor(() => expect(screen.queryByRole('region', { name: 'Issue' })).toBeNull());
    expect(location.search).toBe('');
  });

  it('repeats a failed replacement after a triage from the top, and tells a failed count refresh', async () => {
    let failColumn = false;
    let failFacets = false;
    const api = boardApi(manyBacklog(60), {
      route: (call) => {
        if (call.method !== 'GET' || call.url.pathname !== '/api/v0/tasks') return undefined;
        const column = call.url.searchParams.get('column');
        if (failColumn && column === 'backlog') return json({ error: { code: 'boom', message: 'backlog is broken' } }, 500);
        if (failFacets && column === null) return json({ error: { code: 'boom', message: 'counts are broken' } }, 500);
        return undefined;
      },
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    fireEvent.click(within(cards('backlog')[1]).getByRole('button'));
    const dialog = await screen.findByRole('region', { name: 'Issue' });
    await within(dialog).findByRole('heading', { level: 2 });
    failColumn = true;
    failFacets = true;
    pickPriority(dialog, 'Low');
    await within(column('backlog')).findByRole('alert');
    expect(within(column('backlog')).getByText('backlog is broken')).toBeTruthy();
    expect(cards('backlog').length).toBe(PAGE);
    const notice = await screen.findByText(/counts could not be refreshed/);
    expect(notice.textContent).toContain('counts are broken');
    failColumn = false;
    failFacets = false;
    const before = listCalls(api).length;
    fireEvent.click(within(column('backlog')).getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(listCalls(api).length).toBe(before + 1));
    // The same request again: a replacement from offset 0, not an append from the current length.
    const retry = listCalls(api)[before];
    expect(retry.url.searchParams.get('column')).toBe('backlog');
    expect(retry.url.searchParams.get('offset')).toBe('0');
    expect(retry.url.searchParams.get('limit')).toBe(String(PAGE));
    await waitFor(() => expect(within(column('backlog')).queryByRole('alert')).toBeNull());
    expect(cards('backlog').length).toBe(PAGE);
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(screen.queryByText(/counts could not be refreshed/)).toBeNull());
  });

  it('returns the focus to the content when a triage emptied the board and the issue page closes', async () => {
    history.replaceState(null, '', '/tablo?repo=octomation%2Fmaintainer&state=all');
    const only = page.data.tasks.slice(0, 1);
    let gone = false;
    const api = boardApi(only, {
      route: (call) => {
        if (call.method === 'GET' && call.url.pathname === '/api/v0/tasks' && gone) {
          const limit = Number(call.url.searchParams.get('limit') ?? 100);
          return json({ ...page, data: { ...page.data, total: 0, limit, offset: 0, tasks: [], facets: { ...page.data.facets, columns: [] } } });
        }
        return undefined;
      },
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(1));
    const opener = within(cards('backlog')[0]).getByRole('button');
    opener.focus();
    fireEvent.click(opener);
    const dialog = await screen.findByRole('region', { name: 'Issue' });
    await within(dialog).findByRole('heading', { level: 2 });
    gone = true;
    pickPriority(dialog, 'Low');
    await waitFor(() => expect(document.querySelector('.board, .empty-state')?.textContent).toContain('No tasks match the filters.'));
    expect(opener.isConnected).toBe(false);
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
    await waitFor(() => expect(screen.queryByRole('region', { name: 'Issue' })).toBeNull());
    expect(document.activeElement).not.toBe(document.body);
    expect((document.activeElement as HTMLElement).contains(screen.getByText('No tasks match the filters.'))).toBe(true);
  });

  it('keeps the board on screen while the index is rebuilt and when a refresh fails', async () => {
    let mode: 'ok' | 'indexing' | 'broken' = 'ok';
    const api = boardApi(page.data.tasks, {
      route: (call) => {
        if (call.url.pathname !== '/api/v0/tasks' || call.url.searchParams.get('column') !== null) return undefined;
        if (mode === 'indexing') return json({ ...page, refreshing: true, data: null }, 202);
        if (mode === 'broken') return json({ error: { code: 'boom', message: 'summary is broken' } }, 500);
        return undefined;
      },
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    mode = 'indexing';
    chooseRepository('octomation/maintainer');
    await screen.findByText('Indexing the mirrors…');
    expect(cards('backlog').length).toBe(4);
    mode = 'broken';
    await screen.findByText(/board could not be refreshed/, undefined, { timeout: 3000 });
    expect(cards('backlog').length).toBe(4);
    mode = 'ok';
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(screen.queryByText(/board could not be refreshed/)).toBeNull());
    await waitFor(() => expect(cards('backlog').length).toBe(4));
  });

  it('keeps the previous cards stale after a failed replacement and never appends the new query to them', async () => {
    let failColumn = false;
    const api = mockApi({
      route: (call) => {
        if (call.method !== 'GET' || call.url.pathname !== '/api/v0/tasks') return undefined;
        const params = call.url.searchParams;
        const column = params.get('column');
        const filtered = params.getAll('repo').length > 0;
        if (column === 'backlog' && filtered && failColumn) return json({ error: { code: 'boom', message: 'backlog is broken' } }, 500);
        const all = filtered ? manyBacklog(70).map((t) => ({ ...t, code: t.code.replace('MAIN', 'NEW'), key: `${t.key}-new`, id: `${t.id}-new` })) : manyBacklog(120);
        const selected = column ? all.filter((t) => t.column === column) : all;
        const limit = Number(params.get('limit') ?? 100);
        const offset = Number(params.get('offset') ?? 0);
        return json({ ...page, data: { ...page.data, total: selected.length, limit, offset, tasks: selected.slice(offset, offset + limit) } });
      },
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(PAGE));
    const more = () => within(column('backlog')).getByRole('button', { name: 'Load more in Backlog' }) as HTMLButtonElement;
    expect(more().disabled).toBe(false);
    failColumn = true;
    chooseRepository('octomation/maintainer');
    await within(column('backlog')).findByRole('alert');
    // The old cards stay, marked as the previous results; Load more is gone until the replacement succeeds.
    expect(cards('backlog').length).toBe(PAGE);
    expect(cards('backlog')[0].textContent).toContain('MAIN-1000');
    expect(within(column('backlog')).getByText('Previous results')).toBeTruthy();
    expect(within(column('backlog')).queryByRole('button', { name: 'Load more in Backlog' })).toBeNull();
    failColumn = false;
    fireEvent.click(within(column('backlog')).getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(cards('backlog')[0].textContent).toContain('NEW-1000'));
    expect(cards('backlog').length).toBe(PAGE);
    expect(within(column('backlog')).queryByText('Previous results')).toBeNull();
    expect(more().disabled).toBe(false);
    expect(more().textContent).toContain('20 left');
    fireEvent.click(more());
    await waitFor(() => expect(cards('backlog').length).toBe(70));
    expect(cards('backlog')[50].textContent).toContain('NEW-1050');
  });

  it('shows the error after a 202 was followed by a failure, and is busy from the start of the summary request', async () => {
    let mode: 'ok' | 'indexing' | 'broken' = 'ok';
    const api = boardApi(page.data.tasks, {
      route: (call) => {
        if (call.url.pathname !== '/api/v0/tasks' || call.url.searchParams.get('column') !== null) return undefined;
        if (mode === 'indexing') return json({ ...page, refreshing: true, data: null }, 202);
        if (mode === 'broken') return json({ error: { code: 'boom', message: 'summary is broken' } }, 500);
        return undefined;
      },
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    mode = 'indexing';
    chooseRepository('octomation/maintainer');
    // Busy at once, before any answer: the progress line is up.
    expect(document.querySelector('.progress-line')).toBeTruthy();
    await screen.findByText('Indexing the mirrors…');
    mode = 'broken';
    await screen.findByText(/board could not be refreshed/, undefined, { timeout: 3000 });
    expect(screen.queryByText('Indexing the mirrors…')).toBeNull();
    expect(document.querySelector('.progress-line')).toBeNull();
    expect(cards('backlog').length).toBe(4);
  });

  it('offers Clear filters when a filtered board is empty', async () => {
    history.replaceState(null, '', '/tablo?repo=nobody%2Fnothing&state=all');
    const api = boardApi();
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await screen.findByText('No tasks match the filters.');
    fireEvent.click(screen.getByRole('button', { name: 'Clear filters' }));
    await waitFor(() => expect(location.search).toBe('?state=all'));
    await waitFor(() => expect(cards('backlog').length).toBe(4));
  });

  it('shows the indexing state and then the board', async () => {
    let ready = false;
    const api = boardApi(page.data.tasks, {
      route: (call) => (call.url.pathname === '/api/v0/tasks' && !ready ? json({ ...page, refreshing: true, data: null }, 202) : undefined),
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await screen.findByText(/Indexing/);
    ready = true;
    await waitFor(() => expect(cards('backlog').length).toBe(4), { timeout: 3000 });
  });

  it('shows a column error with a retry and keeps the other columns', async () => {
    let fail = true;
    const api = boardApi(page.data.tasks, {
      route: (call) => (call.url.searchParams.get('column') === 'done' && fail ? json({ error: { code: 'boom', message: 'done is broken' } }, 500) : undefined),
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    await within(column('done')).findByRole('alert');
    expect(within(column('done')).getByText('done is broken')).toBeTruthy();
    fail = false;
    fireEvent.click(within(column('done')).getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(cards('done').length).toBe(1));
  });

  it('orders the columns, hides the empty ones and the card properties from Display', async () => {
    localStorage.removeItem('maintainer.tablo.display');
    // To do is empty: its count is gone from the facets and no task is in it.
    const tasks = page.data.tasks.filter((t) => t.column !== 'todo');
    const facets = { ...page.data.facets, columns: page.data.facets.columns.filter((f) => f.value !== 'todo') };
    const api = boardApi(tasks, {
      route: (call) =>
        call.method === 'GET' && call.url.pathname === '/api/v0/tasks' && call.url.searchParams.get('column') === null
          ? json({ ...page, data: { ...page.data, total: tasks.length, limit: 1, offset: 0, tasks: tasks.slice(0, 1), facets } })
          : undefined,
    });
    render(<TabloPage client={new ApiClient('', api.fetch)} />);
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    fireEvent.click(screen.getByRole('button', { name: 'View' }));
    const panel = screen.getByRole('group', { name: 'Display options' });
    const ordering = within(panel).getByLabelText('Ordering') as HTMLSelectElement;
    expect(ordering.value).toBe('default');
    const before = listCalls(api).length;
    fireEvent.change(ordering, { target: { value: 'updated' } });
    await waitFor(() => expect(location.search).toBe('?state=all&sort=updated'));
    await waitFor(() => expect(listCalls(api).slice(before).filter((c) => c.url.searchParams.get('column') === 'backlog').map((c) => c.url.searchParams.get('sort'))).toEqual(['updated']));
    await waitFor(() => expect(cards('backlog').length).toBe(4));
    expect(screen.getAllByRole('region').length).toBe(COLUMNS.length);
    fireEvent.click(within(screen.getByRole('group', { name: 'Display options' })).getByRole('switch', { name: 'Show empty columns' }));
    await waitFor(() => expect(screen.getAllByRole('region').map((r) => r.getAttribute('aria-label'))).toEqual(COLUMNS.filter((c) => c !== 'todo').map((c) => COLUMN_TITLES[c])));
    // Card properties: without the repository the card does not name it.
    fireEvent.click(within(screen.getByRole('group', { name: 'Display options' })).getByRole('button', { name: 'Repository' }));
    expect(cards('backlog')[0].querySelector('.card__avatar')).toBeNull();
    expect(JSON.parse(localStorage.getItem('maintainer.tablo.display') ?? '{}').properties.repository).toBe(false);
    localStorage.removeItem('maintainer.tablo.display');
  });
});
