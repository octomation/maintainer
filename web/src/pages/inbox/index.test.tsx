import { fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiClient } from '../../api/client';
import { fixtures, json, mockApi, type MockCall } from '../../api/mock';
import type { Envelope, TaskDetail, TaskPage } from '../../api/types';
import { age, date } from '../../components/TaskDetails';
import { groupTasks } from './grouping';
import { InboxPage, PAGE_SIZE } from './index';

const page = fixtures.tasks;
const task = fixtures.task.data;

/** The list requests the mock saw, oldest first. */
function lists(calls: MockCall[]): MockCall[] {
  return calls.filter((c) => c.method === 'GET' && c.url.pathname === '/api/v0/tasks');
}

/** The filters and the sort of a list request, without the paging. */
function filtersOf(call: MockCall): string {
  const params = new URLSearchParams(call.url.search);
  params.delete('limit');
  params.delete('offset');
  return params.toString();
}

/** The Ordering select of the Display menu, opened when it is closed. */
function ordering(): HTMLSelectElement {
  const open = screen.queryByRole('group', { name: 'Display options' });
  if (!open) fireEvent.click(screen.getByRole('button', { name: 'View' }));
  return screen.getByLabelText('Ordering') as HTMLSelectElement;
}

/** Opens the filter menu at a dimension. */
function filterDimension(name: string): HTMLElement {
  if (!screen.queryByRole('group', { name: 'Filters' })) fireEvent.click(screen.getByRole('button', { name: /^Filter/ }));
  const menu = screen.getByRole('group', { name: 'Filters' });
  const back = within(menu).queryByRole('button', { name: 'Back to the filters' });
  if (back) fireEvent.click(back);
  fireEvent.click(within(screen.getByRole('group', { name: 'Filters' })).getByRole('button', { name: new RegExp(`^${name}`) }));
  return within(screen.getByRole('group', { name: 'Filters' })).getByRole('group', { name });
}

/** Picks a priority in the issue page of the pane. */
function pickPriority(pane: HTMLElement, value: string) {
  fireEvent.click(within(pane).getByRole('button', { name: /^Priority:/ }));
  fireEvent.click(within(screen.getByRole('menu', { name: 'Priority' })).getByRole('menuitemradio', { name: value }));
}

/** Picks an estimate in the issue page of the pane. */
function pickEstimate(pane: HTMLElement, value: string) {
  fireEvent.click(within(pane).getByRole('button', { name: /^Estimate:/ }));
  fireEvent.click(within(screen.getByRole('menu', { name: 'Estimate' })).getByRole('menuitemradio', { name: new RegExp(`^${value}`) }));
}

function rows(): HTMLElement[] {
  return within(screen.getByRole('listbox', { name: 'Tasks' })).getAllByRole('option');
}

/** A detail per task of the fixture page, so the pane shows the task that was opened. */
function detailOf(id: string): Envelope<TaskDetail> {
  const row = page.data.tasks.find((t) => t.id === id) ?? page.data.tasks[0];
  return { ...fixtures.task, data: { ...task, ...row, body_html: task.body_html, comments: task.comments, provenance: task.provenance } };
}

/** A mock whose task answers follow the id and whose PATCH can be held back. */
function perTaskApi(hold: { patch?: (release: () => void) => void }) {
  const details = new Map<string, Envelope<TaskDetail>>();
  const detail = (id: string) => details.get(id) ?? detailOf(id);
  return mockApi({
    route: (call) => {
      const triage = call.url.pathname.match(/^\/api\/v0\/tasks\/([^/]+)\/triage$/);
      if (call.method === 'PATCH' && triage) {
        const id = decodeURIComponent(triage[1]);
        const answer = () => {
          const current = detail(id);
          const patch = JSON.parse(call.body ?? '{}') as { impact?: string | null };
          const labels = current.data.labels.filter((l) => !(patch.impact !== undefined && l.startsWith('impact:')));
          if (patch.impact) labels.push(`impact: ${patch.impact}`);
          const next = { ...current, data: { ...current.data, labels, version: `${current.data.version}-2` } };
          details.set(id, next);
          return json(next, 200, { ETag: `"${next.data.version}"` });
        };
        if (hold.patch) {
          const once = hold.patch;
          hold.patch = undefined;
          return new Promise<Response>((resolve) => once(() => resolve(answer())));
        }
        return answer();
      }
      const one = call.url.pathname.match(/^\/api\/v0\/tasks\/([^/]+)$/);
      if (call.method === 'GET' && one) {
        const current = detail(decodeURIComponent(one[1]));
        return json(current, 200, { ETag: `"${current.data.version}"` });
      }
      return undefined;
    },
  });
}

describe('InboxPage', () => {
  beforeEach(() => {
    history.replaceState(null, '', '/inbox');
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('lists the tasks with the repository circle, code, title, repository, milestone, status and age', async () => {
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const all = rows();
    expect(all.length).toBe(page.data.tasks.length);
    const first = page.data.tasks[0];
    expect(all[0].textContent).toContain(first.code);
    expect(all[0].textContent).toContain(first.title);
    expect(within(all[0]).getByTitle(first.repository).textContent).toContain(first.repository.split('/')[1]);
    if (first.milestone) expect(all[0].textContent).toContain(first.milestone.title);
    // No score in the row: it is the ordering and lives in the issue page.
    expect(all[0].textContent).not.toContain(first.score.total.toFixed(2));
    expect(within(all[0]).getByTitle(`Updated ${date(first.updated_at)}`).textContent).toBe(age(first.updated_at));
    // The status sits over the age in the trailing column; the priority is spoken, the reader shows it; no badge without an event.
    expect(all[0].querySelector('.inbox-row__side [aria-label="Backlog"]')).toBeTruthy();
    expect(all[0].querySelector('.repo-avatar__badge')).toBeNull();
    expect(within(all[0]).getByText('Priority: High').className).toBe('sr-only');
    expect(within(all[0]).queryByText('type: feature')).toBeNull();
    expect(all[0].querySelector('.inbox-row__sub')?.textContent).toBe(`${first.repository.split('/')[1]}${first.milestone ? ` · ${first.milestone.title}` : ''}`);
    expect(screen.getByRole('status').textContent).toBe(`${page.data.total} open tasks`);
    const request = lists(api.calls)[0];
    expect(request.url.searchParams.get('limit')).toBe(String(PAGE_SIZE));
    expect(request.url.searchParams.get('offset')).toBe('0');
    expect(screen.getByText('Select a task to read it.')).toBeTruthy();
  });

  it('reads the filters and the sort from the URL and writes the sort back', async () => {
    history.replaceState(null, '', '/inbox?repo=octomation%2Fmaintainer&state=all');
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    expect(filtersOf(lists(api.calls)[0])).toBe('repo=octomation%2Fmaintainer&state=all');
    const sort = ordering();
    expect(sort.value).toBe('score');
    expect(Array.from(sort.options).map((o) => o.value)).toEqual(['score', 'importance', 'updated', 'created']);
    expect(Array.from(sort.options).map((o) => o.textContent)).toEqual(['Score', 'Priority', 'Updated', 'Created']);
    fireEvent.change(sort, { target: { value: 'updated' } });
    await waitFor(() => expect(lists(api.calls).length).toBe(2));
    expect(filtersOf(lists(api.calls)[1])).toBe('repo=octomation%2Fmaintainer&state=all&sort=updated');
    expect(location.search).toBe('?repo=octomation%2Fmaintainer&state=all&sort=updated');
    expect(ordering().value).toBe('updated');
  });

  it('opens a task with Enter, highlights the row and returns the focus with Escape', async () => {
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const row = rows()[0];
    row.focus();
    fireEvent.keyDown(row, { key: 'Enter' });
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    expect(within(pane).getByRole('heading', { level: 2 }).textContent).toBe(task.title);
    expect(row.getAttribute('aria-selected')).toBe('true');
    expect(rows()[1].getAttribute('aria-selected')).toBe('false');
    expect(location.search).toBe('');
    const button = within(pane).getByRole('button', { name: /^Priority:/ });
    button.focus();
    fireEvent.keyDown(button, { key: 'Escape' });
    expect(document.activeElement).toBe(row);
    expect(screen.getByRole('complementary', { name: 'Task' })).toBe(pane);
    fireEvent.keyDown(row, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(rows()[1]);
  });

  it('requests the list again after a triage with the same parameters and keeps the selection', async () => {
    history.replaceState(null, '', '/inbox?state=all&sort=importance');
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    fireEvent.click(rows()[0]);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    expect(lists(api.calls).length).toBe(1);
    pickPriority(pane, 'Low');
    await waitFor(() => expect(lists(api.calls).length).toBe(2));
    const [before, after] = lists(api.calls);
    expect(filtersOf(after)).toBe(filtersOf(before));
    expect(filtersOf(after)).toBe('state=all&sort=importance');
    expect(after.url.searchParams.get('offset')).toBe('0');
    expect(rows()[0].getAttribute('aria-selected')).toBe('true');
    expect(within(rows()[0]).getByText('Priority: Low')).toBeTruthy();
    expect(within(rows()[0]).queryByText('Priority: High')).toBeNull();
    expect(within(pane).getByRole('button', { name: 'Priority: Low' })).toBeTruthy();
  });

  it('discards a list answer that arrives for an older request', async () => {
    const slow = { ...page, data: { ...page.data, total: 1, tasks: [page.data.tasks[1]] } } satisfies Envelope<TaskPage>;
    let release: (() => void) | undefined;
    const api = mockApi({
      route: (call) => {
        if (call.method !== 'GET' || call.url.pathname !== '/api/v0/tasks' || call.url.searchParams.has('sort')) return undefined;
        return new Promise((resolve) => {
          release = () => resolve(json(slow));
        });
      },
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('search');
    await waitFor(() => expect(release).toBeDefined());
    fireEvent.change(ordering(), { target: { value: 'updated' } });
    await screen.findByRole('listbox', { name: 'Tasks' });
    expect(rows().length).toBe(page.data.tasks.length);
    release?.();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(rows().length).toBe(page.data.tasks.length);
    expect(screen.getByRole('status').textContent).toBe(`${page.data.total} open tasks`);
  });

  it('loads more rows after the first page', async () => {
    const api = mockApi({
      route: (call) => {
        if (call.method !== 'GET' || call.url.pathname !== '/api/v0/tasks') return undefined;
        const offset = Number(call.url.searchParams.get('offset') ?? 0);
        const limit = Math.min(Number(call.url.searchParams.get('limit') ?? 100), 3);
        return json({ ...page, data: { ...page.data, offset, limit, tasks: page.data.tasks.slice(offset, offset + limit) } });
      },
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    expect(rows().length).toBe(3);
    expect(screen.getByRole('status').textContent).toBe(`${page.data.total} open tasks · showing 3`);
    fireEvent.click(screen.getByRole('button', { name: `Load more (${page.data.total - 3} left)` }));
    await waitFor(() => expect(rows().length).toBe(6));
    expect(lists(api.calls)[1].url.searchParams.get('offset')).toBe('3');
    fireEvent.click(screen.getByRole('button', { name: /Load more/ }));
    await waitFor(() => expect(rows().length).toBe(page.data.total));
    expect(screen.queryByRole('button', { name: /Load more/ })).toBeNull();
    expect(rows().map((r) => r.getAttribute('data-id'))).toEqual(page.data.tasks.map((t) => t.id));
  });

  it('tells that the file changed on 412 and offers to reload the task', async () => {
    let fail = true;
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' && fail ? json({ error: { code: 'precondition_failed', message: 'the file changed; reload the task' } }, 412) : undefined),
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    fireEvent.click(rows()[0]);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    pickPriority(pane, 'Medium');
    const alert = await within(pane).findByRole('alert');
    expect(alert.textContent).toContain('The file changed');
    expect(lists(api.calls).length).toBe(1);
    fail = false;
    fireEvent.click(within(alert).getByRole('button', { name: 'Reload' }));
    await waitFor(() => expect(within(pane).queryByRole('alert')).toBeNull());
    pickPriority(pane, 'Medium');
    await waitFor(() => expect(lists(api.calls).length).toBe(2));
    expect(within(rows()[0]).getByText('Priority: Medium')).toBeTruthy();
  });

  it('explains duplicate mirrors on 409', async () => {
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' ? json({ error: { code: 'conflict', message: 'the task has duplicate mirrors' } }, 409) : undefined),
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    fireEvent.click(rows()[0]);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    pickEstimate(pane, 'M');
    const alert = await within(pane).findByRole('alert');
    expect(alert.textContent).toContain('mirrored by more than one checkout');
    expect(lists(api.calls).length).toBe(1);
  });

  it('captions the results it shows and marks them as previous after a failed request', async () => {
    history.replaceState(null, '', '/inbox?state=open');
    let fail = false;
    const api = mockApi({
      route: (call) => (call.method === 'GET' && call.url.pathname === '/api/v0/tasks' && fail ? json({ error: { code: 'internal', message: 'index is locked' } }, 500) : undefined),
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    expect(screen.getByRole('status').textContent).toBe(`${page.data.total} open tasks`);
    fail = true;
    fireEvent.click(within(filterDimension('State')).getByRole('radio', { name: 'Closed' }));
    await screen.findByRole('alert');
    // The open rows stay and the caption says so; it does not claim closed tasks.
    expect(rows().length).toBe(page.data.tasks.length);
    expect(screen.getByRole('status').textContent).toBe(`${page.data.total} open tasks · showing the previous results`);
    expect(screen.getByRole('group', { name: 'Active filters' }).textContent).toContain('StateisClosed');
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
    fail = false;
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
    expect(screen.getByRole('status').textContent).toBe(`${page.data.total} closed tasks`);
  });

  it('polls the catalog after a 202 and asks again after a failure once a list arrives, so milestones get their titles', async () => {
    vi.useFakeTimers();
    try {
      let catalogAnswers = 0;
      const api = mockApi({
        route: (call) => {
          if (call.url.pathname !== '/api/v0/catalog') return undefined;
          catalogAnswers++;
          if (catalogAnswers === 1) return json({ ...fixtures.catalog, refreshing: true, data: null }, 202);
          if (catalogAnswers === 2) return json({ error: { code: 'internal', message: 'no catalog' } }, 500);
          return undefined;
        },
      });
      render(<InboxPage client={new ApiClient('', api.fetch)} />);
      await vi.advanceTimersByTimeAsync(50);
      const option = () => {
        const titles = within(filterDimension('Milestone')).getAllByRole('checkbox').map((o) => o.getAttribute('aria-label'));
        fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
        return titles;
      };
      expect(option().some((t) => t?.startsWith('octomation/maintainer/milestones/1'))).toBe(true);
      // The 202 is polled; the failure that follows is told and keeps the keys.
      await vi.advanceTimersByTimeAsync(1100);
      expect(catalogAnswers).toBe(2);
      expect(screen.getByRole('alert').textContent).toContain('catalog could not be loaded');
      // The next successful list request asks for the catalog again.
      fireEvent.change(ordering(), { target: { value: 'updated' } });
      await vi.advanceTimersByTimeAsync(50);
      expect(catalogAnswers).toBe(3);
      expect(screen.queryByRole('alert')).toBeNull();
      expect(option().some((t) => t?.startsWith('Workspace, v0.2'))).toBe(true);
    } finally {
      vi.useRealTimers();
    }
  });

  it('keeps the rows when a refresh fails and tells so where it can be seen', async () => {
    let fail = false;
    const api = mockApi({
      route: (call) => (call.method === 'GET' && call.url.pathname === '/api/v0/tasks' && fail ? json({ error: { code: 'internal', message: 'index is locked' } }, 500) : undefined),
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    fail = true;
    fireEvent.change(ordering(), { target: { value: 'updated' } });
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('index is locked');
    expect(rows().length).toBe(page.data.tasks.length);
    fail = false;
    fireEvent.click(within(alert).getByRole('button', { name: 'Try again' }));
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
    expect(lists(api.calls).length).toBe(3);
  });

  it('tells a refresh failure inside the pane on a narrow screen', async () => {
    vi.stubGlobal('matchMedia', (query: string) => ({ matches: query.includes('max-width'), media: query, addEventListener() {}, removeEventListener() {} }));
    let fail = false;
    const api = mockApi({
      route: (call) => (call.method === 'GET' && call.url.pathname === '/api/v0/tasks' && fail ? json({ error: { code: 'internal', message: 'index is locked' } }, 500) : undefined),
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    fireEvent.click(rows()[0]);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    // A pointer opening focuses the pane too.
    expect(pane.contains(document.activeElement)).toBe(true);
    fail = true;
    pickPriority(pane, 'Low');
    const alert = await within(pane).findByRole('alert');
    expect(alert.textContent).toContain('index is locked');
    expect(rows().length).toBe(page.data.tasks.length);
  });

  it('focuses the pane again when the selected row is activated once more', async () => {
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const row = rows()[0];
    fireEvent.click(row);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    expect(pane.contains(document.activeElement)).toBe(true);
    row.focus();
    expect(pane.contains(document.activeElement)).toBe(false);
    fireEvent.keyDown(row, { key: 'Enter' });
    expect(pane.contains(document.activeElement)).toBe(true);
  });

  it('shows the indexing state before the first snapshot and the empty state without matches', async () => {
    const indexing = mockApi({ listStatus: 202 });
    const { unmount } = render(<InboxPage client={new ApiClient('', indexing.fetch)} />);
    await screen.findByText(/Indexing/);
    unmount();
    const empty = mockApi({ tasks: { ...page, data: { ...page.data, total: 0, tasks: [] } } });
    const collection = render(<InboxPage client={new ApiClient('', empty.fetch)} />);
    await screen.findByText('No tasks in the collection.');
    expect(screen.queryByRole('button', { name: 'Clear filters' })).toBeNull();
    collection.unmount();
    history.replaceState(null, '', '/inbox?repo=octomation%2Fmaintainer&state=closed&sort=updated');
    render(<InboxPage client={new ApiClient('', mockApi({ tasks: { ...page, data: { ...page.data, total: 0, tasks: [] } } }).fetch)} />);
    await screen.findByText('No tasks match the filters.');
    fireEvent.click(screen.getByRole('button', { name: 'Clear filters' }));
    await waitFor(() => expect(location.search).toBe('?state=closed&sort=updated'));
  });

  it('drops a triage of the previous task that completes after another one opened', async () => {
    let release: (() => void) | undefined;
    const api = perTaskApi({ patch: (r) => (release = r) });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const [, b] = page.data.tasks;
    fireEvent.click(rows()[0]);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    pickPriority(pane, 'Low');
    await waitFor(() => expect(release).toBeDefined());
    fireEvent.click(rows()[1]);
    await waitFor(() => expect(within(pane).getByRole('heading', { level: 2 }).textContent).toBe(b.title));
    fireEvent.change(ordering(), { target: { value: 'updated' } });
    await waitFor(() => expect(lists(api.calls).length).toBe(2));
    release?.();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(within(pane).getByRole('heading', { level: 2 }).textContent).toBe(b.title);
    expect(lists(api.calls).length).toBe(2);
    expect(within(rows()[0]).queryByText('Priority: Low')).toBeNull();
    expect(within(rows()[0]).getByText('Priority: High')).toBeTruthy();
    // A triage of the task that is open requeries with the filters of now.
    pickPriority(pane, 'Low');
    await waitFor(() => expect(lists(api.calls).length).toBe(3));
    expect(filtersOf(lists(api.calls)[2])).toBe('sort=updated');
    expect(within(rows()[1]).getByText('Priority: Low')).toBeTruthy();
    expect(rows()[1].getAttribute('aria-selected')).toBe('true');
  });

  it('keeps "Load more" disabled while a replacement request is pending and never appends to rows of another query', async () => {
    let release: (() => void) | undefined;
    const api = mockApi({
      route: (call) => {
        if (call.method !== 'GET' || call.url.pathname !== '/api/v0/tasks') return undefined;
        const offset = Number(call.url.searchParams.get('offset') ?? 0);
        const sorted = call.url.searchParams.get('sort') === 'updated';
        const all = sorted ? [...page.data.tasks].reverse() : page.data.tasks;
        const answer = () => json({ ...page, data: { ...page.data, offset, limit: 3, tasks: all.slice(offset, offset + 3) } });
        if (sorted && !release) return new Promise<Response>((resolve) => (release = () => resolve(answer())));
        return answer();
      },
    });
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const more = () => screen.getByRole('button', { name: /Load more/ }) as HTMLButtonElement;
    expect(more().disabled).toBe(false);
    fireEvent.change(ordering(), { target: { value: 'updated' } });
    await waitFor(() => expect(release).toBeDefined());
    expect(more().disabled).toBe(true);
    fireEvent.click(more());
    expect(lists(api.calls).length).toBe(2);
    release?.();
    await waitFor(() => expect(more().disabled).toBe(false));
    expect(rows().map((r) => r.getAttribute('data-id'))).toEqual([...page.data.tasks].reverse().slice(0, 3).map((t) => t.id));
    fireEvent.click(more());
    await waitFor(() => expect(rows().length).toBe(6));
    const third = lists(api.calls)[2];
    expect(third.url.searchParams.get('sort')).toBe('updated');
    expect(third.url.searchParams.get('offset')).toBe('3');
    expect(rows().map((r) => r.getAttribute('data-id'))).toEqual([...page.data.tasks].reverse().slice(0, 6).map((t) => t.id));
  });

  it('focuses the pane on Enter and closes it with Escape on a narrow screen', async () => {
    vi.stubGlobal('matchMedia', (query: string) => ({ matches: query.includes('max-width'), media: query, addEventListener() {}, removeEventListener() {} }));
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const row = rows()[0];
    row.focus();
    fireEvent.keyDown(row, { key: 'Enter' });
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    expect(pane.contains(document.activeElement)).toBe(true);
    expect(within(pane).getByRole('button', { name: 'Back to the list' })).toBeTruthy();
    const listWrap = screen.getByRole('listbox', { name: 'Tasks' }).closest('.inbox__list-pane') as HTMLElement & { inert: boolean };
    const filtersWrap = screen.getByRole('search').closest('.topbar') as HTMLElement & { inert: boolean };
    expect(listWrap.inert).toBe(true);
    expect(listWrap.hasAttribute('inert')).toBe(true);
    expect(filtersWrap.inert).toBe(true);
    expect((pane as HTMLElement & { inert: boolean }).inert).toBe(false);
    expect(pane.hasAttribute('inert')).toBe(false);
    fireEvent.keyDown(document.activeElement as Element, { key: 'Escape' });
    await waitFor(() => expect(screen.queryByRole('heading', { level: 2 })).toBeNull());
    expect(listWrap.inert).toBe(false);
    expect(listWrap.hasAttribute('inert')).toBe(false);
    expect(filtersWrap.hasAttribute('inert')).toBe(false);
    expect(document.activeElement).toBe(rows()[0]);
    expect(rows()[0].getAttribute('aria-selected')).toBe('false');
  });

  it('lifts the inert state when the layout widens with the pane open', async () => {
    let listener: (() => void) | undefined;
    const media = { matches: true, media: '', addEventListener: (_: string, fn: () => void) => (listener = fn), removeEventListener() {} };
    vi.stubGlobal('matchMedia', () => media);
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    fireEvent.click(rows()[0]);
    await screen.findByRole('complementary', { name: 'Task' });
    const listWrap = screen.getByRole('listbox', { name: 'Tasks' }).closest('.inbox__list-pane') as HTMLElement;
    await waitFor(() => expect(listWrap.hasAttribute('inert')).toBe(true));
    media.matches = false;
    listener?.();
    await waitFor(() => expect(listWrap.hasAttribute('inert')).toBe(false));
    expect(screen.getByRole('complementary', { name: 'Task' })).toBeTruthy();
    expect(rows()[0].getAttribute('aria-selected')).toBe('true');
  });

  it('opens the reader as the full issue page and comes back to the row', async () => {
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const id = rows()[0].dataset.id!;
    fireEvent.click(rows()[0]);
    const pane = await screen.findByRole('complementary', { name: 'Task' });
    await within(pane).findByRole('heading', { level: 2 });
    fireEvent.click(within(pane).getByRole('button', { name: 'Open as a page' }));
    const issue = await screen.findByRole('region', { name: 'Issue' });
    expect(new URLSearchParams(location.search).get('task')).toBe(id);
    await within(issue).findByRole('heading', { level: 2 });
    // The breadcrumb is the way back.
    const back = within(issue).getByRole('button', { name: 'Back to the inbox' });
    expect(back.classList.contains('issue-header__crumb')).toBe(true);
    fireEvent.click(back);
    await waitFor(() => expect(screen.queryByRole('region', { name: 'Issue' })).toBeNull());
    expect(location.search).toBe('');
    expect(rows()[0].getAttribute('aria-selected')).toBe('true');
    await waitFor(() => expect(document.activeElement).toBe(rows()[0]));
  });

  it('has no list/board switch and groups the rows from View, with the group in the URL and the arrows across groups', async () => {
    history.replaceState(null, '', '/inbox');
    const api = mockApi();
    render(<InboxPage client={new ApiClient('', api.fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    expect(screen.queryByRole('navigation', { name: 'Layout' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'View' }));
    fireEvent.change(screen.getByLabelText('Grouping'), { target: { value: 'status' } });
    expect(location.search).toBe('?group=status');
    const listbox = screen.getByRole('listbox', { name: 'Tasks' });
    const groups = within(listbox).getAllByRole('group');
    const expected = groupTasks(page.data.tasks, 'status', page.data.facets, page.data.total);
    expect(groups.map((g) => g.querySelector('.group-header')?.textContent)).toEqual(expected.map((g) => `${g.title}${g.count}`));
    expect(groups.map((g) => within(g).getAllByRole('option').map((o) => o.getAttribute('data-id')))).toEqual(expected.map((g) => g.rows.map((t) => t.id)));
    // The arrows follow the rows as shown, from the last row of a group to the first of the next.
    const lastOfFirst = within(groups[0]).getAllByRole('option').pop() as HTMLElement;
    lastOfFirst.focus();
    fireEvent.keyDown(lastOfFirst, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(within(groups[1]).getAllByRole('option')[0]);
    // Selection works inside a group.
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Enter' });
    await screen.findByRole('complementary', { name: 'Task' });
    expect(within(groups[1]).getAllByRole('option')[0].getAttribute('aria-selected')).toBe('true');
    // The filters keep the grouping in the URL.
    history.replaceState(null, '', '/inbox');
  });

  it('reads the grouping from the URL', async () => {
    history.replaceState(null, '', '/inbox?group=repository');
    render(<InboxPage client={new ApiClient('', mockApi().fetch)} />);
    await screen.findByRole('listbox', { name: 'Tasks' });
    const headers = Array.from(document.querySelectorAll('.inbox__group .group-header__name')).map((h) => h.textContent);
    expect(headers).toEqual(groupTasks(page.data.tasks, 'repository', page.data.facets, page.data.total).map((g) => g.title));
    history.replaceState(null, '', '/inbox');
  });
});

describe('groupTasks', () => {
  const tasks = page.data.tasks;
  const facets = page.data.facets;
  it('orders the priorities and gives the group without one the rest of the total', () => {
    const groups = groupTasks(tasks, 'priority', facets, page.data.total);
    const titles = groups.map((g) => g.title);
    expect(titles.filter((t) => t !== 'No priority')).toEqual(['High', 'Medium', 'Low'].filter((t) => titles.includes(t)));
    if (titles.includes('No priority')) expect(titles[titles.length - 1]).toBe('No priority');
    const high = facets.labels.filter((f) => f.value.toLowerCase().replace(/\s/g, '') === 'impact:high').reduce((sum, f) => sum + f.count, 0);
    expect(groups.find((g) => g.title === 'High')?.count).toBe(high);
    expect(groups.flatMap((g) => g.rows).length).toBe(tasks.length);
  });

  it('keeps the rows in the list order inside a group and puts the rest last', () => {
    const groups = groupTasks(tasks, 'milestone', facets, page.data.total);
    for (const group of groups) expect(group.rows).toEqual(tasks.filter((t) => group.rows.includes(t)));
    const none = groups.findIndex((g) => g.title === 'No milestone');
    if (none >= 0) expect(none).toBe(groups.length - 1);
    expect(groupTasks(tasks, 'none', facets, 9)).toEqual([{ key: 'all', title: '', glyph: null, count: 9, rows: tasks }]);
  });
});
