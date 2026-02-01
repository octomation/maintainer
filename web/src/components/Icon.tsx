/**
 * Inline icons: the Tabler Icons (MIT) paths the hub needs, 16 px, a 1.5 px
 * effective stroke at any size (2.25 units of the 24-unit grid at 16 px),
 * currentColor. An icon is decorative unless given a title; a control
 * that shows only an icon names itself with aria-label. The status,
 * priority and estimate glyphs are not icons: they live in Glyphs.tsx.
 */
const PATHS = {
  search: ['M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0', 'M21 21l-6 -6'],
  adjustments: [
    'M14 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0',
    'M4 6l8 0',
    'M16 6l4 0',
    'M8 12m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0',
    'M4 12l2 0',
    'M10 12l10 0',
    'M17 18m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0',
    'M4 18l11 0',
    'M19 18l1 0',
  ],
  refresh: ['M20 11a8.1 8.1 0 0 0 -15.5 -2m-.5 -4v4h4', 'M4 13a8.1 8.1 0 0 0 15.5 2m.5 4v-4h-4'],
  sun: ['M12 12m-4 0a4 4 0 1 0 8 0a4 4 0 1 0 -8 0', 'M3 12h1m8 -9v1m8 8h1m-9 8v1m-6.4 -15.4l.7 .7m12.1 -.7l-.7 .7m0 11.4l.7 .7m-12.1 -.7l-.7 .7'],
  moon: ['M12 3c.132 0 .263 0 .393 0a7.5 7.5 0 0 0 7.92 12.446a9 9 0 1 1 -8.313 -12.454z'],
  desktop: ['M3 5a1 1 0 0 1 1 -1h16a1 1 0 0 1 1 1v10a1 1 0 0 1 -1 1h-16a1 1 0 0 1 -1 -1v-10z', 'M7 20h10', 'M9 16v4', 'M15 16v4'],
  lock: ['M5 13a2 2 0 0 1 2 -2h10a2 2 0 0 1 2 2v6a2 2 0 0 1 -2 2h-10a2 2 0 0 1 -2 -2v-6z', 'M11 16a1 1 0 1 0 2 0a1 1 0 0 0 -2 0', 'M8 11v-4a4 4 0 1 1 8 0v4'],
  'external-link': ['M12 6h-6a2 2 0 0 0 -2 2v10a2 2 0 0 0 2 2h10a2 2 0 0 0 2 -2v-6', 'M11 13l9 -9', 'M15 4h5v5'],
  x: ['M18 6l-12 12', 'M6 6l12 12'],
  'chevron-down': ['M6 9l6 6l6 -6'],
  'chevron-right': ['M9 6l6 6l-6 6'],
  'chevron-left': ['M15 6l-6 6l6 6'],
  'arrow-up': ['M12 5l0 14', 'M18 11l-6 -6', 'M6 11l6 -6'],
  'arrow-down': ['M12 5l0 14', 'M18 13l-6 6', 'M6 13l6 6'],
  'arrow-left': ['M5 12l14 0', 'M5 12l6 6', 'M5 12l6 -6'],
  flag: ['M5 5a5 5 0 0 1 7 0a5 5 0 0 0 7 0v9a5 5 0 0 1 -7 0a5 5 0 0 0 -7 0v-9z', 'M5 21v-7'],
  check: ['M5 12l5 5l10 -10'],
  plus: ['M12 5l0 14', 'M5 12l14 0'],
  menu: ['M4 6l16 0', 'M4 12l16 0', 'M4 18l16 0'],
  inbox: ['M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z', 'M4 13h3l3 3h4l3 -3h3'],
  board: ['M4 4l6 0', 'M14 4l6 0', 'M4 8m0 2a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v8a2 2 0 0 1 -2 2h-2a2 2 0 0 1 -2 -2z', 'M14 8m0 2a2 2 0 0 1 2 -2h2a2 2 0 0 1 2 2v2a2 2 0 0 1 -2 2h-2a2 2 0 0 1 -2 -2z'],
  list: ['M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v2a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z', 'M4 14m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v2a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z'],
  'git-branch': ['M7 18m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0', 'M7 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0', 'M17 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0', 'M7 8l0 8', 'M9 18h6a2 2 0 0 0 2 -2v-5', 'M14 14l3 -3l3 3'],
  'git-commit': ['M12 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0', 'M12 3l0 6', 'M12 15l0 6'],
  tag: [
    'M7.5 7.5m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0',
    'M3 6v5.172a2 2 0 0 0 .586 1.414l7.71 7.71a2.41 2.41 0 0 0 3.408 0l5.592 -5.592a2.41 2.41 0 0 0 0 -3.408l-7.71 -7.71a2 2 0 0 0 -1.414 -.586h-5.172a3 3 0 0 0 -3 3z',
  ],
  user: ['M8 7a4 4 0 1 0 8 0a4 4 0 0 0 -8 0', 'M6 21v-2a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v2'],
  users: ['M9 7m-4 0a4 4 0 1 0 8 0a4 4 0 1 0 -8 0', 'M3 21v-2a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v2', 'M16 3.13a4 4 0 0 1 0 7.75', 'M21 21v-2a4 4 0 0 0 -3 -3.85'],
  building: ['M3 21l18 0', 'M9 8l1 0', 'M9 12l1 0', 'M9 16l1 0', 'M14 8l1 0', 'M14 12l1 0', 'M14 16l1 0', 'M5 21v-16a2 2 0 0 1 2 -2h10a2 2 0 0 1 2 2v16'],
  category: ['M4 4h6v6h-6z', 'M14 4h6v6h-6z', 'M4 14h6v6h-6z', 'M17 17m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0'],
  repo: ['M19 4v16h-12a2 2 0 0 1 -2 -2v-12a2 2 0 0 1 2 -2h12z', 'M19 16h-12a2 2 0 0 0 -2 2', 'M9 8h6'],
  milestone: ['M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0', 'M12 12m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0', 'M12 12m-5 0a5 5 0 1 0 10 0a5 5 0 1 0 -10 0'],
  'alert-circle': ['M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0', 'M12 8v4', 'M12 16h.01'],
  clock: ['M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0', 'M12 7v5l3 3'],
  state: ['M12 12m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0', 'M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0'],
  folder: ['M5 4h4l3 3h7a2 2 0 0 1 2 2v8a2 2 0 0 1 -2 2h-14a2 2 0 0 1 -2 -2v-11a2 2 0 0 1 2 -2'],
  chart: [
    'M3 13a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v6a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z',
    'M15 9a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v10a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z',
    'M9 5a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v14a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1z',
    'M4 20h14',
  ],
  message: ['M8 9h8', 'M8 13h6', 'M18 4a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-5l-5 3v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12z'],
  history: ['M12 8l0 4l2 2', 'M3.05 11a9 9 0 1 1 .5 4m-.5 5v-5h5'],
  sort: ['M4 6l9 0', 'M4 12l7 0', 'M4 18l7 0', 'M15 15l3 3l3 -3', 'M18 6l0 12'],
  selector: ['M8 9l4 -4l4 4', 'M16 15l-4 4l-4 -4'],
  expand: ['M16 4l4 0l0 4', 'M14 10l6 -6', 'M8 20l-4 0l0 -4', 'M4 20l6 -6'],
  copy: [
    'M7 7m0 2.667a2.667 2.667 0 0 1 2.667 -2.667h8.666a2.667 2.667 0 0 1 2.667 2.667v8.666a2.667 2.667 0 0 1 -2.667 2.667h-8.666a2.667 2.667 0 0 1 -2.667 -2.667z',
    'M4.012 16.737a2.005 2.005 0 0 1 -1.012 -1.737v-10c0 -1.1 .9 -2 2 -2h10c.75 0 1.158 .385 1.5 1',
  ],
  link: ['M9 15l6 -6', 'M11 6l.463 -.536a5 5 0 0 1 7.071 7.072l-.534 .464', 'M13 18l-.397 .534a5.068 5.068 0 0 1 -7.127 0a4.972 4.972 0 0 1 0 -7.071l.524 -.463'],
  file: ['M14 3v4a1 1 0 0 0 1 1h4', 'M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z', 'M9 9l1 0', 'M9 13l6 0', 'M9 17l6 0'],
  hash: ['M5 9l14 0', 'M5 15l14 0', 'M11 4l-4 16', 'M17 4l-4 16'],
  enter: ['M18 6v6a3 3 0 0 1 -3 3h-10l4 -4m0 8l-4 -4'],
  // Three 3 px dots at x=3, 8 and 13 of a 16 px box.
  dots: ['F:M4.5 9.75a2.25 2.25 0 1 1 0 4.5a2.25 2.25 0 0 1 0 -4.5z', 'F:M12 9.75a2.25 2.25 0 1 1 0 4.5a2.25 2.25 0 0 1 0 -4.5z', 'F:M19.5 9.75a2.25 2.25 0 1 1 0 4.5a2.25 2.25 0 0 1 0 -4.5z'],
  // A list: four 2 px bands across about 14×14 px of a 16 px box (drawn for the hub).
  'list-solid': [
    'F:M2.25 1.5h19.5a.75 .75 0 0 1 .75 .75v1.5a.75 .75 0 0 1 -.75 .75h-19.5a.75 .75 0 0 1 -.75 -.75v-1.5a.75 .75 0 0 1 .75 -.75z',
    'F:M2.25 7.5h19.5a.75 .75 0 0 1 .75 .75v1.5a.75 .75 0 0 1 -.75 .75h-19.5a.75 .75 0 0 1 -.75 -.75v-1.5a.75 .75 0 0 1 .75 -.75z',
    'F:M2.25 13.5h19.5a.75 .75 0 0 1 .75 .75v1.5a.75 .75 0 0 1 -.75 .75h-19.5a.75 .75 0 0 1 -.75 -.75v-1.5a.75 .75 0 0 1 .75 -.75z',
    'F:M2.25 19.5h19.5a.75 .75 0 0 1 .75 .75v1.5a.75 .75 0 0 1 -.75 .75h-19.5a.75 .75 0 0 1 -.75 -.75v-1.5a.75 .75 0 0 1 .75 -.75z',
  ],
  // Heavier forms for navigation and menus: filled Tabler shapes, or outlines with a filled interior.
  'inbox-solid': ['M4 4m0 2a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z', 'F:M4 13h3l3 3h4l3 -3h3v5a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2z'],
  'board-solid': [
    'F:M10 3a1 1 0 0 1 0 2h-6a1 1 0 1 1 0 -2z',
    'F:M20 3a1 1 0 0 1 0 2h-6a1 1 0 0 1 0 -2z',
    'F:M8 7a3 3 0 0 1 3 3v8a3 3 0 0 1 -3 3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3z',
    'F:M18 7a3 3 0 0 1 3 3v2a3 3 0 0 1 -3 3h-2a3 3 0 0 1 -3 -3v-2a3 3 0 0 1 3 -3z',
  ],
  'branch-solid': ['F:M7 15.5a2.5 2.5 0 1 1 0 5a2.5 2.5 0 0 1 0 -5z', 'F:M7 3.5a2.5 2.5 0 1 1 0 5a2.5 2.5 0 0 1 0 -5z', 'F:M17 3.5a2.5 2.5 0 1 1 0 5a2.5 2.5 0 0 1 0 -5z', 'M7 8l0 8', 'M9 18h6a2 2 0 0 0 2 -2v-5', 'M14 14l3 -3l3 3'],
  // A label: a small rounded tag of about 12×9 px at 16 px, the source's mark is that light.
  label: ['F:M4 5.5h11.6a1.5 1.5 0 0 1 1.1 .5l4.3 5a1.5 1.5 0 0 1 0 2l-4.3 5a1.5 1.5 0 0 1 -1.1 .5h-11.6a1.5 1.5 0 0 1 -1.5 -1.5v-10a1.5 1.5 0 0 1 1.5 -1.5z'],
  'tag-solid': ['F:M11.172 2a3 3 0 0 1 2.121 .879l7.71 7.71a3.41 3.41 0 0 1 0 4.822l-5.592 5.592a3.41 3.41 0 0 1 -4.822 0l-7.71 -7.71a3 3 0 0 1 -.879 -2.121v-5.172a4 4 0 0 1 4 -4zm-3.672 3.5a2 2 0 0 0 -1.995 1.85l-.005 .15a2 2 0 1 0 2 -2'],
  'user-solid': ['F:M12 2a5 5 0 1 1 -5 5l.005 -.217a5 5 0 0 1 4.995 -4.783z', 'F:M14 14a5 5 0 0 1 5 5v1a2 2 0 0 1 -2 2h-10a2 2 0 0 1 -2 -2v-1a5 5 0 0 1 5 -5h4z'],
  'category-solid': [
    'F:M10 3h-6a1 1 0 0 0 -1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1 -1v-6a1 1 0 0 0 -1 -1z',
    'F:M20 3h-6a1 1 0 0 0 -1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1 -1v-6a1 1 0 0 0 -1 -1z',
    'F:M10 13h-6a1 1 0 0 0 -1 1v6a1 1 0 0 0 1 1h6a1 1 0 0 0 1 -1v-6a1 1 0 0 0 -1 -1z',
    'F:M17 13a4 4 0 1 1 -3.995 4.2l-.005 -.2l.005 -.2a4 4 0 0 1 3.995 -3.8z',
  ],
  'repo-solid': ['F:M9 3a1 1 0 0 1 .608 .206l.1 .087l2.706 2.707h6.586a3 3 0 0 1 2.995 2.824l.005 .176v8a3 3 0 0 1 -2.824 2.995l-.176 .005h-14a3 3 0 0 1 -2.995 -2.824l-.005 -.176v-11a3 3 0 0 1 2.824 -2.995l.176 -.005h4z'],
  'milestone-solid': ['M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0', 'F:M12 7a5 5 0 1 1 0 10a5 5 0 0 1 0 -10z'],
  'search-solid': ['F:M10 5a5 5 0 1 1 0 10a5 5 0 0 1 0 -10z', 'M10 10m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0', 'M21 21l-6 -6'],
  'building-solid': ['M3 21l18 0', 'F:M7 3h10a2 2 0 0 1 2 2v16h-14v-16a2 2 0 0 1 2 -2z'],
} as const;

export type IconName = keyof typeof PATHS;

export interface IconProps {
  name: IconName;
  /** A visible tooltip and the accessible name; without it the icon is decorative. */
  title?: string;
  class?: string;
  size?: number;
  /** Crop the 24-unit grid to its 20-unit artwork, so the icon fills about 14 of 16 px. */
  tight?: boolean;
}

/** A path that starts with F: is filled instead of stroked. */
const FILLED = 'F:';

export function Icon({ name, title, class: cls, size = 16, tight = false }: IconProps) {
  const grid = tight ? 20 : 24;
  return (
    <svg
      class={cls ? `icon ${cls}` : 'icon'}
      width={size}
      height={size}
      viewBox={tight ? '2 2 20 20' : '0 0 24 24'}
      fill="none"
      stroke="currentColor"
      stroke-width={(1.5 * grid) / size}
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden={title ? undefined : 'true'}
      role={title ? 'img' : undefined}
      aria-label={title}
      focusable="false"
    >
      {title && <title>{title}</title>}
      {PATHS[name].map((d: string, i: number) => (d.startsWith(FILLED) ? <path key={i} d={d.slice(FILLED.length)} fill="currentColor" stroke="none" /> : <path key={i} d={d} />))}
    </svg>
  );
}
