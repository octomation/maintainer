import { fireEvent, render, screen, waitFor, within } from '@testing-library/preact';
import { describe, expect, it, vi } from 'vitest';
import { ApiClient } from '../api/client';
import { fixtures, mockApi } from '../api/mock';
import { CommandBar, type IssueCommands } from './CommandBar';

describe('CommandBar', () => {
  const tasks = fixtures.tasks.data.tasks;

  it('lists the pages and the open issues, searches the collection and opens the chosen issue', async () => {
    const api = mockApi();
    const onOpenTask = vi.fn();
    const onClose = vi.fn();
    render(<CommandBar client={new ApiClient('', api.fetch)} open onClose={onClose} onOpenTask={onOpenTask} onGo={() => {}} />);
    const dialog = screen.getByRole('dialog', { name: 'Search' });
    const input = within(dialog).getByRole('combobox', { name: 'Search issues or go to a page' });
    expect(document.activeElement).toBe(input);
    await within(dialog).findByText(tasks[0].title);
    const first = api.calls.find((c) => c.url.pathname === '/api/v0/tasks');
    expect(first?.url.searchParams.get('state')).toBe('open');
    expect(first?.url.searchParams.get('sort')).toBe('score');
    expect(within(dialog).getAllByRole('option').slice(0, 3).map((o) => o.textContent)).toEqual(['Go to Inbox', 'Go to Tablo', 'Go to Status']);
    fireEvent.input(input, { target: { value: 'adopt' } });
    await waitFor(() => expect(api.calls.filter((c) => c.url.searchParams.get('q') === 'adopt').length).toBe(1));
    const search = api.calls.find((c) => c.url.searchParams.get('q') === 'adopt');
    expect(search?.url.searchParams.get('state')).toBe('all');
    // No page matches the text: the first option is an issue.
    await waitFor(() => expect(within(dialog).getAllByRole('option')[0].textContent).toContain(tasks[0].code));
    // Nothing is picked until the arrows pick a result.
    expect(within(dialog).getAllByRole('option').every((o) => o.getAttribute('aria-selected') === 'false')).toBe(true);
    expect(input.getAttribute('aria-activedescendant')).toBeNull();
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    expect(within(dialog).getAllByRole('option')[0].getAttribute('aria-selected')).toBe('true');
    expect(input.getAttribute('aria-activedescendant')).toBe(within(dialog).getAllByRole('option')[0].id);
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    expect(within(dialog).getAllByRole('option')[1].getAttribute('aria-selected')).toBe('true');
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onClose).toHaveBeenCalled();
    expect(onOpenTask).toHaveBeenCalledWith(tasks[0].id);
  });

  it('opens the first result on Enter with nothing picked, goes to a page and closes on Escape', () => {
    const onGo = vi.fn();
    const onClose = vi.fn();
    render(<CommandBar client={new ApiClient('', mockApi().fetch)} open onClose={onClose} onOpenTask={() => {}} onGo={onGo} />);
    const input = screen.getByRole('combobox');
    fireEvent.input(input, { target: { value: 'tab' } });
    expect(screen.getAllByRole('option').every((o) => o.getAttribute('aria-selected') === 'false')).toBe(true);
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onGo).toHaveBeenCalledWith('/tablo');
    fireEvent.keyDown(input, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(2);
  });

  it('is issue-scoped over an open issue: the chip, the actions, the levels, Backspace back to the global search', async () => {
    const triage = vi.fn();
    const copy = vi.fn();
    const onOpenUrl = vi.fn();
    const onClose = vi.fn();
    const issue: IssueCommands = { id: 'x', code: 'MAIN-7', title: 'A long issue title', url: 'https://github.com/o/r/issues/7', path: '/m/.github/issues/MAIN-7.md', impact: 'high', effort: '', blocked: false, triage, copy };
    const { rerender } = render(<CommandBar client={new ApiClient('', mockApi().fetch)} open onClose={onClose} onOpenTask={() => {}} onGo={() => {}} issue={issue} onOpenUrl={onOpenUrl} />);
    const dialog = screen.getByRole('dialog', { name: 'Search' });
    expect(dialog.querySelector('.command-bar__chip')?.textContent).toBe('MAIN-7 — A long issue title');
    const input = within(dialog).getByRole('combobox');
    const titles = () => within(dialog).getAllByRole('option').map((o) => o.querySelector('.command-item__title')?.textContent);
    expect(titles().slice(0, 5)).toEqual(['Change priority…', 'Change estimate…', 'Copy link to the issue', 'Copy file path', 'Open on GitHub']);
    // The actions start right under the input: no group heading before them.
    expect(within(dialog).getByRole('listbox').firstElementChild?.getAttribute('role')).toBe('option');
    // Keycaps only for the working shortcuts; a rule between triage and the copy and open actions.
    const keys = within(dialog).getAllByRole('option').slice(0, 5).map((o) => Array.from(o.querySelectorAll('kbd')).map((k) => k.textContent).join(' '));
    expect(keys).toEqual(['P', 'E', '', '', '']);
    const separator = within(dialog).getByRole('separator');
    expect(separator.previousElementSibling?.textContent).toContain('Change estimate…');
    expect(separator.nextElementSibling?.textContent).toContain('Copy link to the issue');
    // Typing filters the actions; the global results follow.
    fireEvent.input(input, { target: { value: 'copy' } });
    expect(titles().slice(0, 2)).toEqual(['Copy link to the issue', 'Copy file path']);
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(copy).toHaveBeenCalledWith(issue.path, 'path');
    expect(onClose).toHaveBeenCalledTimes(1);
    // Reopened: the levels of the priority, the current one checked, Escape back to the actions.
    rerender(<CommandBar client={new ApiClient('', mockApi().fetch)} open={false} onClose={onClose} onOpenTask={() => {}} onGo={() => {}} issue={issue} onOpenUrl={onOpenUrl} />);
    rerender(<CommandBar client={new ApiClient('', mockApi().fetch)} open onClose={onClose} onOpenTask={() => {}} onGo={() => {}} issue={issue} onOpenUrl={onOpenUrl} />);
    fireEvent.click(screen.getByText('Change priority…'));
    const levels = screen.getAllByRole('option');
    expect(levels.map((o) => [o.querySelector('.command-item__title')?.textContent, o.getAttribute('aria-checked')])).toEqual([
      ['No priority', 'false'],
      ['High', 'true'],
      ['Medium', 'false'],
      ['Low', 'false'],
    ]);
    expect(screen.getByRole('combobox').getAttribute('placeholder')).toBe('Change priority to…');
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Escape' });
    expect(screen.getByText('Change estimate…')).toBeTruthy();
    fireEvent.click(screen.getByText('Change priority…'));
    fireEvent.click(screen.getByText('Low'));
    expect(triage).toHaveBeenCalledWith({ impact: 'low' });
    expect(onClose).toHaveBeenCalledTimes(2);
    // Backspace in the empty field takes the chip off: the global search.
    rerender(<CommandBar client={new ApiClient('', mockApi().fetch)} open={false} onClose={onClose} onOpenTask={() => {}} onGo={() => {}} issue={issue} onOpenUrl={onOpenUrl} />);
    rerender(<CommandBar client={new ApiClient('', mockApi().fetch)} open onClose={onClose} onOpenTask={() => {}} onGo={() => {}} issue={issue} onOpenUrl={onOpenUrl} />);
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Backspace' });
    expect(document.querySelector('.command-bar__chip')).toBeNull();
    expect(screen.queryByText('Change priority…')).toBeNull();
    expect(screen.getByRole('combobox').getAttribute('placeholder')).toBe('Search issues or go to a page…');
  });

  it('stays the global search without an open issue', () => {
    render(<CommandBar client={new ApiClient('', mockApi().fetch)} open onClose={() => {}} onOpenTask={() => {}} onGo={() => {}} />);
    expect(document.querySelector('.command-bar__chip')).toBeNull();
    expect(screen.queryByText('Change priority…')).toBeNull();
  });
});
