import { fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiClient } from '../api/client';
import { fixtures, json, mockApi } from '../api/mock';
import { Sidebar } from './Sidebar';

describe('Sidebar', () => {
  it('reindexes: POSTs /refresh, polls the catalog until the generation changes, then reloads', async () => {
    let generation = 1;
    const api = mockApi({
      route: (call) => {
        if (call.method === 'POST' && call.url.pathname === '/api/v0/refresh') {
          return json({ generation: 1, collected_at: null, refreshing: true, warnings: [], data: null }, 202);
        }
        if (call.method === 'GET' && call.url.pathname === '/api/v0/catalog') {
          generation += 1;
          return json({ ...fixtures.catalog, generation: generation > 3 ? 2 : 1 });
        }
        return undefined;
      },
    });
    const onReindexed = vi.fn();
    render(<Sidebar current="/inbox" theme="system" onTheme={() => {}} status="generation 1" client={new ApiClient('', api.fetch)} onReindexed={onReindexed} pollMs={1} />);
    const button = screen.getByRole('button', { name: 'Reindex' }) as HTMLButtonElement;
    expect(button.disabled).toBe(false);
    fireEvent.click(button);
    await screen.findByRole('button', { name: 'Reindexing…' });
    expect((screen.getByRole('button', { name: 'Reindexing…' }) as HTMLButtonElement).disabled).toBe(true);
    await waitFor(() => expect(onReindexed).toHaveBeenCalledTimes(1));
    const post = api.calls.find((c) => c.method === 'POST');
    expect(post?.url.pathname).toBe('/api/v0/refresh');
    expect(post?.headers.get('X-Requested-With')).toBe('maintainer');
    expect(api.calls.filter((c) => c.url.pathname === '/api/v0/catalog').length).toBe(3);
    await screen.findByRole('button', { name: 'Reindex' });
    expect(screen.getByText('generation 1')).toBeTruthy();
  });

  it('shows a failed reindex in the status text', async () => {
    const api = mockApi({
      route: (call) => (call.method === 'POST' ? json({ error: { code: 'internal', message: 'the mirrors are locked' } }, 500) : undefined),
    });
    const onReindexed = vi.fn();
    render(<Sidebar current="/status" theme="light" onTheme={() => {}} status="collected" client={new ApiClient('', api.fetch)} onReindexed={onReindexed} pollMs={1} />);
    fireEvent.click(screen.getByRole('button', { name: 'Reindex' }));
    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toBe('Reindex failed: the mirrors are locked');
    expect(screen.queryByText('collected')).toBeNull();
    expect((screen.getByRole('button', { name: 'Reindex' }) as HTMLButtonElement).disabled).toBe(false);
    expect(onReindexed).not.toHaveBeenCalled();
  });

  it('hides the button without a client', () => {
    render(<Sidebar current="/status" theme="dark" onTheme={() => {}} />);
    expect(screen.queryByRole('button', { name: 'Reindex' })).toBeNull();
  });

  it('lists the three pages with the current one marked, and names the theme switch', () => {
    const onTheme = vi.fn();
    render(<Sidebar current="/tablo" theme="dark" onTheme={onTheme} client={new ApiClient('', mockApi().fetch)} />);
    const nav = screen.getByRole('navigation', { name: 'Pages' });
    const links = within(within(nav).getByRole('list')).getAllByRole('link');
    expect(links.map((a) => a.textContent)).toEqual(['Inbox', 'Tablo', 'Status']);
    expect(links.map((a) => a.getAttribute('aria-current'))).toEqual([null, 'page', null]);
    // Every page has its icon; the theme switch names itself and shows the theme, no emoji.
    expect(links.every((a) => a.querySelector('svg'))).toBe(true);
    const theme = screen.getByRole('button', { name: 'Theme: dark; switch' });
    expect(theme.textContent?.trim()).toBe('Dark');
    expect(theme.getAttribute('title')).toBe('Theme: dark; switch');
    expect(theme.querySelector('svg')).toBeTruthy();
    fireEvent.click(theme);
    expect(onTheme).toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Reindex' }).querySelector('svg')).toBeTruthy();
    expect(screen.getByRole('complementary', { name: 'Sidebar' }).textContent).not.toMatch(/[\u{1F300}-\u{1FAFF}☀-➿]/u);
  });

  it('opens as an overlay on a narrow screen and closes from its backdrop and its button', () => {
    const onClose = vi.fn();
    render(<Sidebar current="/status" theme="dark" onTheme={() => {}} open onClose={onClose} />);
    expect(screen.getByRole('complementary', { name: 'Sidebar' }).className).toContain('sidebar--open');
    fireEvent.click(screen.getByTestId('sidebar-backdrop'));
    fireEvent.click(screen.getByRole('button', { name: 'Close the sidebar' }));
    expect(onClose).toHaveBeenCalledTimes(2);
  });

  it('puts Reindex and Search in the action row', () => {
    const onSearch = vi.fn();
    render(<Sidebar current="/inbox" theme="dark" onTheme={() => {}} client={new ApiClient('', mockApi().fetch)} onSearch={onSearch} />);
    const actions = document.querySelector('.sidebar__actions') as HTMLElement;
    expect(within(actions).getAllByRole('button').map((b) => b.getAttribute('aria-label') ?? b.textContent)).toEqual(['Reindex', 'Search']);
    fireEvent.click(within(actions).getByRole('button', { name: 'Search' }));
    expect(onSearch).toHaveBeenCalled();
  });

  describe('Repositories tree', () => {
    const repo = (repository: string, open: number, isPrivate = false) => ({ repository, owner: repository.split('/')[0], name: repository.split('/')[1], weight: 1, path: `/w/${repository}`, tasks: open + 1, open, private: isPrivate });
    const repositories = [repo('octomation/tool', 3, true), repo('kamilsk/dotfiles', 2), repo('octomation/maintainer', 7), repo('kamilsk/dotfiles', 2), repo('Acme/zeta', 0)];
    beforeEach(() => localStorage.removeItem('maintainer.sidebar.owners'));

    it('lists every repository by owner, the current owner expanded, with open counts and a lock on private ones', () => {
      render(<Sidebar current="/inbox" theme="dark" onTheme={() => {}} repositories={repositories} currentRepository="octomation/tool" />);
      const tree = screen.getByRole('tree', { name: 'Repositories' });
      const owners = within(tree).getAllByRole('treeitem').filter((i) => i.getAttribute('aria-level') === '1');
      expect(owners.map((o) => [o.querySelector('.repo-tree__name')?.textContent, o.getAttribute('aria-expanded'), o.querySelector('.repo-tree__count')?.textContent])).toEqual([
        ['Acme', 'false', '1'],
        ['kamilsk', 'false', '1'],
        ['octomation', 'true', '2'],
      ]);
      const repos = within(tree).getAllByRole('treeitem').filter((i) => i.getAttribute('aria-level') === '2');
      expect(repos.map((r) => r.querySelector('.repo-tree__name')?.textContent)).toEqual(['maintainer', 'tool']);
      expect(repos[1].getAttribute('aria-selected')).toBe('true');
      expect(repos[1].getAttribute('href')).toBe('/inbox?repo=octomation%2Ftool');
      expect(within(repos[1]).getByRole('img', { name: 'Private' })).toBeTruthy();
      expect(within(repos[0]).queryByRole('img', { name: 'Private' })).toBeNull();
      expect(within(repos[0]).getByLabelText('7 open issues')).toBeTruthy();
    });

    it('toggles owners, remembers them and moves with the keyboard', () => {
      const { unmount } = render(<Sidebar current="/status" theme="dark" onTheme={() => {}} repositories={repositories} />);
      const tree = () => screen.getByRole('tree', { name: 'Repositories' });
      const owner = (name: string) => within(tree()).getAllByRole('treeitem').find((i) => i.getAttribute('aria-level') === '1' && i.textContent?.startsWith(name)) as HTMLElement;
      // Nothing is current: every owner is collapsed.
      expect(within(tree()).getAllByRole('treeitem').every((i) => i.getAttribute('aria-level') === '1')).toBe(true);
      fireEvent.click(owner('kamilsk'));
      expect(owner('kamilsk').getAttribute('aria-expanded')).toBe('true');
      // Duplicate mirrors of a repository are one row.
      expect(within(tree()).getAllByRole('treeitem').filter((i) => i.getAttribute('aria-level') === '2').map((r) => r.textContent)).toEqual(['dotfiles2']);
      expect(JSON.parse(localStorage.getItem('maintainer.sidebar.owners') ?? '[]')).toEqual(['kamilsk']);
      // Keyboard: Down to the repository, Left back to the owner, Left collapses, Down and Right expand the next owner.
      owner('kamilsk').focus();
      fireEvent.keyDown(owner('kamilsk'), { key: 'ArrowDown' });
      expect(document.activeElement?.textContent).toBe('dotfiles2');
      fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'ArrowLeft' });
      expect(document.activeElement).toBe(owner('kamilsk'));
      fireEvent.keyDown(owner('kamilsk'), { key: 'ArrowLeft' });
      expect(owner('kamilsk').getAttribute('aria-expanded')).toBe('false');
      fireEvent.keyDown(owner('kamilsk'), { key: 'ArrowDown' });
      expect(document.activeElement).toBe(owner('octomation'));
      fireEvent.keyDown(owner('octomation'), { key: 'ArrowRight' });
      expect(owner('octomation').getAttribute('aria-expanded')).toBe('true');
      fireEvent.keyDown(owner('octomation'), { key: 'Enter' });
      expect(owner('octomation').getAttribute('aria-expanded')).toBe('false');
      unmount();
      // The choice survives a reload.
      localStorage.setItem('maintainer.sidebar.owners', JSON.stringify(['Acme']));
      render(<Sidebar current="/status" theme="dark" onTheme={() => {}} repositories={repositories} />);
      expect(owner('Acme').getAttribute('aria-expanded')).toBe('true');
    });
  });
});
