'use client'

// The landing's illustrations of the local hub. They are drawn in HTML with the hub's own tokens
// (web/DESIGN.md) on the demo data of the pictures in public/images/hub, so they stay sharp at any
// size and read as the same product. Nothing here talks to a service: it is a picture you can click.
import { useEffect, useRef, useState } from 'react'

function useReducedMotion() {
  const [reduced, setReduced] = useState(false)
  useEffect(() => {
    const query = matchMedia('(prefers-reduced-motion: reduce)')
    const sync = () => setReduced(query.matches)
    sync()
    query.addEventListener('change', sync)
    return () => query.removeEventListener('change', sync)
  }, [])
  return reduced
}

/** True while (or, with `once`, as soon as) the element is at least `threshold` visible. */
function useInView(threshold = 0.3, once = true) {
  const ref = useRef(null)
  const [inView, setInView] = useState(false)
  useEffect(() => {
    const observer = new IntersectionObserver(([entry]) => {
      setInView(entry.isIntersecting)
      if (entry.isIntersecting && once) observer.disconnect()
    }, { threshold })
    observer.observe(ref.current)
    return () => observer.disconnect()
  }, [threshold, once])
  return [ref, inView]
}

/** Fades and lifts its children into place when they scroll into view. */
export function Reveal({ children, className = '' }) {
  const [ref, inView] = useInView(0.2)
  return <div ref={ref} className={`reveal ${className}`} data-inview={inView || undefined}>{children}</div>
}

/* ——— Glyphs of the hub: status, priority, navigation ——— */

const ring = { fill: 'none', strokeWidth: 1.6 }
function State({ kind }) {
  const body = {
    backlog: <circle cx="7" cy="7" r="5.6" {...ring} stroke="#858699" strokeDasharray="1.6 2" />,
    todo: <circle cx="7" cy="7" r="5.6" {...ring} stroke="#d2d3e0" />,
    progress: <><circle cx="7" cy="7" r="5.6" {...ring} stroke="#f2c94c" /><path d="M7 3.6a3.4 3.4 0 0 1 0 6.8z" fill="#f2c94c" /></>,
    done: <><circle cx="7" cy="7" r="6.4" fill="#5e6ad2" /><path d="m4.3 7.2 1.9 1.9 3.6-3.9" fill="none" stroke="#fff" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" /></>,
    wontdo: <><circle cx="7" cy="7" r="6.4" fill="#858699" /><path d="m4.9 4.9 4.2 4.2m0-4.2L4.9 9.1" stroke="#181921" strokeWidth="1.6" strokeLinecap="round" /></>,
    uncommitted: <><circle cx="7" cy="7" r="5.6" {...ring} stroke="#f2c94c" /><circle cx="7" cy="7" r="2.4" fill="#f2c94c" /></>,
    diverged: <><circle cx="7" cy="7" r="5.6" {...ring} stroke="#4ea7fc" /><circle cx="7" cy="7" r="2.4" fill="#4ea7fc" /></>,
  }
  return <svg className="hub-glyph" viewBox="0 0 14 14" aria-hidden="true">{body[kind === 'clean' ? 'done' : kind]}</svg>
}

function Bars({ level = 3 }) {
  return (
    <svg className="hub-glyph" viewBox="0 0 14 14" aria-hidden="true">
      {[0, 1, 2].map(i => <rect key={i} x={2 + i * 4} y={9 - i * 3} width="2.6" height={3 + i * 3} rx="0.8" fill="currentColor" opacity={i < level ? 1 : 0.3} />)}
    </svg>
  )
}

const icons = {
  inbox: 'M2.5 3.5h9v7h-9zM2.5 8h2.6l.8 1.4h2.2L8.9 8h2.6',
  tablo: 'M2.5 3h3.4v8H2.5zM8.1 3h3.4v4.6H8.1z',
  status: 'M4 2.6v8.8M4 8.4c3.6 0 6-1 6-3.6M4 4a1.4 1.4 0 1 0 0-2.8A1.4 1.4 0 0 0 4 4zm6 1a1.4 1.4 0 1 0 0-2.8A1.4 1.4 0 0 0 10 5zm-6 7.8A1.4 1.4 0 1 0 4 10a1.4 1.4 0 0 0 0 2.8z',
  reindex: 'M11.5 7a4.5 4.5 0 0 1-8 2.8M2.5 7a4.5 4.5 0 0 1 8-2.8M10.6 1.8v2.6H8M3.4 12.2V9.6H6',
  search: 'M6.2 10.4a4.2 4.2 0 1 0 0-8.4 4.2 4.2 0 0 0 0 8.4zm3-1.2 2.8 2.8',
  lock: 'M4 6.4h6v5H4zM5.2 6.4V4.6a1.8 1.8 0 0 1 3.6 0v1.8',
  list: 'M2.5 4h9M2.5 7h9M2.5 10h9',
  view: 'M2.5 4.4h9M2.5 9.6h9M5 3v2.8M9 8.2V11',
}
function Icon({ name }) {
  return <svg className="hub-glyph" viewBox="0 0 14 14" aria-hidden="true"><path d={icons[name]} fill="none" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" /></svg>
}

/* ——— Demo data, the same as in the pictures of the service guide ——— */

const tree = [
  { owner: 'alex-rivera', count: 1 },
  { owner: 'lumen-labs', count: 2, repos: [['design-tokens', 4], ['lumen', 6]] },
  { owner: 'northwind', count: 3, repos: [['billing-api', 7, true], ['infra', 5, true], ['web-console', 6, true]] },
]

const checkouts = [
  { state: 'uncommitted', title: 'Uncommitted', rows: [
    ['↑1', 'lumen-labs/lumen', 'fix/render-cache', '+7/-0 · 2f', 'ahead 1', 'public/lumen-labs/lumen'],
    ['↑2', 'northwind/billing-api', 'main*', '+7/-0 · 2f', 'ahead 2', 'private/northwind/billing-api'],
  ] },
  { state: 'diverged', title: 'Ahead or behind', rows: [
    ['↓2', 'alex-rivera/notes-cli', 'main*', '+0/-0', 'behind 2', 'public/alex-rivera/notes-cli'],
    ['↑1 ↓3', 'northwind/mobile-app', 'feat/offline-sync', '+0/-0', 'ahead 1 · behind 3', 'private/northwind/mobile-app'],
    ['↓5', 'northwind/web-console', 'main*', '+0/-0', 'behind 5', 'private/northwind/web-console'],
  ] },
  { state: 'clean', title: 'Clean', rows: [
    ['', 'alex-rivera/dotfiles', 'main*', '+0/-0', 'synced', 'public/alex-rivera/dotfiles'],
    ['', 'alex-rivera/homelab', 'main*', '+0/-0', 'synced', 'private/alex-rivera/homelab'],
    ['', 'lumen-labs/design-tokens', 'main*', '+0/-0', 'synced', 'public/lumen-labs/design-tokens'],
    ['', 'lumen-labs/lumen-docs', 'main*', '+0/-0', 'synced', 'public/lumen-labs/lumen-docs'],
    ['lock', 'northwind/infra', 'main*', '+0/-0', 'synced', 'private/northwind/infra'],
    ['', 'northwind/status-page', 'main*', '+0/-0', 'synced', 'public/northwind/status-page'],
  ] },
]

const repoColor = { B: '#b8891b', I: '#8f8a2a', D: '#7a5cc4', L: '#b06a2c', W: '#3f9a63', N: '#1f9aa5' }
const issues = [
  ['B', 'BILL-1', 'billing: retry webhooks with exponential backoff', 'billing-api · Billing 2.4', 'progress', '1d'],
  ['B', 'BILL-2', 'billing: idempotency keys for POST /charges', 'billing-api · Billing 2.4', 'todo', '3d'],
  ['I', 'INFRA-3', 'infra: staging nodes run out of disk', 'infra · Platform upgrade, November', 'todo', '3h'],
  ['B', 'BILL-8', 'billing: rate limit the public API per account', 'billing-api · Billing 2.5', 'backlog', '2w'],
  ['I', 'INFRA-2', 'infra: rotate the database credentials', 'infra · Platform upgrade, November', 'todo', '8d'],
  ['D', 'TOK-1', 'tokens: export tokens as CSS custom properties', 'design-tokens · Tokens v3', 'todo', '5d'],
  ['L', 'LUM-2', 'lumen: render cache is not invalidated on partial edits', 'lumen · Lumen 1.0', 'progress', '1d'],
  ['W', 'CON-3', 'console: session expires while editing', 'web-console · Console: Q4 polish', 'todo', '1d'],
  ['I', 'INFRA-1', 'infra: upgrade the clusters to Kubernetes 1.34', 'infra · Platform upgrade, November', 'progress', '2d'],
]

const board = [
  { state: 'backlog', title: 'Backlog', count: 16, cards: [
    ['B', 'BILL-8', 'billing: rate limit the public API per account', 'M', 'feature'],
    ['B', 'BILL-5', 'billing: export usage records as CSV', 'S', 'feature'],
    ['B', 'BILL-6', 'billing: drop the legacy v1 Stripe adapter', 'M', 'feature'],
    ['L', 'LUM-8', 'lumen: Windows path separators in permalinks', 'M', 'bug'],
  ] },
  { state: 'todo', title: 'To do', count: 10, cards: [
    ['B', 'BILL-4', 'billing: invoice PDF renders wrong currency symbol for CHF', 'S', 'bug'],
    ['B', 'BILL-2', 'billing: idempotency keys for POST /charges', 'L', 'feature'],
    ['I', 'INFRA-3', 'infra: staging nodes run out of disk', 'M', 'bug'],
  ] },
  { state: 'progress', title: 'In progress', count: 6, cards: [
    ['B', 'BILL-1', 'billing: retry webhooks with exponential backoff', 'M', 'bug'],
    ['L', 'LUM-2', 'lumen: render cache is not invalidated on partial edits', 'M', 'bug'],
    ['I', 'INFRA-1', 'infra: upgrade the clusters to Kubernetes 1.34', 'L', 'feature'],
  ] },
  { state: 'done', title: 'Done', count: 8, cards: [
    ['B', 'BILL-7', 'billing: tax rates are cached for too long', 'S', 'bug'],
    ['B', 'BILL-9', 'billing: dunning emails for failed renewals', 'M', 'feature'],
    ['W', 'CON-5', 'console: empty state for new workspaces', 'S', 'feature'],
  ] },
  { state: 'wontdo', title: 'Won’t do', count: 1, cards: [
    ['W', 'CON-8', 'console: replace moment.js with date-fns', 'M', 'feature'],
  ] },
]

function Avatar({ letter }) {
  return <span className="hub-avatar" style={{ '--tone': repoColor[letter] }} aria-hidden="true">{letter}</span>
}

/* ——— The three pages ——— */

export function StatusList({ groups = checkouts }) {
  return (
    <div className="hub-status">
      {groups.map(group => (
        <div key={group.title}>
          <div className="hub-group"><State kind={group.state} /><b>{group.title}</b><span>{group.rows.length}</span></div>
          {group.rows.map(([sync, repo, branch, changes, status, path]) => (
            <div className="hub-row" key={repo}>
              <span className="hub-sync">{sync === 'lock' ? <Icon name="lock" /> : sync}</span>
              <State kind={group.state} />
              <span className="hub-repo">{repo}</span>
              <span className="hub-branch">{branch}</span>
              <span className="hub-changes">{changes}</span>
              <span className="hub-ahead">{status}</span>
              <span className="hub-path">{path}</span>
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

function InboxPage() {
  return (
    <div className="hub-inbox">
      <div className="hub-inbox-list">
        <p className="hub-count">32 open tasks</p>
        <div className="hub-group"><Bars /><b>High</b><span>12</span></div>
        {issues.map(([letter, id, title, meta, state, age], i) => (
          <div className="hub-issue" key={id} data-selected={i === 0 || undefined}>
            <Avatar letter={letter} />
            <span className="hub-issue-text"><span><i>{id}</i> {title}</span><small>{meta}</small></span>
            <span className="hub-issue-side"><State kind={state} /><small>{age}</small></span>
          </div>
        ))}
      </div>
      <div className="hub-detail">
        <p className="hub-crumb"><Avatar letter="B" /> billing-api › BILL-1</p>
        <h4>billing: retry webhooks with exponential backoff</h4>
        <p>Webhook deliveries to customer endpoints are attempted once. A timeout or a <code>5xx</code> from the receiver drops the event, and support has to replay it by hand.</p>
        <h5>Proposal</h5>
        <ul>
          <li>Retry with exponential backoff: 30 s, 2 min, 10 min, 1 h, 6 h, then give up.</li>
          <li>Add jitter of ±20% so a recovering receiver is not hit by a wave.</li>
        </ul>
        <h5>Acceptance criteria</h5>
        <ul className="hub-checks">
          <li>A failed delivery is retried at least five times over six hours</li>
          <li data-done>The receiver gets an <code>Idempotency-Key</code> header that stays the same across retries</li>
          <li>An alert fires when more than 1% of events exhaust their retries</li>
        </ul>
      </div>
      <dl className="hub-rail">
        <div><dt>Status</dt><dd><State kind="progress" /> In progress</dd></div>
        <div><dt>Priority</dt><dd><Bars /> High</dd></div>
        <div><dt>Estimate</dt><dd><b>M</b> medium</dd></div>
        <div><dt>Labels</dt><dd><span className="hub-chip"><i style={{ background: '#f2c94c' }} />type: bug</span></dd></div>
        <div><dt>Milestone</dt><dd>Billing 2.4</dd></div>
        <div><dt>Repository</dt><dd>northwind/billing-api</dd></div>
        <div className="hub-rail-score"><dt>Score</dt><dd>4.95</dd></div>
      </dl>
    </div>
  )
}

export function Board({ columns = board }) {
  return (
    <div className="hub-board" style={{ '--columns': columns.length }}>
      {columns.map(column => (
        <div className="hub-column" key={column.title}>
          <div className="hub-column-head"><State kind={column.state} /><b>{column.title}</b><span>{column.count}</span></div>
          {column.cards.map(([letter, id, title, estimate, type]) => (
            <div className="hub-card" key={id}>
              <p>{title}<Avatar letter={letter} /></p>
              <span><Bars /><b>{estimate}</b><i>{id}</i><span className="hub-type" data-type={type}>type: {type}</span></span>
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

const pages = [
  { id: 'status', title: 'Status', line: 'Where local work waits' },
  { id: 'inbox', title: 'Inbox', line: 'What matters next' },
  { id: 'tablo', title: 'Tablo', line: 'Where everything stands' },
]

/** The hero: the hub in its window. It walks through its three pages until you pick one. */
export function HubStage() {
  const reduced = useReducedMotion()
  const [ref, inView] = useInView(0.35, false)
  const [page, setPage] = useState('status')
  const [taken, setTaken] = useState(false)

  useEffect(() => {
    if (taken || reduced || !inView) return
    const order = pages.map(p => p.id)
    const timer = setTimeout(() => setPage(order[(order.indexOf(page) + 1) % order.length]), 5200)
    return () => clearTimeout(timer)
  }, [page, taken, reduced, inView])

  const pick = id => { setPage(id); setTaken(true) }
  const current = pages.find(p => p.id === page)

  return (
    <div className="stage-wrap" ref={ref}>
      <div className="stage-tabs" role="tablist" aria-label="Pages of the hub">
        {pages.map(p => (
          <button key={p.id} type="button" role="tab" aria-selected={p.id === page} data-running={(p.id === page && !taken && !reduced && inView) || undefined} onClick={() => pick(p.id)}>
            <b>/{p.id}</b><span>{p.line}</span>
          </button>
        ))}
      </div>
      <div className="stage">
        <div className="hub" role="img" aria-label={`The ${current.title} page of the hub on demo data`}>
          <div className="hub-chrome"><span className="hub-dots"><i /><i /><i /></span><span className="hub-url">maintainer.localhost:17841/{page}</span></div>
          <div className="hub-app">
            <aside className="hub-sidebar">
              <p className="hub-brand"><span className="hub-mark">M</span>maintainer</p>
              <p className="hub-actions"><span><Icon name="reindex" /> Reindex</span><span><Icon name="search" /></span></p>
              <nav>
                {['inbox', 'tablo', 'status'].map(id => (
                  <button key={id} type="button" tabIndex={-1} data-active={id === page || undefined} onClick={() => pick(id)}><Icon name={id} />{id[0].toUpperCase() + id.slice(1)}</button>
                ))}
              </nav>
              <p className="hub-label">Repositories</p>
              <ul className="hub-tree">
                {tree.map(({ owner, count, repos }) => (
                  <li key={owner}>
                    <span><i>{repos ? '⌄' : '›'}</i>{owner}<em>{count}</em></span>
                    {repos && <ul>{repos.map(([name, open, locked]) => <li key={name}><span>{name}{locked && <Icon name="lock" />}<em>{open}</em></span></li>)}</ul>}
                  </li>
                ))}
              </ul>
            </aside>
            <div className="hub-main">
              <div className="hub-topbar">
                <b>{current.title}</b><span className="hub-filter">+ Filter</span>
                <span className="hub-tools"><span data-on={page !== 'tablo' || undefined}><Icon name="list" /></span><span data-on={page === 'tablo' || undefined}><Icon name="tablo" /></span><span className="hub-view"><Icon name="view" /> View</span></span>
              </div>
              <div className="hub-page" key={page}>
                {page === 'status' && <StatusList />}
                {page === 'inbox' && <InboxPage />}
                {page === 'tablo' && <Board />}
              </div>
            </div>
          </div>
        </div>
      </div>
      <p className="stage-note">The hub on demo data, drawn for this page. The real one runs on your computer at <code>maintainer.localhost:17841</code> and reads your own checkouts.</p>
    </div>
  )
}

/* ——— Section illustrations ——— */

const impact = [['High', 1, 3], ['Medium', 0.5, 2], ['Low', 0.25, 1]]
const effort = [['S', 'easy', 1], ['M', 'medium', 0.5], ['L', 'hard', 0.25]]

/** The score of one issue: pick a priority and an estimate, and the sum follows the guide's formula. */
export function ScoreDemo() {
  const [priority, setPriority] = useState(0)
  const [estimate, setEstimate] = useState(1)
  const weight = 1.5
  const terms = [['Impact', impact[priority][1]], ['Ease', effort[estimate][2]], ['Type', 1], ['Age', 0.13], ['Due', 0.67]]
  const score = weight * terms.reduce((sum, [, value]) => sum + value, 0)
  return (
    <figure className="hub-figure score-demo">
      <div className="hub-issue" data-selected>
        <Avatar letter="B" />
        <span className="hub-issue-text"><span><i>BILL-1</i> billing: retry webhooks with exponential backoff</span><small>billing-api · Billing 2.4 · type: bug</small></span>
        <span className="hub-issue-side"><State kind="progress" /></span>
      </div>
      <div className="score-pickers">
        <div role="group" aria-label="Priority">
          <span>Priority</span>
          {impact.map(([name, , level], i) => <button key={name} type="button" aria-pressed={i === priority} onClick={() => setPriority(i)}><Bars level={level} />{name}</button>)}
        </div>
        <div role="group" aria-label="Estimate">
          <span>Estimate</span>
          {effort.map(([name, word], i) => <button key={name} type="button" aria-pressed={i === estimate} onClick={() => setEstimate(i)}><b>{name}</b>{word}</button>)}
        </div>
      </div>
      <dl className="score-terms">
        {terms.map(([name, value]) => (
          <div key={name}><dt>{name}</dt><dd><span style={{ '--value': value }} /></dd><dd>{value.toFixed(2)}</dd></div>
        ))}
        <div className="score-weight"><dt>Repo weight</dt><dd /><dd>× {weight.toFixed(2)}</dd></div>
      </dl>
      <p className="score-total" aria-live="polite"><span>Score</span><b>{score.toFixed(2)}</b></p>
      <figcaption>Every weight is 1 by default; this repository weighs 1.5 in <code>service.toml</code>.</figcaption>
    </figure>
  )
}

export function StatusFigure() {
  return <figure className="hub-figure"><StatusList groups={checkouts.map((g, i) => (i === 2 ? { ...g, rows: g.rows.slice(3) } : g))} /></figure>
}

export function BoardFigure() {
  return <figure className="hub-figure"><Board columns={board.slice(1, 4)} /></figure>
}

// A real run of `suggest --target 25 git/3`: three weeks, the suggested day in brackets.
const weeks = [['#04', [25, 25, 25, 25, 25, 25, 25]], ['#05', [25, 20, 15, 15, 15, 15, 15]], ['#06', [15, 15, 15, 15, 15, 15, 15]]]
const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

export function CalendarFigure() {
  return (
    <figure className="hub-figure calendar-figure">
      <p className="calendar-command"><span>$</span> maintainer github contribution suggest --target 25 git/3</p>
      <div className="calendar-grid" role="img" aria-label="Three weeks of contributions: week 4 at 25 a day, then 15 a day; Monday, 26 January, with 20 is the suggested day">
        <span />{days.map(day => <span key={day} className="calendar-day">{day}</span>)}
        {weeks.map(([week, counts]) => [
          <span key={week} className="calendar-week">{week}</span>,
          ...counts.map((count, i) => <span key={week + i} className="calendar-cell" data-level={count >= 25 ? 3 : count >= 20 ? 2 : 1} data-suggested={(week === '#05' && i === 1) || undefined}>{count}</span>),
        ])}
      </div>
      <p className="calendar-result">Suggestion is <b>2026-01-26T05:53:14Z</b>, 20 → 25</p>
      <figcaption>A real run in a checkout whose latest commit is dated 12 January 2026. The suggested moment is never later than now; your counts and time will differ.</figcaption>
    </figure>
  )
}

/** A command with a copy button. */
export function Command({ children, label }) {
  const [state, setState] = useState('')
  const copy = async () => {
    try { await navigator.clipboard.writeText(children); setState('Copied') }
    catch { setState('Select and copy') }
    setTimeout(() => setState(''), 2000)
  }
  return (
    <div className="install-command">
      <code><span aria-hidden="true">$ </span>{children}</code>
      <button type="button" onClick={copy} aria-label={`Copy: ${label}`}>{state || 'Copy'}</button>
    </div>
  )
}
