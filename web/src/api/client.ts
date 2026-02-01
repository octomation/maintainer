import type {
  ApiErrorBody,
  Catalog,
  CheckoutsData,
  Envelope,
  TaskDetail,
  TaskPage,
  TaskQuery,
  TriagePatch,
} from './types';

/** An answer of the API that failed: the status and the error of the body. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

/** A detail answer carries the ETag the next PATCH must send back. */
export interface Versioned<T> {
  envelope: Envelope<T>;
  etag: string;
}

/** GET /api/v0/checkouts answers 202 without data before the first snapshot. */
export interface CheckoutsAnswer {
  status: number;
  envelope: Envelope<CheckoutsData | null>;
}

/** The query parameters of GET /api/v0/tasks as URLSearchParams. */
export function taskQueryParams(query: TaskQuery): URLSearchParams {
  const params = new URLSearchParams();
  const lists: [string, string[] | undefined][] = [
    ['owner', query.owner],
    ['repo', query.repo],
    ['milestone', query.milestone],
    ['label', query.label],
    ['type', query.type],
    ['assignee', query.assignee],
    ['column', query.column],
  ];
  for (const [name, values] of lists) {
    for (const value of values ?? []) params.append(name, value);
  }
  if (query.state) params.set('state', query.state);
  if (query.q) params.set('q', query.q);
  if (query.sort) params.set('sort', query.sort);
  if (query.limit !== undefined) params.set('limit', String(query.limit));
  if (query.offset !== undefined) params.set('offset', String(query.offset));
  return params;
}

export type Fetch = typeof fetch;

/** The client of /api/v0. Every mutation carries X-Requested-With: maintainer. */
export class ApiClient {
  private readonly base: string;
  private readonly fetcher: Fetch;

  constructor(base = '', fetcher: Fetch = (input, init) => fetch(input, init)) {
    this.base = base.replace(/\/$/, '');
    this.fetcher = fetcher;
  }

  async getCheckouts(params: { q?: string; sort?: string } = {}): Promise<CheckoutsAnswer> {
    const search = new URLSearchParams();
    if (params.q) search.set('q', params.q);
    if (params.sort) search.set('sort', params.sort);
    const suffix = search.size > 0 ? `?${search}` : '';
    const response = await this.fetcher(`${this.base}/api/v0/checkouts${suffix}`, { headers: { Accept: 'application/json' } });
    const envelope = await parse<Envelope<CheckoutsData | null>>(response);
    return { status: response.status, envelope };
  }

  async refreshCheckouts(): Promise<Envelope<null>> {
    return this.mutate<Envelope<null>>('/api/v0/checkouts/refresh', 'POST');
  }

  /** GET /api/v0/tasks; before the first index the answer is 202 with data null. */
  async listTasks(query: TaskQuery): Promise<{ status: number; envelope: Envelope<TaskPage | null> }> {
    const params = taskQueryParams(query);
    const suffix = params.size > 0 ? `?${params}` : '';
    const response = await this.fetcher(`${this.base}/api/v0/tasks${suffix}`, { headers: { Accept: 'application/json' } });
    return { status: response.status, envelope: await parse<Envelope<TaskPage | null>>(response) };
  }

  async getTask(id: string): Promise<Versioned<TaskDetail>> {
    const response = await this.fetcher(`${this.base}/api/v0/tasks/${encodeURIComponent(id)}`, { headers: { Accept: 'application/json' } });
    if (response.status === 202) throw new ApiError(202, 'indexing', 'the index is not ready');
    const envelope = await parse<Envelope<TaskDetail>>(response);
    return { envelope, etag: response.headers.get('ETag') ?? `"${envelope.data.version}"` };
  }

  /** PATCH …/triage with If-Match; a 412 or 409 arrives as an ApiError. */
  async triage(id: string, etag: string, patch: TriagePatch): Promise<Versioned<TaskDetail>> {
    const response = await this.fetcher(`${this.base}/api/v0/tasks/${encodeURIComponent(id)}/triage`, {
      method: 'PATCH',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        'X-Requested-With': 'maintainer',
        'If-Match': etag,
      },
      body: JSON.stringify(patch),
    });
    const envelope = await parse<Envelope<TaskDetail>>(response);
    return { envelope, etag: response.headers.get('ETag') ?? `"${envelope.data.version}"` };
  }

  async getCatalog(): Promise<{ status: number; envelope: Envelope<Catalog | null> }> {
    const response = await this.fetcher(`${this.base}/api/v0/catalog`, { headers: { Accept: 'application/json' } });
    return { status: response.status, envelope: await parse<Envelope<Catalog | null>>(response) };
  }

  async refresh(): Promise<Envelope<null>> {
    return this.mutate<Envelope<null>>('/api/v0/refresh', 'POST');
  }

  private async mutate<T>(path: string, method: string): Promise<T> {
    const response = await this.fetcher(`${this.base}${path}`, {
      method,
      headers: { Accept: 'application/json', 'X-Requested-With': 'maintainer' },
    });
    return parse<T>(response);
  }
}

async function parse<T>(response: Response): Promise<T> {
  const text = await response.text();
  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text) as unknown;
    } catch {
      throw new ApiError(response.status, 'malformed', `the service answered ${response.status} without JSON`);
    }
  }
  if (!response.ok && response.status !== 202) {
    const error = (body as ApiErrorBody | null)?.error;
    throw new ApiError(response.status, error?.code ?? 'error', error?.message ?? `the service answered ${response.status}`);
  }
  return body as T;
}
