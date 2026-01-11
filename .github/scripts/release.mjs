#!/usr/bin/env node
// Release guardrails shared by the pre-push hook, make targets and CI.
// Zero dependencies: Node and git only.
//
//   release.mjs check  <tag> [--local] [--sha <commit>] [--pushed <branch>=<sha>]...
//   release.mjs render <tag> --site-url <url> --out <file>   (prints the title)
//   release.mjs preflight                                    (CI: required secrets are present and usable)
//   release.mjs doctor                                       (on demand: config vs GitHub, with remediation)
//   release.mjs pages  <base-url>                            (CI: the Pages URL matches settings.json pages)
//   release.mjs smoke  <site-url> [--retries <n>]            (CI: the deployed site serves pages and assets)
//   release.mjs vanity                                       (vanity imports in go.mod resolve over verified HTTPS)

import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'

const SETTINGS = '.github/settings.json'
const DEFAULTS = {
  tag_pattern: '^v\\d+\\.\\d+\\.\\d+(-[0-9A-Za-z.-]+)?$',
  notes: 'docs/content/changelog/{tag}.md',
  branches: [],
}

// --- helpers -----------------------------------------------------------------

function git(...args) {
  return execFileSync('git', args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] })
}

function tryGit(...args) {
  try { return git(...args) } catch { return null }
}

function gh(...args) {
  const env = typeof args.at(-1) === 'object' ? { ...process.env, ...args.pop() } : process.env
  try {
    return execFileSync('gh', args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'], env })
  } catch (e) {
    const err = new Error((e.stderr || e.message || '').trim())
    err.status = e.status
    throw err
  }
}

function settings(sha) {
  const raw = sha ? tryGit('show', `${sha}:${SETTINGS}`) : (existsSync(SETTINGS) ? readFileSync(SETTINGS, 'utf8') : null)
  const cfg = raw ? JSON.parse(raw) : {}
  return { release: { ...DEFAULTS, ...(cfg['x-release'] || {}) }, secrets: cfg.secrets || {}, pages: cfg.pages || null }
}

function args(argv) {
  const out = { _: [], pushed: {} }
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a === '--pushed') {
      const [b, s] = argv[++i].split('=')
      out.pushed[b] = s
    } else if (a === '--local') {
      out.local = true
    } else if (a.startsWith('--')) {
      out[a.slice(2)] = argv[++i]
    } else {
      out._.push(a)
    }
  }
  return out
}

// Splits a Markdown document into frontmatter and body, parsing `key: value` lines.
function frontmatter(text) {
  const m = text.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n/)
  if (!m) return { meta: null, body: text }
  const meta = {}
  for (const line of m[1].split(/\r?\n/)) {
    const kv = line.match(/^([A-Za-z_][\w-]*):\s*(.*)$/)
    if (!kv) continue
    let v = kv[2].trim()
    if (/^".*"$/.test(v)) v = JSON.parse(v)
    else if (/^'.*'$/.test(v)) v = v.slice(1, -1).replace(/''/g, "'")
    meta[kv[1]] = v
  }
  return { meta, body: text.slice(m[0].length) }
}

// Calls fn for every piece of prose, skipping fenced code blocks and inline code spans.
function mapProse(body, fn) {
  let fence = null
  return body.split('\n').map((line, i) => {
    const f = line.match(/^\s{0,3}(`{3,}|~{3,})/)
    if (fence) {
      if (f && f[1][0] === fence[0] && f[1].length >= fence.length) fence = null
      return line
    }
    if (f) { fence = f[1]; return line }
    return line.split(/(`+[^`]*`+)/).map((part, j) => (j % 2 ? part : fn(part, i + 1))).join('')
  }).join('\n')
}

function releaseBranch(tag, release, defaultBranch) {
  for (const { match, branch } of release.branches) {
    const m = tag.match(new RegExp(match))
    if (m) return branch.replace(/\$(\d+)/g, (_, n) => m[n] ?? '')
  }
  return defaultBranch
}

function defaultBranch() {
  if (process.env.DEFAULT_BRANCH) return process.env.DEFAULT_BRANCH
  const head = tryGit('symbolic-ref', '--short', 'refs/remotes/origin/HEAD')
  return head ? head.trim().replace(/^origin\//, '') : null
}

// --- check -------------------------------------------------------------------

function check(tag, opts) {
  const errors = []
  const sha = (opts.sha || tryGit('rev-parse', `${tag}^{commit}`) || '').trim()
  if (!sha) return [`${tag}: tag not found; create it first`]

  const { release } = settings(sha)
  if (!new RegExp(release.tag_pattern).test(tag)) {
    errors.push(`${tag}: does not match x-release.tag_pattern ${release.tag_pattern} in ${SETTINGS}`)
  }

  const note = release.notes.replace('{tag}', tag)
  const committed = tryGit('show', `${sha}:${note}`)
  if (committed === null) {
    errors.push(`${note}: missing in ${tag}; write the release note, commit it and move the tag`)
  } else {
    const staged = tryGit('show', `:${note}`)
    const worktree = existsSync(note) ? readFileSync(note, 'utf8') : null
    if (staged !== committed || worktree !== committed) {
      errors.push(`${note}: has changes not in ${tag}; commit them and move the tag`)
    }
    errors.push(...lint(note, committed))
  }

  const base = defaultBranch()
  const branch = releaseBranch(tag, release, base)
  if (!branch) {
    errors.push(`${tag}: cannot resolve the default branch; run: git remote set-head origin --auto`)
  } else {
    // Local preparation may precede a branch push; hooks and CI validate the
    // actual push transaction or the fetched remote branch instead.
    const ref = opts.local ? `refs/heads/${branch}` : `refs/remotes/origin/${branch}`
    const tip = opts.pushed[branch] || tryGit('rev-parse', '--verify', '-q', ref)?.trim()
    if (!tip) {
      errors.push(`${tag}: branch ${branch} is unknown; push it together with the tag: git push --atomic origin ${branch} ${tag}`)
    } else if (tryGit('merge-base', '--is-ancestor', sha, tip) === null) {
      errors.push(`${tag}: its commit is not on ${branch}; push both at once: git push --atomic origin ${branch} ${tag}`)
    }
  }
  return errors
}

function lint(file, text) {
  const errors = []
  const { meta, body } = frontmatter(text)
  if (!meta) return [`${file}: no frontmatter; add title and description`]
  for (const key of ['title', 'description']) {
    if (!meta[key]) errors.push(`${file}: frontmatter has no ${key}`)
  }
  const first = body.split('\n').find((l) => l.trim() !== '')
  if (!first || first.trim() !== `# ${meta.title}`) {
    errors.push(`${file}: the first line must be the H1 "# ${meta.title}", as in frontmatter title`)
  }
  const offset = text.slice(0, text.length - body.length).split('\n').length - 1
  let h1 = 0
  mapProse(body, (prose, n) => {
    const line = n + offset
    if (/^# /.test(prose)) h1++
    if (/^\s*(import|export)\s/.test(prose)) errors.push(`${file}:${line}: MDX import/export is not supported in release notes`)
    if (/<[A-Z][\w.]*[\s/>]/.test(prose)) errors.push(`${file}:${line}: MDX components are not supported in release notes`)
    return prose
  })
  if (h1 > 1) errors.push(`${file}: more than one H1`)
  return errors
}

// --- render ------------------------------------------------------------------

function render(tag, opts) {
  if (!opts['site-url'] || !opts.out) throw new Error('render needs --site-url and --out')
  const { release } = settings()
  const note = release.notes.replace('{tag}', tag)
  const text = readFileSync(note, 'utf8')
  const errors = lint(note, text)
  if (errors.length) throw new Error(errors.join('\n'))

  const { meta, body } = frontmatter(text)
  const site = opts['site-url'].replace(/\/+$/, '')
  const abs = (url) => `${site}${url}`
  const out = mapProse(body.replace(/^\s*# .*\n+/, ''), (prose) => prose
    .replace(/(\]\()(\/(?!\/)[^)\s]*)/g, (_, p, url) => p + abs(url))           // [text](/path) and ![alt](/path)
    .replace(/^(\s*\[[^\]]+\]:\s*)(\/(?!\/)\S*)/, (_, p, url) => p + abs(url)) // [ref]: /path
    .replace(/(\s(?:href|src)=["'])(\/(?!\/)[^"']*)/g, (_, p, url) => p + abs(url)))
  writeFileSync(opts.out, out.trimEnd() + '\n')
  process.stdout.write(meta.title + '\n')
}

// --- pages and smoke ---------------------------------------------------------

// owner/name from GITHUB_REPOSITORY or the origin remote.
function repoSlug() {
  if (process.env.GITHUB_REPOSITORY) return process.env.GITHUB_REPOSITORY
  const m = tryGit('remote', 'get-url', 'origin')?.trim().match(/github\.com[:/]([^/]+\/[^/]+?)(?:\.git)?$/)
  return m ? m[1] : null
}

// The URL the site must be built for: the custom domain, or the default one.
function expectedSite(pages, slug) {
  if (pages?.cname) return `https://${pages.cname}/`
  if (!slug) return null
  const [owner, name] = slug.toLowerCase().split('/')
  return name === `${owner}.github.io` ? `https://${name}/` : `https://${owner}.github.io/${name}/`
}

const slash = (url) => url.replace(/\/*$/, '/')

// The build bakes the base path and the origin in; a domain changed in Settings
// but not in settings.json (or the reverse) must stop the build, not ship it.
function pagesCheck(baseUrl) {
  const expected = expectedSite(settings().pages, repoSlug())
  if (!baseUrl) return ['no Pages base URL; is actions/configure-pages set up?']
  if (!expected) return ['cannot resolve the repository; set GITHUB_REPOSITORY or the origin remote']
  if (slash(baseUrl) === expected) return []
  return [`Pages serves ${slash(baseUrl)}, ${SETTINGS} expects ${expected}; ` +
    'update pages.cname there or Settings → Pages → Custom domain so they agree']
}

async function probe(url, init = {}) {
  try {
    const res = await fetch(url, { redirect: 'manual', signal: AbortSignal.timeout(15000), ...init })
    return { status: res.status, type: res.headers.get('content-type') || '', location: res.headers.get('location'), text: init.method === 'HEAD' ? '' : await res.text() }
  } catch (e) {
    return { status: 0, type: '', location: null, text: '', error: e.cause?.code || e.message }
  }
}

async function smokeOnce(site) {
  const errors = []
  const expect = (cond, msg) => { if (!cond) errors.push(msg) }
  const show = (r) => r.error || `HTTP ${r.status}`

  const home = await probe(site)
  expect(home.status === 200 && home.type.includes('text/html'), `${site}: ${show(home)}, want an HTML page`)
  if (home.status !== 200) return errors

  // A stale base path is exactly this: the page loads, its assets do not.
  const assets = [...new Set(home.text.match(/(?:href|src)="([^"]*\/_next\/static\/[^"]+\.(?:css|js))"/g) || [])]
    .map((a) => new URL(a.replace(/^(?:href|src)="|"$/g, ''), site).href)
  expect(assets.length > 0, `${site}: no /_next/static assets referenced`)
  for (const url of [assets.find((a) => a.endsWith('.css')), assets.find((a) => a.endsWith('.js'))].filter(Boolean)) {
    const r = await probe(url, { method: 'HEAD' })
    expect(r.status === 200, `${url}: ${show(r)}; the site was built for another base path, rebuild it`)
  }

  const og = home.text.match(/<meta[^>]+property="og:image"[^>]+content="([^"]+)"/)?.[1]
  if (og) {
    expect(og.startsWith(site), `og:image ${og} is outside ${site}; SITE_URL is stale, rebuild the site`)
    const r = await probe(og, { method: 'HEAD' })
    expect(r.status === 200 && r.type.startsWith('image/'), `${og}: ${show(r)}, want an image`)
  }

  const nested = await probe(new URL('changelog/', site).href)
  expect(nested.status === 200, `${site}changelog/: ${show(nested)}`)

  const missing = await probe(new URL('smoke-missing-page/', site).href)
  expect(missing.status === 404 && missing.text.includes('/_next/static/'), `${site}smoke-missing-page/: ${show(missing)}, want the site's own 404`)

  // With a custom domain, the default one only redirects, keeping the path.
  const slug = repoSlug()
  const fallback = expectedSite(null, slug)
  if (fallback && fallback !== site) {
    const r = await probe(new URL('changelog/', fallback).href)
    expect([301, 302, 308].includes(r.status) && r.location === new URL('changelog/', site).href,
      `${fallback}changelog/: ${show(r)} to ${r.location}, want a redirect to ${site}changelog/`)
  }
  return errors
}

// Pages may serve the previous deployment for a while after deploy-pages returns.
async function smoke(site, retries) {
  if (!Number.isInteger(retries) || retries < 0) return [`--retries must be a non-negative integer, got ${retries}`]
  let errors = []
  for (let i = 0; i <= retries; i++) {
    errors = await smokeOnce(slash(site))
    if (!errors.length || i === retries) break
    await new Promise((r) => setTimeout(r, 15000))
  }
  return errors
}

// --- vanity imports ----------------------------------------------------------

// GOPRIVATE sends these hosts past the proxy: `go` fetches ?go-get=1 over
// verified HTTPS itself, so a lapsed certificate or a missing go-import tag
// breaks every fresh runner while warm caches keep hiding it.
const VANITY = ['go.octolab.org']

function vanityModules(files = ['go.mod', 'tools/go.mod']) {
  const own = new Set()
  const deps = new Set()
  for (const file of files.filter((f) => existsSync(f))) {
    const text = readFileSync(file, 'utf8')
    const mod = text.match(/^module\s+(\S+)/m)?.[1]
    if (mod) own.add(mod)
    for (const [, path] of text.matchAll(/^\s*(?:require\s+)?([a-z0-9.-]+\.[a-z]+(?:\/\S*)?)\s+v\S+/gm)) {
      if (VANITY.some((h) => path === h || path.startsWith(`${h}/`))) deps.add(path)
    }
  }
  return [...deps].filter((p) => !own.has(p)).sort()
}

async function vanityCheck(path) {
  let res
  try {
    res = await fetch(`https://${path}?go-get=1`, { redirect: 'follow', signal: AbortSignal.timeout(15000) })
  } catch (e) {
    return `https://${path}: ${e.cause?.code || e.message}; the host serves no valid certificate for its name, ` +
      'check Custom domain and Enforce HTTPS in the Pages settings of the repository that hosts it'
  }
  if (res.status !== 200) return `https://${path}?go-get=1: HTTP ${res.status}, want 200`
  const html = await res.text()
  const metas = [...html.matchAll(/<meta\s+name="go-import"\s+content="([^"]+)"/g)].map((m) => m[1].trim().split(/\s+/))
  const hit = metas.find(([prefix]) => path === prefix || path.startsWith(`${prefix}/`))
  if (!hit) return `https://${path}?go-get=1: no go-import meta tag for ${path}`
  const [prefix, vcs, repo] = hit
  if (vcs !== 'git' || !/^https:\/\//.test(repo || '')) return `${path}: go-import "${hit.join(' ')}", want "${prefix} git https://..."`
  return null
}

// [[path, error or null], ...] for every vanity module go.mod and tools/go.mod require.
async function vanity() {
  return Promise.all(vanityModules().map(async (p) => [p, await vanityCheck(p)]))
}

// --- preflight and doctor ----------------------------------------------------

// Reads the first Homebrew repository block from .goreleaser.yml.
function tap() {
  const cfg = existsSync('.goreleaser.yml') ? readFileSync('.goreleaser.yml', 'utf8') : ''
  const block = cfg.match(/^(?:homebrew_casks|brews):[\s\S]*?repository:\s*\n((?:\s{6,}.*\n)+)/m)
  if (!block) return null
  const field = (k) => block[1].match(new RegExp(`^\\s+${k}:\\s*["']?([^"'\\n]+)`, 'm'))?.[1].trim()
  const token = field('token')?.match(/\.Env\.(\w+)/)?.[1]
  return { owner: field('owner'), name: field('name'), token }
}

function secretGuide(name, t) {
  if (t && t.token === name) {
    return `create a fine-grained PAT with resource owner ${t.owner}, repository ${t.owner}/${t.name}, ` +
      `permission Contents: Read and write; save it as secret ${name} ` +
      `(gh secret set ${name} -o <org> or -R <owner>/<repo>)`
  }
  return `save it as secret ${name} (gh secret set ${name} -o <org> or -R <owner>/<repo>)`
}

function preflight() {
  const t = tap()
  const errors = []
  if (t?.token) {
    const token = process.env[t.token]
    if (!token) {
      errors.push(`secret ${t.token} is empty or not passed to this step; ${secretGuide(t.token, t)}`)
    } else {
      try {
        gh('api', `repos/${t.owner}/${t.name}`, '--silent', { GH_TOKEN: token })
      } catch (e) {
        errors.push(`${t.token} cannot read ${t.owner}/${t.name} (${e.message}); ${secretGuide(t.token, t)}`)
      }
    }
  }
  return errors
}

async function doctor() {
  const report = []
  const ok = (m) => report.push(['ok', m])
  const fail = (m) => report.push(['fail', m])
  const skip = (m) => report.push(['unverified', m])

  let repo
  try {
    repo = JSON.parse(gh('repo', 'view', '--json', 'nameWithOwner,defaultBranchRef'))
  } catch (e) {
    fail(`gh cannot reach the repository (${e.message}); run: gh auth login`)
    return report
  }
  const slug = repo.nameWithOwner
  const owner = slug.split('/')[0]

  const local = defaultBranch()
  const remote = repo.defaultBranchRef?.name
  if (local === remote) ok(`default branch ${remote}`)
  else fail(`local origin/HEAD is ${local ?? 'unset'}, GitHub says ${remote}; run: git remote set-head origin --auto`)

  try {
    const workflows = JSON.parse(gh('api', `repos/${slug}/actions/workflows`, '--paginate', '--slurp'))
      .flatMap((page) => page.workflows)
    for (const workflow of workflows.filter((w) => w.path.startsWith('.github/workflows/') && w.state !== 'active')) {
      fail(`${workflow.name} is ${workflow.state}; after pushing the reviewed YAML, run: gh workflow enable ${workflow.path.split('/').at(-1)} -R ${slug}`)
    }
  } catch (e) {
    skip(`workflow activation state is not readable (${e.message})`)
  }

  let pages = null
  try {
    pages = JSON.parse(gh('api', `repos/${slug}/pages`))
  } catch {
    // declared in settings.json means the site must exist: an outage, not a choice
    if (settings().pages) fail(`Pages is not enabled or not readable, but ${SETTINGS} declares it; check https://github.com/${slug}/settings/pages`)
    else skip(`Pages is not enabled or not readable; enable it in https://github.com/${slug}/settings/pages if docs are published`)
  }
  if (pages) {
    const want = settings().pages || {}
    const where = `https://github.com/${slug}/settings/pages`
    if (pages.build_type === (want.build_type || 'workflow')) ok(`Pages publishes ${pages.html_url} from GitHub Actions`)
    else fail(`Pages builds from a branch; set Settings → Pages → Source: GitHub Actions (${where})`)
    if ((pages.cname || null) === (want.cname || null)) ok(`Pages domain ${pages.cname || 'default'} matches ${SETTINGS}`)
    else fail(`Pages domain is ${pages.cname || 'default'}, ${SETTINGS} says ${want.cname || 'default'}; align them (${where}), then rebuild the docs: gh workflow run docs.yml`)
    if (want.https_enforced !== undefined) {
      if (pages.https_enforced === want.https_enforced) ok(`Pages HTTPS enforcement is ${pages.https_enforced ? 'on' : 'off'}`)
      else fail(`Pages HTTPS enforcement is ${pages.https_enforced ? 'on' : 'off'}, ${SETTINGS} wants ${want.https_enforced ? 'on' : 'off'}; toggle Enforce HTTPS (${where})`)
    }
    const errors = await smoke(pages.html_url, 0)
    if (errors.length) errors.forEach((e) => fail(`site: ${e}`))
    else ok(`site ${pages.html_url} serves its pages and assets`)
  }

  for (const [path, error] of await vanity()) {
    if (error) fail(`vanity import ${error}`)
    else ok(`vanity import ${path} resolves over HTTPS`)
  }

  const t = tap()
  const { secrets } = settings()
  const names = new Set([...Object.keys(secrets), ...(t?.token ? [t.token] : [])])
  const workflows = existsSync('.github/workflows')
    ? readdirSync('.github/workflows').filter((f) => /\.ya?ml$/.test(f)).map((f) => `.github/workflows/${f}`)
    : []
  const list = (flag, target) => {
    try { return new Set(JSON.parse(gh('secret', 'list', flag, target, '--json', 'name')).map((s) => s.name)) } catch { return null }
  }
  const repoSecrets = list('-R', slug)
  // organization secrets this repository can use; unlike the organization
  // listing, it needs no admin:org and honours per-repository visibility
  const shared = () => {
    try {
      return new Set(JSON.parse(gh('api', `repos/${slug}/actions/organization-secrets`, '--paginate', '--slurp'))
        .flatMap((page) => page.secrets).map((s) => s.name))
    } catch { return null }
  }
  const orgSecrets = shared() ?? list('-o', owner)
  for (const name of names) {
    const used = workflows.filter((f) => readFileSync(f, 'utf8').includes(`secrets.${name}`))
    if (!used.length) fail(`secret ${name} is declared but no workflow passes it; reference \${{ secrets.${name} }} where it is needed`)
    if (repoSecrets?.has(name)) ok(`secret ${name} is set on ${slug}${used.length ? `, used by ${used.join(', ')}` : ''}`)
    else if (orgSecrets?.has(name)) ok(`secret ${name} is shared by organization ${owner}${used.length ? `, used by ${used.join(', ')}` : ''}`)
    else if (orgSecrets === null) skip(`secret ${name} is not on ${slug}; organization secrets are not readable (gh auth refresh -s admin:org); if missing: ${secretGuide(name, t)}`)
    else fail(`secret ${name} is missing; ${secretGuide(name, t)}`)
  }

  // make tools installs into bin/<os>/<arch>, named as Go names them
  const arch = { x64: 'amd64', ia32: '386' }[process.arch] || process.arch
  const hostArch = execFileSync('uname', ['-m'], { encoding: 'utf8' }).trim()
  const goreleaser = [`bin/${process.platform}/${hostArch}`, `bin/${process.platform}/${arch}`, ...(process.env.PATH || '').split(':')]
  try {
    execFileSync(process.execPath, ['.github/scripts/goreleaser-check.mjs'], { stdio: 'ignore', env: { ...process.env, PATH: goreleaser.join(':') } })
    ok('goreleaser check')
  } catch (e) {
    if (e.code === 'ENOENT' || e.status === 127) skip('goreleaser is not installed; run: make tools')
    else fail('GoReleaser config check failed; run: make release-config-check')
  }
  return report
}

// --- main --------------------------------------------------------------------

const [mode, ...rest] = process.argv.slice(2)
const opts = args(rest)
try {
  switch (mode) {
    case 'check': {
      const errors = check(opts._[0], opts)
      errors.forEach((e) => console.error(`release: ${e}`))
      process.exit(errors.length ? 1 : 0)
    }
    case 'render':
      render(opts._[0], opts)
      break
    case 'preflight': {
      const errors = preflight()
      errors.forEach((e) => console.error(`preflight: ${e}`))
      process.exit(errors.length ? 1 : 0)
    }
    case 'doctor': {
      const report = await doctor()
      for (const [status, msg] of report) console.log(`${status.padEnd(10)} ${msg}`)
      process.exit(report.some(([s]) => s === 'fail') ? 1 : 0)
    }
    case 'site-env': {
      const site = expectedSite(settings().pages, repoSlug())
      if (!site) throw new Error('cannot resolve the Pages URL')
      console.log(`SITE_URL=${site}`)
      console.log(`BASE_PATH=${new URL(site).pathname.replace(/\/$/, '')}`)
      break
    }
    case 'pages': {
      const errors = pagesCheck(opts._[0])
      errors.forEach((e) => console.error(`pages: ${e}`))
      process.exit(errors.length ? 1 : 0)
    }
    case 'smoke': {
      const errors = await smoke(opts._[0], Number(opts.retries ?? 20))
      errors.forEach((e) => console.error(`smoke: ${e}`))
      process.exit(errors.length ? 1 : 0)
    }
    case 'vanity': {
      const results = await vanity()
      for (const [path, error] of results) {
        if (error) console.error(`vanity: ${error}`)
        else console.log(`vanity: ${path} ok`)
      }
      process.exit(results.some(([, e]) => e) ? 1 : 0)
    }
    default:
      console.error('usage: release.mjs check|render|preflight|doctor|pages|smoke|vanity ...')
      process.exit(2)
  }
} catch (e) {
  console.error(`release: ${e.message}`)
  process.exit(1)
}
