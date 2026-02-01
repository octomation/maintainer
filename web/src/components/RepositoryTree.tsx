import { useEffect, useMemo, useRef, useState } from 'preact/hooks';
import type { Repository } from '../api/types';
import { LockGlyph } from './Glyphs';
import { Icon } from './Icon';

/** One repository of the tree: duplicate mirrors of a repository are one row. */
export interface TreeRepository {
  repository: string;
  owner: string;
  name: string;
  open: number;
  private: boolean;
}

export interface TreeOwner {
  owner: string;
  repositories: TreeRepository[];
}

/** Every repository of the catalog grouped by owner, both in alphabetical order. */
export function repositoryTree(repositories: Repository[]): TreeOwner[] {
  const unique = new Map<string, TreeRepository>();
  for (const repo of repositories) {
    const known = unique.get(repo.repository);
    if (known) {
      known.private ||= repo.private;
      continue;
    }
    unique.set(repo.repository, { repository: repo.repository, owner: repo.owner, name: repo.name, open: repo.open, private: repo.private });
  }
  const owners = new Map<string, TreeRepository[]>();
  for (const repo of unique.values()) owners.set(repo.owner, [...(owners.get(repo.owner) ?? []), repo]);
  const byName = (a: string, b: string) => a.localeCompare(b, undefined, { sensitivity: 'base' });
  return Array.from(owners, ([owner, list]) => ({ owner, repositories: list.sort((a, b) => byName(a.name, b.name)) })).sort((a, b) => byName(a.owner, b.owner));
}

const STORAGE = 'maintainer.sidebar.owners';

/** The expanded owners the user chose, or null when none is stored. */
function readExpanded(): string[] | null {
  try {
    const stored = JSON.parse(localStorage.getItem(STORAGE) ?? 'null') as unknown;
    return Array.isArray(stored) ? stored.filter((v): v is string => typeof v === 'string') : null;
  } catch {
    return null;
  }
}

function writeExpanded(owners: string[]): void {
  try {
    localStorage.setItem(STORAGE, JSON.stringify(owners));
  } catch {
    // storage is a convenience
  }
}

export interface RepositoryTreeProps {
  repositories: Repository[];
  /** The repository the inbox is filtered by alone. */
  current: string | null;
}

type Item = { kind: 'owner'; key: string; owner: TreeOwner } | { kind: 'repo'; key: string; repo: TreeRepository };

/**
 * The Repositories tree of the sidebar: owner rows that expand to their
 * repositories, each repository a link to the inbox filtered by it, with its
 * open issues and a lock when it is private. The expanded owners are
 * remembered; by default only the owner of the current repository is open.
 * Keyboard: Up and Down move, Right expands or enters, Left collapses or
 * goes to the owner, Enter opens or toggles, Home and End jump.
 */
export function RepositoryTree({ repositories, current }: RepositoryTreeProps) {
  const tree = useMemo(() => repositoryTree(repositories), [repositories]);
  const currentOwner = current ? current.slice(0, current.indexOf('/')) : null;
  const [expanded, setExpanded] = useState<string[]>(() => readExpanded() ?? (currentOwner ? [currentOwner] : []));
  const [focus, setFocus] = useState<string | null>(null);
  const root = useRef<HTMLUListElement>(null);
  // The owner of a newly current repository opens, so the highlight is visible.
  useEffect(() => {
    if (currentOwner && !expanded.includes(currentOwner)) setExpanded((list) => [...list, currentOwner]);
  }, [currentOwner]);

  const toggle = (owner: string) => {
    const next = expanded.includes(owner) ? expanded.filter((o) => o !== owner) : [...expanded, owner];
    setExpanded(next);
    writeExpanded(next);
  };

  const items: Item[] = [];
  for (const owner of tree) {
    items.push({ kind: 'owner', key: `owner:${owner.owner}`, owner });
    if (expanded.includes(owner.owner)) for (const repo of owner.repositories) items.push({ kind: 'repo', key: `repo:${repo.repository}`, repo });
  }
  // The roving tab stop: the focused row, else the current repository, else the first row.
  const stop = items.find((i) => i.key === focus)?.key ?? items.find((i) => i.kind === 'repo' && i.repo.repository === current)?.key ?? items[0]?.key;

  const move = (key: string) => {
    setFocus(key);
    Array.from(root.current?.querySelectorAll<HTMLElement>('[data-key]') ?? [])
      .find((node) => node.dataset.key === key)
      ?.focus();
  };

  const onKey = (event: KeyboardEvent) => {
    const index = items.findIndex((i) => i.key === (event.target as HTMLElement).dataset.key);
    if (index < 0) return;
    const item = items[index];
    switch (event.key) {
      case 'ArrowDown':
        if (items[index + 1]) move(items[index + 1].key);
        break;
      case 'ArrowUp':
        if (items[index - 1]) move(items[index - 1].key);
        break;
      case 'Home':
        move(items[0].key);
        break;
      case 'End':
        move(items[items.length - 1].key);
        break;
      case 'ArrowRight':
        if (item.kind !== 'owner') return;
        if (!expanded.includes(item.owner.owner)) toggle(item.owner.owner);
        else if (items[index + 1]?.kind === 'repo') move(items[index + 1].key);
        break;
      case 'ArrowLeft':
        if (item.kind === 'owner') {
          if (expanded.includes(item.owner.owner)) toggle(item.owner.owner);
        } else move(`owner:${item.repo.owner}`);
        break;
      case 'Enter':
        if (item.kind === 'owner') toggle(item.owner.owner);
        else return; // the link follows itself
        break;
      default:
        return;
    }
    event.preventDefault();
  };

  if (tree.length === 0) return null;
  return (
    <ul class="repo-tree" role="tree" aria-label="Repositories" ref={root} onKeyDown={onKey}>
      {tree.map((owner) => {
        const open = expanded.includes(owner.owner);
        const key = `owner:${owner.owner}`;
        return (
          <li key={owner.owner} role="none">
            <button
              type="button"
              class="sidebar-item repo-tree__owner"
              role="treeitem"
              aria-expanded={open}
              aria-level={1}
              data-key={key}
              tabIndex={stop === key ? 0 : -1}
              onClick={() => toggle(owner.owner)}
              onFocus={() => setFocus(key)}
            >
              <Icon name={open ? 'chevron-down' : 'chevron-right'} size={12} class="repo-tree__chevron" />
              <span class="repo-tree__name">{owner.owner}</span>
              <span class="repo-tree__count" aria-label={`${owner.repositories.length} repositories`}>
                {owner.repositories.length}
              </span>
            </button>
            {open && (
              <ul class="repo-tree__group" role="group">
                {owner.repositories.map((repo) => {
                  const repoKey = `repo:${repo.repository}`;
                  const selected = repo.repository === current;
                  return (
                    <li key={repo.repository} role="none">
                      <a
                        class="sidebar-item sidebar-item--nested repo-tree__repo"
                        role="treeitem"
                        aria-level={2}
                        aria-selected={selected}
                        aria-current={selected ? 'page' : undefined}
                        href={`/inbox?repo=${encodeURIComponent(repo.repository)}`}
                        data-key={repoKey}
                        tabIndex={stop === repoKey ? 0 : -1}
                        onFocus={() => setFocus(repoKey)}
                        title={`${repo.repository}${repo.private ? ', private' : ''}: ${repo.open} open issues`}
                      >
                        <span class="repo-tree__name">{repo.name}</span>
                        {repo.private && <LockGlyph title="Private" class="repo-tree__lock" />}
                        <span class="repo-tree__count" aria-label={`${repo.open} open issues`}>
                          {repo.open}
                        </span>
                      </a>
                    </li>
                  );
                })}
              </ul>
            )}
          </li>
        );
      })}
    </ul>
  );
}
