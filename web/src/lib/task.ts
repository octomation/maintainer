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

/** Opens the full issue page of a task, or closes it with null; the filters stay. */
export function openTask(id: string | null): void {
  const params = new URLSearchParams(location.search);
  if (id) params.set(TASK_PARAM, id);
  else params.delete(TASK_PARAM);
  const search = params.toString();
  const url = `${location.pathname}${search ? `?${search}` : ''}`;
  if (url === `${location.pathname}${location.search}`) return;
  if (id) history.pushState({ task: id }, '', url);
  else history.replaceState(null, '', url);
  dispatchEvent(new Event(EVENT));
}

/**
 * Closes the full issue page in place: the entry that opened it stays in the
 * history, so the browser's forward opens the issue again.
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
