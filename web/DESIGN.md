---
version: alpha
name: maintainer-hub
description: "The design system of the maintainer hub, the local pages at maintainer.localhost. A visual copy of a dense issue-tracker application: a 220 px sidebar with an action row, the pages and the repositories, a 56 px top bar with the title, a dashed Filter button with its chips, the list/board switch and Display, then grouped 44 px list rows, a five-column board, an inbox and a full issue page with a 340 px properties rail, and a command bar. Dark is the default theme, light is its derived counterpart with the same structure; Inter at 13 px is the interface size; one indigo primary; status, priority and estimate glyphs carry meaning; the maintainer mark in the sidebar is the only brand element."

colors:
  primary: "#5E6AD2"
  primary-hover: "#575BC7"
  on-primary: "#FFFFFF"
  link: "#8C94EA"
  focus-ring: "#5E6AD2"
  bg-app: "#181921"
  bg-sidebar: "#191A23"
  bg-topbar: "#191A23"
  bg-group: "#21232E"
  bg-row-hover: "#1C1D2A"
  bg-row-selected: "#25263A"
  bg-item-selected: "#2F303A"
  bg-item-hover: "#24252F"
  bg-button: "#292A35"
  bg-card: "#1F2130"
  bg-menu: "#1D1E2BDB"
  bg-rail: "#191A23"
  bg-label-chip: "#1A1B26"
  bg-menu-item-hover: "#292B41"
  bg-chip: "#1F2130"
  bg-input: "#1D1E2B"
  border: "#2C2D3C"
  border-divider: "#1C1D2B"
  border-button: "#313248"
  border-button-hover: "#393A4B"
  border-menu: "#00000000"
  border-activity: "#22232E"
  border-rail: "#212332"
  border-rail-toolbar: "#292B41"
  border-card: "#23242F"
  border-label-chip: "#232432"
  separator-menu: "#272935"
  separator-command: "#292A39"
  bg-command-chip: "#2A2B38"
  text-command-chip: "#8E8DA9"
  bg-command-kbd: "#2A2B39"
  separator-command-group: "#2C2D3C80"
  text-primary: "#EEEFFC"
  text-secondary: "#D2D3E0"
  text-muted: "#858699"
  text-muted-raised: "#9293A6"
  text-faint: "#4C4F6B"
  icon-menu: "#8988A2"
  red: "#EB5757"
  orange: "#F2994A"
  yellow: "#F2C94C"
  green: "#4CB782"
  blue: "#4EA7FC"
  purple: "#BB87FC"
  teal: "#00B2BF"
  brand: "#4D7C0F"
  overlay: "#0B0C1280"
  light-bg-app: "#FFFFFF"
  light-bg-sidebar: "#F7F7F9"
  light-bg-topbar: "#FFFFFF"
  light-bg-group: "#F4F4F7"
  light-bg-row-hover: "#F7F7F9"
  light-bg-row-selected: "#EEEFF8"
  light-bg-item-selected: "#E9E9EE"
  light-bg-item-hover: "#EFEFF3"
  light-bg-button: "#FFFFFF"
  light-bg-card: "#FFFFFF"
  light-bg-menu: "#FFFFFFE6"
  light-bg-rail: "#F7F7F9"
  light-bg-label-chip: "#FFFFFF"
  light-bg-menu-item-hover: "#F0F0F5"
  light-bg-chip: "#FFFFFF"
  light-bg-input: "#FFFFFF"
  light-border: "#E6E6EC"
  light-border-divider: "#EFEFF3"
  light-border-button: "#DCDDE4"
  light-border-button-hover: "#C9CAD4"
  light-border-menu: "#E3E4EA"
  light-border-activity: "#EDEDF1"
  light-border-rail: "#E6E6EC"
  light-border-card: "#ECECF1"
  light-border-label-chip: "#E6E6EC"
  light-text-primary: "#1C1D27"
  light-text-secondary: "#3B3C4A"
  light-text-muted: "#6B6F82"
  light-text-muted-raised: "#6B6F82"
  light-text-faint: "#B4B6C4"
  light-icon-menu: "#7A7D90"
  light-link: "#4E5BC5"
  light-red: "#D23B3B"
  light-orange: "#E07B24"
  light-yellow: "#E2B33C"
  light-green: "#2F9E6A"
  light-blue: "#2D8BE0"
  light-purple: "#9A5FE6"
  light-teal: "#00939E"
  light-overlay: "#1C1D2733"

typography:
  title2:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: 500
    lineHeight: 1.25
    letterSpacing: 0px
    fontFeature: "\"cv01\", \"ss03\""
  normal:
    fontFamily: Inter
    fontSize: 15px
    fontWeight: 400
    lineHeight: 22px
    fontFeature: "\"cv01\", \"ss03\""
  normal-medium:
    fontFamily: Inter
    fontSize: 15px
    fontWeight: 500
    lineHeight: 22px
    fontFeature: "\"cv01\", \"ss03\""
  small:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1
    fontFeature: "\"cv01\", \"ss03\""
  small-medium:
    fontFamily: Inter
    fontSize: 13px
    fontWeight: 500
    lineHeight: 1
    fontFeature: "\"cv01\", \"ss03\""
  mini:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1
    fontFeature: "\"cv01\", \"ss03\""
  mini-medium:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 500
    lineHeight: 1
    fontFeature: "\"cv01\", \"ss03\""
  micro:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: 400
    lineHeight: 1.1
    fontFeature: "\"cv01\", \"ss03\""
  number:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1
    fontFeature: "\"tnum\", \"cv01\""
  mono:
    fontFamily: Mono
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1
  prose:
    fontFamily: Inter
    fontSize: 15px
    fontWeight: 400
    lineHeight: 22px
    fontFeature: "\"cv01\", \"ss03\""

rounded:
  none: 0px
  xs: 3px
  sm: 4px
  md: 6px
  lg: 8px
  xl: 16px
  full: 9999px

spacing:
  xxs: 2px
  xs: 4px
  sm: 8px
  gap-row: 9px
  md: 12px
  lg: 16px
  xl: 24px
  xxl: 36px
  inset-issue: 60px
  sidebar: 220px
  topbar: 56px
  group-header: 36px
  row: 44px
  sidebar-item: 27px
  button: 24px
  menu-item: 32px
  properties: 340px
  prop-row: 52px
  rail-toolbar: 56px
  issue-header: 48px
  issue-inset: 60px
  issue-inset-narrow: 24px
  issue-content: 832px
  sidebar-action: 28px
  board-column: 320px
  board-gap: 28px
  inbox-avatar: 36px
  command-bar: 642px
  inbox-list: 420px
  touch: 44px

components:
  sidebar:
    backgroundColor: "{colors.bg-sidebar}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small-medium}"
    padding: 12px 16px
    width: 220px
  sidebar-brand:
    backgroundColor: "{colors.bg-sidebar}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 0 6px
    height: 28px
  sidebar-section-header:
    backgroundColor: "{colors.bg-sidebar}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    padding: 0 8px
    height: 28px
  sidebar-item:
    backgroundColor: "{colors.bg-sidebar}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 5.5px 2px 5.5px 7px
    height: 27px
  sidebar-item-hover:
    backgroundColor: "{colors.bg-item-hover}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 5.5px 2px 5.5px 7px
    height: 27px
  sidebar-item-selected:
    backgroundColor: "{colors.bg-item-selected}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 5.5px 2px 5.5px 7px
    height: 27px
  sidebar-footer-item:
    backgroundColor: "{colors.bg-sidebar}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    rounded: "{rounded.sm}"
    padding: 0 8px
    height: 28px
  top-bar:
    backgroundColor: "{colors.bg-topbar}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    padding: 14px 24px 14px 30px
    height: 56px
  button:
    backgroundColor: "{colors.bg-button}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.mini-medium}"
    rounded: "{rounded.sm}"
    padding: 0 8px
    height: 24px
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.mini-medium}"
    rounded: "{rounded.sm}"
    padding: 0 10px
    height: 24px
  button-ghost:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini-medium}"
    rounded: "{rounded.sm}"
    padding: 0 6px
    height: 24px
  filter-button:
    backgroundColor: "{colors.bg-topbar}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.mini-medium}"
    rounded: "{rounded.sm}"
    padding: 0 8px 0 6px
    height: 24px
  filter-chip:
    backgroundColor: "{colors.bg-chip}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.mini}"
    rounded: "{rounded.sm}"
    padding: 0 6px
    height: 24px
  view-switch:
    backgroundColor: "{colors.bg-group}"
    textColor: "{colors.text-muted-raised}"
    typography: "{typography.mini}"
    rounded: "{rounded.sm}"
    padding: 2px
    width: 64px
    height: 24px
  display-button:
    backgroundColor: "{colors.bg-button}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.mini-medium}"
    rounded: "{rounded.sm}"
    padding: 0 6px
    width: 80px
    height: 24px
  sidebar-action:
    backgroundColor: "{colors.bg-button}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 0 8px
    height: 28px
  sidebar-nested-item:
    backgroundColor: "{colors.bg-sidebar}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 5.5px 2px 5.5px 24px
    height: 27px
  command-bar:
    backgroundColor: "{colors.bg-menu}"
    textColor: "{colors.text-primary}"
    typography: "{typography.normal}"
    rounded: "{rounded.xl}"
    padding: 0 20px
    width: 642px
  kbd:
    backgroundColor: "{colors.bg-button}"
    textColor: "{colors.text-muted-raised}"
    typography: "{typography.micro}"
    rounded: "{rounded.xs}"
    size: 16px
  menu:
    backgroundColor: "{colors.bg-menu}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small}"
    rounded: "{rounded.lg}"
    padding: 4px
    width: 210px
  menu-item:
    backgroundColor: "{colors.bg-menu}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small}"
    rounded: "{rounded.sm}"
    padding: 0 7px
    height: 32px
  menu-item-active:
    backgroundColor: "{colors.bg-menu-item-hover}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    rounded: "{rounded.sm}"
    padding: 0 8px
    height: 32px
  menu-search:
    backgroundColor: "{colors.bg-menu}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    padding: 0 11px
    height: 36px
  display-menu:
    backgroundColor: "{colors.bg-menu}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    rounded: "{rounded.lg}"
    padding: 12px 16px
    width: 340px
  group-header:
    backgroundColor: "{colors.bg-group}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    padding: 0 24px 0 36px
    height: 36px
  list-row:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    padding: 11px 29px 11px 36px
    height: 44px
  list-row-hover:
    backgroundColor: "{colors.bg-row-hover}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    padding: 11px 29px 11px 36px
    height: 44px
  list-row-selected:
    backgroundColor: "{colors.bg-row-selected}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    padding: 11px 29px 11px 36px
    height: 44px
  issue-id:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-muted}"
    typography: "{typography.small}"
    width: 42px
  meta-chip:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    rounded: "{rounded.full}"
    padding: 0 8px
    height: 22px
  label-chip:
    backgroundColor: "{colors.bg-label-chip}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    rounded: "{rounded.full}"
    padding: 0 10px 0 9px
    height: 28px
  label-chip-compact:
    backgroundColor: "{colors.bg-label-chip}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    rounded: "{rounded.full}"
    padding: 0 8px 0 7px
    height: 22px
  estimate:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    rounded: "{rounded.sm}"
    padding: 0 6px 0 4px
    height: 22px
  board-column:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    padding: 0
    width: 320px
  board-column-header:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    padding: 0 8px
    height: 36px
  card:
    backgroundColor: "{colors.bg-card}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.lg}"
    padding: 16px
  card-hover:
    backgroundColor: "{colors.bg-row-selected}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.lg}"
    padding: 16px
  inbox-row:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.md}"
    padding: 10px 12px
    height: 56px
  inbox-row-selected:
    backgroundColor: "{colors.bg-item-selected}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.md}"
    padding: 10px 12px
    height: 56px
  repo-avatar:
    backgroundColor: "{colors.bg-button}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.full}"
    size: 36px
  issue-header:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small}"
    padding: 0 24px 0 60px
    height: 48px
  issue-title:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.title2}"
    padding: 0
  issue-prose:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.prose}"
    padding: 0
  activity-item:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-muted}"
    typography: "{typography.mini}"
    padding: 0
  properties-panel:
    backgroundColor: "{colors.bg-rail}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.small-medium}"
    padding: 26px 24px
    width: 340px
  rail-toolbar:
    backgroundColor: "{colors.bg-rail}"
    textColor: "{colors.text-muted}"
    typography: "{typography.small-medium}"
    padding: 0 24px
    height: 56px
  property-label:
    backgroundColor: "{colors.bg-rail}"
    textColor: "{colors.text-muted}"
    typography: "{typography.small-medium}"
    width: 103px
  property-value:
    backgroundColor: "{colors.bg-rail}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small-medium}"
    rounded: "{rounded.sm}"
    padding: 0 6px
    height: 28px
  peek:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    rounded: "{rounded.lg}"
    padding: 0
    width: 960px
  tooltip:
    backgroundColor: "{colors.bg-menu}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.mini}"
    rounded: "{rounded.sm}"
    padding: 4px 8px
  notice:
    backgroundColor: "{colors.bg-card}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.mini}"
    rounded: "{rounded.md}"
    padding: 10px 12px
  empty-state:
    backgroundColor: "{colors.bg-app}"
    textColor: "{colors.text-muted}"
    typography: "{typography.small}"
    padding: 48px 24px
  skeleton:
    backgroundColor: "{colors.bg-group}"
    textColor: "{colors.bg-group}"
    rounded: "{rounded.sm}"
    height: 10px
  progress-line:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary}"
    height: 2px
  brand-mark:
    backgroundColor: "{colors.brand}"
    textColor: "{colors.on-primary}"
    rounded: "{rounded.xs}"
    size: 18px
  light-surface:
    backgroundColor: "{colors.light-bg-app}"
    textColor: "{colors.light-text-primary}"
    typography: "{typography.small}"
  light-sidebar:
    backgroundColor: "{colors.light-bg-sidebar}"
    textColor: "{colors.light-text-secondary}"
    typography: "{typography.small-medium}"
  light-group-header:
    backgroundColor: "{colors.light-bg-group}"
    textColor: "{colors.light-text-muted}"
    typography: "{typography.small-medium}"
  light-menu:
    backgroundColor: "{colors.light-bg-menu}"
    textColor: "{colors.light-text-secondary}"
    typography: "{typography.small}"
---

## Overview

The hub is a tool the owner opens many times a day: an inbox of issues, a board of the same issues and a list of the local checkouts. It copies the visual language of a dense issue-tracker application, measured from a Figma community design system of that application (the owner's copy, dark mode, file `V3svHx6VW7ar7pQRNbA7QK`) and from screenshots of the application's issue board and inbox. A screenshot of a hub page next to the reference differs in data and in the logo only, with three accepted feature differences: no favorites star (the hub has no favorites), a Reindex action in the place of "New issue", and repositories in the place of teams.

Sources are named next to the values: **F:** a Figma node, **R:** a reference image under `~/Development/research/linear/app-reference/` (outside the repository), **D:** derived here. No icon, logo, avatar or image of the reference enters the repository: the glyphs are our own geometric SVGs, the other icons are Tabler (MIT).

**Key characteristics**

- Dark by default (`{colors.bg-app}` #181921, F `84:3817`); the light theme is derived (D): white and near-white surfaces, the same structure, the same primary, darker decoration colors for contrast on white.
- A 220 px sidebar, a 56 px top bar, a 36 px group header right under it, 44 px list rows (F `47:1331`, `84:3817`, `60:1776`); the issue page with an 830 px content column and a 340 px properties rail (F `96:6792`).
- Inter (OFL), bundled with the build, at 13 px for the interface, 12 px for meta and buttons, 15/22 px for prose, 24 px for the issue title (F `11:24`); metadata is Inter with tabular numerals, monospace only for paths and commits in the properties.
- One primary, indigo `{colors.primary}`, for the done glyph, focus, checked states and links (F `11:23`); decoration colors for glyphs and label dots only. The maintainer mark (`{colors.brand}`) is the only brand element.
- Floating surfaces (menus, the command bar) are a tonal gradient over a backdrop blur, with a quiet edge and a deep shadow (F `84:3840`).
- Icons have a 1.5 px effective stroke and a 13–14 px footprint in a 16 px box.

## Colors

### Surfaces (dark, F `11:23`, `84:3817`, `96:6792`)

- **App** `{colors.bg-app}` #181921: content, rows, the article. **Sidebar** and **top bar** `{colors.bg-sidebar}` / `{colors.bg-topbar}` #191A23. **Properties rail** `{colors.bg-rail}` #191A23 with a `{colors.border-rail}` #212332 left edge and a `{colors.border-rail-toolbar}` #292B41 line under its toolbar (sampled from `figma/issue-page.png`). **Group header** `{colors.bg-group}` #21232E.
- **Row hover** `{colors.bg-row-hover}` #1C1D2A (R `figma/ds-issues.png`), **row selected** `{colors.bg-row-selected}` #25263A (D), **sidebar item selected** `{colors.bg-item-selected}` #2F303A (rgba(133,134,152,.2) on the sidebar, F `47:1331`), hover `{colors.bg-item-hover}` (D).
- **Button** `{colors.bg-button}` #292A35 with `{colors.border-button}` #313248, hover border `{colors.border-button-hover}` #393A4B (F `14:347`). **Card** `{colors.bg-card}` #1F2130 with a quiet `{colors.border-card}` #23242F edge: cards separate by surface, not by outline (R `ref/board-issues.png`). **Label chip** `{colors.bg-label-chip}` #1A1B26 with `{colors.border-label-chip}` #232432.
- **Menu** `{colors.bg-menu}`: a 135° gradient from #1E202B to #1A1C26 at 65 %, both at 92 % opacity (a property menu such as Priority starts at #1F202A at 94 %, the command-surface family), over a 16 px backdrop blur, which reproduces the lighter upper-left and darker lower region sampled in `figma/issues-filter.png` and `figma/command-bar.png` (the command bar from #1F202A); `{colors.bg-menu}` is the flat fallback; edge `{colors.border-menu}` #23242F, separators `{colors.separator-menu}` #272935, item hover `{colors.bg-menu-item-hover}` #292B41.
- **Borders**: `{colors.border}` #2C2D3C for the sidebar's right edge and the top bar's bottom edge; `{colors.border-divider}` #1C1D2B for the line between list rows (sampled in `figma/issues.png`, 0.5 px in the source, 1 px on a 1× screen).

### Text (F `11:23`)

- **Primary** `{colors.text-primary}` #EEEFFC: titles, values, the bright fragments of activity sentences. **Secondary** `{colors.text-secondary}` #D2D3E0: sidebar items, menu items, buttons. **Muted** `{colors.text-muted}` #858699 exactly as the source: identifiers, dates, property labels, section headers; 4.89:1 on the app, 4.84:1 on the sidebar and the rail. On raised fills it fails AA for 13 px (4.36:1 on the group header, 4.45:1 on a card, 3.97:1 on a button), so muted text on those fills uses **muted raised** `{colors.text-muted-raised}` #9293A6 (5.17:1, 5.27:1, 4.70:1; D). One accepted exception: the count in a group header stays #858699 as in the source (4.36:1); it repeats what the list shows and carries no information of its own. Real text keeps AA. **Faint** `{colors.text-faint}` #4C4F6B: disabled text and decorative strokes only. **Menu icons** `{colors.icon-menu}` #8988A2 (sampled in `figma/issues-filter.png`): every icon and glyph of a menu row, the priority bars included; the selected option of a property menu keeps its brighter bars.
- **Link** `{colors.link}` #8C94EA in prose (D: the primary lightened to 6.3:1 on the app; #575BC7 of the source is 3.3:1 there).

### Primary and decoration

- **Primary** `{colors.primary}` #5E6AD2, hover `{colors.primary-hover}` #575BC7: the done glyph, checked checkboxes, the focus ring, the progress line, the primary button.
- **Decoration** (F `11:23`): red #EB5757, orange #F2994A, yellow #F2C94C, green #4CB782, blue #4EA7FC, purple #BB87FC, teal #00B2BF. Yellow is the in-progress glyph; red marks errors; the label dot and the repository avatar take a color from a stable hash over the seven.

### Light theme (D)

The same roles on white: app `{colors.light-bg-app}`, sidebar and rail `{colors.light-bg-sidebar}` / `{colors.light-bg-rail}`, group header `{colors.light-bg-group}`, borders `{colors.light-border}` and `{colors.light-border-divider}`, text `{colors.light-text-primary}` / `{colors.light-text-secondary}` / `{colors.light-text-muted}` (4.97:1 on white, 4.53:1 on the group header, so it serves raised fills too) / `{colors.light-text-faint}`, link `{colors.light-link}` (5.8:1). The primary stays #5E6AD2 (4.7:1 on white). Decoration colors darken (`{colors.light-red}`, `{colors.light-yellow}`, …) so that a glyph keeps 3:1 where it is the only signal. Menus are white at 90 % over the blur with `{colors.light-border-menu}`.

## Typography

- **Inter Variable** from `@fontsource-variable/inter` (OFL), weights 400 and 500 only, subsets latin, latin-ext and cyrillic, served from `/assets` with the build; fallback `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`. `font-feature-settings: "cv01", "ss03"` everywhere; `tnum` on counts, dates and change counts.
- **Mono**: `ui-monospace, "SF Mono", Menlo, Consolas, monospace` at 12 px for paths and commits in the properties panels and for code in prose; never in rows.
- **Scale** (F `11:24`): `{typography.title2}` 24/500 for the issue title; `{typography.normal}` 15/22 for prose; `{typography.small}` 13/400, the interface default and the list-row title; `{typography.small-medium}` for the page title, sidebar items, group headers, inbox titles and property values; `{typography.mini}` 12 px for dates, chips, buttons, the activity; `{typography.micro}` 11 px for keyboard hints and the avatar badge.
- Line height is `normal` (1) inside rows and controls, which centre their content in fixed heights; prose uses 22 px. Sentence case, no uppercase, no exclamation marks, no weight above 500.

## Layout

- The app is a `100dvh` grid: sidebar 220 px | content. The content is a column: the top bar 56 px, then the view, which scrolls inside (`min-height: 0`); the page never scrolls on a desktop.
- **Sidebar** (F `47:1331`): the brand row (28 px, 13 px from the top), a 28 px action row 54 px from the top (a wide Reindex button and a 28 px square search button), 13 px below it the pages (27 px rows), 16 px below them the "Repositories" header and the tree of every repository of the catalog, which scrolls inside the sidebar between the pages and the footer: owner rows (27 px, a 12 px chevron, the owner, the count of its repositories in muted) and under an expanded owner its repositories as nested 27 px rows whose text starts at x=66, with a 1 px guide line at x=48 in `{colors.border-button}` #313248; the footer (12 px muted) at the bottom.
- **Top bar** (F `84:3817`): padding 14 px 24px 14px 30px; the page title 13/500 at 36 px from the content edge, then the dashed **Filter** button and the active filter chips (scrolling sideways when they overflow); on the right only the 64×24 list/board switch where a page has both layouts (the status page: a list and a board of the checkouts; the board: its link to the inbox as the list) and the 80×24 **View** button, 8 px apart. The inbox has no switch: its grouping is in View, the board is /tablo. A 1 px `{colors.border}` line below.
- **List view** (status): the group header (36 px, padding-left 37 px: a 14 px glyph, the name 13/500, the count) starts right under the top bar. Rows are 44 px with fixed leading slots, `grid-template-columns: 16px 42px 14px minmax(0, 1fr) auto` with 9 px gaps from x=36, so the title starts 134 px into the content as in the reference (x≈354 at 1512 px): a 16 px leading slot, a 42 px muted identifier slot, the 14 px state glyph right before the title, the title, then trailing meta (12 px muted Inter, tabular).
- **Inbox**: the list column 420 px (`{spacing.inbox-list}`) with a 1 px border on its right, the reader in the rest. Rows are 56 px, padding 10 px 12 px, radius 6, 8 px between the avatar and the text, a 24 px trailing column, no gap between rows (R `ref/inbox.png`).
- **Issue page** (F `96:6792`): a full page fills the content area (x=220 to the right edge, no overlay). The article has a 48 px header (padding 60 px left and 62 px right, contents 3 px below centre: the 20 px repository mark and the breadcrumb, which is also the way back, and a trailing 24 px issue menu with a 16 px ellipsis) and an 832 px content column at a 75 px inset (the prose 802 px), 35 px under the header; the title ink near y=88 and the description 61 px lower (32 px margin). Activity has its own 832 px wrapper that extends 15 px beyond the prose on each side, padding 30 px 1 px 0 (so its heading and markers start 14 px left of the prose) and a divider inset 5 px in `{colors.border-activity}` #22232E, quieter than the shell's borders. The **properties rail** is 340 px on `{colors.bg-rail}` with its own 56 px toolbar (a 100 px code slot, then 20 px icon actions 20 px apart, the divider on its bottom edge at y=56, inset 24 px on both sides) and the property list 26 px below it, rows 52 px apart, labels 103 px, each value's symbol in a 14×14 slot 10 px before its text, so every value starts on one line. The inbox reader is the same component in its pane with 24 px insets when the article column is under 800 px; when the pane is under 760 px the rail moves under the article header as a grid of properties and its score, dates and provenance go under the activity.
- **Board** (R `ref/board-issues.png`): vertical status columns 320 px wide, 28 px apart, the board padded 24 px; each column scrolls under a 44 px sticky header (glyph, name, count, trailing 24 px actions) with a 1 px divider and 12 px before the first card. Cards 10 px apart.
- **Narrow** (< 900 px): the sidebar becomes an overlay (280 px) from a menu button at the left of the top bar; the inbox reader and the full issue page cover the list with Back. Below 600 px the Filter button is icon-only with the count, Display icon-only, the chips take a second row; touch targets are 44 px; no horizontal page scroll at 390 px.

## Elevation & Depth

| Level | Treatment | Use |
| --- | --- | --- |
| 0 | `{colors.bg-app}` | content, rows, the article |
| 1 | `{colors.bg-sidebar}` / `{colors.bg-rail}` / `{colors.bg-group}`, 1 px edges | sidebar, top bar, properties rail, group headers |
| 2 | `{colors.bg-card}`, quiet `{colors.border-card}`, shadow `0 1px 1px #00000026` | cards, filter chips, buttons (F effect "Button 2") |
| 3 | the menu gradient, `backdrop-filter: blur(16px) saturate(140%)`, no outline in the dark theme (`{colors.border-menu}` is transparent; the light theme keeps a subtle `{colors.light-border-menu}` edge, since white on white needs a boundary), shadow `0 4px 16px #00000026` for menus and popovers, `0 16px 64px #00000080` for the command bar | menus, the Display menu, the command bar, the peek panel, the sidebar overlay |

The peek panel dims the page with `{colors.overlay}`; the command bar does not dim it: a transparent layer catches the click outside and only its own shadow darkens what is around it. Hover changes only the fill, in 80 ms; the peek slides in 150 ms; nothing moves under `prefers-reduced-motion`.

## Shapes

| Token | Value | Use |
| --- | --- | --- |
| `{rounded.xs}` | 3px | the brand mark, keyboard badges, view-switch segments |
| `{rounded.sm}` | 4px | buttons, sidebar items, filter chips, menu items (F `14:347`, `47:1331`) |
| `{rounded.md}` | 6px | inbox rows, comment cards, notices |
| `{rounded.lg}` | 8px | menus, the peek panel, board cards (F `84:3840`, R `ref/board-issues.png`) |
| `{rounded.xl}` | 16px | the command bar (F `84:3840`) |
| `{rounded.full}` | circle | label and meta chips (F `88:3086`), the repository avatar |

## Components

### Sidebar (F `47:1331`)

`sidebar`: 220 px, `{colors.bg-sidebar}`, a 1 px `{colors.border}` on the right, padding 0 16 px. The brand row: the 18 px maintainer mark (`brand-mark`, the favicon's M on `{colors.brand}`) and "maintainer" 13/500. The action row: `sidebar-action` **Reindex** (refresh icon and label, flexible width, 28 px) where the source has "New issue", and a 28 px square search button that opens the command bar; both with a `{colors.border-button-hover}` #393A4B edge, the search icon in `{colors.text-muted}`. The pages: `sidebar-item`s Inbox, Tablo, Status (16 px icon box with about 14 px of artwork and filled interior areas, 10 px gap, 13/500 `{colors.text-secondary}`; the current page `sidebar-item-selected` with `aria-current="page"`: the text brightens, the icon stays `{colors.text-muted}`). The section `sidebar-section-header` "Repositories" in 12 px muted, then a tree (`role="tree"`) of every repository of the catalog, grouped by owner in alphabetical order: an owner row toggles its repositories (`aria-expanded`); a repository row (`sidebar-nested-item`) is a link to the inbox filtered by it, with its count of open issues in muted at the right and, for a repository the fetch state records as private or internal, a 12 px lock of our own drawing in `{colors.text-muted}` after the name (no lock when the visibility is unknown, as for a pin without a record). The current repository is the one the inbox is filtered by alone, highlighted and kept in step when the filter changes. The expanded owners are remembered in the browser; by default only the owner of the current repository is expanded. Keyboard: Up and Down move between the visible rows, Right expands or enters, Left collapses or returns to the owner, Enter opens a repository or toggles an owner, Home and End jump. The footer, like the source's "Invite people / Help & Support": the generation or collection state as a muted 12 px line and the theme item (Dark, Light, System), 12 px muted with 14 px icons.

### Command bar (F `84:3840`)

`command-bar`: 642 px, radius 16, centred on the viewport (not on the content) at the top third, level 3 with its own shadow and no dim; no outline in the dark theme. An input "Search issues or go to a page…" (15 px, 20 px side padding, no leading icon, the placeholder in `{colors.text-faint}`) over a 1 px separator, the panel's top 12.8 % of the viewport height (121 px at 945 px), the input without inner padding, its divider `{colors.separator-command}` #292A39; then grouped results: Pages (Inbox, Tablo, Status) and Issues (the tasks of the collection that match the text, by score, at most eight; the open ones by score while the text is empty), each a 42 px row with its icon or glyph centred in a 16×16 slot and 16 px before the code in muted and the title. No row is highlighted when the bar opens: the arrows or the pointer pick one, and Enter with none picked opens the first result.

Opened while an issue is open (the full issue page or the issue in the inbox reader), the bar is issue-scoped, as the source's: a 103 px header with a context chip "CODE — title" 16 px from the inner top and left (25 px high, radius 3, `{colors.bg-command-chip}`, 12 px text in `{colors.text-command-chip}` #8E8DA9, as wide as its text and ellipsized), the 56 px input region 3 px below it with 20 px side insets; the results start right under the input (4 px list padding, no group heading) with the issue's actions, their icons and glyphs in `{colors.icon-menu}` like every command icon (a checked level keeps its emphasis): Change priority… and Change estimate… (each opens the list of its levels with our glyphs and a check on the current one; a pick triages the label through the issue page, with the same ETag and the same 412, 409 and 422 handling), a 1 px `{colors.separator-command-group}` rule with 5 px above and below, then Copy link to the issue, Copy file path, Open on GitHub. Rows are padded 12 px on the left and 20 px on the right, so a keycap ends 29 px from the panel's edge. An action with a working shortcut shows its keycaps on the right (20×20 at least, radius 3, 11 px in `{colors.text-secondary}`, `{colors.bg-command-kbd}`, 4 px apart): **P** for Change priority and **E** for Change estimate. With an issue open and the focus outside a text field, a menu and a dialog, P and E open the bar issue-scoped right at the priority or the estimate levels. The copy and open actions have no shortcut, so they show no keycaps; typing filters the actions, then the pages and the issues of the collection follow as in the global search. Backspace in the empty field removes the chip and makes the bar the global search; Escape in a list of levels returns to the actions. Arrow keys move, Enter opens, Escape closes and returns the focus. `/` and Cmd/Ctrl+K open it from anywhere but a text field; an issue opens as the full issue page.

### Top bar and filters (F `84:3817`, `84:3852`, `34:691`)

- `top-bar` as in Layout. The title is the page name.
- `filter-button`: "+ Filter", 24 px, 1 px dashed `{colors.border-button}`, 12/500, transparent on the toolbar also while its menu is open (text `{colors.text-secondary}`); the F key opens it. It opens the filter `menu`, 210 px, 1 px under the trigger: a `menu-search` "Filter…" (11 px inset, its divider flush with the panel's top) with a 16 px `kbd` "F" on #2D2F3E at the right, then the dimensions (32 px rows with a 16 px line box, 7 px inset, every icon centred in a fixed 16×16 slot in `{colors.icon-menu}`, 10 px before the label, 14 px icons with filled mass (the label mark smaller, about 12×9 px); Priority, Estimate, State and Status show their own glyphs) in groups separated by `{colors.separator-menu}` rules: State, Status, Content; Priority, Estimate, Labels, Type; Milestone, Repository, Owner, Assignee (a page lists those it supports). The rows have no chevrons and the menu ends after its last row; Clear lives in the top bar. A dimension opens its values in the same menu with a back row: checkboxes (several values are any-of; labels are all-of; Priority and Estimate keep one level) or radios (State), each with its glyph and count; Content is a text field (the code, title, repository or label contains it). Changes apply at once; Escape closes and returns the focus to the trigger.
- `filter-chip`: one per active dimension, segmented by 1 px borders: the icon and the dimension, the operator in muted ("is", "is any of", "include all of", "contains"), the values (opens that dimension), × removes the dimension. "Clear" is a `button-ghost` after the chips.
- `view-switch`: 64×24, two segments, list and board: links between the inbox and the board, buttons on the status page (`layout=board` in the URL), no padding and no gap: two 32×24 segments (border-box), filled icons with about 14 px of artwork in 16 px boxes (the list as four 2 px bands), the inactive one in `{colors.text-faint}` #4C4F6B; the current one on `{colors.bg-button}` with a 1 px `{colors.border-button}` #313248 edge and its icon in `{colors.text-secondary}`.
- `display-button`: exactly 80×24, "View" with the adjustments icon in `{colors.text-muted}` and a chevron, as the source labels it; it opens `display-menu`: rows "Grouping" and "Ordering" with 24 px selects; on the board the ordering within the columns (default: score, the closed columns by update; or one of score, priority, updated, created for all; in the URL as `sort`), switches for the empty columns and for the card properties (priority, estimate, code, labels, repository; remembered in the browser); on the status page also the ordering chain as toggle chips (a click cycles ascending, descending, off and replaces the chain, Shift-click appends), the summary line of the collection, the snapshot time and **Refresh**; "Reset to default" at the bottom.

### List row and group header (F `60:1776`)

`group-header` 36 px on `{colors.bg-group}`: the glyph, the name, the count in muted raised. `list-row` 44 px with the fixed slots of Layout and a 1 px `{colors.border-divider}` under it; the title 13/400 primary; hover `list-row-hover`, the open row `list-row-selected`. The status page also has a **board** of the checkouts (`layout=board`): one column per group of the current grouping (Errors, Uncommitted, Ahead or behind, Orphans, Clean; or the owners; or one column without grouping), a 24×24 header action with a 14 px icon that shows that group in the list; cards in the board-card style with the repository as the title and a footer of the branch, the divergence, the change counts and the lock; a card opens the checkout's peek panel. For a checkout the slots are: the push lock in the leading slot, the divergence ("↑2 ↓1") in the 42 px identifier slot, the state glyph at 12 px at the start of the 14 px slot right before the title (the group header keeps 14 px), then the properties in a column grid aligned from row to row and from group to group: the repository (13/400), the branch, the changes ("+a/-d · nf · ?u", right-aligned, tabular), the status word, and the path, right-aligned in the rest; 12 px muted Inter, 16 px between columns. A column's width is measured from the longest value shown (repository 120–300 px, branch 64–180, changes 72–220, status 56–200), a longer value ends in an ellipsis with the full value in its tooltip, so values of different length never shift the next column. View → Display properties toggles Branch, Changes, Status and Path, remembered in the browser; below 900 px the path, below 600 px the branch and the changes leave the grid.

### Glyphs (D, drawn for the hub; 14 px in rows, 16 px in menus; 1.5 px strokes)

- **Status** (the board columns): backlog, a dashed circle in muted; todo, a circle in `{colors.text-secondary}`; in progress, a circle with its right half filled, yellow; done, a filled primary circle with a check cut out; won't do, a filled muted circle with a cross cut out.
- **Priority** (`impact/*`): high three bars, medium two active of three, low one of three; inactive bars at 35 % opacity, active bars in `{colors.text-secondary}`; none "No priority" three dots in muted. Urgent (an orange rounded square with "!") is defined but unused: no impact level above high exists.
- **Estimate** (`effort/*`): easy S, medium M, hard L: a small triangle of dots and the letter; "No estimate" in the panel only.
- **At 12 px** (list rows, the inbox, chips) the outline glyphs use a 12-unit drawing with a 2 px outline (`r=5`), as the source's row circles; the filled ones scale.
- **Checkout groups** (status page): error, a red circle with "!"; uncommitted, a yellow pie; ahead or behind, a blue circle with arrows; orphan, a dashed muted circle; clean, the primary done glyph.
- **Icons**: Tabler paths (outline and filled, MIT) drawn at a 1.5 px effective stroke. Navigation and menu icons use a tight view box, so their artwork fills about 14 px of the 16 px box, and filled forms or filled interiors where they read as mass (the board, the inbox tray, labels, people, types).

### Label chip (F `88:3086`)

`label-chip`: a 28 px pill on `{colors.bg-label-chip}` with a `{colors.border-label-chip}` edge, an 8 px dot in the label's color and the name in 12 px muted, on the issue page. `label-chip-compact` (22 px, the same colors) where a row needs a chip; on board cards the footer chips are 24 px rounded rectangles (radius 6). `type: x` labels and the issue type are label chips too. Priority and estimate are glyphs, never chips.

### Board (R `ref/board-issues.png`)

`board-column` with a `board-column-header`: status glyph, name, count muted, and at the right a 24 px icon action that opens the column in the inbox (the only column action the hub has; there is no "+" because the hub creates no issues); 44 px, sticky, a 1 px divider, 12 px before the first card. `card` on `{colors.bg-card}` with the quiet `{colors.border-card}` edge, radius 8, padding 16 px, 10 px apart, about 100 px tall at 320 px: a 36 px title region (13/500, two lines clamped) with the repository avatar (16 px circle) at the top right, 8 px, then one 24 px footer that never wraps: the priority glyph, the estimate, the code in 12 px muted, at most one label chip (24 px, radius 6; one is what fits beside the code at 320 px) and a compact "+N". The milestone and the full repository name are on the issue page. Hover lifts the edge to `{colors.border-button}`. "Load more" is a `button-ghost` at the column's end; a column's error stays inside it. A card opens the full issue page.

### Inbox (R `ref/inbox.png`)

View → Grouping: none (the default), status, priority, estimate, repository, owner, milestone, type, as `group` in the URL. A group is the `group-header` of the list (36 px on `{colors.bg-group}`, the glyph of the value: status, priority or estimate glyph, the repository's circle, an owner, milestone or type icon; the name; the count in muted) over its rows; the groups follow the order of their values (the workflow order for status; high, medium, low, none; S, M, L, none; the facets' order for the others; the group without a value last), the ordering applies inside a group, and the arrow keys and the selection follow the rows as shown, across the groups. The count of a group is its total in the selection, from the facets of the list answer (status from the columns, priority and estimate from the impact and effort labels, the others from their facets; a group without a value takes the rest of the total); "Load more" adds rows to the groups, so a group can show fewer rows than its count until the list is loaded.

`inbox-row` 56 px, padding 10 px 12 px, gap 8: a 36 px circular `repo-avatar` (the repository's initial on a tint of its hash color) with a badge at its bottom right only when there is an event to show (a flag when the issue is flagged; the list carries no comment events); a text stack of two 16 px lines 4 px apart: "CODE title" (code muted, title 13/500), then one plain ellipsized 12 px muted sentence (repository · milestone); a 24 px trailing column, right-aligned, with the 12 px status glyph over the age. The selected row is `inbox-row-selected`.

### Issue page and properties rail (F `96:6792`, `88:3086`)

`issue-header` 48 px over the article column only: the 20 px repository mark, 14 px, then the breadcrumb "repository › code" in 12 px with an inline separator, which is a button back to the list or the board where the page covers one ("Back to the board", "Back to the inbox"), in the inbox reader an expand action, and a trailing 24 px "…" menu (three 3 px dots at x=3, 8 and 13 of a 16 px box) with Copy link, Copy issue code, Copy file path and Open on GitHub. `issue-title`, the flags as meta chips, `issue-prose` 32 px under the title, then **Activity** in its wider wrapper: one chronological stack of events and comments, 16 px circular markers, 20 px between items, sentences in 12/500 muted with the values (actors, dates, reasons) in `{colors.text-primary}`, comments as cards, the divider in `{colors.border-activity}`. Accepted feature differences: no comment composer and no subscription actions, because the hub cannot post comments or subscribe; no dead controls are drawn. `properties-panel` (the rail): the `rail-toolbar` with the code in a 100 px slot and three 20 px icon actions 20 px apart (copy the GitHub link, copy the file path, open on GitHub), then the properties 52 px apart, labels 13/500 muted and values 12/400 (the menus' triggers too): Status, Priority (a `menu` with No priority, High, Medium, Low, each with its glyph and a check on the current), Estimate (No estimate, S, M, L); option text 13/400, Labels (chips), Type, Milestone, Assignees, Repository, State; then Score with its terms, Dates and Provenance in 12 px. Saving shows "Saving…" then "Saved" under the property; a `412` disables Priority and Estimate until Reload; `409` and `422` explain themselves in a notice. The full page is addressed by `?task=<id>` in the URL of the inbox or the board; Back, Escape and the browser's back return to the list or the board with the focus on the issue that was opened.

### Peek panel

`peek`: the status page's checkout opens in a 480 px modal panel from the right, 8 px from the viewport's edges, level 3, radius 8. It contains the focus, makes the rest inert, closes on Escape, the backdrop and Close, and returns the focus.

### States

- **Indexing / collecting**: a 2 px `progress-line` under the top bar and the state in the sidebar footer; skeleton rows in the layout's shape.
- **Empty** vs **no matches**: `empty-state`, one muted sentence, Clear filters for no matches.
- **Recoverable failure**: a `notice` with a red icon and Try again at the top of the view.
- **Focus**: a 2 px `{colors.focus-ring}` outline, offset 1 px outside or inset on rows inside a scroll container. Every icon-only control has an accessible name and a tooltip.

## Do's and Don'ts

### Do

- Keep the measured sizes: 220, 56, 36, 44; 340 and 52 in the rail; 13 px text; 24 px buttons.
- Carry state with the glyphs: status, priority and estimate are shapes before they are words.
- Keep the primary for done, focus, checked and links.
- Design every state of every view in both themes and at 390 px.

### Don't

- No logo, name, illustration, avatar or icon file of the reference; no green outside the maintainer mark.
- No score in rows: the score is the default ordering and lives in the properties rail.
- No monospace in rows; no weight above 500, no uppercase, no emoji; no gradients except the blur behind floating surfaces.
