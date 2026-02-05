import assert from 'node:assert/strict'
import { test } from 'node:test'

import { compare, parseProfile, parseSnapshot } from './contribution-profile-diff.mjs'

const page = `
<table>
  <td tabindex="0" data-ix="0" aria-selected="false" data-date="2025-01-03" id="contribution-day-component-5-0" data-level="2" class="day"></td>
  <tool-tip id="t1" for="contribution-day-component-5-0" popover="manual">15 contributions on January 3rd.</tool-tip>
  <td data-date="2025-01-04" id="contribution-day-component-6-0" class="day"></td>
  <tool-tip for="contribution-day-component-6-0" popover="manual">1,203 contributions on January 4th.</tool-tip>
  <td data-date="2025-01-05" id="contribution-day-component-0-1" class="day"></td>
  <tool-tip for="contribution-day-component-0-1" popover="manual">No contributions on January 5th.</tool-tip>
  <td data-date="2025-01-06" id="contribution-day-component-1-1" class="day"></td>
  <tool-tip for="contribution-day-component-1-1" popover="manual">1 contribution on January 6th.</tool-tip>
</table>`

test('reads the count of every cell from its tooltip', () => {
  assert.deepEqual(parseProfile(page), { '2025-01-03': 15, '2025-01-04': 1203, '2025-01-05': 0, '2025-01-06': 1 })
})

test('refuses a page without the calendar or with an unknown tooltip', () => {
  assert.throws(() => parseProfile('<html></html>'), /no calendar cells/)
  assert.throws(
    () => parseProfile('<td data-date="2025-01-03" id="a"></td><tool-tip for="a">Something else</tool-tip>'),
    /unreadable tooltip for 2025-01-03/,
  )
})

test('takes the dates of a snapshot', () => {
  assert.deepEqual(parseSnapshot('{"2025-01-03T00:00:00Z":15,"2025-01-04T00:00:00Z":0}'), { '2025-01-03': 15, '2025-01-04': 0 })
})

test('lists the differing days of the year up to the end of the snapshot', () => {
  const snapshot = { '2025-01-03': 15, '2025-01-04': 22, '2025-01-05': 0 }
  const profile = { '2024-12-31': 4, '2025-01-03': 15, '2025-01-04': 15, '2025-01-05': 7, '2025-01-06': 9 }
  assert.deepEqual(compare(snapshot, profile), [
    { date: '2025-01-04', snapshot: 22, profile: 15 },
    { date: '2025-01-05', snapshot: 0, profile: 7 },
  ])
})
