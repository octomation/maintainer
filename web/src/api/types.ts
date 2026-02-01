// The TypeScript side of the /api/v0 contract. The Go types live in
// internal/service/hub and internal/service/web; the JSON fixtures under
// ./fixtures are generated from them and decoded back in a Go test, so a
// change on either side shows up there.

export interface Envelope<T> {
  generation: number;
  collected_at: string | null;
  refreshing: boolean;
  warnings: string[];
  data: T;
}

export interface ApiErrorBody {
  error: { code: string; message: string };
}

export type TaskState = 'open' | 'closed';
export type StateReason = 'completed' | 'not_planned' | 'duplicate' | null;
export type Column = 'backlog' | 'todo' | 'in_progress' | 'done' | 'wont_do';
export const COLUMNS: Column[] = ['backlog', 'todo', 'in_progress', 'done', 'wont_do'];
export const COLUMN_TITLES: Record<Column, string> = {
  backlog: 'Backlog',
  todo: 'To do',
  in_progress: 'In progress',
  done: 'Done',
  wont_do: "Won't do",
};

export type Impact = 'high' | 'medium' | 'low';
export type Effort = 'easy' | 'medium' | 'hard';
export const IMPACTS: Impact[] = ['high', 'medium', 'low'];
export const EFFORTS: Effort[] = ['easy', 'medium', 'hard'];

export interface MilestoneRef {
  key: string;
  title: string;
}

export interface Membership {
  board: string;
  section: string;
}

export interface Flags {
  ambiguous: boolean;
  contradiction: boolean;
  duplicate_mirror: boolean;
  stale: boolean;
}

export interface Score {
  total: number;
  repo_weight: number;
  impact: number;
  ease: number;
  type: number;
  age: number;
  due: number;
}

export interface Task {
  key: string;
  id: string;
  repository: string;
  owner: string;
  code: string;
  number: number | null;
  url: string | null;
  title: string;
  state: TaskState;
  state_reason: StateReason;
  type: string | null;
  labels: string[];
  assignees: string[];
  milestone: MilestoneRef | null;
  issue_fields: Record<string, unknown>;
  created_at: string | null;
  updated_at: string | null;
  closed_at: string | null;
  boards: Membership[];
  column: Column;
  flags: Flags;
  score: Score;
  version: string;
}

export interface RenderedComment {
  at: string;
  url: string;
  html: string;
}

export interface Provenance {
  path: string;
  root: string;
}

export interface TaskDetail extends Task {
  body_html: string;
  comments: RenderedComment[];
  provenance: Provenance;
}

export interface Facet {
  value: string;
  count: number;
}

export interface Facets {
  owners: Facet[];
  repositories: Facet[];
  milestones: Facet[];
  labels: Facet[];
  types: Facet[];
  assignees: Facet[];
  columns: Facet[];
  states: Facet[];
}

export interface TaskPage {
  total: number;
  limit: number;
  offset: number;
  facets: Facets;
  tasks: Task[];
}

export interface Repository {
  repository: string;
  owner: string;
  name: string;
  weight: number;
  path: string;
  tasks: number;
  /** The open tasks of the repository. */
  open: number;
  /** Not public on GitHub (private or internal) by the fetch state; false when unknown. */
  private: boolean;
}

export interface Milestone {
  key: string;
  repository: string;
  code: string;
  number: number;
  title: string;
  state: string;
  due_on: string | null;
  url: string | null;
}

export interface Board {
  key: string;
  title: string;
  url: string | null;
  repositories: string[];
  sections: string[];
}

export interface Catalog {
  repositories: Repository[];
  milestones: Milestone[];
  boards: Board[];
  labels: string[];
  columns: Column[];
}

export interface CheckoutRow {
  id?: number;
  repository: string;
  path: string;
  branch: string;
  default_branch?: string;
  commit?: string;
  upstream?: string;
  added: number;
  deleted: number;
  changed_files: number;
  untracked: number;
  binary: number;
  conflicts: number;
  ahead: number;
  behind: number;
  status: string;
  push_locked: boolean;
  pinned: boolean;
  /** public, private or internal, from the fetch state; absent when unknown. */
  visibility?: string;
  orphan_reason?: string;
  active_path?: string;
  remote_checked_at?: string;
  error?: string;
  cells: string[];
  display_path: string;
  placeholder: boolean;
}

export interface CheckoutSummary {
  checkouts: number;
  uncommitted: number;
  ahead: number;
  behind: number;
  locked: number;
  orphan: number;
  errors: number;
}

export interface SortSpec {
  column: string;
  desc: boolean;
}

export interface CheckoutsData {
  rows: CheckoutRow[];
  summary: CheckoutSummary;
  total: number;
  sort: SortSpec[];
  query: string;
  error?: string;
}

/** The columns of GET /api/v0/checkouts, in the order of the TUI. */
export const CHECKOUT_COLUMNS = ['repository', 'branch', 'uncommitted', 'status', 'lock', 'path'] as const;
export type CheckoutColumn = (typeof CHECKOUT_COLUMNS)[number];
export const CHECKOUT_COLUMN_TITLES: Record<CheckoutColumn, string> = {
  repository: 'Repository',
  branch: 'Branch',
  uncommitted: 'Uncommitted',
  status: 'Status',
  lock: 'Lock',
  path: 'Path',
};

export type TaskSort = 'score' | 'importance' | 'updated' | 'created';
export type StateFilter = 'open' | 'closed' | 'all';

/** The query parameters of GET /api/v0/tasks. */
export interface TaskQuery {
  owner?: string[];
  repo?: string[];
  state?: StateFilter;
  milestone?: string[];
  label?: string[];
  type?: string[];
  assignee?: string[];
  column?: Column[];
  q?: string;
  sort?: TaskSort;
  limit?: number;
  offset?: number;
}

/** The body of PATCH /api/v0/tasks/{id}/triage. */
export interface TriagePatch {
  impact?: Impact | null;
  effort?: Effort | null;
}
