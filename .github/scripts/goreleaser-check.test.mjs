import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { chmodSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import test from 'node:test'

const script = resolve('.github/scripts/goreleaser-check.mjs')
test('config validation permits only the intentional brews deprecation', () => {
  const cwd = mkdtempSync(join(tmpdir(), 'maintainer-goreleaser-'))
  try {
    const executable = join(cwd, 'goreleaser')
    writeFileSync(executable, '#!/bin/sh\nprintf "%s\n" "$CHECK_OUTPUT" >&2\nexit "$CHECK_STATUS"\n')
    chmodSync(executable, 0o755)
    const run = (status, output) => spawnSync(process.execPath, [script], {
      cwd, encoding: 'utf8',
      env: { ...process.env, PATH: `${cwd}:${process.env.PATH}`, CHECK_STATUS: String(status), CHECK_OUTPUT: output },
    }).status
    assert.equal(run(0, 'configuration valid'), 0)
    assert.equal(run(2, 'DEPRECATED:  brews should not be used anymore'), 0)
    assert.equal(run(1, 'DEPRECATED:  brews\nconfiguration is invalid'), 1)
    assert.equal(run(2, 'DEPRECATED:  brews\nDEPRECATED:  another_property'), 2)
    assert.equal(run(2, 'unrecognized failure'), 2)
  } finally { rmSync(cwd, { recursive: true, force: true }) }
})
