import { fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiClient } from '../../api/client';
import { fixtures, json, mockApi, type MockApi } from '../../api/mock';
import type { CheckoutsData, Envelope } from '../../api/types';
import { StatusPage } from './index';
import { CHECKOUT_COLUMNS } from '../../api/types';
import { columnTemplate, cycleSort, formatTime, groupOf, groupRows, sortParam, sortSummary, summaryLine } from './model';

const checkouts = fixtures.checkouts as Envelope<CheckoutsData>;

/** Preact runs effects after the next frame: with fake timers, this much flushes them. */
const EFFECT_MS = 150;

function checkoutCalls(api: MockApi) {
  return api.calls.filter((call) => call.method === 'GET' && call.url.pathname === '/api/v0/checkouts');
}

function lastSort(api: MockApi): string | null {
  const calls = checkoutCalls(api);
  return calls[calls.length - 1]?.url.searchParams.get('sort') ?? null;
}

function bodyRows(): HTMLElement[] {
  return within(screen.getByRole('listbox', { name: 'Checkouts' })).getAllByRole('option');
}

/** The server cells of a row, in the order of the columns. */
function cellsOf(row: Element): string[] {
  return CHECKOUT_COLUMNS.map((column) => row.querySelector(`[data-cell="${column}"]`)?.textContent ?? '');
}

/** The Display panel, opened when it is closed. */
function display(): HTMLElement {
  if (!screen.queryByRole('group', { name: 'Display options' })) fireEvent.click(screen.getByRole('button', { name: 'View' }));
  return screen.getByRole('group', { name: 'Display options' });
}

/** The search field of the Filter menu, opened when it is closed. */
function searchBox(): HTMLInputElement {
  if (!screen.queryByRole('group', { name: 'Search the checkouts' })) fireEvent.click(screen.getByRole('button', { name: /Filter/ }));
  return screen.getByRole('searchbox', { name: 'Search' }) as HTMLInputElement;
}

/** An ordering key of the Display panel. */
function orderingKey(title: string): HTMLButtonElement {
  return within(within(display()).getByRole('group', { name: 'Ordering' })).getByRole('button', { name: new RegExp(`^${title}`) }) as HTMLButtonElement;
}

async function renderPage(api: MockApi = mockApi()) {
  const onStatus = vi.fn<(status: string) => void>();
  render(<StatusPage client={new ApiClient('', api.fetch)} onStatus={onStatus} />);
  await screen.findByRole('listbox', { name: 'Checkouts' });
  return { api, onStatus };
}

afterEach(() => {
  vi.useRealTimers();
});

describe('StatusPage', () => {
  it('renders the cells, the order and the summary the server gives for the fixture', async () => {
    const { api, onStatus } = await renderPage();
    // Grouped by state: the groups in their order, each keeping the server's order of its rows.
    const groups = within(screen.getByRole('listbox', { name: 'Checkouts' })).getAllByRole('group');
    const expected = groupRows(checkouts.data.rows, 'state');
    expect(groups.map((g) => g.querySelector('.group-header')?.textContent)).toEqual(expected.map((g) => `${g.title}${g.rows.length}`));
    expect(bodyRows().map((row) => cellsOf(row)[0])).toEqual(expected.flatMap((g) => g.rows.map((r) => r.repository)));
    // Without grouping the rows are the server's, cell for cell; the lock is an icon named after its meaning.
    fireEvent.change(within(display()).getByLabelText('Grouping'), { target: { value: 'none' } });
    const withoutLock = (cells: string[]) => cells.filter((_, i) => i !== 4);
    expect(bodyRows().map((row) => withoutLock(cellsOf(row)))).toEqual(checkouts.data.rows.map((row) => withoutLock(row.cells)));
    expect(bodyRows().map((row) => row.querySelector('[data-cell="lock"] [aria-label="Push locked"]') !== null)).toEqual(checkouts.data.rows.map((row) => row.cells[4] === '🔒'));
    // The TUI header over the shown rows, as `maintainer status --format=plain` prints it.
    expect(within(display()).getByTestId('summary').textContent).toBe('4 checkouts · 1 uncommitted · 2 ahead · 1 behind · 1 locked · 1 orphan · 1 error · local refs (no fetch)');
    expect(within(display()).getByTestId('summary').textContent).toBe(summaryLine(checkouts.data.summary));
    // The ordering shows the chain the server echoed: the fixture was answered for sort=-status.
    const status = orderingKey('Status');
    expect(status.getAttribute('aria-pressed')).toBe('true');
    expect(status.getAttribute('aria-label')).toBe('Status, sorted descending, priority 1');
    expect(status.querySelector('svg')).toBeTruthy();
    expect(status.getAttribute('title')).toContain('Shift');
    expect(orderingKey('Repository').getAttribute('aria-pressed')).toBe('false');
    expect(within(display()).getByTestId('sort-summary').textContent).toBe('Sort: Status ↓1');
    // The first request carries neither a sort nor a query.
    expect(checkoutCalls(api)[0].url.search).toBe('');
    await waitFor(() => expect(onStatus).toHaveBeenLastCalledWith('collected 2026-10-20 12:00:00Z'));
    expect(within(display()).getByText('snapshot 2026-10-20 12:00:00Z')).toBeTruthy();
  });

  it('marks the error and orphan rows as the TUI does', async () => {
    await renderPage();
    const rows = bodyRows();
    const broken = rows.find((row) => row.textContent?.includes('broken'));
    expect(broken?.className).toContain('is-error');
    const cells = cellsOf(broken!);
    expect(cells[1]).toBe('—');
    expect(cells[2]).toBe('?');
    expect(cells[3]).toBe('error: git status: not a git repository');
    expect(within(broken!).getByLabelText('Errors')).toBeTruthy();
    const orphan = rows.find((row) => row.textContent?.includes('dotfiles'));
    expect(orphan?.className).toContain('is-orphan');
    expect(groupOf(checkouts.data.rows.find((r) => r.repository.includes('dotfiles'))!)).toBe('orphan');
    const maintainer = rows.find((row) => row.textContent?.includes('octomation/maintainer'))!;
    expect(cellsOf(maintainer)[1]).toBe('main*');
    expect(rows.find((row) => row.textContent?.includes('octomation/tool'))?.querySelector('[data-cell="lock"] svg')?.getAttribute('aria-label')).toBe('Push locked');
    // The lock cell of an unlocked row stays empty; the status words keep the server text.
    expect(cellsOf(maintainer)[4]).toBe('');
  });

  it('sends the sort chain: a click cycles one column, a shift-click appends', async () => {
    const { api } = await renderPage();
    fireEvent.click(orderingKey('Repository'));
    await waitFor(() => expect(lastSort(api)).toBe('repository'));
    fireEvent.click(orderingKey('Repository'));
    await waitFor(() => expect(lastSort(api)).toBe('-repository'));
    fireEvent.click(orderingKey('Path'), { shiftKey: true });
    await waitFor(() => expect(lastSort(api)).toBe('-repository,path'));
    fireEvent.click(orderingKey('Path'), { shiftKey: true });
    await waitFor(() => expect(lastSort(api)).toBe('-repository,-path'));
    fireEvent.click(orderingKey('Path'), { shiftKey: true });
    await waitFor(() => expect(lastSort(api)).toBe('-repository'));
    fireEvent.click(orderingKey('Repository'));
    await waitFor(() => expect(lastSort(api)).toBeNull());
    // Only reads: no sort click posts anything.
    expect(api.calls.every((call) => call.method === 'GET')).toBe(true);
  });

  it('groups by owner and resets the ordering to the default', async () => {
    const { api } = await renderPage();
    fireEvent.change(within(display()).getByLabelText('Grouping'), { target: { value: 'owner' } });
    const headers = Array.from(document.querySelectorAll('.group-header__name')).map((h) => h.textContent);
    expect(headers).toEqual(groupRows(checkouts.data.rows, 'owner').map((g) => g.title));
    fireEvent.click(orderingKey('Repository'));
    await waitFor(() => expect(lastSort(api)).toBe('repository'));
    fireEvent.click(within(display()).getByRole('button', { name: 'Reset to default' }));
    await waitFor(() => expect(lastSort(api)).toBeNull());
  });

  it('sends the query after a pause and never collects for it', async () => {
    vi.useFakeTimers();
    const api = mockApi();
    render(<StatusPage client={new ApiClient('', api.fetch)} />);
    await vi.advanceTimersByTimeAsync(0);
    expect(checkoutCalls(api)).toHaveLength(1);
    fireEvent.input(searchBox(), { target: { value: 'tool' } });
    await vi.advanceTimersByTimeAsync(100);
    expect(checkoutCalls(api)).toHaveLength(1);
    await vi.advanceTimersByTimeAsync(30);
    expect(checkoutCalls(api)).toHaveLength(2);
    expect(checkoutCalls(api)[1].url.searchParams.get('q')).toBe('tool');
    expect(api.calls.every((call) => call.method === 'GET')).toBe(true);
  });

  it('shows the scope and the relevance order of a filtered answer', async () => {
    const filtered: Envelope<CheckoutsData> = {
      ...checkouts,
      data: { ...checkouts.data, rows: checkouts.data.rows.slice(0, 1), sort: [], query: 'tool', summary: { ...checkouts.data.summary, checkouts: 1 } },
    };
    const api = mockApi({ route: (call) => (call.url.pathname === '/api/v0/checkouts' ? json(filtered) : undefined) });
    await renderPage(api);
    expect(within(display()).getByTestId('summary').textContent).toContain('shown 1 / 4 · 1 checkout ·');
    expect(within(display()).getByTestId('sort-summary').textContent).toBe('Sort: relevance');
    expect(bodyRows()).toHaveLength(1);
  });

  it('opens the drawer of a row and closes it with Escape', async () => {
    await renderPage();
    const tool = bodyRows().find((row) => row.textContent?.includes('octomation/tool')) as HTMLElement;
    fireEvent.click(tool);
    const dialog = await screen.findByRole('dialog', { name: 'octomation/tool' });
    const text = dialog.textContent ?? '';
    expect(text).toContain('/home/me/Development/private/octomation/tool');
    expect(text).toContain('origin/feature');
    expect(text).toContain('abc1234000000000000000000000000000000000');
    expect(text).toContain('diverged · ahead 2 · behind 1');
    expect(text).toContain('locked');
    expect(text).not.toContain('🔒');
    expect(text).toContain('Pinnedyes');
    expect(tool.getAttribute('aria-selected')).toBe('true');
    fireEvent.keyDown(document, { key: 'Escape' });
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(tool.getAttribute('aria-selected')).toBe('false');
  });

  it('opens the drawer with Enter on a focused row and shows orphan and error details', async () => {
    await renderPage();
    const dotfiles = bodyRows().find((row) => row.textContent?.includes('dotfiles')) as HTMLElement;
    expect(dotfiles.getAttribute('tabindex')).toBe('0');
    dotfiles.focus();
    fireEvent.keyDown(dotfiles, { key: 'Enter' });
    let dialog = await screen.findByRole('dialog', { name: 'kamilsk/dotfiles' });
    expect(dialog.textContent).toContain('duplicate-pin');
    expect(dialog.textContent).toContain('/home/me/Development/public/kamilsk/dotfiles');
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    const broken = bodyRows().find((row) => row.textContent?.includes('broken')) as HTMLElement;
    fireEvent.keyDown(broken, { key: 'Enter' });
    dialog = await screen.findByRole('dialog', { name: 'broken' });
    expect(dialog.textContent).toContain('error: git status: not a git repository');
    expect(dialog.textContent).toContain('Errorgit status: not a git repository');
  });

  it('reads until the first snapshot exists', async () => {
    vi.useFakeTimers();
    let answers = 0;
    const api = mockApi({
      route: (call) => {
        if (call.url.pathname !== '/api/v0/checkouts') return undefined;
        answers++;
        return answers < 3 ? json(fixtures.checkoutsPending, 202) : undefined;
      },
    });
    const onStatus = vi.fn<(status: string) => void>();
    render(<StatusPage client={new ApiClient('', api.fetch)} onStatus={onStatus} />);
    await vi.advanceTimersByTimeAsync(0);
    expect(screen.getByText('Reading local repository status…')).toBeTruthy();
    expect(screen.queryByRole('listbox')).toBeNull();
    expect(within(display()).getByRole('button', { name: /Collecting/ }).hasAttribute('disabled')).toBe(true);
    expect(within(display()).getByText('no snapshot yet')).toBeTruthy();
    await vi.advanceTimersByTimeAsync(1000);
    expect(screen.queryByRole('listbox')).toBeNull();
    await vi.advanceTimersByTimeAsync(1000);
    expect(screen.getByRole('listbox', { name: 'Checkouts' })).toBeTruthy();
    expect(checkoutCalls(api)).toHaveLength(3);
    expect(within(display()).getByRole('button', { name: 'Refresh' }).hasAttribute('disabled')).toBe(false);
    await vi.advanceTimersByTimeAsync(EFFECT_MS);
    expect(onStatus).toHaveBeenLastCalledWith('collected 2026-10-20 12:00:00Z');
  });

  it('refreshes without blocking the table and polls until the collection ends', async () => {
    vi.useFakeTimers();
    let refreshing = false;
    let reads = 0;
    const later: Envelope<CheckoutsData> = { ...checkouts, generation: 1, collected_at: '2026-10-20T12:05:00Z' };
    const api = mockApi({
      route: (call) => {
        if (call.method === 'POST' && call.url.pathname === '/api/v0/checkouts/refresh') {
          refreshing = true;
          return json({ ...checkouts, refreshing: true, data: null }, 202);
        }
        if (call.method === 'GET' && call.url.pathname === '/api/v0/checkouts') {
          if (!refreshing) return undefined;
          reads++;
          if (reads < 2) return json({ ...checkouts, refreshing: true });
          refreshing = false;
          return json(later);
        }
        return undefined;
      },
    });
    const onStatus = vi.fn<(status: string) => void>();
    render(<StatusPage client={new ApiClient('', api.fetch)} onStatus={onStatus} />);
    await vi.advanceTimersByTimeAsync(0);
    fireEvent.click(within(display()).getByRole('button', { name: 'Refresh' }));
    await vi.advanceTimersByTimeAsync(EFFECT_MS);
    const post = api.calls.find((call) => call.method === 'POST');
    expect(post?.url.pathname).toBe('/api/v0/checkouts/refresh');
    expect(post?.headers.get('X-Requested-With')).toBe('maintainer');
    // The table stays while the service collects.
    expect(screen.getByRole('listbox', { name: 'Checkouts' })).toBeTruthy();
    expect(within(display()).getByRole('button', { name: /Collecting/ }).hasAttribute('disabled')).toBe(true);
    expect(onStatus).toHaveBeenLastCalledWith('collected 2026-10-20 12:00:00Z · collecting…');
    expect(reads).toBe(1);
    await vi.advanceTimersByTimeAsync(1000 + EFFECT_MS);
    expect(within(display()).getByRole('button', { name: 'Refresh' })).toBeTruthy();
    expect(within(display()).getByText('snapshot 2026-10-20 12:05:00Z')).toBeTruthy();
    expect(onStatus).toHaveBeenLastCalledWith('collected 2026-10-20 12:05:00Z');
    expect(reads).toBe(2);
    // No further polling once the collection is over.
    const before = api.calls.length;
    await vi.advanceTimersByTimeAsync(3000);
    expect(api.calls.length).toBe(before);
  });

  it('lets Refresh fail without leaving the page collecting, and tries the collection again', async () => {
    let fail = true;
    const api = mockApi({
      route: (call) => (call.method === 'POST' && fail ? json({ error: { code: 'busy', message: 'a collection is already running' } }, 503) : undefined),
    });
    await renderPage(api);
    fireEvent.click(within(display()).getByRole('button', { name: 'Refresh' }));
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('a collection is already running');
    // The snapshot stays and Refresh is usable again: nothing is collecting.
    expect(screen.getByRole('listbox', { name: 'Checkouts' })).toBeTruthy();
    expect(within(display()).getByRole('button', { name: 'Refresh' }).hasAttribute('disabled')).toBe(false);
    expect(within(display()).queryByRole('button', { name: /Collecting/ })).toBeNull();
    fail = false;
    fireEvent.click(within(alert).getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(api.calls.filter((c) => c.method === 'POST').length).toBe(2));
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
  });

  it('offers to clear a search that matches nothing', async () => {
    const none: Envelope<CheckoutsData> = { ...checkouts, data: { ...checkouts.data, rows: [], query: 'zzz' } };
    const api = mockApi({ route: (call) => (call.url.searchParams.get('q') === 'zzz' ? json(none) : undefined) });
    await renderPage(api);
    fireEvent.input(searchBox(), { target: { value: 'zzz' } });
    await screen.findByText('No checkout matches the search.');
    expect(screen.getByRole('group', { name: 'Active filters' }).textContent).toContain('Searchmatches“zzz”');
    fireEvent.click(screen.getByRole('button', { name: 'Clear search' }));
    await screen.findByRole('listbox', { name: 'Checkouts' });
    expect((searchBox() as HTMLInputElement).value).toBe('');
  });

  it('keeps the stale snapshot and says why the last collection failed', async () => {
    const failed: Envelope<CheckoutsData> = { ...checkouts, warnings: ['git: boom'], data: { ...checkouts.data, error: 'git: boom' } };
    const api = mockApi({ route: (call) => (call.url.pathname === '/api/v0/checkouts' ? json(failed) : undefined) });
    await renderPage(api);
    const alerts = screen.getAllByRole('alert');
    expect(alerts).toHaveLength(1);
    expect(alerts[0].textContent).toContain('The last collection failed: git: boom');
    expect(bodyRows()).toHaveLength(4);
  });

  it('offers a retry when the service does not answer', async () => {
    let fail = true;
    const api = mockApi({
      route: (call) => (call.url.pathname === '/api/v0/checkouts' && fail ? json({ error: { code: 'unavailable', message: 'the checkouts are not collected by this service' } }, 503) : undefined),
    });
    render(<StatusPage client={new ApiClient('', api.fetch)} />);
    await screen.findByText('the checkouts are not collected by this service');
    fail = false;
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    await screen.findByRole('listbox', { name: 'Checkouts' });
  });

  it('shows the checkouts as a board of the current grouping, in the URL, and opens a card in the peek panel', async () => {
    history.replaceState(null, '', '/status');
    await renderPage();
    const layout = screen.getByRole('group', { name: 'Layout' });
    expect(within(layout).getByRole('button', { name: 'List' }).getAttribute('aria-pressed')).toBe('true');
    fireEvent.click(within(layout).getByRole('button', { name: 'Board' }));
    expect(location.search).toBe('?layout=board');
    expect(screen.queryByRole('listbox', { name: 'Checkouts' })).toBeNull();
    const expected = groupRows(checkouts.data.rows, 'state');
    const columns = screen.getAllByRole('region');
    expect(columns.map((c) => c.getAttribute('aria-label'))).toEqual(expected.map((g) => g.title));
    expect(columns.map((c) => within(c).getAllByRole('listitem').length)).toEqual(expected.map((g) => g.rows.length));
    // A card: the repository as the title, the branch and the counts in its footer.
    const maintainer = screen.getByText('octomation/maintainer').closest('button') as HTMLElement;
    expect(maintainer.querySelector('[data-cell="branch"]')?.textContent).toBe('main*');
    fireEvent.click(maintainer);
    await screen.findByRole('dialog', { name: 'octomation/maintainer' });
    fireEvent.keyDown(document, { key: 'Escape' });
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    // The board follows the grouping: without one it is a single column.
    fireEvent.change(within(display()).getByLabelText('Grouping'), { target: { value: 'none' } });
    expect(screen.getAllByRole('region').map((c) => c.getAttribute('aria-label'))).toEqual(['Checkouts']);
    fireEvent.click(within(screen.getByRole('group', { name: 'Layout' })).getByRole('button', { name: 'List' }));
    expect(location.search).toBe('');
    expect(screen.getByRole('listbox', { name: 'Checkouts' })).toBeTruthy();
  });

  it('shows a board column in the list from its header action', async () => {
    history.replaceState(null, '', '/status?layout=board');
    render(<StatusPage client={new ApiClient('', mockApi().fetch)} />);
    await waitFor(() => expect(screen.getAllByRole('region').length).toBeGreaterThan(0));
    const expected = groupRows(checkouts.data.rows, 'state');
    const last = expected[expected.length - 1];
    fireEvent.click(screen.getByRole('button', { name: `Show ${last.title} in the list` }));
    expect(location.search).toBe('');
    const list = await screen.findByRole('listbox', { name: 'Checkouts' });
    await waitFor(() => expect(document.activeElement?.textContent).toContain(last.rows[0].repository));
    expect(list.contains(document.activeElement)).toBe(true);
  });

  it('opens the board from the URL', async () => {
    history.replaceState(null, '', '/status?layout=board');
    render(<StatusPage client={new ApiClient('', mockApi().fetch)} />);
    await waitFor(() => expect(screen.getAllByRole('region').length).toBeGreaterThan(0));
    history.replaceState(null, '', '/status');
  });

  it('lays the properties out in one column grid for every group, measured from the values, and hides columns from View', async () => {
    localStorage.removeItem('maintainer.status.columns');
    await renderPage();
    const list = screen.getByRole('listbox', { name: 'Checkouts' });
    const template = list.style.getPropertyValue('--checkout-columns');
    // The leading slots, repository, branch, changes, status, then the path in the rest.
    expect(template.split(' ').slice(0, 3)).toEqual(['16px', '42px', '14px']);
    expect(template.endsWith('minmax(0, 1fr)')).toBe(true);
    expect(template.match(/\d+px/g)?.length).toBe(7);
    expect(template).toBe(columnTemplate(checkouts.data.rows, { branch: true, changes: true, status: true, path: true }));
    // Every row of every group has the same cells in the same order.
    expect(new Set(bodyRows().map((row) => Array.from(row.querySelectorAll('[data-cell]')).map((c) => c.getAttribute('data-cell')).join()))).toEqual(new Set(['lock,repository,branch,uncommitted,status,path']));
    const properties = within(display()).getByRole('group', { name: 'Display properties' });
    fireEvent.click(within(properties).getByRole('button', { name: 'Branch' }));
    expect(bodyRows()[0].querySelector('[data-cell="branch"]')).toBeNull();
    expect(list.style.getPropertyValue('--checkout-columns').match(/\d+px/g)?.length).toBe(6);
    expect(JSON.parse(localStorage.getItem('maintainer.status.columns') ?? '{}').branch).toBe(false);
    localStorage.removeItem('maintainer.status.columns');
  });
});

describe('model', () => {
  it('cycles the sort as the TUI does', () => {
    expect(cycleSort([], 'status', false)).toEqual([{ column: 'status', desc: false }]);
    expect(cycleSort([{ column: 'status', desc: false }], 'status', false)).toEqual([{ column: 'status', desc: true }]);
    expect(cycleSort([{ column: 'status', desc: true }], 'status', false)).toEqual([]);
    expect(cycleSort([{ column: 'status', desc: true }], 'path', false)).toEqual([{ column: 'path', desc: false }]);
    const chain = [
      { column: 'status', desc: true },
      { column: 'path', desc: false },
    ];
    expect(cycleSort(chain, 'repository', true)).toEqual([...chain, { column: 'repository', desc: false }]);
    expect(cycleSort(chain, 'status', true)).toEqual([{ column: 'path', desc: false }]);
    expect(cycleSort(chain, 'path', true)).toEqual([
      { column: 'status', desc: true },
      { column: 'path', desc: true },
    ]);
    expect(sortParam(chain)).toBe('-status,path');
    expect(sortParam([])).toBe('');
  });

  it('writes the header lines of the TUI', () => {
    expect(summaryLine({ checkouts: 1, uncommitted: 0, ahead: 0, behind: 0, locked: 0, orphan: 0, errors: 0 })).toBe(
      '1 checkout · 0 uncommitted · 0 ahead · 0 behind · 0 locked · 0 orphan · 0 errors · local refs (no fetch)',
    );
    expect(sortSummary([], '')).toBe('Sort: default (repository / path)');
    expect(sortSummary([], 'x')).toBe('Sort: relevance');
    expect(
      sortSummary(
        [
          { column: 'repository', desc: false },
          { column: 'lock', desc: true },
        ],
        'x',
      ),
    ).toBe('Sort: Repository ↑1  →  Lock ↓2');
    expect(formatTime('2026-10-20T12:00:00.5Z')).toBe('2026-10-20 12:00:00Z');
    expect(formatTime('2026-10-20T15:00:00+03:00')).toBe('2026-10-20 12:00:00Z');
    expect(formatTime(null)).toBe('');
    expect(formatTime('later')).toBe('later');
  });
});
