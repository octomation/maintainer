import { describe, expect, it } from 'vitest';
import { ApiClient, ApiError, taskQueryParams } from './client';
import { fixtures, json, mockApi } from './mock';

describe('ApiClient', () => {
  it('lists tasks with the query parameters of the contract', async () => {
    const api = mockApi();
    const client = new ApiClient('', api.fetch);
    const { status, envelope } = await client.listTasks({ repo: ['octomation/maintainer'], state: 'all', label: ['a', 'b'], limit: 5, sort: 'updated' });
    expect(status).toBe(200);
    expect(envelope.data?.total).toBe(fixtures.tasks.data.total);
    expect(api.calls[0].url.pathname).toBe('/api/v0/tasks');
    expect(api.calls[0].url.searchParams.getAll('label')).toEqual(['a', 'b']);
    expect(api.calls[0].url.searchParams.get('repo')).toBe('octomation/maintainer');
    expect(api.calls[0].url.searchParams.get('limit')).toBe('5');
    expect(api.calls[0].url.searchParams.get('sort')).toBe('updated');
  });

  it('reports the indexing state as 202 with data null', async () => {
    const client = new ApiClient('', mockApi({ listStatus: 202 }).fetch);
    const { status, envelope } = await client.listTasks({});
    expect(status).toBe(202);
    expect(envelope.refreshing).toBe(true);
    expect(envelope.data).toBeNull();
  });

  it('keeps the ETag of a task and sends it back on triage with the mutation headers', async () => {
    const api = mockApi();
    const client = new ApiClient('', api.fetch);
    const task = await client.getTask(fixtures.task.data.id);
    expect(task.etag).toBe(`"${fixtures.task.data.version}"`);
    const updated = await client.triage(task.envelope.data.id, task.etag, { impact: 'low', effort: null });
    expect(updated.envelope.data.labels).toContain('impact: low');
    expect(updated.envelope.data.labels.some((l) => l.startsWith('effort:'))).toBe(false);
    expect(updated.etag).not.toBe(task.etag);
    const call = api.calls[1];
    expect(call.method).toBe('PATCH');
    expect(call.headers.get('X-Requested-With')).toBe('maintainer');
    expect(call.headers.get('Content-Type')).toBe('application/json');
    expect(call.headers.get('If-Match')).toBe(task.etag);
    expect(JSON.parse(call.body ?? '')).toEqual({ impact: 'low', effort: null });
  });

  it('throws an ApiError with the status and code of a failed answer', async () => {
    const api = mockApi();
    const client = new ApiClient('', api.fetch);
    await expect(client.triage(fixtures.task.data.id, '"sha256-stale"', { impact: 'low' })).rejects.toMatchObject({
      name: 'ApiError',
      status: 412,
      code: 'precondition_failed',
    });
    const conflict = mockApi({ route: (call) => (call.method === 'PATCH' ? json({ error: { code: 'conflict', message: 'duplicate mirrors' } }, 409) : undefined) });
    const error = await new ApiClient('', conflict.fetch).triage('x', '"v"', { impact: 'low' }).catch((e: unknown) => e);
    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).status).toBe(409);
    expect((error as ApiError).message).toBe('duplicate mirrors');
  });

  it('sends the mutation header on refreshes', async () => {
    const api = mockApi();
    const client = new ApiClient('', api.fetch);
    await client.refresh();
    await client.refreshCheckouts();
    expect(api.calls.map((c) => [c.method, c.url.pathname, c.headers.get('X-Requested-With')])).toEqual([
      ['POST', '/api/v0/refresh', 'maintainer'],
      ['POST', '/api/v0/checkouts/refresh', 'maintainer'],
    ]);
  });

  it('reads checkouts with the sort chain and the query', async () => {
    const api = mockApi();
    const client = new ApiClient('', api.fetch);
    const { status, envelope } = await client.getCheckouts({ q: 'octo', sort: '-status,repository' });
    expect(status).toBe(200);
    expect(envelope.data?.rows.length).toBeGreaterThan(0);
    expect(envelope.data?.rows[0].cells).toHaveLength(6);
    expect(api.calls[0].url.search).toBe('?q=octo&sort=-status%2Crepository');
  });

  it('builds query params only for the set fields', () => {
    expect(taskQueryParams({}).toString()).toBe('');
    expect(taskQueryParams({ column: ['todo', 'done'], q: 'x', offset: 0 }).toString()).toBe('column=todo&column=done&q=x&offset=0');
  });
});
