import assert from 'node:assert/strict'
import { spawn, spawnSync } from 'node:child_process'
import { chmodSync, copyFileSync, existsSync, mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { createServer } from 'node:http'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const script = resolve('.github/scripts/release.mjs')
// Fixtures run Git in disposable repositories only: drop the variables that
// select another repository or configuration, such as GIT_DIR exported to a
// hook, and ignore the system and global config, such as tag.gpgSign.
const isolated = {
  ...Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.startsWith('GIT_'))),
  GIT_CONFIG_NOSYSTEM: '1',
  GIT_CONFIG_GLOBAL: '/dev/null',
}
function fixture(fn) {
  const cwd = mkdtempSync(join(tmpdir(), 'maintainer-release-'))
  mkdirSync(join(cwd, '.github'))
  const settings = (pages = {}) => writeFileSync(join(cwd, '.github/settings.json'), JSON.stringify({ pages }))
  let extra = {}
  const run = (...args) => spawnSync(process.execPath, [script, ...args], {
    cwd, encoding: 'utf8',
    env: { ...isolated, GITHUB_REPOSITORY: 'octomation/maintainer', GITHUB_OUTPUT: '', OCTOLAB_RELEASER_CLIENT_ID: '', OCTOLAB_RELEASER_KEY: '', ...extra },
  })
  const withEnv = (env) => { extra = env }
  try { fn({ cwd, settings, run, withEnv }) } finally { rmSync(cwd, { recursive: true, force: true }) }
}

test('Pages preview resolves default and custom domains and rejects drift', () => fixture(({ settings, run }) => {
  settings()
  assert.equal(run('pages', 'https://octomation.github.io/maintainer/').status, 0)
  assert.equal(run('pages', 'https://octomation.github.io/indexit/').status, 1)
  assert.equal(run('site-env').stdout, 'SITE_URL=https://octomation.github.io/maintainer/\nBASE_PATH=/maintainer\n')
  settings({ cname: 'maintainer.example.com' })
  assert.equal(run('pages', 'https://maintainer.example.com').status, 0)
  assert.equal(run('pages', 'https://octomation.github.io/maintainer/').status, 1)
  assert.equal(run('site-env').stdout, 'SITE_URL=https://maintainer.example.com/\nBASE_PATH=\n')
}))

test('release notes render site links but preserve code examples', () => fixture(({ cwd, settings, run }) => {
  settings()
  mkdirSync(join(cwd, 'docs/content/changelog'), { recursive: true })
  const note = join(cwd, 'docs/content/changelog/v9.9.9.md')
  writeFileSync(note, '---\ntitle: Example release\ndescription: Test note.\n---\n\n# Example release\n\n[Guide](/contributions/)\n\n```md\n[Example](/keep-this/)\n```\n')
  const output = join(cwd, 'rendered.md')
  const result = run('render', 'v9.9.9', '--site-url', 'https://octomation.github.io/maintainer/', '--out', output)
  assert.equal(result.status, 0, result.stderr)
  assert.equal(result.stdout, 'Example release\n')
  const text = readFileSync(output, 'utf8')
  assert.match(text, /https:\/\/octomation.github.io\/maintainer\/contributions\//)
  assert.match(text, /\[Example\]\(\/keep-this\/\)/)
  assert.doesNotMatch(text, /# Example release|description:/)
  writeFileSync(note, '---\ntitle: Example release\ndescription: Test.\n---\n\n# Wrong title\n')
  assert.equal(run('render', 'v9.9.9', '--site-url', 'https://example.com', '--out', output).status, 1)
}))

test('preflight refuses missing App credentials and hands the tap to the token step', () => fixture(({ cwd, run, withEnv }) => {
  writeFileSync(join(cwd, '.goreleaser.yml'), 'brews:\n  - name: maintainer\n    repository:\n      owner: octolab\n      name: homebrew-tap\n      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"\n')
  const output = join(cwd, 'github-output')
  withEnv({ GITHUB_OUTPUT: output })
  const missing = run('preflight')
  assert.equal(missing.status, 1)
  assert.match(missing.stderr, /OCTOLAB_RELEASER_CLIENT_ID is empty/)
  assert.match(missing.stderr, /OCTOLAB_RELEASER_KEY is empty/)
  assert.match(missing.stderr, /installed on octolab\/homebrew-tap, octolab\/skills, octolab\/vanity only/)
  assert.equal(existsSync(output), false)
  withEnv({ GITHUB_OUTPUT: output, OCTOLAB_RELEASER_CLIENT_ID: 'Iv1.fixture', OCTOLAB_RELEASER_KEY: 'fixture' })
  const ready = run('preflight')
  assert.equal(ready.status, 0, ready.stderr)
  assert.equal(readFileSync(output, 'utf8'), 'tap-owner=octolab\ntap-name=homebrew-tap\n')
}))

// A local site stands in for Pages: smoke must not reach the network, so no
// repository is set and the default-domain redirect check is skipped.
test('smoke requires an og:image on the site', async () => {
  let og = ''
  const server = createServer((req, res) => {
    const html = (status, body) => { res.writeHead(status, { 'content-type': 'text/html' }); res.end(body) }
    if (req.url === '/') return html(200, `<link href="/_next/static/a.css"><script src="/_next/static/b.js"></script>${og}`)
    if (req.url.startsWith('/_next/static/')) { res.writeHead(200); return res.end() }
    if (req.url === '/og.png') { res.writeHead(200, { 'content-type': 'image/png' }); return res.end() }
    if (req.url === '/changelog/') return html(200, 'changelog')
    html(404, '<script src="/_next/static/b.js"></script>')
  })
  await new Promise((done) => server.listen(0, '127.0.0.1', done))
  const site = `http://127.0.0.1:${server.address().port}/`
  const cwd = mkdtempSync(join(tmpdir(), 'maintainer-smoke-'))
  const smoke = () => new Promise((done) => {
    const child = spawn(process.execPath, [script, 'smoke', site, '--retries', '0'], {
      cwd, env: { ...isolated, GITHUB_REPOSITORY: '' },
    })
    let stderr = ''
    child.stderr.on('data', (chunk) => { stderr += chunk })
    child.on('close', (status) => done({ status, stderr }))
  })
  try {
    const missing = await smoke()
    assert.equal(missing.status, 1)
    assert.match(missing.stderr, /no og:image/)
    og = `<meta property="og:image" content="${site}og.png"/>`
    const present = await smoke()
    assert.equal(present.status, 0, present.stderr)
  } finally {
    server.close()
    rmSync(cwd, { recursive: true, force: true })
  }
})

test('missing release tags fail without creating Git refs', () => fixture(({ run }) => {
  const result = run('check', 'v9.9.9')
  assert.equal(result.status, 1)
  assert.match(result.stderr, /tag not found/)
}))

// All Git objects and refs below belong to disposable fixtures, never this checkout.
function releaseFixture(fn) {
  fixture(({ cwd, settings, run }) => {
    const git = (...args) => {
      const result = spawnSync('git', args, { cwd, env: isolated, encoding: 'utf8' })
      assert.equal(result.status, 0, result.stderr)
      return result.stdout.trim()
    }
    git('init', '--initial-branch=main')
    git('config', 'user.name', 'Release tests')
    git('config', 'user.email', 'tests@example.invalid')
    git('config', 'commit.gpgsign', 'false')
    git('config', 'core.hooksPath', '/dev/null')
    settings()
    git('add', '.')
    git('commit', '-m', 'Base fixture')
    const base = git('rev-parse', 'HEAD')
    git('update-ref', 'refs/remotes/origin/main', base)
    git('symbolic-ref', 'refs/remotes/origin/HEAD', 'refs/remotes/origin/main')
    const note = 'docs/content/changelog/v9.9.9.md'
    mkdirSync(join(cwd, 'docs/content/changelog'), { recursive: true })
    writeFileSync(join(cwd, note), '---\ntitle: Release fixture\ndescription: Fixture.\n---\n\n# Release fixture\n')
    git('add', note)
    git('commit', '-m', 'Release fixture')
    const sha = git('rev-parse', 'HEAD')
    git('tag', '-a', 'v9.9.9', '-m', 'Annotated release fixture')
    fn({ cwd, git, run, note, base, sha })
  })
}

test('local release preparation precedes the branch push; CI still requires the remote branch', () => releaseFixture(({ git, run, sha, note, cwd }) => {
  assert.equal(run('check', 'v9.9.9', '--local').status, 0)
  assert.match(run('check', 'v9.9.9').stderr, /its commit is not on main/)
  assert.equal(run('check', 'v9.9.9', '--pushed', `main=${sha}`).status, 0)
  assert.equal(run('check', 'v9.9.9', '--pushed', `unrelated=${sha}`).status, 1)
  git('update-ref', 'refs/remotes/origin/main', sha)
  assert.equal(run('check', 'v9.9.9').status, 0)
  writeFileSync(join(cwd, note), 'Uncommitted release note')
  assert.match(run('check', 'v9.9.9', '--local').stderr, /has changes not in v9.9.9/)
}))

test('a release tag requires the version of every skill of its commit', () => releaseFixture(({ cwd, git, run }) => {
  const skill = (name, version) => {
    mkdirSync(join(cwd, 'skills', name), { recursive: true })
    writeFileSync(join(cwd, 'skills', name, 'SKILL.md'), `---\nname: ${name}\nmetadata:\n  author: octomation\n  version: "${version}"\n---\n\n# ${name}\n`)
  }
  const tag = () => {
    git('add', 'skills')
    git('commit', '-m', 'Skills')
    git('tag', '-f', '-a', 'v9.9.9', '-m', 'Annotated release fixture')
  }
  skill('maintainer', '9.9.8')
  skill('review', '9.9.9')
  writeFileSync(join(cwd, 'skills/skills.go'), 'package skills\n')
  tag()
  const stderr = run('check', 'v9.9.9', '--local').stderr
  assert.match(stderr, /skills\/maintainer\/SKILL\.md: metadata\.version is 9\.9\.8, not 9\.9\.9/)
  assert.doesNotMatch(stderr, /review|skills\.go/)
  skill('maintainer', '9.9.9')
  skill('review', '9.9.8')
  tag()
  assert.match(run('check', 'v9.9.9', '--local').stderr, /skills\/review\/SKILL\.md: metadata\.version is 9\.9\.8/)
  skill('review', '9.9.9')
  tag()
  assert.equal(run('check', 'v9.9.9', '--local').status, 0)

  // Two skills need a skills-action that publishes more than one.
  const pin = (version) => {
    mkdirSync(join(cwd, '.github/workflows'), { recursive: true })
    writeFileSync(join(cwd, '.github/workflows/cd.yml'), `jobs:\n  skill:\n    steps:\n      - uses: octomation/skills-action@${version}\n`)
    git('add', '.github/workflows')
    tag()
  }
  for (const old of ['v0.1.1', 'v0.0.9']) {
    pin(old)
    assert.match(run('check', 'v9.9.9', '--local').stderr,
      new RegExp(`cd\\.yml: octomation/skills-action@${old.replace(/\./g, '\\.')} publishes one skill, skills holds 2; pin octomation/skills-action v0\\.2\\.0 or later`))
  }
  for (const ok of ['v0.2.0', 'v0.10.0', 'v1.0.0']) {
    pin(ok)
    assert.equal(run('check', 'v9.9.9', '--local').status, 0, ok)
  }
  pin('v0.1.1')
  rmSync(join(cwd, 'skills/review'), { recursive: true })
  git('add', '-A', 'skills')
  tag()
  assert.equal(run('check', 'v9.9.9', '--local').status, 0, 'one skill is what v0.1.1 publishes')
}))

test('skill compares the catalog with the source or the binary and needs release archives', () => fixture(({ cwd, run, withEnv }) => {
  // The test vector shared with internal/skill and the octolab/skills catalog.
  const vector = 'sha256:64ce802d5cef2c789ce95d306c9fcce6779c6e8166339ae47634e7f8b265a34b'
  mkdirSync(join(cwd, 'skills/maintainer/references'), { recursive: true })
  writeFileSync(join(cwd, 'skills/maintainer/SKILL.md'), 'hello\n')
  writeFileSync(join(cwd, 'skills/maintainer/references/a.md'), 'a')
  mkdirSync(join(cwd, 'skills/testdata'))
  writeFileSync(join(cwd, 'skills/skills.go'), 'package skills\n')

  // A fake gh serves the catalog tree from catalog.json and blobs by their sha.
  const bin = join(cwd, 'bin')
  mkdirSync(bin)
  const catalog = (files, extra = []) => writeFileSync(join(cwd, 'catalog.json'), JSON.stringify({
    truncated: false,
    tree: [
      { path: 'plugins/maintainer/.claude-plugin/plugin.json', type: 'blob', mode: '100644', sha: 'p' },
      ...Object.keys(files).map((path) => ({ path: `plugins/maintainer/skills/${path}`, type: 'blob', mode: '100644', sha: Buffer.from(files[path]).toString('hex') || 'empty' })),
      ...extra,
    ],
  }))
  writeFileSync(join(bin, 'gh'), `#!/usr/bin/env node
const [, , api, path] = process.argv
if (api !== 'api') process.exit(2)
if (path.startsWith('repos/octolab/skills/git/trees/maintainer--v9.9.9?')) { process.stdout.write(require('fs').readFileSync(${JSON.stringify(join(cwd, 'catalog.json'))})); process.exit(0) }
const blob = path.match(/^repos\\/octolab\\/skills\\/git\\/blobs\\/(.+)$/)
if (blob) { const hex = blob[1] === 'empty' ? '' : blob[1]; process.stdout.write(JSON.stringify({ encoding: 'base64', content: Buffer.from(hex, 'hex').toString('base64') })); process.exit(0) }
process.stderr.write('HTTP 404: Not Found'); process.exit(1)
`)
  chmodSync(join(bin, 'gh'), 0o755)
  writeFileSync(join(bin, 'maintainer'), `#!/bin/sh\nprintf '%s\\n' '{"name":"maintainer","digest":"${vector}"}' '{"name":"review","digest":"sha256:r"}'\n`)
  chmodSync(join(bin, 'maintainer'), 0o755)
  withEnv({ PATH: `${bin}:${process.env.PATH}` })

  catalog({ 'maintainer/SKILL.md': 'hello\n', 'maintainer/references/a.md': 'a' })
  const ok = run('skill', '--catalog', 'maintainer--v9.9.9')
  assert.equal(ok.status, 0, ok.stderr)
  assert.match(ok.stdout, new RegExp(`the source and the catalog at maintainer--v9.9.9: maintainer ${vector}`))
  assert.match(run('skill', '--catalog', 'maintainer--v9.9.9', '--binary', join(bin, 'maintainer')).stderr,
    /octolab\/skills@maintainer--v9\.9\.9: no review skill, the binary has sha256:r/)
  assert.match(run('skill', '--catalog', 'maintainer--v0.0.0').stderr, /not readable: HTTP 404/)
  assert.match(run('skill', '--catalog', '').stderr, /--catalog needs the catalog tag/)

  catalog({ 'maintainer/SKILL.md': 'hello\n', 'maintainer/references/a.md': 'b', 'review/SKILL.md': 'r' })
  const stderr = run('skill', '--catalog', 'maintainer--v9.9.9').stderr
  assert.match(stderr, /the maintainer skill is sha256:\w+, the source has sha256:64ce/)
  assert.match(stderr, /the review skill is not in the source/)

  catalog({ 'maintainer/SKILL.md': 'hello\n' }, [{ path: 'plugins/maintainer/skills/maintainer/link', type: 'blob', mode: '120000', sha: 'l' }])
  assert.match(run('skill', '--catalog', 'maintainer--v9.9.9').stderr, /symbolic links and submodules are not allowed/)
  catalog({ 'maintainer/SKILL.md': 'hello\n', 'maintainer/.hidden': '' })
  assert.match(run('skill', '--catalog', 'maintainer--v9.9.9').stderr, /dotfiles and symbolic links are not allowed/)

  assert.match(run('skill').stderr, /dist: no release archives/)
}))

test('pre-push checks annotated tags and destination branches, deletions and tag-only pushes', () => releaseFixture(({ cwd, git, sha, base }) => {
  mkdirSync(join(cwd, '.github/scripts'))
  copyFileSync(script, join(cwd, '.github/scripts/release.mjs'))
  // release.mjs imports the GoReleaser check that accepts the brews deprecation
  copyFileSync(resolve('.github/scripts/goreleaser-check.mjs'), join(cwd, '.github/scripts/goreleaser-check.mjs'))
  mkdirSync(join(cwd, 'test-bin'))
  const makeLog = join(cwd, 'make.log')
  writeFileSync(join(cwd, 'test-bin/make'), '#!/bin/sh\nprintf "%s\n" "$*" >> "$MAKE_LOG"\n')
  chmodSync(join(cwd, 'test-bin/make'), 0o755)
  const hook = resolve('.github/hooks/pre-push')
  const push = (input) => {
    rmSync(makeLog, { force: true })
    return spawnSync('/bin/bash', [hook, 'origin', 'unused'], {
      cwd, input, encoding: 'utf8',
      env: { ...isolated, DEFAULT_BRANCH: 'main', MAKE_LOG: makeLog, PATH: `${join(cwd, 'test-bin')}:${process.env.PATH}` },
    })
  }
  const zero = '0'.repeat(40)
  const tag = `refs/tags/v9.9.9 ${git('rev-parse', 'v9.9.9')} refs/tags/v9.9.9 ${zero}\n`
  // A tag alone cannot introduce a commit absent from the remote branch.
  assert.equal(push(tag).status, 1)
  assert.equal(push(`refs/heads/release ${sha} refs/heads/main ${base}\n${tag}`).status, 0)
  assert.equal(readFileSync(makeLog, 'utf8').trim(), 'go-deps-tidy verify')
  assert.equal(push(`refs/heads/main ${sha} refs/heads/unrelated ${base}\n${tag}`).status, 1)
  assert.equal(push(`(delete) ${zero} refs/tags/v9.9.9 ${sha}\n`).status, 0)
  assert.equal(readFileSync(makeLog, 'utf8').trim(), 'go-deps-tidy verify')
  git('update-ref', 'refs/remotes/origin/main', sha)
  assert.equal(push(tag).status, 0)
  assert.equal(readFileSync(makeLog, 'utf8').trim(), 'go-deps-tidy verify')
}))

test('hooks are wired in place through the local core.hooksPath and leave shared hooks alone', () => releaseFixture(({ cwd, git }) => {
  mkdirSync(join(cwd, '.github/hooks'), { recursive: true })
  mkdirSync(join(cwd, 'shared-hooks'))
  for (const name of ['pre-commit', 'pre-push', 'post-merge']) {
    copyFileSync(resolve('.github/hooks', name), join(cwd, '.github/hooks', name))
    writeFileSync(join(cwd, 'shared-hooks', name), 'shared hook: keep me')
  }
  const globalConfig = join(cwd, 'global.gitconfig')
  writeFileSync(globalConfig, `[core]\n\thooksPath = ${join(cwd, 'shared-hooks')}\n`)
  git('config', '--local', '--unset', 'core.hooksPath')
  const env = { ...isolated, GIT_CONFIG_GLOBAL: globalConfig }
  const make = (target) => spawnSync('make', ['-f', resolve('Makefile'), target], { cwd, env, encoding: 'utf8' })
  const hooksPath = () => spawnSync('git', ['config', 'core.hooksPath'], { cwd, env, encoding: 'utf8' }).stdout.trim()
  assert.equal(hooksPath(), join(cwd, 'shared-hooks'))
  const installed = make('git-hooks')
  assert.equal(installed.status, 0, installed.stderr)
  assert.equal(hooksPath(), '.github/hooks')
  assert.equal(make('git-unhook').status, 0)
  assert.equal(hooksPath(), join(cwd, 'shared-hooks'))
  assert.equal(make('git-unhook').status, 0)
  assert.match(readFileSync(globalConfig, 'utf8'), /shared-hooks/)
  for (const name of ['pre-commit', 'pre-push', 'post-merge']) {
    assert.equal(readFileSync(join(cwd, 'shared-hooks', name), 'utf8'), 'shared hook: keep me')
  }
}))

test('fixtures ignore the Git repository and config the caller selects', () => {
  const victim = mkdtempSync(join(tmpdir(), 'maintainer-victim-'))
  try {
    const git = (...args) => spawnSync('git', ['-C', victim, ...args], { env: isolated, encoding: 'utf8' }).stdout
    git('init', '-q')
    const config = git('config', '--local', '--list')
    // a nested run of one fixture test, as if a hook ran the suite; without
    // NODE_TEST_CONTEXT it reports to stdout instead of to this runner
    const { NODE_TEST_CONTEXT, ...caller } = process.env
    const inner = spawnSync(process.execPath, ['--test', '--test-reporter=tap', '--test-name-pattern=^local release preparation', fileURLToPath(import.meta.url)], {
      encoding: 'utf8', env: { ...caller, GIT_DIR: join(victim, '.git') },
    })
    assert.equal(inner.status, 0, inner.stdout + inner.stderr)
    assert.match(inner.stdout, /^# pass 1$/m)
    assert.equal(git('config', '--local', '--list'), config)
    assert.equal(git('for-each-ref'), '')
  } finally {
    rmSync(victim, { recursive: true, force: true })
  }
})
