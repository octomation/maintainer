import type { ComponentChildren } from 'preact';
import type { Effort, Impact } from '../api/types';
import { EFFORTS, IMPACTS } from '../api/types';
import { decorationOf, ESTIMATE_SIZE, EstimateGlyph, PriorityGlyph } from './Glyphs';

export const IMPACT_TITLE: Record<Impact, string> = { high: 'High', medium: 'Medium', low: 'Low' };
export const EFFORT_TITLE: Record<Effort, string> = { easy: 'Easy', medium: 'Medium', hard: 'Hard' };

export interface MetaChipProps {
  children: ComponentChildren;
  tone?: 'neutral' | 'flag';
  title?: string;
  label?: string;
}

/** A small pill for meta such as a milestone or a flag; flags carry the orange dot. */
export function MetaChip({ children, tone = 'neutral', title, label }: MetaChipProps) {
  return (
    <span class={tone === 'flag' ? 'meta-chip meta-chip--flag' : 'meta-chip'} title={title} aria-label={label}>
      {children}
    </span>
  );
}

/** A label as a pill with a colored dot; the color is a stable hash of the label. */
export function LabelChip({ label, compact = false }: { label: string; compact?: boolean }) {
  return (
    <span class={compact ? 'label-chip label-chip--compact' : 'label-chip'} title={label}>
      <span class="label-chip__dot" style={{ '--dot': `var(--${decorationOf(label)})` }} aria-hidden="true" />
      <span class="label-chip__text">{label}</span>
    </span>
  );
}

/** Every label of a task. */
export function LabelChips({ labels }: { labels: string[] }) {
  if (labels.length === 0) return null;
  return (
    <span class="label-chips">
      {labels.map((label) => (
        <LabelChip key={label} label={label} />
      ))}
    </span>
  );
}

const TRIAGE = /^(impact|effort)\s*:/i;

/** The value of the single `prefix: value` label, or '' when there is none or more than one. */
export function labelValue(labels: string[], prefix: string): string {
  let value = '';
  for (const label of labels) {
    const lower = label.toLowerCase();
    if (lower.startsWith(`${prefix}:`)) {
      if (value) return '';
      value = lower.slice(prefix.length + 1).trim();
    }
  }
  return value;
}

/** The impact of a task as a priority; '' without a known one. */
export function impactOf(labels: string[]): Impact | '' {
  const value = labelValue(labels, 'impact');
  return IMPACTS.includes(value as Impact) ? (value as Impact) : '';
}

/** The effort of a task as an estimate; '' without a known one. */
export function effortOf(labels: string[]): Effort | '' {
  const value = labelValue(labels, 'effort');
  return EFFORTS.includes(value as Effort) ? (value as Effort) : '';
}

/** The labels that are neither priority nor estimate. */
export function otherLabels(labels: string[]): string[] {
  return labels.filter((label) => !TRIAGE.test(label));
}

export function priorityTitle(impact: Impact | ''): string {
  return impact ? `Priority: ${IMPACT_TITLE[impact]}` : 'No priority';
}

/** The priority glyph of a task, named for assistive technology. */
export function Priority({ labels, showNone = false }: { labels: string[]; showNone?: boolean }) {
  const impact = impactOf(labels);
  if (!impact && !showNone) return null;
  return (
    <span class="priority" title={priorityTitle(impact)}>
      <PriorityGlyph level={impact || 'none'} title={priorityTitle(impact)} />
    </span>
  );
}

/** The estimate of a task: the glyph and S, M or L. */
export function Estimate({ labels }: { labels: string[] }) {
  const effort = effortOf(labels);
  if (!effort) return null;
  const name = `Estimate: ${ESTIMATE_SIZE[effort]} (${EFFORT_TITLE[effort].toLowerCase()})`;
  return (
    <span class="estimate" title={name} aria-label={name} role="img">
      <EstimateGlyph />
      <span aria-hidden="true">{ESTIMATE_SIZE[effort]}</span>
    </span>
  );
}

/**
 * The meta of a row or a card: the priority glyph, the estimate and the
 * other labels as chips, up to `max` of them and `+N` for the rest, which
 * the issue page lists in full.
 */
export function TaskMeta({ labels, max = 2, showNone = false, priority = true, estimate = true, others = true }: { labels: string[]; max?: number; showNone?: boolean; priority?: boolean; estimate?: boolean; others?: boolean }) {
  const rest = others ? otherLabels(labels) : [];
  const shown = rest.slice(0, max);
  const hidden = rest.slice(max);
  return (
    <span class="task-meta">
      {priority && <Priority labels={labels} showNone={showNone} />}
      {estimate && <Estimate labels={labels} />}
      {shown.map((label) => (
        <LabelChip key={label} label={label} compact />
      ))}
      {hidden.length > 0 && (
        <span class="meta-chip" title={hidden.join(', ')} aria-label={`${hidden.length} more labels: ${hidden.join(', ')}`}>
          +{hidden.length}
        </span>
      )}
    </span>
  );
}
