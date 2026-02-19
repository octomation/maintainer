import { afterEach, describe, expect, it, vi } from 'vitest';
import { closeTask, openTask } from './task';

describe('openTask and closeTask', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    history.replaceState(null, '', '/inbox');
  });

  it('keeps one entry for the issues opened over the list and goes back to the list', () => {
    history.replaceState(null, '', '/before');
    history.pushState(null, '', '/inbox?state=all');
    const length = history.length;
    openTask('id-1');
    openTask('id-2');
    expect(location.search).toBe('?state=all&task=id-2');
    expect(history.length).toBe(length + 1);
    closeTask();
    expect(`${location.pathname}${location.search}`).toBe('/inbox?state=all');
    history.forward();
    expect(location.search).toBe('?state=all&task=id-2');
  });

  it('goes back once for two closes before the popstate', () => {
    history.pushState(null, '', '/inbox');
    openTask('id-1');
    // A browser traverses the history after the call; happy-dom does it at once.
    const back = vi.spyOn(history, 'back').mockImplementation(() => {});
    closeTask();
    closeTask();
    expect(back).toHaveBeenCalledTimes(1);
    dispatchEvent(new PopStateEvent('popstate'));
    closeTask();
    expect(back).toHaveBeenCalledTimes(2);
    dispatchEvent(new PopStateEvent('popstate'));
  });

  it('keeps the filters changed over the issue: the close replaces it with the new list', () => {
    history.replaceState(null, '', '/before');
    history.pushState(null, '', '/inbox?label=type%3A%20bug');
    const length = history.length;
    openTask('id-1');
    // A list in another spelling of the same query is still the list under the issue.
    history.replaceState(history.state, '', '/inbox?label=type%3A+bug&task=id-1');
    closeTask();
    expect(location.search).toBe('?label=type%3A%20bug');
    expect(history.length).toBe(length + 1);
    openTask('id-1');
    // The filters change over the issue, as pushQuery does it: the entry keeps its state.
    history.replaceState(history.state, '', '/inbox?state=all&task=id-1');
    closeTask();
    expect(location.search).toBe('?state=all');
    history.back();
    expect(location.search).toBe('?label=type%3A%20bug');
  });

  it('replaces an issue opened from a link with the list in place', () => {
    history.replaceState(null, '', '/inbox?task=id-1');
    const length = history.length;
    closeTask();
    expect(location.search).toBe('');
    expect(history.length).toBe(length);
  });
});
