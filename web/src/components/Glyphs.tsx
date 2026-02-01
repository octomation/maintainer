import type { Column, Effort, Impact } from '../api/types';

/**
 * The glyphs of the hub, drawn for it: workflow statuses for the board
 * columns, priorities for `impact: *`, estimates for `effort: *`, the
 * groups of the checkout list and the maintainer mark. Geometric shapes on
 * a 14 or 16 px grid in currentColor or a decoration token; a glyph is
 * decorative unless it gets a title.
 */

interface GlyphProps {
  /** The accessible name and tooltip; without it the glyph is decorative. */
  title?: string;
  size?: number;
  class?: string;
}

function a11y(title?: string) {
  return title ? { role: 'img' as const, 'aria-label': title } : { 'aria-hidden': 'true' as const };
}

/**
 * The outline geometry of a status glyph: a 14-unit drawing with a 1.5 px
 * outline, or at 12 px and below a 12-unit drawing with a 2 px outline
 * (r=5), as the source's row circles. Filled shapes scale from the 14-unit
 * drawing.
 */
function geometry(size: number) {
  return size <= 12
    ? { unit: 12, c: 6, r: 5, stroke: 2, inner: 2.5, dash: '1.6 1.54', scale: 'scale(0.857143)' }
    : { unit: 14, c: 7, r: 6, stroke: 1.5, inner: 3.5, dash: '1.4 1.74', scale: undefined };
}

/** The workflow status of a board column. */
export function StatusGlyph({ column, title, size = 14, class: cls }: GlyphProps & { column: Column }) {
  const g = geometry(size);
  const common = { class: `glyph glyph--status glyph--${column}${cls ? ` ${cls}` : ''}`, width: size, height: size, viewBox: `0 0 ${g.unit} ${g.unit}`, focusable: 'false' as const, ...a11y(title) };
  const ring = (dashed = false) => <circle cx={g.c} cy={g.c} r={g.r} fill="none" stroke="currentColor" stroke-width={g.stroke} stroke-dasharray={dashed ? g.dash : undefined} />;
  switch (column) {
    case 'backlog':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          {ring(true)}
        </svg>
      );
    case 'todo':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          {ring()}
        </svg>
      );
    case 'in_progress':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          {ring()}
          <path d={`M${g.c} ${g.c - g.inner}a${g.inner} ${g.inner} 0 0 1 0 ${2 * g.inner}z`} fill="currentColor" />
        </svg>
      );
    case 'done':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          <g transform={g.scale}>
            <circle cx="7" cy="7" r="7" fill="currentColor" />
            <path d="M4.2 7.2l1.9 1.9l3.8 -3.9" fill="none" stroke="var(--glyph-knockout)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
          </g>
        </svg>
      );
    case 'wont_do':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          <g transform={g.scale}>
            <circle cx="7" cy="7" r="7" fill="currentColor" />
            <path d="M4.8 4.8l4.4 4.4m0 -4.4l-4.4 4.4" fill="none" stroke="var(--glyph-knockout)" stroke-width="1.5" stroke-linecap="round" />
          </g>
        </svg>
      );
  }
}

export type PriorityLevel = Impact | 'none' | 'urgent';

const BARS: Record<Impact, number> = { high: 3, medium: 2, low: 1 };

/** Three bars for high, two for medium, one for low; three dots for none. */
export function PriorityGlyph({ level, title, size = 16, class: cls }: GlyphProps & { level: PriorityLevel }) {
  const common = { class: `glyph glyph--priority glyph--priority-${level}${cls ? ` ${cls}` : ''}`, width: size, height: size, viewBox: '0 0 16 16', focusable: 'false' as const, ...a11y(title) };
  if (level === 'none') {
    return (
      <svg {...common}>
        {title && <title>{title}</title>}
        <circle cx="3" cy="8" r="1.25" fill="currentColor" />
        <circle cx="8" cy="8" r="1.25" fill="currentColor" />
        <circle cx="13" cy="8" r="1.25" fill="currentColor" />
      </svg>
    );
  }
  if (level === 'urgent') {
    return (
      <svg {...common}>
        {title && <title>{title}</title>}
        <rect x="1.5" y="1.5" width="13" height="13" rx="3" fill="currentColor" />
        <path d="M8 4.5v4.5" stroke="var(--glyph-knockout)" stroke-width="1.75" stroke-linecap="round" />
        <circle cx="8" cy="11.5" r="1" fill="var(--glyph-knockout)" />
      </svg>
    );
  }
  const active = BARS[level];
  const bars = [
    { x: 1.5, y: 9, h: 5 },
    { x: 6.5, y: 5.5, h: 8.5 },
    { x: 11.5, y: 2, h: 12 },
  ];
  return (
    <svg {...common}>
      {title && <title>{title}</title>}
      {bars.map((bar, i) => (
        <rect key={i} x={bar.x} y={bar.y} width="3" height={bar.h} rx="1" fill="currentColor" opacity={i < active ? 1 : 0.35} />
      ))}
    </svg>
  );
}

/** The estimate glyph: three dots on a triangle, the points of a size. */
export function EstimateGlyph({ title, size = 14, class: cls }: GlyphProps) {
  return (
    <svg class={`glyph glyph--estimate${cls ? ` ${cls}` : ''}`} width={size} height={size} viewBox="0 0 14 14" focusable="false" {...a11y(title)}>
      {title && <title>{title}</title>}
      <circle cx="7" cy="3.5" r="1.6" fill="currentColor" />
      <circle cx="3.5" cy="10" r="1.6" fill="currentColor" />
      <circle cx="10.5" cy="10" r="1.6" fill="currentColor" />
    </svg>
  );
}

export const ESTIMATE_SIZE: Record<Effort, string> = { easy: 'S', medium: 'M', hard: 'L' };

/** The groups of the checkout list. */
export type CheckoutGroup = 'error' | 'uncommitted' | 'diverged' | 'orphan' | 'clean';

export function CheckoutGlyph({ group, title, size = 14, class: cls }: GlyphProps & { group: CheckoutGroup }) {
  const g = geometry(size);
  const common = { class: `glyph glyph--checkout glyph--${group}${cls ? ` ${cls}` : ''}`, width: size, height: size, viewBox: `0 0 ${g.unit} ${g.unit}`, focusable: 'false' as const, ...a11y(title) };
  const ring = (dashed = false) => <circle cx={g.c} cy={g.c} r={g.r} fill="none" stroke="currentColor" stroke-width={g.stroke} stroke-dasharray={dashed ? g.dash : undefined} />;
  switch (group) {
    case 'error':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          <g transform={g.scale}>
            <circle cx="7" cy="7" r="7" fill="currentColor" />
            <path d="M7 3.6v4" stroke="var(--glyph-knockout)" stroke-width="1.6" stroke-linecap="round" />
            <circle cx="7" cy="10.2" r=".95" fill="var(--glyph-knockout)" />
          </g>
        </svg>
      );
    case 'uncommitted':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          {ring()}
          <path d={`M${g.c} ${g.c - g.inner}a${g.inner} ${g.inner} 0 1 1 -${g.inner} ${g.inner}h${g.inner}z`} fill="currentColor" />
        </svg>
      );
    case 'diverged':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          {ring()}
          <g transform={g.scale}>
            <path d="M4.8 6l2.2 -2.2l2.2 2.2m-4.4 2l2.2 2.2l2.2 -2.2" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" />
          </g>
        </svg>
      );
    case 'orphan':
      return (
        <svg {...common}>
          {title && <title>{title}</title>}
          {ring(true)}
          <path d={`M${g.c - 2} ${g.c}h4`} stroke="currentColor" stroke-width={g.stroke} stroke-linecap="round" />
        </svg>
      );
    case 'clean':
      return <StatusGlyph column="done" title={title} size={size} class={`glyph--clean${cls ? ` ${cls}` : ''}`} />;
  }
}

/** A small padlock, drawn for the hub: a private repository. */
export function LockGlyph({ title, size = 12, class: cls }: GlyphProps) {
  return (
    <svg class={`glyph glyph--lock${cls ? ` ${cls}` : ''}`} width={size} height={size} viewBox="0 0 12 12" focusable="false" {...a11y(title)}>
      {title && <title>{title}</title>}
      <path d="M3.75 5.25v-1.5a2.25 2.25 0 0 1 4.5 0v1.5" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" />
      <rect x="2.25" y="5.25" width="7.5" height="5.25" rx="1.25" fill="currentColor" />
    </svg>
  );
}

/** The maintainer mark of the favicon: the only brand element. */
export function BrandMark({ size = 18 }: { size?: number }) {
  return (
    <svg class="brand-mark" width={size} height={size} viewBox="0 0 16 16" aria-hidden="true" focusable="false">
      <rect width="16" height="16" rx="3" fill="var(--brand)" />
      <path d="M4 12V4.5l4 4 4-4V12" fill="none" stroke="#fff" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
    </svg>
  );
}

/** The decoration palette, in the order the hash picks from. */
export const DECORATION = ['red', 'orange', 'yellow', 'green', 'blue', 'purple', 'teal'] as const;
export type Decoration = (typeof DECORATION)[number];

/** A stable color for a name: the same label or repository always gets the same one. */
export function decorationOf(name: string): Decoration {
  let hash = 5381;
  for (let i = 0; i < name.length; i++) hash = ((hash << 5) + hash + name.charCodeAt(i)) | 0;
  return DECORATION[Math.abs(hash) % DECORATION.length];
}

/** A repository as a circle with its initial, tinted by the hash of its name; an optional badge sits at its bottom right. */
export function RepoAvatar({ repository, size = 36, badge }: { repository: string; size?: number; badge?: preact.ComponentChildren }) {
  const name = repository.slice(repository.indexOf('/') + 1) || repository;
  return (
    <span class="repo-avatar" style={{ '--deco': `var(--${decorationOf(repository)})`, width: `${size}px`, height: `${size}px`, fontSize: `${Math.max(9, Math.round(size * 0.38))}px` }} aria-hidden={badge ? undefined : 'true'}>
      <span aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
      {badge && <span class="repo-avatar__badge">{badge}</span>}
    </span>
  );
}
