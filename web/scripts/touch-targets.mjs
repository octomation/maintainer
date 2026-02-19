#!/usr/bin/env node
// Checks the touch targets of the hub at 390 px in headless Chrome: every
// interactive element of /status, /inbox, /tablo and the issue page, and of
// each menu, panel and dialog they open, must offer a target of at least
// 44×44 px. The target is what the hit test answers, so a control may look
// smaller than its target (a padding or a ::before around it).
//
//   node web/scripts/touch-targets.mjs --binary <maintainer> [--chrome <path>] [--shots <dir>]
//
// It starts the hub from the binary on a free port with a temporary config,
// state and workspace (one checkout with the issues of this repository), and
// never touches the hub on 17841 or your own configuration.
import { execFileSync, spawn } from 'node:child_process';
import { cpSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const MIN = 44;
const WIDTH = 390;
const HEIGHT = 844;
const repo = resolve(dirname(fileURLToPath(import.meta.url)), '../..');

function args() {
  const out = { chrome: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' };
  const argv = process.argv.slice(2);
  for (let i = 0; i < argv.length; i += 2) {
    const name = argv[i].replace(/^--/, '');
    if (!['binary', 'chrome', 'shots'].includes(name) || argv[i + 1] === undefined) {
      throw new Error(`usage: touch-targets.mjs --binary <maintainer> [--chrome <path>] [--shots <dir>]`);
    }
    out[name] = argv[i + 1];
  }
  if (!out.binary) throw new Error('--binary is required');
  return out;
}

function freePort() {
  return new Promise((ok, fail) => {
    const server = createServer();
    server.once('error', fail);
    server.listen(0, '127.0.0.1', () => {
      const { port } = server.address();
      server.close(() => ok(port));
    });
  });
}

const wait = (ms) => new Promise((ok) => setTimeout(ok, ms));

async function until(what, check, timeout = 30000) {
  const end = Date.now() + timeout;
  for (;;) {
    const value = await check();
    if (value) return value;
    if (Date.now() > end) throw new Error(`timed out waiting for ${what}`);
    await wait(100);
  }
}

/** The environment of the scratch hub and its Git: no config, state, token or signing key of the user. */
function isolated(root) {
  const env = {
    ...process.env,
    HOME: root,
    XDG_CONFIG_HOME: join(root, 'config'),
    XDG_STATE_HOME: join(root, 'state'),
    XDG_CACHE_HOME: join(root, 'cache'),
    GIT_CONFIG_GLOBAL: '/dev/null',
    GIT_CONFIG_NOSYSTEM: '1',
    GIT_AUTHOR_NAME: 'touch',
    GIT_AUTHOR_EMAIL: 'touch@localhost',
    GIT_COMMITTER_NAME: 'touch',
    GIT_COMMITTER_EMAIL: 'touch@localhost',
  };
  for (const name of ['MAINTAINER_FETCH_CONFIG', 'GITHUB_TOKEN', 'GH_TOKEN', 'GIT_CONFIG', 'GIT_CONFIG_COUNT', 'GIT_CONFIG_PARAMETERS']) delete env[name];
  return env;
}

/** A scratch workspace: one checkout of octomation/maintainer with the issue mirror of this tree. */
function workspace(root) {
  const checkout = join(root, 'ws/public/octomation/maintainer');
  mkdirSync(join(checkout, '.github'), { recursive: true });
  for (const dir of ['issues', 'roadmap']) cpSync(join(repo, '.github', dir), join(checkout, '.github', dir), { recursive: true });
  const git = (...a) => execFileSync('git', a, { cwd: checkout, env: isolated(root), stdio: 'ignore' });
  git('init', '-q', '-b', 'main');
  git('remote', 'add', 'origin', 'https://github.com/octomation/maintainer.git');
  git('add', '-A');
  git('commit', '-qm', 'mirror');
  const config = join(root, 'config/maintainer');
  mkdirSync(config, { recursive: true });
  writeFileSync(join(config, 'fetch.toml'), `[workspace]\nroot = ${JSON.stringify(join(root, 'ws'))}\npath = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"\n`);
  writeFileSync(join(config, 'service.toml'), 'access = "open"\n');
}

/** Stops a child process and resolves once it has exited: SIGTERM, then SIGKILL after 5 s. */
async function stop(proc) {
  if (proc.exitCode !== null || proc.signalCode !== null) return;
  const exited = new Promise((ok) => proc.once('exit', ok));
  proc.kill('SIGTERM');
  if ((await Promise.race([exited.then(() => true), wait(5000).then(() => false)])) === false) {
    proc.kill('SIGKILL');
    await exited;
  }
}

function hub(binary, root, port) {
  return spawn(resolve(binary), ['service', 'run', '--listen', `127.0.0.1:${port}`], { cwd: root, env: isolated(root), stdio: ['ignore', 'ignore', 'pipe'] });
}

/** A minimal Chrome DevTools Protocol client over the browser's WebSocket. */
async function browser(chrome, root, track) {
  const profile = join(root, 'chrome');
  const proc = spawn(chrome, ['--headless=new', '--remote-debugging-port=0', `--user-data-dir=${profile}`, '--no-first-run', '--no-default-browser-check', '--hide-scrollbars', 'about:blank'], { stdio: 'ignore' });
  track(proc);
  const file = join(profile, 'DevToolsActivePort');
  const [port, path] = (await until('Chrome', () => existsSync(file) && readFileSync(file, 'utf8').trim().split('\n').length === 2 && readFileSync(file, 'utf8').trim().split('\n'))).map((s) => s.trim());
  const ws = new WebSocket(`ws://127.0.0.1:${port}${path}`);
  await new Promise((ok, fail) => {
    ws.onopen = ok;
    ws.onerror = fail;
  });
  let id = 0;
  const pending = new Map();
  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    const call = pending.get(message.id);
    if (!call) return;
    pending.delete(message.id);
    if (message.error) call.fail(new Error(`${call.method}: ${message.error.message}`));
    else call.ok(message.result);
  };
  const send = (method, params = {}, sessionId) =>
    new Promise((ok, fail) => {
      const n = ++id;
      pending.set(n, { ok, fail, method });
      ws.send(JSON.stringify({ id: n, method, params, sessionId }));
    });
  const { targetId } = await send('Target.createTarget', { url: 'about:blank' });
  const { sessionId } = await send('Target.attachToTarget', { targetId, flatten: true });
  const page = (method, params) => send(method, params, sessionId);
  await page('Page.enable');
  await page('Runtime.enable');
  await page('Emulation.setDeviceMetricsOverride', { width: WIDTH, height: HEIGHT, deviceScaleFactor: 2, mobile: true });
  return {
    page,
    async evaluate(fn, ...values) {
      const expression = `(${fn})(...${JSON.stringify(values)})`;
      const { result, exceptionDetails } = await page('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true });
      if (exceptionDetails) throw new Error(exceptionDetails.exception?.description ?? exceptionDetails.text);
      return result.value;
    },
    close() {
      ws.close();
    },
  };
}

// The code below runs in the page.

/** The page has loaded its data: no skeleton, no progress line, nothing busy. */
function settled() {
  return document.readyState === 'complete' && document.querySelector('#app *') !== null && !document.querySelector('.skeleton, .progress-line, [aria-busy="true"]');
}

const INTERACTIVE = 'a[href], button, input:not([type="hidden"]), select, textarea, summary, [role="button"], [role="link"], [role="tab"], [role="option"], [role="menuitem"], [role="menuitemradio"], [role="menuitemcheckbox"], [role="treeitem"], [role="switch"], [role="checkbox"], [role="radio"], [tabindex]:not([tabindex="-1"])';
// What opens a menu, a panel or a dialog, and the dimensions of the Filter menu, which open their values.
const OPENERS = '[aria-haspopup], .topbar__menu, .filter-menu button.menu-item';

/** The openers that are reachable now, as their index among all of them. */
function triggers(selector) {
  return [...document.querySelectorAll(selector)].map((el, i) => [el, i]).filter(([el]) => el.getClientRects().length && !el.closest('[inert]') && !el.disabled).map(([, i]) => i);
}

function click(selector, index) {
  const el = document.querySelectorAll(selector)[index];
  if (!el) return false;
  el.scrollIntoView({ block: 'center', inline: 'center' });
  el.click();
  return true;
}

/**
 * Measures every interactive element the hit test reaches: the extent of the
 * target along the lines through its center, and whether a circle of 44 px
 * around the center belongs to it, so a rounded 44 px button passes and a
 * neighbor that overlaps the target does not. An element under another one is
 * not reachable in this state and is counted apart; one whose center is
 * covered while a part of it answers fails. With mark, it only marks
 * what it would measure; afterwards the marked elements are skipped, so the
 * state of an open menu measures what the menu brought.
 */
function measure(selector, min, mark) {
  const owns = (el, x, y) => {
    if (x < 0 || y < 0 || x >= innerWidth || y >= innerHeight) return false;
    const hit = document.elementFromPoint(x, y);
    return !!hit && el.contains(hit);
  };
  const describe = (el) => {
    const label = el.getAttribute('aria-label') || el.getAttribute('title') || el.textContent.trim().replace(/\s+/g, ' ').slice(0, 40);
    const cls = [...el.classList].slice(0, 2).map((c) => `.${c}`).join('');
    return `${el.tagName.toLowerCase()}${cls} "${label}"`;
  };
  const results = [];
  let covered = 0;
  for (const el of document.querySelectorAll(selector)) {
    if (el.dataset.touchSeen) continue;
    if (el.closest('[inert], [aria-hidden="true"]') || el.disabled) continue;
    if (!el.getClientRects().length) continue;
    const style = getComputedStyle(el);
    if (style.visibility === 'hidden' || style.pointerEvents === 'none') continue;
    // A link in the Markdown of an issue or a comment follows its line of text: the inline exception of WCAG 2.5.8.
    if (el.tagName === 'A' && el.closest('.markdown')) continue;
    el.scrollIntoView({ block: 'center', inline: 'center' });
    // A field inside its label takes the label as its target.
    const target = (el.tagName === 'INPUT' && el.closest('label')) || el;
    const r = target.getBoundingClientRect();
    if (r.width === 0 || r.height === 0) continue;
    const cx = r.left + r.width / 2;
    const cy = r.top + r.height / 2;
    if (!owns(target, cx, cy)) {
      // Under another element at its center: unreachable, unless a part of it still answers the hit test.
      let part = false;
      for (let i = 0; i <= 8 && !part; i++) {
        for (let j = 0; j <= 8 && !part; j++) part = owns(target, r.left + 0.5 + ((r.width - 1) * i) / 8, r.top + 0.5 + ((r.height - 1) * j) / 8);
      }
      if (!part) {
        covered++;
        continue;
      }
      if (mark) el.dataset.touchSeen = '1';
      else results.push({ what: describe(el), box: `${Math.round(r.width)}×${Math.round(r.height)}`, target: 'a part beside its center', ok: false });
      continue;
    }
    if (mark) {
      el.dataset.touchSeen = '1';
      continue;
    }
    const reach = (dx, dy) => {
      let d = 0;
      while (d < 60 && owns(target, cx + dx * (d + 1), cy + dy * (d + 1))) d++;
      return d;
    };
    const span = (a, b) => (a === 60 || b === 60 ? `${a + b + 1}+` : `${a + b + 1}`);
    const [left, right, up, down] = [reach(-1, 0), reach(1, 0), reach(0, -1), reach(0, 1)];
    let circle = true;
    // One pixel short of the half: the layout places boxes at fractions of a pixel.
    for (const radius of [min / 2 - 1, min / 4]) {
      for (let k = 0; k < 24 && circle; k++) {
        const angle = (Math.PI * 2 * k) / 24;
        if (!owns(target, cx + radius * Math.cos(angle), cy + radius * Math.sin(angle))) circle = false;
      }
    }
    const ok = circle && left + right + 1 >= min && up + down + 1 >= min;
    results.push({ what: describe(el), box: `${Math.round(r.width)}×${Math.round(r.height)}`, target: `${span(left, right)}×${span(up, down)}`, ok });
  }
  return { results, covered };
}

async function main() {
  const opts = args();
  const root = mkdtempSync(join(tmpdir(), 'touch-targets-'));
  let service;
  let chrome;
  let chromeProcess;
  // The same cleanup on SIGINT and SIGTERM as at the end: Chrome and the hub exit, then the scratch folder goes.
  const cleanup = async () => {
    chrome?.close();
    await Promise.all([chromeProcess && stop(chromeProcess), service && stop(service)]);
    rmSync(root, { recursive: true, force: true, maxRetries: 5, retryDelay: 100 });
  };
  for (const [signal, code] of [['SIGINT', 130], ['SIGTERM', 143]]) {
    process.once(signal, () => cleanup().finally(() => process.exit(code)));
  }
  try {
    workspace(root);
    const port = await freePort();
    service = hub(opts.binary, root, port);
    let log = '';
    service.stderr.on('data', (chunk) => (log += chunk));
    const base = `http://maintainer.localhost:${port}`;
    await until('the hub', async () => (await fetch(`${base}/status`).catch(() => null))?.ok);
    const tasks = await until('the index', async () => {
      const answer = await (await fetch(`${base}/api/v0/tasks?limit=1`)).json();
      return answer.data?.tasks?.length ? answer.data.tasks : null;
    });
    await until('the checkouts', async () => (await (await fetch(`${base}/api/v0/checkouts`)).json()).data);
    const task = encodeURIComponent(tasks[0].id);
    chrome = await browser(opts.chrome, root, (proc) => (chromeProcess = proc));
    // The pages, each also with every menu, panel and dialog one click away and one click inside it.
    const pages = {
      status: '/status',
      'status board': '/status?layout=board',
      inbox: '/inbox',
      'inbox filtered': '/inbox?repo=octomation%2Fmaintainer&label=type%3A%20bug&state=all&q=hub',
      tablo: '/tablo',
      'tablo filtered': '/tablo?repo=octomation%2Fmaintainer&state=all',
      issue: `/inbox?task=${task}`,
    };
    // The panels that a row, a card or a tree item opens.
    const panels = [
      ['status › checkout', '/status', [['.checkout', 0]]],
      ['status board › checkout', '/status?layout=board', [['.checkout-card', 0]]],
      ['inbox › reader', '/inbox', [['.inbox-row', 0]]],
      ['inbox › sidebar › owner', '/inbox', [['.topbar__menu', 0], ['[role="treeitem"][aria-expanded]', 0]]],
    ];
    const open = async (path, theme) => {
      await chrome.page('Page.navigate', { url: `${base}${path}` });
      await wait(200);
      if (theme) {
        await chrome.evaluate((t) => localStorage.setItem('maintainer.theme', t), theme);
        await chrome.page('Page.reload');
        await wait(200);
      }
      await until(`${path} to load`, () => chrome.evaluate(settled));
      await wait(300);
    };
    let failures = 0;
    let total = 0;
    const report = (name, { results, covered }) => {
      total += results.length;
      const bad = results.filter((r) => !r.ok);
      failures += bad.length;
      console.log(`${bad.length ? 'FAIL' : 'ok  '} ${name}: ${results.length} targets, ${covered} under another layer`);
      for (const r of bad) console.log(`       ${r.what}: box ${r.box}, target ${r.target}`);
      return results;
    };
    // Opens the path, clicks the steps and measures what they brought: what any earlier state reached is
    // marked; null when a step is gone. The name is the one of the last step before its click.
    const state = async (path, steps) => {
      await open(path);
      let last = '';
      for (const [selector, index] of steps) {
        await chrome.evaluate(measure, INTERACTIVE, MIN, true);
        last = await name(selector, index);
        if (!(await chrome.evaluate(click, selector, index))) return null;
        await wait(300);
      }
      return { last, ...(await chrome.evaluate(measure, INTERACTIVE, MIN)) };
    };
    const name = (selector, index) =>
      chrome.evaluate((s, n) => {
        const el = document.querySelectorAll(s)[n];
        return el?.getAttribute('aria-label') || el?.textContent.trim().replace(/\s+/g, ' ').slice(0, 30);
      }, selector, index);
    const poc = [];
    for (const [title, path] of Object.entries(pages)) {
      const results = report(title, await state(path, []));
      poc.push(...results.filter((r) => /view-switch__item|issue-header__crumb--back|prop__button/.test(r.what)).map((r) => `${title}: ${r.what}: target ${r.target}`));
      // The view switch has no empty tail: its frame is as wide as its segments.
      const tail = await chrome.evaluate(() => {
        const frame = document.querySelector('.view-switch');
        if (!frame) return null;
        return Math.round(frame.getBoundingClientRect().width - [...frame.children].reduce((w, c) => w + c.getBoundingClientRect().width, 0));
      });
      if (tail !== null) {
        poc.push(`${title}: the view switch has a tail of ${tail} px`);
        if (tail > 0) failures++;
      }
      if (opts.shots) {
        mkdirSync(opts.shots, { recursive: true });
        for (const theme of ['dark', 'light']) {
          await open(path, theme);
          const { data } = await chrome.page('Page.captureScreenshot', { format: 'png' });
          writeFileSync(join(opts.shots, `${title.replace(/ /g, '-')}-${theme}.png`), Buffer.from(data, 'base64'));
        }
        await chrome.evaluate(() => localStorage.removeItem('maintainer.theme'));
        await open(path);
      }
      const outer = await chrome.evaluate(triggers, OPENERS);
      for (const i of outer) {
        const first = await state(path, [[OPENERS, i]]);
        if (!first) continue;
        const label = first.last;
        report(`${title} › ${label}`, first);
        const inner = (await chrome.evaluate(triggers, OPENERS)).filter((j) => !outer.includes(j));
        for (const j of inner) {
          const second = await state(path, [[OPENERS, i], [OPENERS, j]]);
          if (second) report(`${title} › ${label} › ${second.last}`, second);
        }
      }
    }
    for (const [title, path, steps] of panels) {
      const result = await state(path, steps);
      if (!result) throw new Error(`${title}: nothing to click`);
      report(title, result);
    }
    // The command bar, from the keyboard.
    await open('/inbox');
    await chrome.evaluate(measure, INTERACTIVE, MIN, true);
    await chrome.page('Input.dispatchKeyEvent', { type: 'keyDown', key: '/', text: '/' });
    await chrome.page('Input.dispatchKeyEvent', { type: 'keyUp', key: '/' });
    await wait(300);
    report('inbox › command bar', await chrome.evaluate(measure, INTERACTIVE, MIN));
    console.log('\nThe controls of the issue:');
    for (const line of poc) console.log(`  ${line}`);
    console.log(failures ? `${failures} failures: a target under ${MIN}×${MIN} px or a tail of the view switch, of ${total} targets at ${WIDTH} px` : `every one of ${total} targets is at least ${MIN}×${MIN} px at ${WIDTH} px, and the view switch has no tail`);
    if (/panic|fatal/i.test(log)) console.error(log);
    process.exitCode = failures ? 1 : 0;
  } finally {
    await cleanup();
  }
}

main().catch((error) => {
  console.error(error.message);
  process.exit(2);
});
