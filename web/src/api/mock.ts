// A fetch of /api/v0 over the JSON fixtures, for component tests: the pages
// render against the same answers the Go service gives for its fixture
// mirror, and a test can override any route or make a mutation fail.

import type { Fetch } from './client';
import type { Envelope, TaskDetail, TaskPage, TriagePatch } from './types';
import catalogFixture from './fixtures/catalog.json';
import checkoutsFixture from './fixtures/checkouts.json';
import checkoutsPendingFixture from './fixtures/checkouts-pending.json';
import errorFixture from './fixtures/error.json';
import taskFixture from './fixtures/task.json';
import tasksFixture from './fixtures/tasks.json';

export const fixtures = {
  catalog: catalogFixture,
  checkouts: checkoutsFixture,
  checkoutsPending: checkoutsPendingFixture,
  error: errorFixture,
  task: taskFixture as Envelope<TaskDetail>,
  tasks: tasksFixture as Envelope<TaskPage>,
};

export interface MockCall {
  method: string;
  url: URL;
  headers: Headers;
  body: string | null;
}

export interface MockOptions {
  /** Answer a route yourself: return undefined to fall back to the fixtures. */
  route?: (call: MockCall) => Response | undefined | Promise<Response | undefined>;
  /** The tasks every list answer returns; the fixture page by default. */
  tasks?: Envelope<TaskPage>;
  /** The detail every task answer returns; the fixture task by default. */
  task?: Envelope<TaskDetail>;
  /** The status the first list answers get, such as 202 for the indexing state. */
  listStatus?: number;
}

export interface MockApi {
  fetch: Fetch;
  calls: MockCall[];
}

export function json(body: unknown, status = 200, headers: Record<string, string> = {}): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json', ...headers } });
}

/** A mock fetch that answers like the service over the fixtures. */
export function mockApi(options: MockOptions = {}): MockApi {
  const calls: MockCall[] = [];
  let detail = clone(options.task ?? fixtures.task);
  const fetcher: Fetch = async (input, init) => {
    const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, 'http://maintainer.localhost');
    const method = init?.method ?? 'GET';
    const headers = new Headers(init?.headers);
    const body = typeof init?.body === 'string' ? init.body : null;
    const call: MockCall = { method, url, headers, body };
    calls.push(call);
    const custom = await options.route?.(call);
    if (custom) return custom;
    const path = url.pathname;
    if (method === 'GET' && path === '/api/v0/tasks') {
      const page = clone(options.tasks ?? fixtures.tasks);
      if (options.listStatus === 202) return json({ ...page, refreshing: true, data: null }, 202);
      return json(page);
    }
    if (method === 'GET' && path === '/api/v0/catalog') return json(fixtures.catalog);
    if (method === 'GET' && path === '/api/v0/checkouts') return json(fixtures.checkouts);
    if (method === 'POST' && (path === '/api/v0/refresh' || path === '/api/v0/checkouts/refresh')) {
      return json({ generation: detail.generation, collected_at: detail.collected_at, refreshing: true, warnings: [], data: null }, 202);
    }
    const triage = path.match(/^\/api\/v0\/tasks\/([^/]+)\/triage$/);
    if (method === 'PATCH' && triage) {
      if (headers.get('If-Match') !== `"${detail.data.version}"`) {
        return json({ error: { code: 'precondition_failed', message: 'the file changed; reload the task' } }, 412);
      }
      const patch = JSON.parse(body ?? '{}') as TriagePatch;
      const labels = detail.data.labels.filter((label) => {
        const lower = label.toLowerCase();
        return !((patch.impact !== undefined && lower.startsWith('impact:')) || (patch.effort !== undefined && lower.startsWith('effort:')));
      });
      if (patch.impact) labels.push(`impact: ${patch.impact}`);
      if (patch.effort) labels.push(`effort: ${patch.effort}`);
      detail = { ...detail, generation: detail.generation + 1, data: { ...detail.data, labels, version: `sha256-${detail.generation + 1}` } };
      return json(detail, 200, { ETag: `"${detail.data.version}"` });
    }
    const one = path.match(/^\/api\/v0\/tasks\/([^/]+)$/);
    if (method === 'GET' && one) return json(detail, 200, { ETag: `"${detail.data.version}"` });
    return json(fixtures.error, 404);
  };
  return { fetch: fetcher, calls };
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}
