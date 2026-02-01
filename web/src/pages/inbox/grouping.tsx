import type { ComponentChildren } from 'preact';
import type { Catalog, Column, Effort, Facet, Facets, Impact, Task } from '../../api/types';
import { COLUMN_TITLES, COLUMNS, EFFORTS, IMPACTS } from '../../api/types';
import { EFFORT_TITLE, effortOf, IMPACT_TITLE, impactOf } from '../../components/Chip';
import { ESTIMATE_SIZE, EstimateGlyph, PriorityGlyph, RepoAvatar, StatusGlyph } from '../../components/Glyphs';
import { Icon } from '../../components/Icon';

/** How the inbox groups its rows; `group` in the URL. */
export type TaskGrouping = 'none' | 'status' | 'priority' | 'estimate' | 'repository' | 'owner' | 'milestone' | 'type';

export const GROUPINGS: { value: TaskGrouping; label: string }[] = [
  { value: 'none', label: 'No grouping' },
  { value: 'status', label: 'Status' },
  { value: 'priority', label: 'Priority' },
  { value: 'estimate', label: 'Estimate' },
  { value: 'repository', label: 'Repository' },
  { value: 'owner', label: 'Owner' },
  { value: 'milestone', label: 'Milestone' },
  { value: 'type', label: 'Type' },
];

/** The grouping of a URL search string; none when it names no known one. */
export function groupingFromSearch(search: string): TaskGrouping {
  const value = new URLSearchParams(search).get('group');
  return GROUPINGS.some((g) => g.value === value && value !== 'none') ? (value as TaskGrouping) : 'none';
}

export interface TaskGroup {
  key: string;
  title: string;
  glyph: ComponentChildren;
  /** The total of the group in the selection, from the facets; the rows shown may be fewer until the list is loaded. */
  count: number;
  rows: Task[];
}

const NONE = '';

/** The value of a task in a grouping; '' when it has none. */
function valueOf(task: Task, grouping: TaskGrouping): string {
  switch (grouping) {
    case 'status':
      return task.column;
    case 'priority':
      return impactOf(task.labels);
    case 'estimate':
      return effortOf(task.labels);
    case 'repository':
      return task.repository;
    case 'owner':
      return task.owner;
    case 'milestone':
      return task.milestone?.key ?? NONE;
    case 'type':
      return task.type ?? NONE;
    default:
      return 'all';
  }
}

/** The level of a triage label, such as `high` of `impact: high`. */
function levelOf(label: string, prefix: string): string {
  const lower = label.toLowerCase();
  return lower.startsWith(`${prefix}:`) ? lower.slice(prefix.length + 1).trim() : '';
}

/**
 * The values of a grouping in their order with their totals in the
 * selection: the workflow order for status, the levels for priority and
 * estimate, the facets' order for the rest; the group without a value last,
 * with what the others leave of the total.
 */
function valuesOf(grouping: TaskGrouping, facets: Facets | null, total: number): { value: string; count: number }[] {
  const facet = (list: Facet[] | undefined, value: string) => list?.find((f) => f.value === value)?.count ?? 0;
  const withRest = (values: { value: string; count: number }[]) => {
    const rest = Math.max(0, total - values.reduce((sum, v) => sum + v.count, 0));
    return [...values, { value: NONE, count: rest }];
  };
  switch (grouping) {
    case 'status':
      return COLUMNS.map((c) => ({ value: c, count: facet(facets?.columns, c) }));
    case 'priority':
    case 'estimate': {
      const prefix = grouping === 'priority' ? 'impact' : 'effort';
      const levels: string[] = grouping === 'priority' ? IMPACTS : EFFORTS;
      return withRest(levels.map((level) => ({ value: level, count: (facets?.labels ?? []).filter((f) => levelOf(f.value, prefix) === level).reduce((sum, f) => sum + f.count, 0) })));
    }
    case 'repository':
      return (facets?.repositories ?? []).map((f) => ({ value: f.value, count: f.count }));
    case 'owner':
      return (facets?.owners ?? []).map((f) => ({ value: f.value, count: f.count }));
    case 'milestone':
      return withRest((facets?.milestones ?? []).map((f) => ({ value: f.value, count: f.count })));
    case 'type':
      return withRest((facets?.types ?? []).map((f) => ({ value: f.value, count: f.count })));
    default:
      return [{ value: 'all', count: total }];
  }
}

function present(grouping: TaskGrouping, value: string, sample: Task | undefined, catalog: Catalog | null): { title: string; glyph: ComponentChildren } {
  switch (grouping) {
    case 'status':
      return { title: COLUMN_TITLES[value as Column] ?? value, glyph: <StatusGlyph column={value as Column} /> };
    case 'priority':
      return value ? { title: IMPACT_TITLE[value as Impact], glyph: <PriorityGlyph level={value as Impact} size={14} /> } : { title: 'No priority', glyph: <PriorityGlyph level="none" size={14} /> };
    case 'estimate':
      return value ? { title: `${ESTIMATE_SIZE[value as Effort]} · ${EFFORT_TITLE[value as Effort].toLowerCase()}`, glyph: <EstimateGlyph /> } : { title: 'No estimate', glyph: <EstimateGlyph class="glyph--empty" /> };
    case 'repository':
      return { title: value, glyph: <RepoAvatar repository={value} size={14} /> };
    case 'owner':
      return { title: value, glyph: <Icon name="building-solid" size={14} tight /> };
    case 'milestone':
      return value
        ? { title: sample?.milestone?.title ?? catalog?.milestones.find((m) => m.key === value)?.title ?? value.replace(/^github:/, ''), glyph: <Icon name="milestone-solid" size={14} tight /> }
        : { title: 'No milestone', glyph: <Icon name="milestone" size={14} /> };
    case 'type':
      return value ? { title: value, glyph: <Icon name="category-solid" size={14} tight /> } : { title: 'No type', glyph: <Icon name="category" size={14} /> };
    default:
      return { title: '', glyph: null };
  }
}

/**
 * The loaded rows in groups: the groups in the order of their values, each
 * with its total from the facets and its rows in the list's order; a group
 * shows only when it has loaded rows. A value the facets do not list (rare,
 * as for a task of another selection) gets a group after the listed ones,
 * counted from its rows.
 */
export function groupTasks(tasks: Task[], grouping: TaskGrouping, facets: Facets | null, total: number, catalog: Catalog | null = null): TaskGroup[] {
  if (grouping === 'none') return [{ key: 'all', title: '', glyph: null, count: total, rows: tasks }];
  const rows = new Map<string, Task[]>();
  for (const task of tasks) {
    const value = valueOf(task, grouping);
    rows.set(value, [...(rows.get(value) ?? []), task]);
  }
  const ordered = valuesOf(grouping, facets, total);
  const known = new Set(ordered.map((v) => v.value));
  const extra = Array.from(rows.keys())
    .filter((value) => !known.has(value))
    .map((value) => ({ value, count: rows.get(value)?.length ?? 0 }));
  const none = ordered.filter((v) => v.value === NONE);
  const values = [...ordered.filter((v) => v.value !== NONE), ...extra, ...none];
  return values
    .filter((v) => (rows.get(v.value)?.length ?? 0) > 0)
    .map((v) => {
      const list = rows.get(v.value) ?? [];
      const { title, glyph } = present(grouping, v.value, list[0], catalog);
      return { key: `${grouping}:${v.value || 'none'}`, title, glyph, count: Math.max(v.count, list.length), rows: list };
    });
}
