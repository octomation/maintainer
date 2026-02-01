import type { ComponentChildren } from 'preact';
import { createContext } from 'preact';
import { useContext } from 'preact/hooks';
import type { IssueCommands } from './CommandBar';
import { Icon } from './Icon';

/** What the shell offers the pages: the narrow layout opens the sidebar from the top bar. */
export interface Shell {
  openSidebar: () => void;
  /**
   * An open issue offers its commands to the command bar; the one registered
   * last is the issue in front. The returned function withdraws it.
   */
  registerIssue: (get: () => IssueCommands | null) => () => void;
}

export const ShellContext = createContext<Shell>({ openSidebar: () => {}, registerIssue: () => () => {} });

export interface TopBarProps {
  title: string;
  /** After the title: the Filter button, the chips, the search. */
  children?: ComponentChildren;
  /** On the right: the view switch, Display, Refresh. */
  actions?: ComponentChildren;
  inert?: boolean;
}

/**
 * The 56 px bar over a view: the menu button on a narrow screen, the page
 * title, the filters and, on the right, the actions. It wraps on a narrow
 * screen: the chips and the search take a second row.
 */
export function TopBar({ title, children, actions, inert }: TopBarProps) {
  const shell = useContext(ShellContext);
  return (
    <header class="topbar" inert={inert}>
      <button type="button" class="icon-button topbar__menu" aria-label="Open the sidebar" title="Pages" onClick={shell.openSidebar}>
        <Icon name="menu" />
      </button>
      <h1 class="topbar__title">{title}</h1>
      {children}
      {actions && <div class="topbar__actions">{actions}</div>}
    </header>
  );
}

export interface ViewSwitchProps {
  current: 'list' | 'board';
  /** Links: the search of the current filters, carried to the inbox (list) or the board. */
  search?: string;
  /** Buttons instead of links: the layout of one page, such as the status page. */
  onChange?: (layout: 'list' | 'board') => void;
}

/**
 * List and board: links between the inbox and the board, which are two
 * layouts of the same issues, or buttons for a page with both layouts.
 */
export function ViewSwitch({ current, search = '', onChange }: ViewSwitchProps) {
  if (onChange) {
    return (
      <div class="view-switch" role="group" aria-label="Layout">
        {(['list', 'board'] as const).map((layout) => (
          <button
            type="button"
            key={layout}
            class="view-switch__item"
            aria-pressed={current === layout}
            aria-current={current === layout ? 'page' : undefined}
            aria-label={layout === 'list' ? 'List' : 'Board'}
            title={layout === 'list' ? 'List' : 'Board'}
            onClick={() => onChange(layout)}
          >
            <Icon name={layout === 'list' ? 'list-solid' : 'board-solid'} size={16} tight={layout !== 'list'} />
          </button>
        ))}
      </div>
    );
  }
  return (
    <nav class="view-switch" aria-label="Layout">
      <a class="view-switch__item" href={`/inbox${search}`} aria-current={current === 'list' ? 'page' : undefined} aria-label="List" title="List: the inbox">
        <Icon name="list-solid" size={16} />
      </a>
      <a class="view-switch__item" href={`/tablo${search}`} aria-current={current === 'board' ? 'page' : undefined} aria-label="Board" title="Board: the tablo">
        <Icon name="board-solid" size={16} tight />
      </a>
    </nav>
  );
}
