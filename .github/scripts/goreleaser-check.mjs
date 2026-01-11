#!/usr/bin/env node
// GoReleaser 2.18.2 exits 2 for a valid config with deprecated properties.
// Keep only the intentional brews warning non-blocking while publishing both
// Formula and Cask. Invalid config (exit 1) and other deprecations still fail.
import { spawnSync } from 'node:child_process'

const result = spawnSync('goreleaser', ['check'], { encoding: 'utf8' })
if (result.error) {
  console.error(result.error.message)
  process.exit(result.error.code === 'ENOENT' ? 127 : 1)
}
process.stdout.write(result.stdout)
process.stderr.write(result.stderr)
const output = `${result.stdout}\n${result.stderr}`.replace(/\x1b\[[0-9;]*m/g, '')
const deprecated = [...output.matchAll(/DEPRECATED:\s+(\S+)/g)].map((match) => match[1])
if (result.status === 2 && deprecated.length > 0 && deprecated.every((key) => key === 'brews')) {
  console.log('Accepted brews deprecation: Formula and Cask are intentionally published together.')
  process.exit(0)
}
process.exit(result.status ?? 1)
