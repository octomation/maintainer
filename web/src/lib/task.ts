import { useEffect, useState } from 'preact/hooks';

/**
 * The full issue page is addressed by `?task=<id>` in the URL of the inbox
 * or the board, beside the filters: a reload or a link opens it again, and
 * the browser's back returns to the list.
 */
export const TASK_PARAM = 'task';
const EVENT = 'maintainer:task';

/** The task of the full issue page in a URL search string, or null. */
export function taskFromSearch(search: string): string | null {
  return new URLSearchParams(search).get(TASK_PARAM) || null;
}

/** A close went back to the list, and its popstate has not come yet. */
let leaving = false;

/** The address of the list that openTask pushed the current entry over, or null. */
function listUnder(): string | null {
  const state: unknown = history.state;
  if (typeof state !== 'object' || state === null) return null;
  const { task, list } = state as { task?: unknown; list?: unknown };
  return typeof task === 'string' && typeof list === 'string' ? list : null;
}

/** Opens the full issue page of a task, or closes it with null; the filters stay. */
export function openTask(id: string | null): void {
  const address = (params: URLSearchParams) => `${location.pathname}${params.toString() ? `?${params}` : ''}`;
  const params = new URLSearchParams(location.search);
  params.delete(TASK_PARAM);
  // The list in the same spelling as the URLs below, so a close can compare them.
  const under = address(params);
  if (id) params.set(TASK_PARAM, id);
  const url = address(params);
  if (url === `${location.pathname}${location.search}`) return;
  const list = listUnder();
  if (!id && list === url) {
    // The list is the entry before: go back to it, popstate tells the page.
    // A second close before that popstate would go back past the list.
    if (leaving) return;
    leaving = true;
    addEventListener('popstate', () => (leaving = false), { once: true });
    history.back();
    return;
  }
  // Another issue over an open one takes its entry, so the way back stays one step.
  if (!id) history.replaceState(null, '', url);
  else if (taskFromSearch(location.search) === null) history.pushState({ task: id, list: under }, '', url);
  else history.replaceState(list === null ? null : { task: id, list }, '', url);
  dispatchEvent(new Event(EVENT));
}

/**
 * Closes the full issue page. When openTask pushed it over the list, and the
 * list is still the one under it, the close goes back to that entry, so the
 * history has no second copy of the list and the browser's forward opens the
 * issue again. Otherwise the page is replaced by the list in place: a page
 * opened from a link or a new tab has no list before it, and filters changed
 * over the page stay.
 */
export function closeTask(): void {
  openTask(null);
}

/** The task of the full issue page as state: it follows openTask and the browser's back and forward. */
export function useOpenTask(): string | null {
  const [task, setTask] = useState(() => (typeof location !== 'undefined' ? taskFromSearch(location.search) : null));
  useEffect(() => {
    const update = () => setTask(taskFromSearch(location.search));
    addEventListener(EVENT, update);
    addEventListener('popstate', update);
    return () => {
      removeEventListener(EVENT, update);
      removeEventListener('popstate', update);
    };
  }, []);
  return task;
}
