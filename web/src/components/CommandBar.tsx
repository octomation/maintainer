import type { ComponentChildren } from 'preact';
import { useEffect, useId, useRef, useState } from 'preact/hooks';
import type { ApiClient } from '../api/client';
import type { Effort, Impact, Task, TriagePatch } from '../api/types';
import { EFFORTS, IMPACTS } from '../api/types';
import { EFFORT_TITLE, IMPACT_TITLE } from './Chip';
import { ESTIMATE_SIZE, EstimateGlyph, PriorityGlyph, StatusGlyph } from './Glyphs';
import { Icon, type IconName } from './Icon';
import { ROUTES } from './Sidebar';

/** What the open issue offers the command bar: its identity and its actions. */
export interface IssueCommands {
  id: string;
  code: string;
  title: string;
  url: string | null;
  path: string;
  impact: Impact | '';
  effort: Effort | '';
  /** Triage is blocked: saving, a 412 waiting for Reload, duplicate mirrors. */
  blocked: boolean;
  /** Triages through the issue page: its ETag, its 412, 409 and 422 handling. */
  triage: (patch: TriagePatch) => void;
  copy: (text: string, what: 'link' | 'path') => void;
}

export interface CommandBarProps {
  client: ApiClient;
  open: boolean;
  onClose: () => void;
  /** Opens the full issue page of a task. */
  onOpenTask: (id: string) => void;
  /** Goes to a page; a navigation by default. */
  onGo?: (path: string) => void;
  /** The issue open when the bar opened: the bar is issue-scoped until Backspace removes it. */
  issue?: IssueCommands | null;
  /** Opens a URL in a new tab; window.open by default. */
  onOpenUrl?: (url: string) => void;
  /** Opens issue-scoped right at a list of levels, as the P and E shortcuts do. */
  initialMode?: Mode;
}

/** How long the input waits before it asks the service. */
export const COMMAND_DEBOUNCE_MS = 120;
/** The issues the bar lists at most. */
const LIMIT = 8;

type Item =
  | { kind: 'action'; key: string; title: string; icon: ComponentChildren; run: () => void; keepOpen?: boolean; keys?: string[]; group?: 'triage' | 'links' }
  | { kind: 'page'; path: string; title: string; icon: IconName }
  | { kind: 'task'; task: Task };

export type Mode = 'root' | 'priority' | 'estimate';

/**
 * The command bar: a 642 px dialog over the page with a search field and
 * grouped results. Global, it lists the pages and the issues of the
 * collection that contain the text (the open ones by score while it is
 * empty). Opened over an open issue it is issue-scoped: a context chip, then
 * the issue's actions (priority, estimate, copy, open), then the global
 * results; Backspace in the empty field removes the chip. Arrow keys move,
 * Enter opens, Escape steps back from a list of levels or closes.
 */
export function CommandBar({ client, open, onClose, onOpenTask, onGo = (path) => location.assign(path), issue = null, onOpenUrl = (url) => void window.open(url, '_blank', 'noopener'), initialMode = 'root' }: CommandBarProps) {
  const [text, setText] = useState('');
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [failed, setFailed] = useState('');
  // No result is picked when the bar opens; the arrows or the pointer pick one.
  const [active, setActive] = useState(-1);
  const [context, setContext] = useState<IssueCommands | null>(null);
  const [mode, setMode] = useState<Mode>('root');
  const input = useRef<HTMLInputElement>(null);
  const previous = useRef<HTMLElement | null>(null);
  const sequence = useRef(0);
  const id = useId();

  useEffect(() => {
    if (!open) return;
    previous.current = document.activeElement as HTMLElement | null;
    setText('');
    setActive(-1);
    setMode(issue ? initialMode : 'root');
    setContext(issue);
    input.current?.focus();
    return () => {
      if (previous.current?.isConnected) previous.current.focus();
    };
  }, [open]);

  useEffect(() => {
    if (!open || mode !== 'root') return;
    const seq = ++sequence.current;
    setLoading(true);
    const timer = setTimeout(() => {
      const q = text.trim();
      client
        .listTasks({ q: q || undefined, state: q ? 'all' : 'open', sort: 'score', limit: LIMIT })
        .then(({ envelope }) => {
          if (seq !== sequence.current) return;
          setTasks(envelope.data?.tasks ?? []);
          setFailed('');
          setLoading(false);
        })
        .catch((err: unknown) => {
          if (seq !== sequence.current) return;
          setTasks([]);
          setFailed(err instanceof Error ? err.message : String(err));
          setLoading(false);
        });
    }, COMMAND_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [open, text, client, mode]);

  if (!open) return null;

  const needle = text.trim().toLowerCase();
  const matches = (title: string) => !needle || title.toLowerCase().includes(needle);
  const enter = (next: Mode) => {
    setMode(next);
    setText('');
    setActive(-1);
    input.current?.focus();
  };

  // The issue's actions: the levels of a property in a nested list, the rest at once.
  let actions: Item[] = [];
  if (context && mode === 'root') {
    actions = [
      { kind: 'action', key: 'priority', title: 'Change priority…', icon: <PriorityGlyph level={context.impact || 'none'} />, run: () => enter('priority'), keepOpen: true, keys: ['P'], group: 'triage' },
      { kind: 'action', key: 'estimate', title: 'Change estimate…', icon: <EstimateGlyph />, run: () => enter('estimate'), keepOpen: true, keys: ['E'], group: 'triage' },
      ...(context.url ? [{ kind: 'action' as const, key: 'link', title: 'Copy link to the issue', icon: <Icon name="link" />, run: () => context.copy(context.url!, 'link'), group: 'links' as const }] : []),
      { kind: 'action', key: 'path', title: 'Copy file path', icon: <Icon name="copy" />, run: () => context.copy(context.path, 'path'), group: 'links' },
      ...(context.url ? [{ kind: 'action' as const, key: 'github', title: 'Open on GitHub', icon: <Icon name="external-link" />, run: () => onOpenUrl(context.url!), group: 'links' as const }] : []),
    ];
    actions = actions.filter((a) => a.kind === 'action' && matches(a.title));
  } else if (context && mode === 'priority') {
    actions = (
      [
        { kind: 'action', key: 'none', title: 'No priority', icon: <PriorityGlyph level="none" />, run: () => context.triage({ impact: null }) },
        ...IMPACTS.map((level: Impact) => ({ kind: 'action' as const, key: level, title: IMPACT_TITLE[level], icon: <PriorityGlyph level={level} />, run: () => context.triage({ impact: level }) })),
      ] as Item[]
    ).filter((a) => a.kind === 'action' && matches(a.title));
  } else if (context && mode === 'estimate') {
    actions = (
      [
        { kind: 'action', key: 'none', title: 'No estimate', icon: <EstimateGlyph class="glyph--empty" />, run: () => context.triage({ effort: null }) },
        ...EFFORTS.map((level: Effort) => ({ kind: 'action' as const, key: level, title: `${ESTIMATE_SIZE[level]} · ${EFFORT_TITLE[level].toLowerCase()}`, icon: <EstimateGlyph />, run: () => context.triage({ effort: level }) })),
      ] as Item[]
    ).filter((a) => a.kind === 'action' && matches(a.title));
  }
  const checked = (item: Item) => {
    if (!context || item.kind !== 'action') return false;
    if (mode === 'priority') return item.key === (context.impact || 'none');
    if (mode === 'estimate') return item.key === (context.effort || 'none');
    return false;
  };

  const nested = mode !== 'root';
  const pages: Item[] = nested ? [] : ROUTES.filter((r) => matches(r.title)).map((r) => ({ kind: 'page', path: r.path, title: `Go to ${r.title}`, icon: r.icon }));
  const issues: Item[] = nested ? [] : tasks.map((task) => ({ kind: 'task', task }));
  const items = [...actions, ...pages, ...issues];
  const current = active < 0 ? -1 : Math.min(active, items.length - 1);

  const choose = (item: Item) => {
    if (item.kind === 'action') {
      if (nested && context?.blocked) return;
      item.run();
      if (!item.keepOpen) onClose();
      return;
    }
    onClose();
    if (item.kind === 'page') onGo(item.path);
    else onOpenTask(item.task.id);
  };

  const onKey = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      event.preventDefault();
      event.stopPropagation();
      if (nested) enter('root');
      else onClose();
    } else if (event.key === 'Backspace' && text === '' && context) {
      // Backspace in the empty field: out of a list of levels, then off the issue.
      event.preventDefault();
      if (nested) enter('root');
      else {
        setContext(null);
        setActive(-1);
      }
    } else if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      if (items.length) setActive(current < 0 ? (event.key === 'ArrowDown' ? 0 : items.length - 1) : (current + (event.key === 'ArrowDown' ? 1 : items.length - 1)) % items.length);
    } else if (event.key === 'Enter') {
      event.preventDefault();
      // Enter with nothing picked opens the first result.
      const item = items[current < 0 ? 0 : current];
      if (item) choose(item);
    } else if (event.key === 'Tab') {
      // The dialog keeps the focus on its field; the list is chosen with the arrows.
      event.preventDefault();
    }
  };

  const optionId = (index: number) => `${id}-option-${index}`;
  const row = (item: Item, index: number) => (
    <li
      key={item.kind === 'page' ? item.path : item.kind === 'task' ? item.task.id : `action-${item.key}`}
      id={optionId(index)}
      role="option"
      aria-selected={index === current}
      aria-checked={nested ? checked(item) : undefined}
      aria-disabled={item.kind === 'action' && nested && context?.blocked ? true : undefined}
      class={`command-item${index === current ? ' is-active' : ''}`}
      onMouseMove={() => setActive(index)}
      onMouseDown={(event) => event.preventDefault()}
      onClick={() => choose(item)}
    >
      {item.kind === 'action' ? (
        <>
          <span class="command-item__icon">{item.icon}</span>
          <span class="command-item__title">{item.title}</span>
          {checked(item) && <Icon name="check" size={14} class="command-item__check" />}
          {item.keys && (
            <span class="command-item__keys" aria-label={`Shortcut ${item.keys.join(' ')}`}>
              {item.keys.map((key) => (
                <kbd class="command-kbd" key={key}>
                  {key}
                </kbd>
              ))}
            </span>
          )}
        </>
      ) : item.kind === 'page' ? (
        <>
          <span class="command-item__icon">
            <Icon name={item.icon} tight />
          </span>
          <span class="command-item__title">{item.title}</span>
        </>
      ) : (
        <>
          <span class="command-item__icon">
            <StatusGlyph column={item.task.column} />
          </span>
          <span class="command-item__code">{item.task.code}</span>
          <span class="command-item__title">{item.task.title}</span>
          <span class="command-item__hint">{item.task.repository}</span>
        </>
      )}
    </li>
  );

  const placeholder = mode === 'priority' ? 'Change priority to…' : mode === 'estimate' ? 'Change estimate to…' : context ? 'Type a command or search…' : 'Search issues or go to a page…';
  return (
    <>
      <div class="command-bar__backdrop" onClick={onClose} data-testid="command-backdrop" />
      <div class={`command-bar${context ? ' command-bar--issue' : ''}`} role="dialog" aria-modal="true" aria-label="Search" onKeyDown={onKey}>
        <div class="command-bar__header">
          {context && (
            <div class="command-bar__context" title={`${context.code} — ${context.title}`}>
              <span class="command-bar__chip">
                {context.code} — {context.title}
              </span>
            </div>
          )}
          <div class="command-bar__field">
            <input
              ref={input}
              type="text"
              role="combobox"
              aria-label={context ? `Commands for ${context.code}, or search` : 'Search issues or go to a page'}
              aria-expanded="true"
              aria-controls={`${id}-list`}
              aria-activedescendant={current >= 0 && items.length ? optionId(current) : undefined}
              aria-autocomplete="list"
              placeholder={placeholder}
              value={text}
              onInput={(event) => {
                setText((event.target as HTMLInputElement).value);
                setActive(-1);
              }}
            />
          </div>
        </div>
        <ul class="command-bar__list" id={`${id}-list`} role="listbox" aria-label="Results" aria-busy={loading}>
          {actions.length > 0 && nested && (
            <li class="command-bar__group" role="presentation">
              {mode === 'priority' ? 'Priority' : 'Estimate'}
            </li>
          )}
          {actions.map((item, index) => {
            // A rule between the triage actions and the copy and open ones.
            const before = actions[index - 1];
            const rule = !nested && index > 0 && item.kind === 'action' && before?.kind === 'action' && before.group !== item.group;
            return rule ? [<li class="command-separator" role="separator" key={`rule-${index}`} />, row(item, index)] : row(item, index);
          })}
          {nested && context?.blocked && (
            <li class="command-bar__empty" role="presentation">
              The issue cannot be triaged now; see its page.
            </li>
          )}
          {pages.length > 0 && (
            <li class="command-bar__group" role="presentation">
              Pages
            </li>
          )}
          {pages.map((item, index) => row(item, actions.length + index))}
          {!nested && (
            <li class="command-bar__group" role="presentation">
              Issues
            </li>
          )}
          {issues.map((item, index) => row(item, actions.length + pages.length + index))}
          {!nested && !loading && issues.length === 0 && (
            <li class="command-bar__empty" role="presentation">
              {failed ? `The issues could not be searched: ${failed}` : needle ? 'No issue contains the text.' : 'No open issues.'}
            </li>
          )}
        </ul>
      </div>
    </>
  );
}
