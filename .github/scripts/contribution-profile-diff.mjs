#!/usr/bin/env node
// Compares a year of `maintainer github contribution snapshot` with the
// contribution calendar of the same year on the public GitHub profile, day by
// day. The commands promise the days of the profile (MAIN-159), so any day
// listed here is a defect or a limit to explain. Zero dependencies: Node only.
// The profile is read anonymously, as anyone sees it; no token is used.
//
//   maintainer github contribution snapshot 2026 > 2026.json
//   contribution-profile-diff.mjs <login> 2026.json
//
// Prints each differing day with both counts and exits 1 if there is one.
// The profile counts only what it shows publicly: with private contributions
// hidden on the profile, their days differ.

import { readFileSync } from 'node:fs'
import { pathToFileURL } from 'node:url'

// The calendar cells carry their date and an id; the tooltip that names the
// count refers to that id.
export function parseProfile(html) {
  const cells = new Map()
  for (const [tag] of html.matchAll(/<td\b[^>]*\bdata-date="[^"]+"[^>]*>/g)) {
    const date = tag.match(/\bdata-date="(\d{4}-\d\d-\d\d)"/)?.[1]
    const id = tag.match(/\bid="([^"]+)"/)?.[1]
    if (date && id) cells.set(id, date)
  }
  const days = {}
  for (const [, id, text] of html.matchAll(/<tool-tip\b[^>]*\bfor="([^"]+)"[^>]*>([^<]*)<\/tool-tip>/g)) {
    const date = cells.get(id)
    if (!date) continue
    const count = text.match(/^\s*([\d,]+) contributions?\b/)
    if (count) days[date] = Number(count[1].replaceAll(',', ''))
    else if (/^\s*No contributions\b/.test(text)) days[date] = 0
    else throw new Error(`unreadable tooltip for ${date}: ${text.trim()}`)
  }
  if (cells.size === 0) throw new Error('no calendar cells in the profile page')
  return days
}

// A snapshot maps RFC 3339 midnights in UTC to counts.
export function parseSnapshot(json) {
  const days = {}
  for (const [stamp, count] of Object.entries(JSON.parse(json))) days[stamp.slice(0, 10)] = count
  return days
}

// The days of the snapshot's year that differ; a day only one side has counts
// as zero on the other, as in `maintainer github contribution diff`.
export function compare(snapshot, profile) {
  const years = new Set(Object.keys(snapshot).map((date) => date.slice(0, 4)))
  const dates = new Set([...Object.keys(snapshot), ...Object.keys(profile)].filter((date) => years.has(date.slice(0, 4))))
  const last = Object.keys(snapshot).sort().at(-1)
  return [...dates]
    .filter((date) => date <= last)
    .sort()
    .filter((date) => (snapshot[date] ?? 0) !== (profile[date] ?? 0))
    .map((date) => ({ date, snapshot: snapshot[date] ?? 0, profile: profile[date] ?? 0 }))
}

async function main([login, file]) {
  if (!login || !file) {
    console.error('usage: contribution-profile-diff.mjs <login> <snapshot.json>')
    return 2
  }
  const snapshot = parseSnapshot(readFileSync(file, 'utf8'))
  const year = Object.keys(snapshot).sort()[0]?.slice(0, 4)
  if (!year) throw new Error(`${file}: an empty snapshot`)
  const url = `https://github.com/users/${encodeURIComponent(login)}/contributions?from=${year}-01-01&to=${year}-12-31`
  const response = await fetch(url, { headers: { 'User-Agent': 'contribution-profile-diff' } })
  if (!response.ok) throw new Error(`${url}: ${response.status} ${response.statusText}`)
  const diff = compare(snapshot, parseProfile(await response.text()))
  for (const { date, snapshot: ours, profile } of diff) console.log(`${date}  snapshot ${ours}  profile ${profile}`)
  console.error(diff.length ? `${diff.length} day(s) of ${year} differ from the profile` : `every day of ${year} agrees with the profile`)
  return diff.length ? 1 : 0
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  main(process.argv.slice(2)).then(
    (code) => process.exit(code),
    (error) => {
      console.error(error.message)
      process.exit(2)
    },
  )
}
