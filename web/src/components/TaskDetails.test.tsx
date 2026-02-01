import { fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { describe, expect, it, vi } from 'vitest';
import { ApiClient } from '../api/client';
import { fixtures, json, mockApi } from '../api/mock';
import type { TaskDetail } from '../api/types';
import { DECORATION, decorationOf } from './Glyphs';
import { age, TaskDetails, withoutTitle } from './TaskDetails';

/** The property button of Priority or Estimate. */
const property = (name: string) => screen.getByRole('button', { name: new RegExp(`^${name}:`) }) as HTMLButtonElement;
/** Opens a property's menu and picks a value by its title. */
function pick(name: string, value: string) {
  fireEvent.click(property(name));
  fireEvent.click(within(screen.getByRole('menu', { name })).getByRole('menuitemradio', { name: new RegExp(`^${value}`) }));
}

describe('TaskDetails', () => {
  const task = fixtures.task.data;

  it('renders the task, its body, comments, score and provenance', async () => {
    const api = mockApi();
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} />);
    expect(screen.getByText('Loading the task…')).toBeTruthy();
    await screen.findByRole('heading', { level: 2 });
    expect(screen.getByRole('heading', { level: 2 }).textContent).toBe(task.title);
    // The code in the breadcrumb and in the toolbar of the properties rail.
    expect(screen.getAllByText(task.code)).toHaveLength(2);
    expect(within(screen.getByLabelText('Properties')).getByText(task.code)).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Open on GitHub' }).getAttribute('href')).toBe(task.url);
    // The leading H1 repeats the title, which the page shows above the body.
    expect(screen.getByLabelText('Body').innerHTML).not.toContain('<h1>');
    expect(screen.getByLabelText('Body').innerHTML).toContain('<p>');
    // One chronological stack: the comments are events of the activity, with no subsection row.
    expect(screen.queryByText(/Comments \(/)).toBeNull();
    expect(screen.getByLabelText('Activity').querySelectorAll('.activity__item--comment').length).toBe(task.comments.length);
    expect(screen.getByLabelText('Provenance').textContent).toContain(task.provenance.path);
    expect(screen.getByLabelText('Score').textContent).toContain(`Score ${task.score.total.toFixed(2)}`);
    for (const label of task.labels.filter((l) => !/^(impact|effort):/.test(l))) expect(screen.getAllByText(label).length).toBeGreaterThan(0);
    expect(screen.getByLabelText('Dates').textContent).toContain('Created');
    expect(screen.getByLabelText('Properties').textContent).toContain(task.assignees[0]);
    // impact: high and effort: easy are the priority and the estimate.
    expect(property('Priority').getAttribute('aria-label')).toBe('Priority: High');
    expect(property('Estimate').getAttribute('aria-label')).toBe('Estimate: S');
    fireEvent.click(property('Priority'));
    const menu = screen.getByRole('menu', { name: 'Priority' });
    expect(within(menu).getAllByRole('menuitemradio').map((item) => [item.textContent, item.getAttribute('aria-checked')])).toEqual([
      ['No priority', 'false'],
      ['High', 'true'],
      ['Medium', 'false'],
      ['Low', 'false'],
    ]);
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
    expect(screen.queryByRole('menu')).toBeNull();
    expect(document.activeElement).toBe(property('Priority'));
  });

  it('triages with the ETag, says so beside the group and reports the new detail', async () => {
    const api = mockApi();
    const onChanged = vi.fn<(detail: TaskDetail) => void>();
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} onChanged={onChanged} />);
    await screen.findByRole('heading', { level: 2 });
    pick('Priority', 'Low');
    expect(screen.getByRole('status').textContent).toBe('Saving…');
    await waitFor(() => expect(onChanged).toHaveBeenCalled());
    expect(screen.getByRole('status').textContent).toBe('Saved');
    expect(onChanged.mock.calls[0][0].labels).toContain('impact: low');
    expect(property('Priority').getAttribute('aria-label')).toBe('Priority: Low');
    const patch = api.calls.find((c) => c.method === 'PATCH');
    expect(patch?.headers.get('If-Match')).toBe(`"${task.version}"`);
    expect(JSON.parse(patch?.body ?? '')).toEqual({ impact: 'low' });
    pick('Estimate', 'No estimate');
    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(2));
    const second = api.calls.filter((c) => c.method === 'PATCH')[1];
    expect(JSON.parse(second.body ?? '')).toEqual({ effort: null });
    expect(second.headers.get('If-Match')).not.toBe(`"${task.version}"`);
    expect(property('Estimate').getAttribute('aria-label')).toBe('Estimate: No estimate');
    fireEvent.click(property('Estimate'));
    const none = within(screen.getByRole('menu', { name: 'Estimate' })).getByRole('menuitemradio', { name: 'No estimate' }) as HTMLButtonElement;
    expect(none.getAttribute('aria-checked')).toBe('true');
    expect(none.disabled).toBe(true);
  });

  it('offers to reload on 412, blocks every triage action until then, and reloads the task', async () => {
    let fail = true;
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' && fail ? json({ error: { code: 'precondition_failed', message: 'the file changed; reload the task' } }, 412) : undefined),
    });
    const onChanged = vi.fn();
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} onChanged={onChanged} />);
    await screen.findByRole('heading', { level: 2 });
    pick('Priority', 'Medium');
    await screen.findByRole('alert');
    expect(screen.getByRole('alert').textContent).toContain('The file changed');
    expect(onChanged).not.toHaveBeenCalled();
    for (const name of ['Priority', 'Estimate']) expect(property(name).disabled).toBe(true);
    fail = false;
    fireEvent.click(screen.getByRole('button', { name: 'Reload' }));
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
    expect(api.calls.filter((c) => c.method === 'GET').length).toBe(2);
    expect(property('Priority').disabled).toBe(false);
    pick('Priority', 'Medium');
    await waitFor(() => expect(onChanged).toHaveBeenCalled());
  });

  it('keeps the triage blocked and says "Reloading…" until the reloaded task is in', async () => {
    let release: (() => void) | undefined;
    let reloads = 0;
    const api = mockApi({
      route: (call) => {
        if (call.method === 'PATCH') return json({ error: { code: 'precondition_failed', message: 'the file changed' } }, 412);
        if (call.method === 'GET' && call.url.pathname.startsWith('/api/v0/tasks/') && ++reloads === 2) {
          return new Promise<Response>((resolve) => (release = () => resolve(json(fixtures.task, 200, { ETag: '"fresh"' }))));
        }
        return undefined;
      },
    });
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} />);
    await screen.findByRole('heading', { level: 2 });
    pick('Priority', 'Medium');
    await screen.findByRole('alert');
    fireEvent.click(screen.getByRole('button', { name: 'Reload' }));
    await waitFor(() => expect(release).toBeDefined());
    // The old detail stays on screen, every triage action stays blocked, the reload is announced.
    expect(screen.getByRole('heading', { level: 2 }).textContent).toBe(task.title);
    expect(screen.getByText('Reloading…')).toBeTruthy();
    expect(property('Priority').disabled).toBe(true);
    release?.();
    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
    expect(property('Priority').disabled).toBe(false);
    pick('Priority', 'Medium');
    await waitFor(() => expect(api.calls.filter((c) => c.method === 'PATCH').length).toBe(2));
    expect(api.calls.filter((c) => c.method === 'PATCH')[1].headers.get('If-Match')).toBe('"fresh"');
  });

  it('explains duplicate mirrors on 409', async () => {
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' ? json({ error: { code: 'conflict', message: 'the task has duplicate mirrors; edit the file of one of them' } }, 409) : undefined),
    });
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} />);
    await screen.findByRole('heading', { level: 2 });
    pick('Estimate', 'M');
    await screen.findByRole('alert');
    expect(screen.getByRole('alert').textContent).toContain('mirrored by more than one checkout');
  });

  it('explains an unsupported label layout on 422', async () => {
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' ? json({ error: { code: 'unprocessable', message: 'labels are not a list' } }, 422) : undefined),
    });
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} />);
    await screen.findByRole('heading', { level: 2 });
    pick('Estimate', 'L');
    await screen.findByRole('alert');
    expect(screen.getByRole('alert').textContent).toContain('cannot edit');
    expect(screen.getByRole('alert').textContent).toContain('labels are not a list');
    expect(property('Estimate').disabled).toBe(false);
  });

  it('shows an error state with a retry', async () => {
    let fail = true;
    const api = mockApi({ route: (call) => (call.method === 'GET' && fail ? json({ error: { code: 'not_found', message: 'no such task' } }, 404) : undefined) });
    render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} />);
    await screen.findByRole('alert');
    expect(screen.getByText('no such task')).toBeTruthy();
    fail = false;
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    await screen.findByRole('heading', { level: 2 });
  });

  it('drops a triage that completes after the pane was unmounted', async () => {
    let release: (() => void) | undefined;
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' ? new Promise<Response>((resolve) => (release = () => resolve(json(fixtures.task, 200, { ETag: '"x"' })))) : undefined),
    });
    const onChanged = vi.fn();
    const { unmount } = render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} onChanged={onChanged} />);
    await screen.findByRole('heading', { level: 2 });
    pick('Priority', 'Low');
    await waitFor(() => expect(release).toBeDefined());
    unmount();
    release?.();
    await new Promise((resolve) => setTimeout(resolve, 10));
    expect(onChanged).not.toHaveBeenCalled();
  });

  it('shows Back and Expand when asked, and copies the link and the path from the rail', async () => {
    const api = mockApi();
    const onBack = vi.fn();
    const onExpand = vi.fn();
    const writeText = vi.fn(() => Promise.resolve());
    vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText } });
    try {
      render(<TaskDetails client={new ApiClient('', api.fetch)} id={task.id} onBack={onBack} onExpand={onExpand} />);
      await screen.findByRole('heading', { level: 2 });
      // The breadcrumb is the way back, a button the keyboard reaches.
      const back = screen.getByRole('button', { name: 'Back to the list' });
      expect(back.classList.contains('issue-header__crumb')).toBe(true);
      expect(back.textContent).toContain(task.code);
      back.focus();
      expect(document.activeElement).toBe(back);
      fireEvent.click(back);
      expect(onBack).toHaveBeenCalled();
      fireEvent.click(screen.getByRole('button', { name: 'Open as a page' }));
      expect(onExpand).toHaveBeenCalled();
      const rail = screen.getByLabelText('Properties');
      fireEvent.click(within(rail).getByRole('button', { name: 'Copy the link to the issue' }));
      expect(writeText).toHaveBeenLastCalledWith(task.url);
      await within(rail).findByText('Copied');
      fireEvent.click(within(rail).getByRole('button', { name: 'Copy the file path' }));
      expect(writeText).toHaveBeenLastCalledWith(task.provenance.path);
      expect(within(rail).getByRole('link', { name: 'Open on GitHub' }).getAttribute('href')).toBe(task.url);
      // The issue menu of the header holds the same actions and the code.
      fireEvent.click(screen.getByRole('button', { name: 'Issue actions' }));
      const menu = screen.getByRole('menu', { name: 'Issue actions' });
      expect(within(menu).getAllByRole('menuitem').map((item) => item.textContent)).toEqual(['Copy link', 'Copy issue code', 'Copy file path', 'Open on GitHub']);
      fireEvent.click(within(menu).getByRole('menuitem', { name: 'Copy issue code' }));
      expect(writeText).toHaveBeenLastCalledWith(task.code);
      expect(screen.queryByRole('menu')).toBeNull();
    } finally {
      vi.unstubAllGlobals();
    }
  });

  it('writes the activity with the values bright inside muted sentences', async () => {
    render(<TaskDetails client={new ApiClient('', mockApi().fetch)} id={task.id} />);
    await screen.findByRole('heading', { level: 2 });
    const activity = screen.getByLabelText('Activity');
    const created = activity.querySelector('.activity__item');
    expect(created?.querySelector('.activity__marker')).toBeTruthy();
    expect(created?.querySelector('.activity__value')?.textContent).toBe(task.repository.split('/')[1]);
    expect(created?.textContent).toContain('created the issue');
  });
});

describe('helpers', () => {
  it('drops a leading H1 only when it repeats the title', () => {
    expect(withoutTitle('<h1 id="x">A &amp; B</h1>\n<p>body</p>', 'A & B')).toBe('<p>body</p>');
    expect(withoutTitle('<h1>Other</h1><p>body</p>', 'A')).toBe('<h1>Other</h1><p>body</p>');
    expect(withoutTitle('<p>body</p><h1>A</h1>', 'A')).toBe('<p>body</p><h1>A</h1>');
  });

  it('writes ages as the inbox shows them', () => {
    const now = Date.parse('2026-10-21T12:00:00Z');
    expect(age('2026-10-21T11:55:00Z', now)).toBe('5m');
    expect(age('2026-10-21T09:00:00Z', now)).toBe('3h');
    expect(age('2026-10-18T12:00:00Z', now)).toBe('3d');
    expect(age('2026-09-30T12:00:00Z', now)).toBe('3w');
    expect(age('2026-06-21T12:00:00Z', now)).toBe('4mo');
    expect(age('2024-10-21T12:00:00Z', now)).toBe('2y');
    expect(age(null, now)).toBe('—');
  });

  it('gives a label the same color every time', () => {
    expect(decorationOf('scope: code')).toBe(decorationOf('scope: code'));
    expect(DECORATION).toContain(decorationOf('type: bug'));
  });
});
