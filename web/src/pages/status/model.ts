// The browser side of the status view: the sort chain the page sends, and
// the lines of the TUI header. The order of the rows, the cells and the
// counts come from the server; nothing here sorts or filters.

import type { CheckoutRow, CheckoutSummary, SortSpec } from '../../api/types';
import type { CheckoutGroup } from '../../components/Glyphs';
import { CHECKOUT_COLUMN_TITLES, type CheckoutColumn } from '../../api/types';

/**
 * Mirrors cycleSort of internal/service/status: a column cycles
 * absent → ascending → descending → absent. A plain click replaces the chain
 * with that column; a shift-click edits the column in place, as Shift+s does
 * in the TUI, so the precedence of the other keys stays.
 */
export function cycleSort(keys: SortSpec[], column: string, additive: boolean): SortSpec[] {
  const index = keys.findIndex((key) => key.column === column);
  const next: SortSpec = { column, desc: index >= 0 ? !keys[index].desc : false };
  const remove = index >= 0 && keys[index].desc;
  if (!additive) return remove ? [] : [next];
  const result = keys.slice();
  if (index < 0) result.push(next);
  else if (remove) result.splice(index, 1);
  else result[index] = next;
  return result;
}

/** The `sort` parameter of GET /api/v0/checkouts: a comma list, `-` for descending. */
export function sortParam(keys: SortSpec[]): string {
  return keys.map((key) => (key.desc ? '-' : '') + key.column).join(',');
}

function plural(n: number, noun: string): string {
  return n === 1 ? `1 ${noun}` : `${n} ${noun}s`;
}

/** The header of `maintainer status --format=plain` after "Repository status · ". */
export function summaryLine(summary: CheckoutSummary): string {
  return [
    plural(summary.checkouts, 'checkout'),
    `${summary.uncommitted} uncommitted`,
    `${summary.ahead} ahead`,
    `${summary.behind} behind`,
    `${summary.locked} locked`,
    `${summary.orphan} orphan`,
    plural(summary.errors, 'error'),
    'local refs (no fetch)',
  ].join(' · ');
}

/** The "Sort:" line of the TUI header. */
export function sortSummary(keys: SortSpec[], query: string): string {
  if (keys.length === 0) return query.trim() !== '' ? 'Sort: relevance' : 'Sort: default (repository / path)';
  return (
    'Sort: ' +
    keys
      .map((key, i) => {
        const title = CHECKOUT_COLUMN_TITLES[key.column as CheckoutColumn] ?? key.column;
        return `${title} ${key.desc ? '↓' : '↑'}${i + 1}`;
      })
      .join('  →  ')
  );
}

/** An RFC 3339 time as `YYYY-MM-DD HH:MM:SSZ` in UTC; the input when it does not parse. */
export function formatTime(value: string | null | undefined): string {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return `${date.toISOString().slice(0, 19).replace('T', ' ')}Z`;
}

/** How the list groups the rows; the order inside a group stays the server's. */
export type Grouping = 'state' | 'owner' | 'none';

/** The state of a checkout as the list groups it, worst first. */
export function groupOf(row: CheckoutRow): CheckoutGroup {
  if (row.error) return 'error';
  if (row.orphan_reason || row.placeholder) return 'orphan';
  if (row.added + row.deleted + row.changed_files + row.untracked + row.binary + row.conflicts > 0) return 'uncommitted';
  if (row.ahead > 0 || row.behind > 0 || /ahead|behind|diverged|unborn/.test(row.status)) return 'diverged';
  return 'clean';
}

export const CHECKOUT_GROUPS: CheckoutGroup[] = ['error', 'uncommitted', 'diverged', 'orphan', 'clean'];
export const CHECKOUT_GROUP_TITLES: Record<CheckoutGroup, string> = {
  error: 'Errors',
  uncommitted: 'Uncommitted',
  diverged: 'Ahead or behind',
  orphan: 'Orphans',
  clean: 'Clean',
};

export interface RowGroup {
  key: string;
  /** The header; none when the list is not grouped. */
  title: string;
  glyph?: CheckoutGroup;
  rows: CheckoutRow[];
}

/**
 * The rows in groups: by state in the order of CHECKOUT_GROUPS, by owner in
 * the order the owners first appear; empty groups are left out and every
 * group keeps the server's order of its rows.
 */
export function groupRows(rows: CheckoutRow[], grouping: Grouping): RowGroup[] {
  if (grouping === 'none') return [{ key: 'all', title: '', rows }];
  if (grouping === 'owner') {
    const groups = new Map<string, CheckoutRow[]>();
    for (const row of rows) {
      const slash = row.repository.indexOf('/');
      const owner = slash > 0 ? row.repository.slice(0, slash) : 'No owner';
      groups.set(owner, [...(groups.get(owner) ?? []), row]);
    }
    return Array.from(groups, ([owner, list]) => ({ key: `owner-${owner}`, title: owner, rows: list }));
  }
  return CHECKOUT_GROUPS.map((group) => ({ key: group, title: CHECKOUT_GROUP_TITLES[group], glyph: group, rows: rows.filter((row) => groupOf(row) === group) })).filter((group) => group.rows.length > 0);
}

/** The status page as a list or a board of the checkouts. */
export type Layout = 'list' | 'board';

/** The layout of a URL search string: the board with layout=board, the list otherwise. */
export function layoutFromSearch(search: string): Layout {
  return new URLSearchParams(search).get('layout') === 'board' ? 'board' : 'list';
}

/** The property columns of a checkout row that View can hide; the repository always shows. */
export type StatusColumn = 'branch' | 'changes' | 'status' | 'path';
export type StatusColumns = Record<StatusColumn, boolean>;
export const STATUS_COLUMNS: { key: StatusColumn; title: string }[] = [
  { key: 'branch', title: 'Branch' },
  { key: 'changes', title: 'Changes' },
  { key: 'status', title: 'Status' },
  { key: 'path', title: 'Path' },
];
const COLUMNS_KEY = 'maintainer.status.columns';
const ALL_COLUMNS: StatusColumns = { branch: true, changes: true, status: true, path: true };

/** The visible columns, remembered in the browser; all of them when nothing is stored. */
export function readColumns(): StatusColumns {
  try {
    const stored = JSON.parse(localStorage.getItem(COLUMNS_KEY) ?? 'null') as Partial<StatusColumns> | null;
    const columns = { ...ALL_COLUMNS };
    for (const { key } of STATUS_COLUMNS) if (typeof stored?.[key] === 'boolean') columns[key] = stored[key] as boolean;
    return columns;
  } catch {
    return { ...ALL_COLUMNS };
  }
}

export function writeColumns(columns: StatusColumns): void {
  try {
    localStorage.setItem(COLUMNS_KEY, JSON.stringify(columns));
  } catch {
    // storage is a convenience
  }
}

/** The width of a text in pixels: measured in the page's font where a canvas can, estimated from its length otherwise. */
let context: CanvasRenderingContext2D | null | undefined;
export function textWidth(text: string, font: string, perChar: number): number {
  if (context === undefined) {
    try {
      context = typeof document !== 'undefined' ? document.createElement('canvas').getContext('2d') : null;
    } catch {
      context = null;
    }
  }
  if (context && typeof context.measureText === 'function') {
    context.font = font;
    const width = context.measureText(text).width;
    if (width > 0) return width;
  }
  return text.length * perChar;
}

const clamp = (value: number, min: number, max: number) => Math.round(Math.min(max, Math.max(min, value)));

/**
 * The grid of a checkout row: the leading slots, then each visible property
 * as a column as wide as its longest value (within bounds), the path in the
 * rest. One template for the whole list, so the columns line up from row to
 * row and from group to group.
 */
export function columnTemplate(rows: CheckoutRow[], columns: StatusColumns): string {
  const longest = (index: number, font: string, perChar: number) => rows.reduce((max, row) => Math.max(max, textWidth(row.cells[index] ?? '', font, perChar)), 0);
  const ui = "'Inter Variable', Inter, sans-serif";
  const tracks = ['16px', '42px', '14px', `${clamp(longest(0, `400 13px ${ui}`, 7.4) + 2, 120, 300)}px`];
  if (columns.branch) tracks.push(`${clamp(longest(1, `400 12px ${ui}`, 6.8) + 2, 64, 180)}px`);
  if (columns.changes) tracks.push(`${clamp(longest(2, `400 12px ${ui}`, 7) + 2, 72, 220)}px`);
  if (columns.status) tracks.push(`${clamp(longest(3, `400 12px ${ui}`, 6.8) + 2, 56, 200)}px`);
  tracks.push('minmax(0, 1fr)');
  return tracks.join(' ');
}
