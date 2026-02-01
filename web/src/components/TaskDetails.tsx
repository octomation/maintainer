import type { ComponentChildren } from 'preact';
import { useCallback, useContext, useEffect, useLayoutEffect, useRef, useState } from 'preact/hooks';
import { ApiClient, ApiError } from '../api/client';
import type { Effort, Impact, TaskDetail, TriagePatch } from '../api/types';
import { COLUMN_TITLES, EFFORTS, IMPACTS } from '../api/types';
import { EFFORT_TITLE, effortOf, IMPACT_TITLE, impactOf, LabelChip, MetaChip, otherLabels } from './Chip';
import { ESTIMATE_SIZE, EstimateGlyph, PriorityGlyph, RepoAvatar, StatusGlyph } from './Glyphs';
import { Icon } from './Icon';
import { Popover } from './Popover';
import { ErrorState, Loading } from './States';
import type { IssueCommands } from './CommandBar';
import { ShellContext } from './TopBar';

export interface TaskDetailsProps {
  client: ApiClient;
  /** The HTTP id of the task. */
  id: string;
  /** Called with the new detail after a successful triage. */
  onChanged?: (detail: TaskDetail) => void;
  /** Bump to reload the task from the service. */
  reloadKey?: number;
  /** Back, in the header, when the page covers the list or the board: the breadcrumb leads there. */
  onBack?: () => void;
  /** The accessible name of the way back, such as "Back to the board". */
  backLabel?: string;
  /** Open as the full issue page, from the inbox reader. */
  onExpand?: () => void;
  /** The reader beside the inbox list or the full issue page. */
  variant?: 'split' | 'full';
}

type Conflict = { kind: 'changed' } | { kind: 'duplicate'; message: string } | { kind: 'layout'; message: string } | { kind: 'error'; message: string };
type Feedback = { kind: 'saving' | 'saved'; of: 'impact' | 'effort' } | null;

/**
 * The issue page: the title, the rendered body and the activity on the
 * left, the properties on the right, with Priority (impact) and Estimate
 * (effort) as menus that triage the file. A 412 offers to reload the task
 * and blocks both menus until then; a 409 explains the duplicate mirrors; a
 * 422 the label layout the service cannot edit.
 */
export function TaskDetails({ client, id, onChanged, reloadKey = 0, onBack, backLabel = 'Back to the list', onExpand, variant = 'split' }: TaskDetailsProps) {
  const [detail, setDetail] = useState<TaskDetail | null>(null);
  const [etag, setEtag] = useState('');
  const [error, setError] = useState<unknown>(null);
  const [busy, setBusy] = useState(false);
  const [conflict, setConflict] = useState<Conflict | null>(null);
  const [feedback, setFeedback] = useState<Feedback>(null);
  const [attempt, setAttempt] = useState(0);
  // A reload after a 412: the triage stays blocked until the new detail and ETag are in.
  const [reloading, setReloading] = useState(false);
  // The task the page shows now and whether it is still mounted: a triage
  // that completes for another task, or after the page went away, is
  // dropped. The pages also key the page by the task id, so a new task is a
  // new instance; this guards the instance itself.
  const current = useRef(id);
  const alive = useRef(true);
  const changed = useRef(onChanged);
  current.current = id;
  changed.current = onChanged;
  // A layout effect: its cleanup runs in the commit, while a passive one
  // would wait for the next paint and let a completion slip through.
  useLayoutEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
    };
  }, []);

  const [copied, setCopied] = useState<'link' | 'path' | 'code' | null>(null);
  const copy = (text: string, what: 'link' | 'path' | 'code') => {
    void navigator.clipboard
      ?.writeText(text)
      .then(() => {
        if (!alive.current) return;
        setCopied(what);
        setTimeout(() => alive.current && setCopied(null), 1500);
      })
      .catch(() => {});
  };

  const load = useCallback(() => {
    setReloading(true);
    setAttempt((n) => n + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;
    setError(null);
    setFeedback(null);
    client
      .getTask(id)
      .then(({ envelope, etag }) => {
        if (cancelled) return;
        setDetail(envelope.data);
        setEtag(etag);
        setConflict(null);
        setReloading(false);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setError(err);
        setReloading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [client, id, reloadKey, attempt]);

  // The issue offers its commands to the command bar of the shell while it is open.
  const shell = useContext(ShellContext);
  const commands = useRef<IssueCommands | null>(null);
  useEffect(() => shell.registerIssue(() => commands.current), [shell]);

  const triage = async (patch: TriagePatch) => {
    if (!detail || busy) return;
    const target = id;
    const of = patch.impact !== undefined ? 'impact' : 'effort';
    setBusy(true);
    setConflict(null);
    setFeedback({ kind: 'saving', of });
    const stale = () => !alive.current || current.current !== target;
    try {
      const { envelope, etag: next } = await client.triage(detail.id, etag, patch);
      if (stale()) return;
      setDetail(envelope.data);
      setEtag(next);
      setFeedback({ kind: 'saved', of });
      changed.current?.(envelope.data);
    } catch (err) {
      if (stale()) return;
      setFeedback(null);
      if (err instanceof ApiError && err.status === 412) setConflict({ kind: 'changed' });
      else if (err instanceof ApiError && err.status === 409) setConflict({ kind: 'duplicate', message: err.message });
      else if (err instanceof ApiError && err.status === 422) setConflict({ kind: 'layout', message: err.message });
      else setConflict({ kind: 'error', message: err instanceof Error ? err.message : String(err) });
    } finally {
      if (!stale()) setBusy(false);
    }
  };

  const head = (
    <header class="issue-header">
      {(() => {
        const crumb = (
          <>
            {detail && <RepoAvatar repository={detail.repository} size={20} />}
            <span class="issue-header__path">
              {detail && (
                <>
                  <span class="issue-header__repo" title={detail.repository}>
                    {detail.repository.slice(detail.repository.indexOf('/') + 1)}
                  </span>
                  <span class="issue-header__sep" aria-hidden="true">
                    {' › '}
                  </span>
                </>
              )}
              <span class="identifier">{detail?.code ?? '…'}</span>
            </span>
          </>
        );
        // The breadcrumb is the way back where the page covers a list or the board.
        return onBack ? (
          <button type="button" class="issue-header__crumb issue-header__crumb--back" onClick={onBack} aria-label={backLabel} title={`${backLabel} (Escape)`}>
            {crumb}
          </button>
        ) : (
          <span class="issue-header__crumb">{crumb}</span>
        );
      })()}
      <span class="issue-header__spacer" />
      {onExpand && (
        <button type="button" class="icon-button" onClick={onExpand} aria-label="Open as a page" title="Open as a page">
          <Icon name="expand" />
        </button>
      )}
      {detail && (
        <Popover
          role="menu"
          label="Issue actions"
          align="end"
          trigger={(props) => (
            <button type="button" {...props} class="icon-button issue-header__menu" aria-label="Issue actions" title="Issue actions">
              <Icon name="dots" size={16} />
            </button>
          )}
        >
          {(close) => (
            <div class="issue-menu">
              {detail.url && (
                <button type="button" class="menu-item" role="menuitem" onClick={() => (copy(detail.url!, 'link'), close(true))}>
                  <Icon name="link" />
                  <span class="menu-item__text">Copy link</span>
                </button>
              )}
              <button type="button" class="menu-item" role="menuitem" onClick={() => (copy(detail.code, 'code'), close(true))}>
                <Icon name="hash" />
                <span class="menu-item__text">Copy issue code</span>
              </button>
              <button type="button" class="menu-item" role="menuitem" onClick={() => (copy(detail.provenance.path, 'path'), close(true))}>
                <Icon name="copy" />
                <span class="menu-item__text">Copy file path</span>
              </button>
              {detail.url && (
                <a class="menu-item" role="menuitem" href={detail.url} target="_blank" rel="noreferrer noopener" onClick={() => close(false)}>
                  <Icon name="external-link" />
                  <span class="menu-item__text">Open on GitHub</span>
                </a>
              )}
            </div>
          )}
        </Popover>
      )}
    </header>
  );

  commands.current = detail
    ? {
        id: detail.id,
        code: detail.code,
        title: detail.title,
        url: detail.url,
        path: detail.provenance.path,
        impact: impactOf(detail.labels),
        effort: effortOf(detail.labels),
        blocked: busy || reloading || conflict?.kind === 'changed' || detail.flags.duplicate_mirror,
        triage: (patch) => void triage(patch),
        copy: (text, what) => copy(text, what),
      }
    : null;

  if (error || !detail) {
    return (
      <div class={`issue-host issue-host--${variant}`}>
        <div class="issue issue--state">
          <div class="issue__article">
            {head}
            <div class="issue__scroll">{error ? <ErrorState error={error} onRetry={load} /> : <Loading shape="prose">Loading the task…</Loading>}</div>
          </div>
        </div>
      </div>
    );
  }

  const impact = impactOf(detail.labels);
  const effort = effortOf(detail.labels);
  const flags = flagList(detail);
  const labels = otherLabels(detail.labels);
  const repo = detail.repository.slice(detail.repository.indexOf('/') + 1);
  // A 412 blocks every triage action, the empty value included, until the reloaded task is in.
  const blocked = busy || reloading || conflict?.kind === 'changed' || detail.flags.duplicate_mirror;
  const feedbackOf = (of: 'impact' | 'effort') =>
    feedback?.of === of && (
      <span class="prop__feedback" role="status" aria-live="polite">
        {feedback.kind === 'saving' ? 'Saving…' : 'Saved'}
      </span>
    );

  return (
    <div class={`issue-host issue-host--${variant}`}>
      <article class="issue" aria-busy={busy}>
        <div class="issue__article">
          {head}
          <div class="issue__scroll">
            <div class="issue__main">
              <h2 class="issue__title">{detail.title}</h2>
              {flags.length > 0 && (
                <div class="issue__flags">
                  {flags.map((flag) => (
                    <MetaChip key={flag.title} tone="flag" title={flag.reason}>
                      <Icon name="flag" size={12} />
                      {flag.title}
                    </MetaChip>
                  ))}
                </div>
              )}
              <section class="prose" aria-label="Body">
                <div class="markdown" dangerouslySetInnerHTML={{ __html: withoutTitle(detail.body_html, detail.title) }} />
              </section>
              <section class="activity" aria-label="Activity">
                <h3 class="activity__title">Activity</h3>
                <ol class="activity__list">
                  <li class="activity__item">
                    <span class="activity__marker">
                      <Icon name="history" size={10} />
                    </span>
                    <span class="activity__line">
                      <span class="activity__value">{repo}</span> created the issue <span class="activity__time">{date(detail.created_at)}</span>
                    </span>
                  </li>
                  {detail.comments.map((comment) => (
                    <li class="activity__item activity__item--comment" key={comment.url}>
                      <span class="activity__marker">
                        <Icon name="message" size={10} />
                      </span>
                      <div class="activity__comment">
                        <div class="activity__meta">
                          <span class="activity__value">Comment</span> ·{' '}
                          <a href={comment.url} target="_blank" rel="noreferrer noopener">
                            {date(comment.at, true)}
                          </a>
                        </div>
                        <div class="markdown prose" dangerouslySetInnerHTML={{ __html: comment.html }} />
                      </div>
                    </li>
                  ))}
                  {detail.closed_at && (
                    <li class="activity__item">
                      <span class="activity__marker activity__marker--glyph">
                        <StatusGlyph column={detail.column === 'wont_do' ? 'wont_do' : 'done'} size={16} />
                      </span>
                      <span class="activity__line">
                        <span class="activity__value">{repo}</span> closed the issue
                        {detail.state_reason ? (
                          <>
                            {' '}
                            as <span class="activity__value">{detail.state_reason.replace('_', ' ')}</span>
                          </>
                        ) : null}{' '}
                        <span class="activity__time">{date(detail.closed_at)}</span>
                      </span>
                    </li>
                  )}
                </ol>
              </section>
            </div>
          </div>
        </div>

        <aside class="props" aria-label="Properties">
          <div class="rail-toolbar">
            <span class="rail-toolbar__code">{detail.code}</span>
            {detail.url && (
              <button type="button" class="icon-button" onClick={() => copy(detail.url!, 'link')} aria-label="Copy the link to the issue" title="Copy the GitHub link">
                <Icon name="link" />
              </button>
            )}
            <button type="button" class="icon-button" onClick={() => copy(detail.provenance.path, 'path')} aria-label="Copy the file path" title="Copy the file path">
              <Icon name="copy" />
            </button>
            {detail.url && <GitHubLink url={detail.url} />}
            {copied && (
              <span class="rail-toolbar__copied" role="status">
                Copied
              </span>
            )}
          </div>
          <div class="props__scroll">
            {conflict?.kind === 'changed' && (
              <div class="notice notice--warn" role="alert">
                <p>The file changed since it was loaded. Reload the task and try again.</p>
                {reloading ? (
                  <span class="prop__feedback" role="status" aria-live="polite">
                    Reloading…
                  </span>
                ) : (
                  <button type="button" class="button" onClick={load}>
                    Reload
                  </button>
                )}
              </div>
            )}
            {conflict?.kind === 'duplicate' && (
              <div class="notice notice--warn" role="alert">
                <p>This task is mirrored by more than one checkout, so the service does not know which file to edit. {conflict.message}</p>
              </div>
            )}
            {conflict?.kind === 'layout' && (
              <div class="notice notice--warn" role="alert">
                <p>The labels of this file are laid out in a way the service cannot edit; change them in the file. {conflict.message}</p>
              </div>
            )}
            {conflict?.kind === 'error' && (
              <div class="notice notice--error" role="alert">
                <p>{conflict.message}</p>
              </div>
            )}
            <dl class="props__list">
              <Prop name="Status">
                <span class="prop__value">
                  <StatusGlyph column={detail.column} />
                  <span class="prop__text">{COLUMN_TITLES[detail.column]}</span>
                </span>
              </Prop>
              <Prop name="Priority">
                <ChoiceMenu
                  name="Priority"
                  disabled={blocked}
                  current={impact}
                  empty="No priority"
                  options={IMPACTS.map((level) => ({ value: level, title: IMPACT_TITLE[level], glyph: <PriorityGlyph level={level} /> }))}
                  emptyGlyph={<PriorityGlyph level="none" />}
                  onPick={(value) => void triage({ impact: (value || null) as Impact | null })}
                />
                {feedbackOf('impact')}
              </Prop>
              <Prop name="Estimate">
                <ChoiceMenu
                  name="Estimate"
                  disabled={blocked}
                  current={effort}
                  empty="No estimate"
                  options={EFFORTS.map((level) => ({ value: level, title: ESTIMATE_SIZE[level], hint: EFFORT_TITLE[level].toLowerCase(), glyph: <EstimateGlyph /> }))}
                  emptyGlyph={<EstimateGlyph class="glyph--empty" />}
                  onPick={(value) => void triage({ effort: (value || null) as Effort | null })}
                />
                {feedbackOf('effort')}
              </Prop>
              <Prop name="Labels">
                {labels.length > 0 ? (
                  <span class="label-chips">
                    {labels.map((label) => (
                      <LabelChip key={label} label={label} />
                    ))}
                  </span>
                ) : (
                  <span class="prop__value prop__value--empty">No labels</span>
                )}
              </Prop>
              <Prop name="Type">
                <span class={`prop__value${detail.type ? '' : ' prop__value--empty'}`}>
                  <Icon name="category" />
                  <span class="prop__text">{detail.type ?? 'No type'}</span>
                </span>
              </Prop>
              <Prop name="Milestone">
                <span class={`prop__value${detail.milestone ? '' : ' prop__value--empty'}`} title={detail.milestone?.key}>
                  <Icon name="milestone" />
                  <span class="prop__text">{detail.milestone ? detail.milestone.title : 'No milestone'}</span>
                </span>
              </Prop>
              <Prop name="Assignees">
                <span class={`prop__value${detail.assignees.length ? '' : ' prop__value--empty'}`}>
                  <Icon name="user" />
                  <span class="prop__text">{detail.assignees.length > 0 ? detail.assignees.join(', ') : 'Unassigned'}</span>
                </span>
              </Prop>
              <Prop name="Repository">
                <span class="prop__value" title={detail.repository}>
                  <Icon name="repo" />
                  <span class="prop__text">{detail.repository}</span>
                </span>
              </Prop>
              <Prop name="State">
                <span class="prop__value">
                  <span class="prop__text">
                    {detail.state}
                    {detail.state_reason ? ` (${detail.state_reason.replace('_', ' ')})` : ''}
                  </span>
                </span>
              </Prop>
            </dl>

            <section class="props__section" aria-label="Score">
              <h3 class="props__heading">
                Score <span class="number">{detail.score.total.toFixed(2)}</span>
              </h3>
              <dl class="terms">
                {(
                  [
                    ['Impact', detail.score.impact],
                    ['Ease', detail.score.ease],
                    ['Type', detail.score.type],
                    ['Age', detail.score.age],
                    ['Due', detail.score.due],
                    ['Repo weight', detail.score.repo_weight],
                  ] as const
                ).map(([name, value]) => (
                  <Term key={name} name={name} value={value.toFixed(2)} />
                ))}
              </dl>
            </section>
            <section class="props__section" aria-label="Dates">
              <h3 class="props__heading">Dates</h3>
              <dl class="terms">
                <Term name="Created" value={date(detail.created_at)} />
                <Term name="Updated" value={date(detail.updated_at)} />
                <Term name="Closed" value={date(detail.closed_at)} />
              </dl>
            </section>
            <section class="props__section" aria-label="Provenance">
              <h3 class="props__heading">Provenance</h3>
              <dl class="terms terms--stacked">
                <dt>File</dt>
                <dd class="mono">{detail.provenance.path}</dd>
                <dt>Root</dt>
                <dd class="mono">{detail.provenance.root}</dd>
                <dt>Boards</dt>
                <dd>{detail.boards.length > 0 ? detail.boards.map((b) => `${b.board} · ${b.section}`).join('; ') : '—'}</dd>
              </dl>
            </section>
          </div>
        </aside>
      </article>
    </div>
  );
}

/** Open on GitHub as an icon link: the header of the page and of the peek panel. */
export function GitHubLink({ url }: { url: string }) {
  return (
    <a class="icon-button" href={url} target="_blank" rel="noreferrer noopener" aria-label="Open on GitHub" title="Open on GitHub">
      <Icon name="external-link" />
    </a>
  );
}

function Prop({ name, children }: { name: string; children: ComponentChildren }) {
  return (
    <div class="prop">
      <dt class="prop__name">{name}</dt>
      <dd class="prop__body">{children}</dd>
    </div>
  );
}

function Term({ name, value }: { name: string; value: string }) {
  return (
    <>
      <dt>{name}</dt>
      <dd class="number">{value}</dd>
    </>
  );
}

interface Choice {
  value: string;
  title: string;
  hint?: string;
  glyph: ComponentChildren;
}

interface ChoiceMenuProps {
  /** The property: the accessible name of the button and the menu. */
  name: string;
  current: string;
  /** The title of the empty value, which removes the label. */
  empty: string;
  emptyGlyph: ComponentChildren;
  options: Choice[];
  disabled: boolean;
  onPick: (value: string) => void;
}

/** A property value that opens a menu of its values, the current one checked. */
function ChoiceMenu({ name, current, empty, emptyGlyph, options, disabled, onPick }: ChoiceMenuProps) {
  const chosen = options.find((o) => o.value === current);
  const all: Choice[] = [{ value: '', title: empty, glyph: emptyGlyph }, ...options];
  return (
    <Popover
      role="menu"
      label={name}
      disabled={disabled}
      trigger={(props) => (
        <button type="button" {...props} class={`prop__button${chosen ? '' : ' prop__value--empty'}`} disabled={disabled} aria-label={`${name}: ${chosen ? chosen.title : empty}`} title={`Change ${name.toLowerCase()}`}>
          {chosen ? chosen.glyph : emptyGlyph}
          <span>{chosen ? chosen.title : empty}</span>
          {chosen?.hint && <span class="prop__hint">{chosen.hint}</span>}
        </button>
      )}
    >
      {(close) =>
        all.map((option) => {
          const checked = option.value === current;
          return (
            <button
              type="button"
              key={option.value || 'none'}
              class="menu-item"
              role="menuitemradio"
              aria-checked={checked}
              disabled={disabled || (checked && option.value === '')}
              onClick={() => {
                close(true);
                if (!checked) onPick(option.value);
              }}
            >
              <span class="menu-icon">{option.glyph}</span>
              <span class="menu-item__text">{option.title}</span>
              {option.hint && <span class="menu-item__hint">{option.hint}</span>}
              <span class="menu-item__check">{checked && <Icon name="check" size={14} />}</span>
            </button>
          );
        })
      }
    </Popover>
  );
}

/**
 * The body without a leading H1 that repeats the title: the mirrors start
 * the body with it, and the page shows the title above the body already.
 */
export function withoutTitle(html: string, title: string): string {
  const match = /^\s*<h1[^>]*>([\s\S]*?)<\/h1>\s*/i.exec(html);
  if (!match) return html;
  const text = match[1]
    .replace(/<[^>]+>/g, '')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&amp;/g, '&')
    .trim();
  return text === title.trim() ? html.slice(match[0].length) : html;
}

/** Human-readable flags of a task with the reason for a tooltip. */
export function flagList(task: { flags: TaskDetail['flags'] }): { title: string; reason: string }[] {
  const out: { title: string; reason: string }[] = [];
  if (task.flags.ambiguous) out.push({ title: 'ambiguous', reason: 'Its boards disagree about the column.' });
  if (task.flags.contradiction) out.push({ title: 'contradiction', reason: 'Open, but listed under Done or Won’t do.' });
  if (task.flags.duplicate_mirror) out.push({ title: 'duplicate mirror', reason: 'More than one checkout mirrors this task.' });
  if (task.flags.stale) out.push({ title: 'stale', reason: 'The file is malformed now; this is the previous reading.' });
  return out;
}

export function date(value: string | null, withTime = false): string {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  const day = d.toISOString().slice(0, 10);
  return withTime ? `${day} ${d.toISOString().slice(11, 16)}Z` : day;
}

/** The age of a date as the inbox shows it: 5m, 3h, 2d, 4w, 7mo, 2y. */
export function age(value: string | null, now: number = Date.now()): string {
  if (!value) return '—';
  const t = new Date(value).getTime();
  if (Number.isNaN(t)) return value;
  const minutes = Math.max(0, Math.round((now - t) / 60_000));
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.round(hours / 24);
  if (days < 14) return `${days}d`;
  if (days < 60) return `${Math.round(days / 7)}w`;
  if (days < 365) return `${Math.round(days / 30)}mo`;
  return `${Math.round(days / 365)}y`;
}
