import type { ComponentChildren } from 'preact';
import { Icon } from './Icon';

export type SkeletonShape = 'rows' | 'table' | 'cards' | 'prose';

/** Grey lines in the shape of the content that is on its way. */
export function Skeleton({ shape = 'prose', count = 6 }: { shape?: SkeletonShape; count?: number }) {
  const items = Array.from({ length: count }, (_, i) => i);
  if (shape === 'rows' || shape === 'table') {
    return (
      <div class={`skeleton-rows${shape === 'table' ? ' skeleton-rows--dense' : ''}`} aria-hidden="true">
        {items.map((i) => (
          <div class="skeleton-rows__row" key={i}>
            <span class="skeleton" />
            <span class="skeleton" />
            {shape === 'table' && <span class="skeleton" />}
          </div>
        ))}
      </div>
    );
  }
  if (shape === 'cards') {
    return (
      <div class="skeleton-cards" aria-hidden="true">
        {items.map((i) => (
          <div class="skeleton-card" key={i}>
            <span class="skeleton" />
            <span class="skeleton" />
            <span class="skeleton" />
          </div>
        ))}
      </div>
    );
  }
  return (
    <div class="skeleton-prose" aria-hidden="true">
      <span class="skeleton" />
      <span class="skeleton" />
      <span class="skeleton" />
      <span class="skeleton" />
    </div>
  );
}

/** The 2 px line under the top bar while the service indexes or collects. */
export function ProgressLine() {
  return <div class="progress-line" aria-hidden="true" />;
}

export interface LoadingProps {
  children?: ComponentChildren;
  /** The layout the content will have: the skeleton takes its shape. */
  shape?: SkeletonShape;
  count?: number;
}

/** Something is being read from the service: a caption and a skeleton. */
export function Loading({ children = 'Loading…', shape = 'prose', count }: LoadingProps) {
  return (
    <div class="loading" role="status" aria-live="polite">
      <p class="state-text">{children}</p>
      <Skeleton shape={shape} count={count} />
    </div>
  );
}

/** The index has no snapshot yet: the pages show it until the first one. */
export function Indexing({ what = 'the mirrors', shape = 'rows' }: { what?: string; shape?: SkeletonShape }) {
  return (
    <div role="status" aria-live="polite">
      <ProgressLine />
      <p class="state-text">Indexing {what}… The first snapshot is being built; the page updates by itself.</p>
      <Skeleton shape={shape} />
    </div>
  );
}

export interface ErrorStateProps {
  error: unknown;
  onRetry?: () => void;
}

/** A request failed and can be tried again. */
export function ErrorState({ error, onRetry }: ErrorStateProps) {
  const message = error instanceof Error ? error.message : String(error);
  return (
    <div class="notice notice--error" role="alert">
      <Icon name="alert-circle" size={14} class="notice__icon" />
      <p>
        <strong>Something went wrong.</strong> {message}
      </p>
      {onRetry && (
        <button type="button" class="button" onClick={onRetry}>
          Try again
        </button>
      )}
    </div>
  );
}

export interface EmptyProps {
  children?: ComponentChildren;
  /** An action under the sentence, such as Clear filters. */
  action?: ComponentChildren;
}

/** Nothing here, or nothing matches. */
export function Empty({ children = 'Nothing here.', action }: EmptyProps) {
  return (
    <div class="empty-state">
      <p>{children}</p>
      {action}
    </div>
  );
}
