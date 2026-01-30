import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import test from 'node:test'
import { toolPins } from './tool-pins-check.mjs'

const script = resolve('.github/scripts/tool-pins-check.mjs')
// A repository whose pins agree, in the forms the workflows use; each test
// overrides some of its files.
const agreed = {
  'go.mod': 'module example.com/app\n\ngo 1.27.0\n\nrequire github.com/spf13/cobra v1.8.0\n',
  'tools/go.mod': [
    'module tools', '', 'go 1.27.1', '',
    'tool (', '\tgithub.com/golangci/golangci-lint/v2/cmd/golangci-lint', '\tgithub.com/goreleaser/goreleaser/v2', ')', '',
    'require (',
    '\tgithub.com/golangci/golangci-lint/v2 v2.14.0 // indirect',
    '\tgithub.com/goreleaser/goreleaser/v2 v2.18.2 // indirect',
    ')', '',
  ].join('\n'),
  '.github/workflows/ci.yml': `jobs:
  lint:
    steps:
      - name: Set up Go environment
        uses: actions/setup-go@v7.0.0
        with:
          go-version: 1.27.x
          check-latest: true
      - name: Run linter
        uses: golangci/golangci-lint-action@v9.3.0
        with:
          version: v2.14.0
          args: --fast-only
      - name: Set up Node.js environment
        uses: actions/setup-node@v7.1.0
        with: { node-version: 24, package-manager-cache: false }
  test:
    strategy:
      matrix:
        go:
          - 1.x
          - 1.27.x
    steps:
      - uses: actions/setup-go@v7.0.0
        with:
          go-version: '\${{ matrix.go }}'
      - name: Run tests with coverage report
        run: make test-with-coverage
        if: matrix.go == '1.27.x'
`,
  '.github/workflows/cd.yml': `jobs:
  release:
    steps:
      - name: Publish the release
        uses: goreleaser/goreleaser-action@v7.3.0
        env:
          GITHUB_TOKEN: \${{ github.token }}
        with:
          distribution: goreleaser
          version: v2.18.2
          args: release --clean
      - uses: actions/setup-node@v7.1.0
        with:
          node-version: 24
`,
  'Makefile': 'GO_VERSIONS   = 1.27\n',
  'README.md': 'Install the latest release with Go 1.27 or newer.\n',
  'docs/content/guide.mdx': 'Build it with Go 1.27 or newer.\n',
  'docs/content/changelog/v0.1.0.md': 'Requires Go 1.26 or newer.\n',
}

function fixture(files, fn) {
  const cwd = mkdtempSync(join(tmpdir(), 'tool-pins-'))
  try {
    for (const [file, text] of Object.entries({ ...agreed, ...files })) {
      if (text === null) continue
      mkdirSync(dirname(join(cwd, file)), { recursive: true })
      writeFileSync(join(cwd, file), text)
    }
    const run = () => spawnSync(process.execPath, [script, cwd], { encoding: 'utf8' })
    fn({ cwd, run })
  } finally { rmSync(cwd, { recursive: true, force: true }) }
}

const bump = (file, from, to) => ({ [file]: agreed[file].replaceAll(from, to) })

test('agreeing pins pass and are counted', () => fixture({}, ({ cwd, run }) => {
  assert.deepEqual(toolPins(cwd), {
    errors: [],
    summary: 'Go 1.27 (8), goreleaser v2.18.2 (1), golangci-lint v2.14.0 (1), Node.js 24 (2)',
  })
  const result = run()
  assert.equal(result.status, 0, result.stderr)
  assert.equal(result.stdout, 'tool-pins: Go 1.27 (8), goreleaser v2.18.2 (1), golangci-lint v2.14.0 (1), Node.js 24 (2)\n')
}))

test('a tool moved in tools/go.mod fails every action that pins the old release', () => fixture({
  ...bump('tools/go.mod', 'v2.18.2', 'v2.18.3'),
}, ({ cwd, run }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    '.github/workflows/cd.yml:10: goreleaser v2.18.2, tools/go.mod requires github.com/goreleaser/goreleaser/v2 v2.18.3',
  ])
  const result = run()
  assert.equal(result.status, 1)
  assert.equal(result.stdout, '')
  assert.equal(result.stderr, 'tool-pins: .github/workflows/cd.yml:10: goreleaser v2.18.2, tools/go.mod requires github.com/goreleaser/goreleaser/v2 v2.18.3\n')
}))

test('a new major version of a tool is found by its module path', () => fixture({
  ...bump('tools/go.mod', 'github.com/goreleaser/goreleaser/v2 v2.18.2', 'github.com/goreleaser/goreleaser/v3 v3.0.0'),
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    '.github/workflows/cd.yml:10: goreleaser v2.18.2, tools/go.mod requires github.com/goreleaser/goreleaser/v3 v3.0.0',
  ])
}))

test('a new Go line in go.mod fails every Go version left on the old one', () => fixture({
  ...bump('go.mod', 'go 1.27.0', 'go 1.28.0'),
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    'tools/go.mod:3: go 1.27.1, go.mod is on Go 1.28',
    '.github/workflows/ci.yml:7: go-version 1.27.x, go.mod is on Go 1.28; pin 1.28.x',
    '.github/workflows/ci.yml:22: Go 1.27.x in the matrix, go.mod is on Go 1.28; use 1.28.x',
    '.github/workflows/ci.yml:29: Go 1.27.x in the matrix, go.mod is on Go 1.28; use 1.28.x',
    'Makefile:1: GO_VERSIONS 1.27 has no 1.28, the Go line of go.mod',
    'README.md:1: "Go 1.27 or newer", go.mod is on Go 1.28',
    'docs/content/guide.mdx:1: "Go 1.27 or newer", go.mod is on Go 1.28',
  ])
}))

test('a toolchain line and a list of Go versions in the Makefile follow go.mod', () => fixture({
  'go.mod': `${agreed['go.mod']}\ntoolchain go1.28.1\n`,
  'Makefile': 'GO_VERSIONS = 1.26 1.27\n',
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, ['go.mod:7: toolchain 1.28.1, go.mod is on Go 1.27'])
}))

test('floating, missing and unpinned versions fail', () => fixture({
  ...bump('.github/workflows/ci.yml', `        with:
          version: v2.14.0
          args: --fast-only`, `        with: { args: --fast-only }`),
  ...bump('.github/workflows/cd.yml', 'version: v2.18.2', "version: '~> v2'"),
  'tools/go.mod': agreed['tools/go.mod'].replace('1.27.1', '1.27.1\n\ntoolchain go1.27.2'),
  '.github/workflows/docs.yml': `jobs:
  build:
    steps:
      - uses: actions/setup-go@v7.0.0
        with: { go-version: stable }
      - uses: actions/setup-node@v7.1.0 # pinned by hand
        with:
          # the docs build needs the next major
          node-version: '26'
`,
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    '.github/workflows/cd.yml:10: goreleaser ~> v2, tools/go.mod requires github.com/goreleaser/goreleaser/v2 v2.18.2',
    '.github/workflows/ci.yml:10: golangci/golangci-lint-action without a version installs the latest golangci-lint; pin v2.14.0 from tools/go.mod',
    '.github/workflows/docs.yml:5: go-version stable, go.mod is on Go 1.27; pin 1.27.x',
    '.github/workflows/docs.yml:9: node-version 26, the other workflows use 24',
  ])
}))

test('the inputs of a step are read before its uses key too', () => fixture({
  '.github/workflows/lint.yml': `jobs:
  lint:
    steps:
      - name: Run linter
        with:
          version: v2.13.0
        uses: golangci/golangci-lint-action@v9.3.0
      - uses: actions/checkout@v7.0.1
        with:
          version: v0.0.1
`,
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    '.github/workflows/lint.yml:6: golangci-lint v2.13.0, tools/go.mod requires github.com/golangci/golangci-lint/v2 v2.14.0',
  ])
}))

test('flow matrices, reversed comparisons and comments are read as YAML', () => fixture({
  '.github/workflows/matrix.yml': `jobs:
  inline:
    strategy:
      matrix: { go: ['1.x', '1.26.x'] }
  scalar:
    strategy:
      matrix: { os: ubuntu-24.04, go: 1.21.x }
  multiline:
    strategy:
      matrix:
        go: [
          '1.x',
          '1.25.x'
        ]
    steps:
      # was: matrix.go == '1.24.x'
      - run: make test # matrix.go == '1.23.x'
        if: \${{ '1.22.x' == matrix.go }}
      - run: make test
        if: \${{ inputs.reason != 'see #175' || matrix.go == '1.20.x' }}
  mapped:
    strategy:
      matrix:
        include:
          - {
              go: '1.27.x',
              os: ubuntu-24.04
            }
          - {
              os: macos-15,
              go: '1.19.x'
            }
`,
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    '.github/workflows/matrix.yml:4: Go 1.26.x in the matrix, go.mod is on Go 1.27; use 1.27.x',
    '.github/workflows/matrix.yml:7: Go 1.21.x in the matrix, go.mod is on Go 1.27; use 1.27.x',
    '.github/workflows/matrix.yml:13: Go 1.25.x in the matrix, go.mod is on Go 1.27; use 1.27.x',
    '.github/workflows/matrix.yml:18: Go 1.22.x in the matrix, go.mod is on Go 1.27; use 1.27.x',
    '.github/workflows/matrix.yml:20: Go 1.20.x in the matrix, go.mod is on Go 1.27; use 1.27.x',
    '.github/workflows/matrix.yml:31: Go 1.19.x in the matrix, go.mod is on Go 1.27; use 1.27.x',
  ])
}))

test('GO_VERSIONS is read as make reads it, without comments', () => {
  fixture({ 'Makefile': 'GO_VERSIONS = 1.27\nGO_VERSIONS += 1.26 1.25\nGO_VERSIONS ?= 1.24\n' }, ({ cwd }) => {
    assert.deepEqual(toolPins(cwd).errors, [])
  })
  fixture({ 'Makefile': 'GO_VERSIONS = 1.27\nGO_VERSIONS := 1.26 # TODO: 1.27\n' }, ({ cwd }) => {
    assert.deepEqual(toolPins(cwd).errors, ['Makefile:2: GO_VERSIONS 1.26 has no 1.27, the Go line of go.mod'])
  })
})

test('an action needs its tool in tools/go.mod', () => fixture({
  ...bump('tools/go.mod', '\tgithub.com/golangci/golangci-lint/v2 v2.14.0 // indirect\n', ''),
}, ({ cwd }) => {
  assert.deepEqual(toolPins(cwd).errors, [
    '.github/workflows/ci.yml:10: golangci/golangci-lint-action installs golangci-lint, but tools/go.mod does not require github.com/golangci/golangci-lint',
  ])
}))

test('a go.mod without a go line cannot be checked', () => fixture({ 'go.mod': 'module example.com/app\n' }, ({ run }) => {
  const result = run()
  assert.equal(result.status, 1)
  assert.equal(result.stderr, 'tool-pins: go.mod: no go line to check the Go versions against\n')
}))
