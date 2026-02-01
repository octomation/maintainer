import { Fragment } from 'preact';
import { useEffect, useMemo, useState } from 'preact/hooks';
import type { Catalog, Column, Effort, Facet, Facets, Impact, StateFilter, TaskQuery } from '../api/types';
import { COLUMN_TITLES, COLUMNS, EFFORTS, IMPACTS } from '../api/types';
import { TOUCH, useMediaQuery } from '../lib/media';
import { EFFORT_TITLE, IMPACT_TITLE, LabelChip } from './Chip';
import { ESTIMATE_SIZE, EstimateGlyph, PriorityGlyph, StatusGlyph } from './Glyphs';
import { Icon, type IconName } from './Icon';
import { Popover } from './Popover';

export interface TaskFiltersProps {
  /** The current filters; paging is ignored. */
  value: TaskQuery;
  /** The facets of the current selection; options and counts come from them. */
  facets: Facets | null;
  /** The catalog gives titles to milestones; optional. */
  catalog?: Catalog | null;
  onChange: (next: TaskQuery) => void;
  /** Offer the state dimension; a page that fixes the state hides it. */
  showState?: boolean;
  /** The page's default state: no chip for it, and removing the State chip returns to it. */
  defaultState?: StateFilter;
  /** Offer the board column as Status; the board itself has the columns. */
  showColumn?: boolean;
}

/** How long the search box waits before it sends `q`; the service answers in a few ms. */
export const SEARCH_DEBOUNCE_MS = 120;

const STATES: { value: StateFilter; label: string }[] = [
  { value: 'open', label: 'Open' },
  { value: 'closed', label: 'Closed' },
  { value: 'all', label: 'All' },
];

type ListName = 'owner' | 'repo' | 'milestone' | 'label' | 'type' | 'assignee';
export type Dimension = 'state' | 'column' | 'content' | 'priority' | 'estimate' | ListName;

const TITLE: Record<Dimension, string> = {
  state: 'State',
  column: 'Status',
  content: 'Content',
  priority: 'Priority',
  estimate: 'Estimate',
  label: 'Labels',
  type: 'Type',
  milestone: 'Milestone',
  repo: 'Repository',
  owner: 'Owner',
  assignee: 'Assignee',
};
/** The singular of a dimension in the names of its controls. */
const NOUN: Record<Dimension, string> = { ...TITLE, label: 'Label' } as Record<Dimension, string>;
const ICON: Record<Exclude<Dimension, 'state' | 'column' | 'priority' | 'estimate'>, IconName> = {
  content: 'search-solid',
  label: 'label',
  type: 'category-solid',
  milestone: 'milestone-solid',
  repo: 'repo-solid',
  owner: 'building-solid',
  assignee: 'user-solid',
};

/** The mark of a dimension: its own glyph where the hub draws one, else a heavy icon with about 14 px of artwork. */
function DimensionIcon({ d, size = 16 }: { d: Dimension; size?: number }) {
  if (d === 'state') return <StatusGlyph column="todo" size={size - 2} class="dimension-icon" />;
  if (d === 'column') return <StatusGlyph column="in_progress" size={size - 2} class="dimension-icon" />;
  if (d === 'priority') return <PriorityGlyph level="high" size={size} class="dimension-icon" />;
  if (d === 'estimate') return <EstimateGlyph size={size - 2} class="dimension-icon" />;
  return <Icon name={ICON[d]} size={size} tight={d !== 'label'} class="dimension-icon" />;
}
/** The dimensions in the groups of the menu. */
const GROUPS: Dimension[][] = [
  ['state', 'column', 'content'],
  ['priority', 'estimate', 'label', 'type'],
  ['milestone', 'repo', 'owner', 'assignee'],
];
const ORDER: Dimension[] = GROUPS.flat();

/** A repository as "name · owner": the distinguishing part first; the query keeps owner/name. */
export function repositoryTitle(value: string): string {
  const slash = value.indexOf('/');
  return slash > 0 ? `${value.slice(slash + 1)} · ${value.slice(0, slash)}` : value;
}

const PREFIX = { priority: 'impact', estimate: 'effort' } as const;
const TRIAGE = /^(impact|effort)\s*:/i;

/** The level of a triage label, such as `high` of `impact: high`. */
function levelOf(label: string, prefix: string): string {
  const lower = label.toLowerCase();
  return lower.startsWith(`${prefix}:`) ? lower.slice(prefix.length + 1).trim() : '';
}

export function hasFilters(value: TaskQuery): boolean {
  return Boolean(value.owner?.length || value.repo?.length || value.milestone?.length || value.label?.length || value.type?.length || value.assignee?.length || value.column?.length || value.q);
}

interface Option {
  value: string;
  title: string;
  count?: number;
  glyph?: preact.ComponentChildren;
}

/**
 * The filters of the inbox and the board in the top bar: the dashed Filter
 * button opens a menu of dimensions, a dimension opens its values with
 * counts from the facets; every active dimension becomes a chip next to the
 * button; the search field sits on the right. Values of one dimension are
 * any-of, labels are all-of, priority and estimate pick one level; the query
 * and its URL parameters do not change with the layout.
 */
export function TaskFilters({ value, facets, catalog, onChange, showState = true, defaultState = 'open', showColumn = false }: TaskFiltersProps) {
  const [q, setQ] = useState(value.q ?? '');
  useEffect(() => setQ(value.q ?? ''), [value.q]);
  useEffect(() => {
    if ((value.q ?? '') === q) return;
    const timer = setTimeout(() => onChange({ ...value, q: q || undefined }), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [q, value, onChange]);

  const touch = useMediaQuery(TOUCH);
  const [open, setOpen] = useState(false);
  const [dim, setDim] = useState<Dimension | null>(null);
  const [needle, setNeedle] = useState('');
  const dimensions = ORDER.filter((d) => (d === 'state' ? showState : d === 'column' ? showColumn : true));

  const milestoneTitle = (key: string) => catalog?.milestones.find((m) => m.key === key)?.title ?? key.replace(/^github:/, '');
  const labels = value.label ?? [];
  const state = value.state ?? defaultState;

  const setList = (name: ListName | 'column', values: string[]) => {
    const next: TaskQuery = { ...value };
    if (values.length === 0) delete next[name];
    else (next as Record<string, unknown>)[name] = values;
    onChange(next);
  };
  /** Priority and estimate are one label each: a level replaces the previous one. */
  const setLevel = (which: 'priority' | 'estimate', label: string | null) => {
    const prefix = PREFIX[which];
    const kept = labels.filter((l) => !levelOf(l, prefix));
    setList('label', label ? [...kept, label] : kept);
  };
  const clear = () => onChange({ state: value.state, sort: value.sort });

  /** The selected values of a dimension, as the query holds them. */
  const selected = (d: Dimension): string[] => {
    if (d === 'content') return value.q ? [value.q] : [];
    if (d === 'state') return showState && value.state && value.state !== defaultState ? [value.state] : [];
    if (d === 'priority' || d === 'estimate') return labels.filter((l) => levelOf(l, PREFIX[d]));
    if (d === 'label') return labels.filter((l) => !TRIAGE.test(l));
    if (d === 'column') return value.column ?? [];
    return value[d] ?? [];
  };

  const titleOf = (d: Dimension, v: string): string => {
    if (d === 'state') return STATES.find((s) => s.value === v)?.label ?? v;
    if (d === 'column') return COLUMN_TITLES[v as Column] ?? v;
    if (d === 'priority') return IMPACT_TITLE[levelOf(v, 'impact') as Impact] ?? v;
    if (d === 'estimate') {
      const effort = levelOf(v, 'effort') as Effort;
      return ESTIMATE_SIZE[effort] ? `${ESTIMATE_SIZE[effort]} (${EFFORT_TITLE[effort].toLowerCase()})` : v;
    }
    if (d === 'milestone') return milestoneTitle(v);
    if (d === 'repo') return repositoryTitle(v);
    if (d === 'content') return `“${v}”`;
    return v;
  };

  /** The options of a dimension: the facets, plus a selected value the facets no longer list. */
  const options = (d: Dimension): Option[] => {
    const withSelected = (facetList: Facet[] | undefined, chosen: string[]): Facet[] => {
      const known = new Set((facetList ?? []).map((f) => f.value));
      return [...(facetList ?? []), ...chosen.filter((v) => !known.has(v)).map((v) => ({ value: v, count: 0 }))];
    };
    switch (d) {
      case 'content':
        return [];
      case 'state':
        return STATES.map((s) => ({ value: s.value, title: s.label, count: s.value === 'all' ? undefined : facets ? (facets.states.find((f) => f.value === s.value)?.count ?? 0) : undefined }));
      case 'column':
        return COLUMNS.map((c) => ({ value: c, title: COLUMN_TITLES[c], count: facets?.columns.find((f) => f.value === c)?.count ?? 0, glyph: <StatusGlyph column={c} /> }));
      case 'priority':
      case 'estimate': {
        const prefix = PREFIX[d];
        const levels: string[] = d === 'priority' ? IMPACTS : EFFORTS;
        return levels.map((level) => {
          const facet = facets?.labels.find((f) => levelOf(f.value, prefix) === level);
          const chosen = labels.find((l) => levelOf(l, prefix) === level);
          const label = chosen ?? facet?.value ?? `${prefix}: ${level}`;
          return {
            value: label,
            title: d === 'priority' ? IMPACT_TITLE[level as Impact] : `${ESTIMATE_SIZE[level as Effort]} (${EFFORT_TITLE[level as Effort].toLowerCase()})`,
            count: facet?.count ?? 0,
            glyph: d === 'priority' ? <PriorityGlyph level={level as Impact} /> : <EstimateGlyph />,
          };
        });
      }
      case 'label':
        return withSelected(
          facets?.labels.filter((f) => !TRIAGE.test(f.value)),
          selected('label'),
        ).map((f) => ({ value: f.value, title: f.value, count: f.count }));
      case 'milestone':
        return withSelected(facets?.milestones, selected(d)).map((f) => ({ value: f.value, title: milestoneTitle(f.value), count: f.count }));
      case 'repo':
        return withSelected(facets?.repositories, selected(d)).map((f) => ({ value: f.value, title: repositoryTitle(f.value), count: f.count }));
      case 'owner':
        return withSelected(facets?.owners, selected(d)).map((f) => ({ value: f.value, title: f.value, count: f.count }));
      case 'type':
        return withSelected(facets?.types, selected(d)).map((f) => ({ value: f.value, title: f.value, count: f.count }));
      case 'assignee':
        return withSelected(facets?.assignees, selected(d)).map((f) => ({ value: f.value, title: f.value, count: f.count }));
    }
  };

  const toggle = (d: Dimension, v: string) => {
    if (d === 'state') {
      onChange({ ...value, state: v as StateFilter });
      return;
    }
    if (d === 'priority' || d === 'estimate') {
      setLevel(d, selected(d).includes(v) ? null : v);
      return;
    }
    const current = d === 'label' ? labels : d === 'column' ? (value.column ?? []) : (value[d as ListName] ?? []);
    setList(d === 'column' ? 'column' : (d as ListName), current.includes(v) ? current.filter((x) => x !== v) : [...current, v]);
  };

  const remove = (d: Dimension) => {
    if (d === 'content') {
      setQ('');
      onChange({ ...value, q: undefined });
    } else if (d === 'state') onChange({ ...value, state: defaultState });
    else if (d === 'priority' || d === 'estimate') setLevel(d, null);
    else if (d === 'label') setList('label', labels.filter((l) => TRIAGE.test(l)));
    else setList(d as ListName | 'column', []);
  };

  const active = dimensions.filter((d) => selected(d).length > 0);
  const operator = (d: Dimension, n: number) => (d === 'content' ? 'contains' : d === 'label' ? (n > 1 ? 'include all of' : 'includes') : n > 1 ? 'is any of' : 'is');

  const openAt = (d: Dimension | null) => {
    setDim(d);
    setNeedle('');
    setOpen(true);
  };
  const onOpenChange = (next: boolean) => {
    setOpen(next);
    if (!next) {
      setDim(null);
      setNeedle('');
    }
  };

  const shown = useMemo(() => dimensions.filter((d) => TITLE[d].toLowerCase().includes(needle.trim().toLowerCase())), [dimensions.join(), needle]);
  const count = active.length;

  // F opens the menu, as the shortcut badge in its search row says, unless a text field has the focus.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key !== 'f' && event.key !== 'F') return;
      if (event.metaKey || event.ctrlKey || event.altKey || event.defaultPrevented) return;
      const target = event.target as HTMLElement | null;
      if (target?.closest('input, textarea, select, [contenteditable="true"], [aria-modal="true"]')) return;
      event.preventDefault();
      openAt(null);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  return (
    <form class="filters" role="search" aria-label="Task filters" onSubmit={(event) => event.preventDefault()}>
      <Popover
        role="group"
        label="Filters"
        open={open}
        onOpenChange={onOpenChange}
        class="filters__menu"
        trigger={(props) => (
          <button type="button" {...props} class={`filter-button${touch ? ' filter-button--icon' : ''}`} aria-label={touch ? (count ? `Filter, ${count} active` : 'Filter') : undefined} title="Filter by state, priority, labels, milestone…" onClick={() => (open ? onOpenChange(false) : openAt(null))}>
            <Icon name="plus" size={12} />
            {!touch && 'Filter'}
            {touch && count > 0 && <span class="filter-button__count">{count}</span>}
          </button>
        )}
      >
        {() => (
          <div class="filter-menu" onKeyDown={(event) => arrows(event)}>
            <div class="menu-search">
              {dim && (
                <button type="button" class="icon-button menu-search__back" onClick={() => openAt(null)} aria-label="Back to the filters" title="Back">
                  <Icon name="arrow-left" size={14} />
                </button>
              )}
              {dim === 'content' ? (
                <input type="text" data-autofocus aria-label="Content contains" placeholder="Code, title, repository, label…" value={q} onInput={(event) => setQ((event.target as HTMLInputElement).value)} />
              ) : (
                <input type="text" data-autofocus aria-label={dim ? `Filter ${TITLE[dim].toLowerCase()}` : 'Filter dimensions'} placeholder={dim ? `${TITLE[dim]}…` : 'Filter…'} value={needle} onInput={(event) => setNeedle((event.target as HTMLInputElement).value)} />
              )}
              {!dim && (
                <kbd class="kbd" title="Press F to open the filters">
                  F
                </kbd>
              )}
            </div>
            {dim === null ? (
              <ul class="menu-list" aria-label="Dimensions">
                {GROUPS.map((group) => group.filter((d) => shown.includes(d))).filter((group) => group.length > 0).map((group, index) => (
                  <Fragment key={group[0]}>
                    {index > 0 && <li class="menu-separator" role="presentation" />}
                    {group.map((d) => (
                      <li key={d}>
                        <button type="button" class="menu-item" data-menu-item onClick={() => openAt(d)}>
                          <span class="menu-icon">
                            <DimensionIcon d={d} />
                          </span>
                          <span class="menu-item__text">{TITLE[d]}</span>
                          {selected(d).length > 0 && <span class="menu-item__hint">{d === 'content' ? '1' : selected(d).length}</span>}
                        </button>
                      </li>
                    ))}
                  </Fragment>
                ))}
                {shown.length === 0 && <li class="menu-empty">No such filter</li>}
              </ul>
            ) : dim === 'content' ? (
              <div class="menu-list" role="group" aria-label={TITLE[dim]}>
                <div class="menu-heading">Issues whose code, title, repository or a label contains the text.</div>
              </div>
            ) : (
              <div class="menu-list" role="group" aria-label={TITLE[dim]}>
                {dim === 'label' && <div class="menu-heading">Every checked label must match</div>}
                {options(dim)
                  .filter((o) => o.title.toLowerCase().includes(needle.trim().toLowerCase()))
                  .map((o) => {
                    const checked = dim === 'state' ? state === o.value : selected(dim).includes(o.value);
                    return (
                      <label class="menu-item menu-item--option" key={o.value}>
                        {dim === 'state' ? (
                          <input type="radio" name="filter-state" data-menu-item aria-label={o.title} checked={checked} onChange={() => toggle(dim, o.value)} />
                        ) : (
                          <input type="checkbox" data-menu-item aria-label={o.title} checked={checked} onChange={() => toggle(dim, o.value)} />
                        )}
                        {o.glyph && <span class="menu-icon">{o.glyph}</span>}
                        <span class="menu-item__text">{dim === 'label' ? <LabelChip label={o.title} compact /> : o.title}</span>
                        {o.count !== undefined && <span class="menu-item__hint">{o.count}</span>}
                      </label>
                    );
                  })}
                {options(dim).length === 0 && <div class="menu-empty">No {TITLE[dim].toLowerCase()} in the selection.</div>}
              </div>
            )}
          </div>
        )}
      </Popover>
      <div class="filters__chips" aria-label="Active filters" role={active.length ? 'group' : undefined} hidden={active.length === 0}>
        {active.map((d) => {
          const values = selected(d);
          const titles = values.map((v) => titleOf(d, v));
          const text = titles.length <= 2 ? titles.join(', ') : `${titles.length} ${TITLE[d].toLowerCase()}`;
          return (
            <span class="filter-chip" key={d}>
              <span class="filter-chip__dim">
                <DimensionIcon d={d} size={12} />
                {TITLE[d]}
              </span>
              <span class="filter-chip__op">{operator(d, values.length)}</span>
              <button type="button" class="filter-chip__value" onClick={() => openAt(d)} title={`${TITLE[d]}: ${titles.join(', ')}`}>
                {d === 'priority' && <PriorityGlyph level={levelOf(values[0], 'impact') as Impact} size={12} />}
                {d === 'column' && values.length === 1 && <StatusGlyph column={values[0] as Column} size={12} />}
                {text}
              </button>
              <button type="button" class="filter-chip__remove" aria-label={`Remove ${NOUN[d].toLowerCase()} filter`} title={`Remove ${NOUN[d].toLowerCase()} filter`} onClick={() => remove(d)}>
                <Icon name="x" size={12} />
              </button>
            </span>
          );
        })}
        {hasFilters(value) && (
          <button type="button" class="button button--ghost filters__clear" onClick={clear}>
            Clear
          </button>
        )}
      </div>
      <span class="filters__spacer" />
    </form>
  );
}

/** Arrow keys move between the dimension rows and the options. */
function arrows(event: KeyboardEvent) {
  if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return;
  const panel = event.currentTarget as HTMLElement;
  const items = Array.from(panel.querySelectorAll<HTMLElement>('input[data-autofocus], [data-menu-item]'));
  const index = items.indexOf(document.activeElement as HTMLElement);
  const next = items[(index + (event.key === 'ArrowDown' ? 1 : items.length - 1)) % items.length];
  if (next) {
    event.preventDefault();
    next.focus();
  }
}
