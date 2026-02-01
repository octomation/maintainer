import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { describe, expect, it, vi } from 'vitest';
import { ApiClient } from './api/client';
import { fixtures, json, mockApi } from './api/mock';
import { App } from './app';
import { pushQuery, QUERY_EVENT, queryFromSearch, searchFromQuery } from './lib/query';

describe('App', () => {
  it('renders the sidebar with the current page and the three routes', async () => {
    const client = new ApiClient('', mockApi().fetch);
    render(<App path="/inbox" client={client} />);
    const nav = screen.getByRole('navigation', { name: 'Pages' });
    expect(nav.textContent).toContain('Status');
    expect(nav.textContent).toContain('Inbox');
    expect(nav.textContent).toContain('Tablo');
    expect(screen.getByRole('link', { name: 'Inbox' }).getAttribute('aria-current')).toBe('page');
    expect(screen.getByRole('link', { name: 'Status' }).getAttribute('aria-current')).toBeNull();
    await screen.findByRole('search');
  });

  it('shows the indexing state until the first snapshot', async () => {
    const client = new ApiClient('', mockApi({ listStatus: 202 }).fetch);
    render(<App path="/tablo" client={client} />);
    await screen.findByText(/Indexing/);
  });

  it('starts dark and switches the theme to light, system and back', async () => {
    localStorage.removeItem('maintainer.theme');
    const client = new ApiClient('', mockApi().fetch);
    render(<App path="/status" client={client} />);
    await waitFor(() => expect(document.documentElement.getAttribute('data-theme')).toBe('dark'));
    fireEvent.click(screen.getByRole('button', { name: /Theme: dark/ }));
    await waitFor(() => expect(document.documentElement.getAttribute('data-theme')).toBe('light'));
    expect(localStorage.getItem('maintainer.theme')).toBe('light');
    fireEvent.click(screen.getByRole('button', { name: /Theme: light/ }));
    await waitFor(() => expect(document.documentElement.getAttribute('data-theme')).toBe('system'));
    fireEvent.click(screen.getByRole('button', { name: /Theme: system/ }));
    await waitFor(() => expect(document.documentElement.getAttribute('data-theme')).toBe('dark'));
    expect(localStorage.getItem('maintainer.theme')).toBeNull();
  });

  it('opens the sidebar from the top bar on a narrow screen', async () => {
    vi.stubGlobal('matchMedia', (query: string) => ({ matches: query.includes('max-width'), media: query, addEventListener() {}, removeEventListener() {} }));
    try {
      const client = new ApiClient('', mockApi().fetch);
      render(<App path="/status" client={client} />);
      const sidebar = screen.getByRole('complementary', { name: 'Sidebar' });
      expect(sidebar.className).not.toContain('sidebar--open');
      const menu = screen.getByRole('button', { name: 'Open the sidebar' });
      menu.focus();
      fireEvent.click(menu);
      await waitFor(() => expect(sidebar.className).toContain('sidebar--open'));
      await waitFor(() => expect(document.activeElement).toBe(screen.getByRole('link', { name: 'Status' })));
      expect((document.getElementById('main') as HTMLElement & { inert: boolean }).inert).toBe(true);
      fireEvent.keyDown(document, { key: 'Escape' });
      await waitFor(() => expect(sidebar.className).not.toContain('sidebar--open'));
      await waitFor(() => expect(document.activeElement).toBe(menu));
    } finally {
      vi.unstubAllGlobals();
    }
  });
});

describe('App shortcuts', () => {
  it('opens the command bar with / and Ctrl K, not while typing', async () => {
    const client = new ApiClient('', mockApi().fetch);
    render(<App path="/status" client={client} />);
    const field = document.createElement('input');
    document.body.appendChild(field);
    fireEvent.keyDown(field, { key: '/' });
    expect(screen.queryByRole('dialog', { name: 'Search' })).toBeNull();
    field.remove();
    fireEvent.keyDown(document.body, { key: '/' });
    expect(screen.getByRole('dialog', { name: 'Search' })).toBeTruthy();
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Escape' });
    expect(screen.queryByRole('dialog', { name: 'Search' })).toBeNull();
    fireEvent.keyDown(document.body, { key: 'k', ctrlKey: true });
    expect(screen.getByRole('dialog', { name: 'Search' })).toBeTruthy();
  });
});

describe('App issue commands', () => {
  const task = fixtures.task.data;

  it('opens the command bar over the open issue and triages it through the issue page, 412 included', async () => {
    history.replaceState(null, '', `/inbox?task=${encodeURIComponent(task.id)}`);
    let fail = true;
    const api = mockApi({
      route: (call) => (call.method === 'PATCH' && fail ? json({ error: { code: 'precondition_failed', message: 'the file changed' } }, 412) : undefined),
    });
    render(<App path="/inbox" client={new ApiClient('', api.fetch)} />);
    const page = await screen.findByRole('region', { name: 'Issue' });
    await within(page).findByRole('heading', { level: 2 });
    fireEvent.keyDown(document.body, { key: '/' });
    const dialog = screen.getByRole('dialog', { name: 'Search' });
    expect(dialog.querySelector('.command-bar__chip')?.textContent).toBe(`${task.code} — ${task.title}`);
    fireEvent.click(within(dialog).getByText('Change priority…'));
    fireEvent.click(within(dialog).getByText('Low'));
    await waitFor(() => expect(api.calls.filter((c) => c.method === 'PATCH').length).toBe(1));
    const patch = api.calls.find((c) => c.method === 'PATCH');
    expect(patch?.headers.get('If-Match')).toBe(`"${task.version}"`);
    expect(JSON.parse(patch?.body ?? '')).toEqual({ impact: 'low' });
    // The 412 is told on the issue page, which blocks the triage until Reload.
    const alert = await within(page).findByRole('alert');
    expect(alert.textContent).toContain('The file changed');
    fail = false;
    fireEvent.click(within(alert).getByRole('button', { name: 'Reload' }));
    await waitFor(() => expect(within(page).queryByRole('alert')).toBeNull());
    fireEvent.keyDown(document.body, { key: '/' });
    fireEvent.click(within(screen.getByRole('dialog', { name: 'Search' })).getByText('Change estimate…'));
    fireEvent.click(within(screen.getByRole('dialog', { name: 'Search' })).getByText(/^L ·/));
    await waitFor(() => expect(api.calls.filter((c) => c.method === 'PATCH').length).toBe(2));
    expect(JSON.parse(api.calls.filter((c) => c.method === 'PATCH')[1].body ?? '')).toEqual({ effort: 'hard' });
    history.replaceState(null, '', '/');
  });

  it('opens the levels with P and E on an open issue, not while typing', async () => {
    history.replaceState(null, '', `/inbox?task=${encodeURIComponent(task.id)}`);
    render(<App path="/inbox" client={new ApiClient('', mockApi().fetch)} />);
    const page = await screen.findByRole('region', { name: 'Issue' });
    await within(page).findByRole('heading', { level: 2 });
    fireEvent.keyDown(document.body, { key: 'p' });
    let dialog = screen.getByRole('dialog', { name: 'Search' });
    expect(within(dialog).getByRole('combobox').getAttribute('placeholder')).toBe('Change priority to…');
    expect(within(dialog).getAllByRole('option').map((o) => o.getAttribute('aria-checked'))).toEqual(['false', 'true', 'false', 'false']);
    fireEvent.keyDown(within(dialog).getByRole('combobox'), { key: 'Escape' });
    fireEvent.keyDown(within(dialog).getByRole('combobox'), { key: 'Escape' });
    expect(screen.queryByRole('dialog', { name: 'Search' })).toBeNull();
    fireEvent.keyDown(document.body, { key: 'E' });
    dialog = screen.getByRole('dialog', { name: 'Search' });
    expect(within(dialog).getByRole('combobox').getAttribute('placeholder')).toBe('Change estimate to…');
    fireEvent.keyDown(within(dialog).getByRole('combobox'), { key: 'Escape' });
    fireEvent.keyDown(within(dialog).getByRole('combobox'), { key: 'Escape' });
    // Not from a text field, and not without an open issue.
    const field = document.createElement('input');
    document.body.appendChild(field);
    fireEvent.keyDown(field, { key: 'p' });
    expect(screen.queryByRole('dialog', { name: 'Search' })).toBeNull();
    field.remove();
    history.replaceState(null, '', '/');
  });

  it('ignores P and E without an open issue', async () => {
    render(<App path="/status" client={new ApiClient('', mockApi().fetch)} />);
    fireEvent.keyDown(document.body, { key: 'p' });
    expect(screen.queryByRole('dialog', { name: 'Search' })).toBeNull();
  });

  it('is the global search where no issue is open', async () => {
    render(<App path="/status" client={new ApiClient('', mockApi().fetch)} />);
    fireEvent.keyDown(document.body, { key: '/' });
    const dialog = screen.getByRole('dialog', { name: 'Search' });
    expect(dialog.querySelector('.command-bar__chip')).toBeNull();
    expect(within(dialog).queryByText('Change priority…')).toBeNull();
  });
});

describe('App sidebar tree', () => {
  it('marks the repository the inbox is filtered by, with its lock, and follows the filters', async () => {
    localStorage.removeItem('maintainer.sidebar.owners');
    history.replaceState(null, '', '/inbox?repo=octomation%2Ftool');
    render(<App path="/inbox" client={new ApiClient('', mockApi().fetch)} />);
    const tree = await screen.findByRole('tree', { name: 'Repositories' });
    const selected = () => within(tree).getAllByRole('treeitem').find((i) => i.getAttribute('aria-selected') === 'true');
    await waitFor(() => expect(selected()?.textContent).toContain('tool'));
    expect(within(selected() as HTMLElement).getByRole('img', { name: 'Private' })).toBeTruthy();
    act(() => pushQuery({ repo: ['octomation/maintainer'] }));
    await waitFor(() => expect(selected()?.textContent).toContain('maintainer'));
    expect(within(selected() as HTMLElement).queryByRole('img', { name: 'Private' })).toBeNull();
    history.replaceState(null, '', '/');
  });
});

describe('query', () => {
  it('keeps the view parameters when the filters change and tells the shell', () => {
    history.replaceState(null, '', '/inbox?group=status&task=x&repo=a%2Fb');
    const heard = vi.fn();
    addEventListener(QUERY_EVENT, heard);
    pushQuery({ repo: ['c/d'], state: 'all' });
    removeEventListener(QUERY_EVENT, heard);
    const params = new URLSearchParams(location.search);
    expect(params.getAll('repo')).toEqual(['c/d']);
    expect(params.get('group')).toBe('status');
    expect(params.get('task')).toBe('x');
    expect(heard).toHaveBeenCalledTimes(1);
    history.replaceState(null, '', '/');
  });

  it('round-trips the filters through the URL and ignores paging', () => {
    const search = searchFromQuery({ repo: ['a/b', 'c/d'], state: 'all', label: ['impact: high'], sort: 'updated', q: 'x y', limit: 50, offset: 100, column: ['todo'] });
    expect(search).toBe('?repo=a%2Fb&repo=c%2Fd&label=impact%3A+high&column=todo&state=all&q=x+y&sort=updated');
    expect(queryFromSearch(search)).toEqual({ repo: ['a/b', 'c/d'], state: 'all', label: ['impact: high'], sort: 'updated', q: 'x y', column: ['todo'] });
    expect(queryFromSearch('?state=bogus&sort=title&column=nope&repo=')).toEqual({});
    expect(searchFromQuery({})).toBe('');
  });
});
