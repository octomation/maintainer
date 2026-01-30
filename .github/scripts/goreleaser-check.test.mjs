import { test } from 'node:test'
import assert from 'node:assert/strict'
import { acceptedBrewsDeprecation } from './goreleaser-check.mjs'

const output = `  • DEPRECATED:  brews  should not be used anymore
  • .goreleaser.yml error=configuration is valid, but uses deprecated properties
  ⨯ check failed error=1 out of 1 configuration file(s) have issues`

test('accepts the deliberate brews deprecation only', () => {
  assert.equal(acceptedBrewsDeprecation(2, output), true)
  assert.equal(acceptedBrewsDeprecation(1, output), false)
  assert.equal(acceptedBrewsDeprecation(null, output), false)
  assert.equal(acceptedBrewsDeprecation(2, output.replace('brews ', 'archives.format ')), false)
  assert.equal(acceptedBrewsDeprecation(2, output + '\nDEPRECATED: another setting'), false)
  assert.equal(acceptedBrewsDeprecation(2, 'yaml: unknown field'), false)
  assert.equal(acceptedBrewsDeprecation(2, output.replace('configuration is valid,', 'configuration is invalid,')), false)
})
