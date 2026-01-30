#!/usr/bin/env node
// Checks that the tool versions pinned outside tools/go.mod follow it, and the
// Go versions pinned outside go.mod follow its go line. make update moves only
// the modules: an action that installs its own copy of a tool, or a Go version
// in a workflow, the Makefile or the docs, would keep the old one quietly.
// Zero dependencies: Node only. It reads the YAML forms the workflows use, not
// all of YAML.
//
//   tool-pins-check.mjs [<dir>]   (prints each stale pin as file:line, exits 1)

import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'

// Actions that install a tool themselves instead of taking it from make tools,
// by the module tools/go.mod requires for it: the version input must name the
// same release.
const ACTIONS = {
  'golangci/golangci-lint-action': { tool: 'golangci-lint', module: 'github.com/golangci/golangci-lint' },
  'goreleaser/goreleaser-action': { tool: 'goreleaser', module: 'github.com/goreleaser/goreleaser' },
}
// setup-go versions that follow new Go releases instead of a Go line: a matrix
// may test the latest Go with them, a step that pins Go may not
const FLOATING = new Set(['1.x', 'stable', 'oldstable'])
const WORKFLOWS = ['.github/workflows', '.github/actions']
const DOCS = ['README.md', 'docs/content']
const NOTES = 'docs/content/changelog/'

const indent = (text) => text.match(/^\s*/)[0].length
const blank = (text) => !text.trim()
const unquote = (value) => value.trim().replace(/^(['"])(.*)\1$/, '$2')
const expression = (value) => value.startsWith('${{')
// 1.27 for go 1.27.0, 1.27.x, v1.27, 1.27rc1; null for anything else
const goLine = (version) => version.match(/^v?(\d+\.\d+)(?:\.\d+|\.x)?(?:(?:rc|beta)\d+)?$/)?.[1] ?? null
const semver = (version) => `v${version.replace(/^v/, '')}`

// A YAML line without its comment: a # that starts the line or follows a
// space, outside quotes.
function uncomment(line) {
  let quote = null
  for (let i = 0; i < line.length; i++) {
    const c = line[i]
    if (quote === '"' && c === '\\') i++
    else if (quote) quote = c === quote ? null : quote
    else if (c === '"' || c === "'") quote = c
    else if (c === '#' && (i === 0 || /\s/.test(line[i - 1]))) return line.slice(0, i).trimEnd()
  }
  return line
}

function lines(root, file) {
  const path = join(root, file)
  return existsSync(path) ? readFileSync(path, 'utf8').split('\n') : null
}

// The files and the files under the directories of paths, relative to root,
// that have one of the extensions.
function files(root, paths, extensions) {
  return paths.filter((path) => existsSync(join(root, path))).flatMap((path) => {
    const names = statSync(join(root, path)).isDirectory()
      ? readdirSync(join(root, path), { recursive: true }).map((name) => `${path}/${name}`).sort()
      : [path]
    return names.filter((name) => extensions.some((ext) => name.endsWith(ext)))
  })
}

// The go and toolchain lines and the required modules of a go.mod.
function goMod(text) {
  const mod = { go: null, toolchain: null, modules: new Map() }
  let block = false
  text.forEach((line, i) => {
    const directive = line.match(/^(go|toolchain)\s+(\S+)/)
    if (directive) mod[directive[1]] = { version: directive[2].replace(/^go/, ''), line: i + 1 }
    if (/^require\s*\(\s*$/.test(line)) block = true
    else if (block && /^\)/.test(line)) block = false
    const require = block ? line.match(/^\s*(\S+)\s+(v\S+)/) : line.match(/^require\s+(\S+)\s+(v\S+)/)
    if (require) mod.modules.set(require[1], { version: require[2], line: i + 1 })
  })
  return mod
}

// The required module of a tool, by its path without the major version suffix.
function required(mod, base) {
  for (const [path, req] of mod.modules) {
    if (path === base || new RegExp(`^${base.replace(/\./g, '\\.')}/v\\d+$`).test(path)) return { path, ...req }
  }
  return null
}

// The steps of a workflow that use an action, with their with: inputs in the
// block or the flow form.
function steps(text) {
  const out = []
  text.forEach((line, i) => {
    const uses = line.match(/^(\s*)(-\s+)?uses:\s*['"]?([^@\s'"]+)/)
    if (!uses) return
    const key = uses[1].length + (uses[2]?.length ?? 0)
    let start = i
    for (let j = i - 1; !uses[2] && j >= 0; j--) {
      if (blank(text[j]) || indent(text[j]) >= key) continue
      if (/^\s*-\s/.test(text[j])) start = j
      break
    }
    let end = i + 1
    while (end < text.length && (blank(text[end]) || indent(text[end]) >= key)) end++
    const inputs = {}
    for (let j = start; j < end; j++) {
      const w = text[j].match(/^(\s*)(-\s+)?with:\s*(.*)$/)
      if (!w || w[1].length + (w[2]?.length ?? 0) !== key) continue
      const flow = w[3].trim()
      if (flow.startsWith('{')) {
        for (const [, name, value] of flow.slice(1, flow.lastIndexOf('}')).matchAll(/([\w-]+)\s*:\s*([^,]*)/g)) {
          inputs[name] = { value: unquote(value), line: j + 1 }
        }
        continue
      }
      let child = null
      for (let k = j + 1; k < end && (blank(text[k]) || indent(text[k]) > key); k++) {
        if (blank(text[k])) continue
        child ??= indent(text[k])
        const input = indent(text[k]) === child && text[k].match(/^\s*([\w-]+):\s*(.*)$/)
        if (input) inputs[input[1]] = { value: unquote(input[2]), line: k + 1 }
      }
    }
    out.push({ action: uses[3], line: i + 1, inputs })
  })
  return out
}

// The Go versions of a build matrix: the go key in the block, flow or list
// item form, also inside a flow mapping or across lines, and the literals
// compared with matrix.go on either side.
function matrix(text) {
  const out = []
  const joined = text.join('\n')
  const at = (index) => joined.slice(0, index).split('\n').length
  for (const m of joined.matchAll(/(?<![\w.-])go:\s*\[([^\]]*)\]/g)) {
    let index = m.index + m[0].indexOf('[') + 1
    for (const value of m[1].split(',')) {
      if (value.trim()) out.push({ value: unquote(value), line: at(index + value.search(/\S/)) })
      index += value.length + 1
    }
  }
  // the go key of a flow mapping, maybe on a line of its own: not block form
  const flow = new Set()
  for (const m of joined.matchAll(/[{,]\s*go:\s*([^\s,}[][^,}]*)/g)) {
    const line = at(m.index + m[0].indexOf('go:'))
    flow.add(line)
    out.push({ value: unquote(m[1]), line })
  }
  for (const m of joined.matchAll(/matrix\.go\s*[!=]=\s*(['"])(.*?)\1|(['"])([^'"]*)\3\s*[!=]=\s*matrix\.go\b/g)) {
    out.push({ value: m[2] ?? m[4], line: at(m.index) })
  }
  text.forEach((line, i) => {
    const go = line.match(/^(\s*)(-\s+)?go:\s*(.*)$/)
    if (!go || go[3].startsWith('[') || flow.has(i + 1)) return
    if (go[3]) return out.push({ value: unquote(go[3]), line: i + 1 })
    const key = go[1].length + (go[2]?.length ?? 0)
    for (let j = i + 1; j < text.length; j++) {
      if (blank(text[j])) continue
      const item = indent(text[j]) >= key && text[j].match(/^\s*-\s+(.*)$/)
      if (!item) break
      out.push({ value: unquote(item[1]), line: j + 1 })
    }
  })
  return out.sort((a, b) => a.line - b.line)
}

// { errors, summary }: every pin that disagrees as "file:line: problem", and
// the pins that agree, counted by tool and version.
export function toolPins(root = '.') {
  const found = []
  const order = []
  const read = (file) => {
    order.push(file)
    return lines(root, file)
  }
  const seen = new Map()
  const pin = (label) => seen.set(label, (seen.get(label) ?? 0) + 1)
  const at = (file, line, problem) => found.push({ file, line, problem })

  const app = read('go.mod')
  const mods = app ? { 'go.mod': goMod(app) } : {}
  const want = mods['go.mod']?.go && goLine(mods['go.mod'].go.version)
  if (!want) return { errors: ['go.mod: no go line to check the Go versions against'], summary: '' }
  const onLine = (version) => goLine(version) === want
  const goPin = () => pin(`Go ${want}`)

  const tools = read('tools/go.mod')
  if (tools) mods['tools/go.mod'] = goMod(tools)
  for (const [file, mod] of Object.entries(mods)) {
    for (const directive of ['go', 'toolchain']) {
      const d = mod[directive]
      if (!d) continue
      if (onLine(d.version)) goPin()
      else at(file, d.line, `${directive} ${d.version}, go.mod is on Go ${want}`)
    }
  }

  const node = []
  for (const file of files(root, WORKFLOWS, ['.yml', '.yaml'])) {
    const text = read(file).map(uncomment)
    for (const step of steps(text)) {
      const action = step.action
      const version = step.inputs.version
      if (ACTIONS[action]) {
        const { tool, module: base } = ACTIONS[action]
        const req = mods['tools/go.mod'] && required(mods['tools/go.mod'], base)
        if (!req) at(file, step.line, `${action} installs ${tool}, but tools/go.mod does not require ${base}`)
        else if (!version) at(file, step.line, `${action} without a version installs the latest ${tool}; pin ${req.version} from tools/go.mod`)
        else if (expression(version.value)) at(file, version.line, `${tool} version ${version.value} is not a literal; pin ${req.version} from tools/go.mod`)
        else if (semver(version.value) !== req.version) at(file, version.line, `${tool} ${version.value}, tools/go.mod requires ${req.path} ${req.version}`)
        else pin(`${tool} ${req.version}`)
      }
      const go = step.inputs['go-version']
      if (action === 'actions/setup-go' && go && !expression(go.value)) {
        if (onLine(go.value)) goPin()
        else at(file, go.line, `go-version ${go.value}, go.mod is on Go ${want}; pin ${want}.x`)
      }
      const nodeVersion = step.inputs['node-version']
      if (action === 'actions/setup-node' && nodeVersion && !expression(nodeVersion.value)) node.push({ file, ...nodeVersion })
    }
    for (const entry of matrix(text)) {
      if (expression(entry.value) || FLOATING.has(entry.value)) continue
      if (onLine(entry.value)) goPin()
      else at(file, entry.line, `Go ${entry.value} in the matrix, go.mod is on Go ${want}; use ${want}.x`)
    }
  }

  // no file pins Node.js for the workflows: they only have to agree
  const counts = new Map()
  for (const { value } of node) counts.set(value, (counts.get(value) ?? 0) + 1)
  const common = [...counts].sort((a, b) => b[1] - a[1])[0]?.[0]
  for (const { file, line, value } of node) {
    if (value === common) pin(`Node.js ${value}`)
    else at(file, line, `node-version ${value}, the other workflows use ${common}`)
  }

  // = and := set the list, += extends it, ?= sets it once; conditionals are not
  // followed
  let goVersions = null
  read('Makefile')?.forEach((line, i) => {
    const m = line.replace(/(?<!\\)#.*$/, '').match(/^(?:(?:export|override)\s+)*GO_VERSIONS\s*(\+=|\?=|:{0,3}=)\s*(.*)$/)
    if (!m || (m[1] === '?=' && goVersions)) return
    const versions = m[2].trim().split(/\s+/).filter(Boolean)
    goVersions = { versions: m[1] === '+=' ? [...(goVersions?.versions ?? []), ...versions] : versions, line: i + 1 }
  })
  if (goVersions?.versions.some(onLine)) goPin()
  else if (goVersions) at('Makefile', goVersions.line, `GO_VERSIONS ${goVersions.versions.join(' ')} has no ${want}, the Go line of go.mod`)

  for (const file of files(root, DOCS, ['.md', '.mdx']).filter((f) => !f.startsWith(NOTES))) {
    read(file).forEach((line, i) => {
      for (const [claim, version] of line.matchAll(/\bGo (\d+\.\d+(?:\.\d+)?) or newer\b/g)) {
        if (onLine(version)) goPin()
        else at(file, i + 1, `"${claim}", go.mod is on Go ${want}`)
      }
    })
  }

  // in the order the files were read, by line within a file
  const errors = found
    .sort((a, b) => order.indexOf(a.file) - order.indexOf(b.file) || a.line - b.line)
    .map(({ file, line, problem }) => `${file}:${line}: ${problem}`)
  return { errors, summary: [...seen].map(([label, n]) => `${label} (${n})`).join(', ') }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const { errors, summary } = toolPins(process.argv[2] || '.')
  errors.forEach((e) => console.error(`tool-pins: ${e}`))
  if (!errors.length) console.log(`tool-pins: ${summary}`)
  process.exitCode = errors.length ? 1 : 0
}
