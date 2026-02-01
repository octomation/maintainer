import { spawnSync } from 'node:child_process'
import { pathToFileURL } from 'node:url'

// GoReleaser v2.18.2 supports the Formula service block only through brews,
// but check exits 2 for it. Accept only this deliberate exception, never
// arbitrary validation failures.
export function acceptedBrewsDeprecation(status, output) {
  const plain = output.replace(/\x1b\[[0-9;]*m/g, '')
  const warnings = plain.split('\n').filter(line => line.includes('DEPRECATED:'))
  return status === 2
    && plain.includes('configuration is valid, but uses deprecated properties')
    && warnings.length === 1
    && /DEPRECATED:\s+brews\s+should not be used anymore/.test(warnings[0])
    && /error=1 out of 1 configuration file\(s\) have issues/.test(plain)
}

export function checkGoreleaser(env = process.env) {
  const result = spawnSync('goreleaser', ['check', '.goreleaser.yml'], {
    encoding: 'utf8', env,
  })
  if (result.error) throw result.error
  const output = (result.stdout || '') + (result.stderr || '')
  const accepted = acceptedBrewsDeprecation(result.status, output)
  return { ok: result.status === 0 || accepted, accepted, output }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const result = checkGoreleaser()
  process.stdout.write(result.output)
  if (result.accepted) console.log('Accepted: brews is required for the Formula service block.')
  process.exitCode = result.ok ? 0 : 1
}
