import { fireEvent, render, screen, within } from '@testing-library/preact';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { fixtures } from '../api/mock';
import type { TaskQuery } from '../api/types';
import { TaskFilters } from './TaskFilters';

/** A viewport where every query up to the given width matches, as a narrow screen does. */
function stubWidth(width: number) {
  vi.stubGlobal('matchMedia', (query: string) => {
    const max = /max-width:\s*(\d+)px/.exec(query);
    return { matches: max ? width <= Number(max[1]) : false, media: query, addEventListener() {}, removeEventListener() {} };
  });
}

function openFilter() {
  fireEvent.click(screen.getByRole('button', { name: /Filter/ }));
  return screen.getByRole('group', { name: 'Filters' });
}

/** Opens the filter menu at a dimension and returns the group of its values. */
function openDimension(name: string) {
  const menu = screen.queryByRole('group', { name: 'Filters' }) ?? openFilter();
  fireEvent.click(within(menu).getByRole('button', { name: new RegExp(`^${name}`) }));
  return within(screen.getByRole('group', { name: 'Filters' })).getByRole('group', { name });
}

describe('TaskFilters', () => {
  const facets = fixtures.tasks.data.facets;
  afterEach(() => vi.unstubAllGlobals());

  it('lists the dimensions and offers the facets as values with their counts', () => {
    render(<TaskFilters value={{}} facets={facets} onChange={() => {}} showColumn />);
    const menu = openFilter();
    const dimensions = within(within(menu).getByRole('list', { name: 'Dimensions' })).getAllByRole('button').map((b) => b.textContent);
    expect(dimensions).toEqual(['State', 'Status', 'Content', 'Priority', 'Estimate', 'Labels', 'Type', 'Milestone', 'Repository', 'Owner', 'Assignee']);
    // The groups are separated, and the search row shows the F shortcut.
    expect(menu.querySelectorAll('.menu-separator').length).toBe(2);
    expect(menu.querySelector('kbd')?.textContent).toBe('F');
    // No chevrons on the dimensions and no footer: Clear lives in the top bar.
    expect(menu.querySelector('.menu-item__chevron, .menu-footer')).toBeNull();
    const repo = openDimension('Repository');
    // The name first, then the owner; the value stays owner/name.
    const option = within(repo).getByRole('checkbox', { name: 'maintainer · octomation' });
    expect(option.closest('label')?.textContent).toContain(String(facets.repositories[0].count));
    const state = openFilterAt('State');
    expect((within(state).getByRole('radio', { name: 'Open' }) as HTMLInputElement).checked).toBe(true);
    expect(state.textContent).toContain(`Open${facets.states.find((f) => f.value === 'open')?.count ?? 0}`);
  });

  it('narrows the dimensions with the menu search and steps back', () => {
    render(<TaskFilters value={{}} facets={facets} onChange={() => {}} />);
    const menu = openFilter();
    fireEvent.input(within(menu).getByRole('textbox', { name: 'Filter dimensions' }), { target: { value: 'mile' } });
    expect(within(within(menu).getByRole('list', { name: 'Dimensions' })).getAllByRole('button').map((b) => b.textContent)).toEqual(['Milestone']);
    openDimension('Milestone');
    fireEvent.click(screen.getByRole('button', { name: 'Back to the filters' }));
    expect(within(screen.getByRole('group', { name: 'Filters' })).getByRole('list', { name: 'Dimensions' })).toBeTruthy();
  });

  it('reports values of one dimension as any-of and keeps the others', () => {
    const onChange = vi.fn<(next: TaskQuery) => void>();
    const { rerender } = render(<TaskFilters value={{ state: 'all', repo: ['octomation/tool'] }} facets={facets} onChange={onChange} />);
    fireEvent.click(within(openDimension('Owner')).getByRole('checkbox', { name: 'octomation' }));
    expect(onChange).toHaveBeenLastCalledWith({ state: 'all', repo: ['octomation/tool'], owner: ['octomation'] });
    fireEvent.click(within(openFilterAt('Repository')).getByRole('checkbox', { name: 'maintainer · octomation' }));
    expect(onChange).toHaveBeenLastCalledWith({ state: 'all', repo: ['octomation/tool', 'octomation/maintainer'] });
    fireEvent.click(within(openFilterAt('State')).getByRole('radio', { name: 'Closed' }));
    expect(onChange).toHaveBeenLastCalledWith({ state: 'closed', repo: ['octomation/tool'] });
    rerender(<TaskFilters value={{ state: 'all', repo: ['octomation/tool', 'octomation/maintainer'] }} facets={facets} onChange={onChange} />);
    const chip = within(screen.getByRole('group', { name: 'Active filters' })).getByText('Repository').closest('.filter-chip') as HTMLElement;
    expect(chip.textContent).toContain('is any of');
    expect(chip.textContent).toContain('tool · octomation, maintainer · octomation');
  });

  it('selects several labels, all of which must match, and removes them from the chip', () => {
    const onChange = vi.fn<(next: TaskQuery) => void>();
    const { rerender } = render(<TaskFilters value={{}} facets={facets} onChange={onChange} />);
    const labels = openDimension('Labels');
    // Priority and estimate are dimensions of their own, not labels.
    expect(within(labels).queryByRole('checkbox', { name: 'impact: high' })).toBeNull();
    fireEvent.click(within(labels).getByRole('checkbox', { name: 'scope: code' }));
    expect(onChange).toHaveBeenLastCalledWith({ label: ['scope: code'] });
    rerender(<TaskFilters value={{ label: ['scope: code'] }} facets={facets} onChange={onChange} />);
    fireEvent.click(within(screen.getByRole('group', { name: 'Labels' })).getByRole('checkbox', { name: 'type: feature' }));
    expect(onChange).toHaveBeenLastCalledWith({ label: ['scope: code', 'type: feature'] });
    rerender(<TaskFilters value={{ label: ['impact: high', 'scope: code', 'type: feature'] }} facets={facets} onChange={onChange} />);
    const strip = screen.getByRole('group', { name: 'Active filters' });
    expect(strip.textContent).toContain('include all of');
    fireEvent.click(within(strip).getByRole('button', { name: 'Remove label filter' }));
    // The priority stays: it is a filter of its own.
    expect(onChange).toHaveBeenLastCalledWith({ label: ['impact: high'] });
  });

  it('picks one priority and one estimate, each replacing the previous level', () => {
    const onChange = vi.fn<(next: TaskQuery) => void>();
    const { rerender } = render(<TaskFilters value={{ label: ['scope: code'] }} facets={facets} onChange={onChange} />);
    const priority = openDimension('Priority');
    expect(within(priority).getAllByRole('checkbox').map((c) => c.getAttribute('aria-label'))).toEqual(['High', 'Medium', 'Low']);
    fireEvent.click(within(priority).getByRole('checkbox', { name: 'High' }));
    expect(onChange).toHaveBeenLastCalledWith({ label: ['scope: code', 'impact: high'] });
    rerender(<TaskFilters value={{ label: ['scope: code', 'impact: high'] }} facets={facets} onChange={onChange} />);
    fireEvent.click(within(screen.getByRole('group', { name: 'Priority' })).getByRole('checkbox', { name: 'Low' }));
    expect(onChange).toHaveBeenLastCalledWith({ label: ['scope: code', 'impact: low'] });
    fireEvent.click(within(openFilterAt('Estimate')).getByRole('checkbox', { name: 'S (easy)' }));
    expect(onChange).toHaveBeenLastCalledWith({ label: ['scope: code', 'impact: high', 'effort: easy'] });
    const strip = screen.getByRole('group', { name: 'Active filters' });
    expect(strip.textContent).toContain('PriorityisHigh');
    fireEvent.click(within(strip).getByRole('button', { name: 'Remove priority filter' }));
    expect(onChange).toHaveBeenLastCalledWith({ label: ['scope: code'] });
  });

  it('opens the menu from the trigger and closes it with Escape, focus back on the trigger', () => {
    render(<TaskFilters value={{}} facets={facets} onChange={() => {}} />);
    const trigger = screen.getByRole('button', { name: /Filter/ });
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
    const menu = openFilter();
    expect(trigger.getAttribute('aria-expanded')).toBe('true');
    expect(menu.contains(document.activeElement)).toBe(true);
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
    expect(screen.queryByRole('group', { name: 'Filters' })).toBeNull();
    expect(document.activeElement).toBe(trigger);
    openFilter();
    fireEvent.mouseDown(document.body);
    expect(screen.queryByRole('group', { name: 'Filters' })).toBeNull();
  });

  it('opens the dimension of a chip from its values', () => {
    render(<TaskFilters value={{ milestone: ['github:octomation/maintainer/milestones/1'] }} facets={facets} onChange={() => {}} />);
    const strip = screen.getByRole('group', { name: 'Active filters' });
    fireEvent.click(within(strip).getByTitle(/^Milestone:/));
    expect(within(screen.getByRole('group', { name: 'Filters' })).getByRole('group', { name: 'Milestone' })).toBeTruthy();
  });

  it('keeps a selected value that the facets no longer list', () => {
    render(<TaskFilters value={{ milestone: ['github:x/y/milestones/9'] }} facets={facets} onChange={() => {}} />);
    const option = within(openDimension('Milestone')).getByRole('checkbox', { name: 'x/y/milestones/9' }) as HTMLInputElement;
    expect(option.checked).toBe(true);
    expect(option.closest('label')?.textContent).toContain('0');
  });

  it('debounces the content search, chips it and clears every filter but the state and the sort', async () => {
    vi.useFakeTimers();
    try {
      const onChange = vi.fn<(next: TaskQuery) => void>();
      const { rerender } = render(<TaskFilters value={{ repo: ['octomation/tool'], state: 'closed', sort: 'updated' }} facets={facets} onChange={onChange} />);
      openDimension('Content');
      fireEvent.input(screen.getByRole('textbox', { name: 'Content contains' }), { target: { value: 'fetch' } });
      expect(onChange).not.toHaveBeenCalled();
      vi.advanceTimersByTime(100);
      expect(onChange).not.toHaveBeenCalled();
      vi.advanceTimersByTime(30);
      expect(onChange).toHaveBeenLastCalledWith({ repo: ['octomation/tool'], state: 'closed', sort: 'updated', q: 'fetch' });
      rerender(<TaskFilters value={{ repo: ['octomation/tool'], state: 'closed', sort: 'updated', q: 'fetch' }} facets={facets} onChange={onChange} />);
      expect(screen.getByRole('group', { name: 'Active filters' }).textContent).toContain('Contentcontains“fetch”');
      fireEvent.click(screen.getByRole('button', { name: 'Clear' }));
      expect(onChange).toHaveBeenLastCalledWith({ state: 'closed', sort: 'updated' });
    } finally {
      vi.useRealTimers();
    }
  });

  it('opens with the F key unless a text field has the focus', () => {
    render(
      <>
        <input aria-label="Elsewhere" />
        <TaskFilters value={{}} facets={facets} onChange={() => {}} />
      </>,
    );
    fireEvent.keyDown(screen.getByLabelText('Elsewhere'), { key: 'f' });
    expect(screen.queryByRole('group', { name: 'Filters' })).toBeNull();
    fireEvent.keyDown(document.body, { key: 'f' });
    expect(screen.getByRole('group', { name: 'Filters' })).toBeTruthy();
  });

  it('hides the state when the page fixes it', () => {
    render(<TaskFilters value={{ state: 'all' }} facets={facets} onChange={() => {}} showState={false} />);
    const menu = openFilter();
    expect(within(menu).queryByRole('button', { name: /^State/ })).toBeNull();
    expect(screen.queryByRole('group', { name: 'Active filters' })).toBeNull();
  });

  it('chips a non-default state and removing it returns to the default', () => {
    const onChange = vi.fn<(next: TaskQuery) => void>();
    const { rerender } = render(<TaskFilters value={{ state: 'all', repo: ['octomation/tool'] }} facets={facets} onChange={onChange} />);
    const strip = screen.getByRole('group', { name: 'Active filters' });
    expect(strip.textContent).toContain('StateisAll');
    fireEvent.click(within(strip).getByRole('button', { name: 'Remove state filter' }));
    expect(onChange).toHaveBeenLastCalledWith({ state: 'open', repo: ['octomation/tool'] });
    rerender(<TaskFilters value={{ state: 'open', repo: ['octomation/tool'] }} facets={facets} onChange={onChange} />);
    expect(screen.getByRole('group', { name: 'Active filters' }).textContent).not.toContain('State');
  });

  it('uses the page default state for the chip: the board defaults to All and chips an explicit Open', () => {
    const onChange = vi.fn<(next: TaskQuery) => void>();
    const { rerender } = render(<TaskFilters value={{ state: 'all' }} facets={facets} onChange={onChange} defaultState="all" />);
    expect(screen.queryByRole('group', { name: 'Active filters' })).toBeNull();
    rerender(<TaskFilters value={{ state: 'open' }} facets={facets} onChange={onChange} defaultState="all" />);
    const strip = screen.getByRole('group', { name: 'Active filters' });
    expect(strip.textContent).toContain('StateisOpen');
    fireEvent.click(within(strip).getByRole('button', { name: 'Remove state filter' }));
    expect(onChange).toHaveBeenLastCalledWith({ state: 'all' });
  });

  it('turns the Filter button into an icon with the count below 600 px', () => {
    stubWidth(390);
    render(<TaskFilters value={{ repo: ['octomation/tool'], label: ['scope: code'] }} facets={facets} onChange={() => {}} />);
    const trigger = screen.getByRole('button', { name: 'Filter, 2 active' });
    expect(trigger.textContent).toBe('2');
  });
});

/** Opens the menu again at another dimension, as a chip or the back row would. */
function openFilterAt(name: string) {
  const open = screen.queryByRole('group', { name: 'Filters' });
  if (open) fireEvent.click(screen.getByRole('button', { name: /^Filter/ }));
  return openDimension(name);
}
