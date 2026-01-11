import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { chmodSync, copyFileSync, mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import test from 'node:test'

const script = resolve('.github/scripts/release.mjs')
function fixture(fn) {
  const cwd = mkdtempSync(join(tmpdir(), 'maintainer-release-'))
  mkdirSync(join(cwd, '.github'))
  const settings = (pages = {}) => writeFileSync(join(cwd, '.github/settings.json'), JSON.stringify({ pages }))
  const run = (...args) => spawnSync(process.execPath, [script, ...args], {
    cwd, encoding: 'utf8',
    env: { ...process.env, GITHUB_REPOSITORY: 'octomation/maintainer', HOMEBREW_TAP_TOKEN: '' },
  })
  try { fn({ cwd, settings, run }) } finally { rmSync(cwd, { recursive: true, force: true }) }
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

test('preflight refuses a missing tap token before contacting GitHub', () => fixture(({ cwd, run }) => {
  writeFileSync(join(cwd, '.goreleaser.yml'), 'homebrew_casks:\n  - name: maintainer\n    repository:\n      owner: octolab\n      name: homebrew-tap\n      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"\n')
  const result = run('preflight')
  assert.equal(result.status, 1)
  assert.match(result.stderr, /HOMEBREW_TAP_TOKEN is empty/)
}))

test('missing release tags fail without creating Git refs', () => fixture(({ run }) => {
  const result = run('check', 'v9.9.9')
  assert.equal(result.status, 1)
  assert.match(result.stderr, /tag not found/)
}))

// All Git objects and refs below belong to disposable fixtures, never this checkout.
function releaseFixture(fn) {
  fixture(({ cwd, settings, run }) => {
    const git = (...args) => {
      const result = spawnSync('git', args, { cwd, encoding: 'utf8' })
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

test('pre-push checks annotated tags and destination branches, deletions and tag-only pushes', () => releaseFixture(({ cwd, git, sha, base }) => {
  mkdirSync(join(cwd, '.github/scripts'))
  copyFileSync(script, join(cwd, '.github/scripts/release.mjs'))
  mkdirSync(join(cwd, 'test-bin'))
  const makeLog = join(cwd, 'make.log')
  writeFileSync(join(cwd, 'test-bin/make'), '#!/bin/sh\nprintf "%s\n" "$*" >> "$MAKE_LOG"\n')
  chmodSync(join(cwd, 'test-bin/make'), 0o755)
  const hook = resolve('.github/hooks/pre-push')
  const push = (input) => {
    rmSync(makeLog, { force: true })
    return spawnSync('/bin/bash', [hook, 'origin', 'unused'], {
      cwd, input, encoding: 'utf8',
      env: { ...process.env, DEFAULT_BRANCH: 'main', MAKE_LOG: makeLog, PATH: `${join(cwd, 'test-bin')}:${process.env.PATH}` },
    })
  }
  const zero = '0'.repeat(40)
  const tag = `refs/tags/v9.9.9 ${git('rev-parse', 'v9.9.9')} refs/tags/v9.9.9 ${zero}\n`
  // A tag alone cannot introduce a commit absent from the remote branch.
  assert.equal(push(tag).status, 1)
  assert.equal(push(`refs/heads/release ${sha} refs/heads/main ${base}\n${tag}`).status, 0)
  assert.equal(readFileSync(makeLog, 'utf8').trim(), 'deps-tidy verify')
  assert.equal(push(`refs/heads/main ${sha} refs/heads/unrelated ${base}\n${tag}`).status, 1)
  assert.equal(push(`(delete) ${zero} refs/tags/v9.9.9 ${sha}\n`).status, 0)
  assert.equal(readFileSync(makeLog, 'utf8').trim(), 'deps-tidy verify')
  git('update-ref', 'refs/remotes/origin/main', sha)
  assert.equal(push(tag).status, 0)
  assert.equal(readFileSync(makeLog, 'utf8').trim(), 'deps-tidy verify')
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
  const env = { ...process.env, GIT_CONFIG_GLOBAL: globalConfig }
  const make = (target) => spawnSync('make', ['-f', resolve('Makefile'), target], { cwd, env, encoding: 'utf8' })
  const hooksPath = () => spawnSync('git', ['config', 'core.hooksPath'], { cwd, env, encoding: 'utf8' }).stdout.trim()
  assert.equal(hooksPath(), join(cwd, 'shared-hooks'))
  const installed = make('hooks')
  assert.equal(installed.status, 0, installed.stderr)
  assert.equal(hooksPath(), '.github/hooks')
  assert.equal(make('unhook').status, 0)
  assert.equal(hooksPath(), join(cwd, 'shared-hooks'))
  assert.equal(make('unhook').status, 0)
  assert.match(readFileSync(globalConfig, 'utf8'), /shared-hooks/)
  for (const name of ['pre-commit', 'pre-push', 'post-merge']) {
    assert.equal(readFileSync(join(cwd, 'shared-hooks', name), 'utf8'), 'shared hook: keep me')
  }
}))
