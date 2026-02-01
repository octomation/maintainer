import type { Column, StateFilter, TaskQuery, TaskSort } from '../api/types';
import { COLUMNS } from '../api/types';

const LISTS = ['owner', 'repo', 'milestone', 'label', 'type', 'assignee'] as const;
const SORTS: TaskSort[] = ['score', 'importance', 'updated', 'created'];
const STATES: StateFilter[] = ['open', 'closed', 'all'];

/** Reads the filters and the sort of a task query from a URL search string. */
export function queryFromSearch(search: string): TaskQuery {
  const params = new URLSearchParams(search);
  const query: TaskQuery = {};
  for (const name of LISTS) {
    const values = params.getAll(name).filter((v) => v !== '');
    if (values.length > 0) query[name] = values;
  }
  const columns = params.getAll('column').filter((v): v is Column => COLUMNS.includes(v as Column));
  if (columns.length > 0) query.column = columns;
  const state = params.get('state');
  if (state && STATES.includes(state as StateFilter)) query.state = state as StateFilter;
  const q = params.get('q');
  if (q) query.q = q;
  const sort = params.get('sort');
  if (sort && SORTS.includes(sort as TaskSort)) query.sort = sort as TaskSort;
  return query;
}

/** Writes the filters and the sort of a task query to a URL search string; paging stays out. */
export function searchFromQuery(query: TaskQuery): string {
  const params = new URLSearchParams();
  for (const name of LISTS) {
    for (const value of query[name] ?? []) params.append(name, value);
  }
  for (const value of query.column ?? []) params.append('column', value);
  if (query.state) params.set('state', query.state);
  if (query.q) params.set('q', query.q);
  if (query.sort) params.set('sort', query.sort);
  const s = params.toString();
  return s ? `?${s}` : '';
}

/** The parameters of a page's own view, which the filters keep: the open issue, the grouping, the layout. */
const VIEW_PARAMS = ['task', 'group', 'layout'];

/** Tells the shell that the search of the URL changed without a navigation. */
export const QUERY_EVENT = 'maintainer:query';

/** Replaces the search of the current URL without a navigation; the page's view parameters stay in it. */
export function pushQuery(query: TaskQuery): void {
  const current = new URLSearchParams(location.search);
  const params = new URLSearchParams(searchFromQuery(query));
  for (const name of VIEW_PARAMS) {
    const value = current.get(name);
    if (value) params.set(name, value);
  }
  replaceSearch(params);
}

/** Replaces the search of the current URL with these parameters and tells the shell. */
export function replaceSearch(params: URLSearchParams): void {
  const search = params.toString() ? `?${params.toString()}` : '';
  const url = `${location.pathname}${search}`;
  if (url === `${location.pathname}${location.search}`) return;
  history.replaceState(history.state, '', url);
  dispatchEvent(new Event(QUERY_EVENT));
}

/** Two queries select the same tasks in the same order. */
export function sameQuery(a: TaskQuery, b: TaskQuery): boolean {
  return searchFromQuery(a) === searchFromQuery(b);
}
