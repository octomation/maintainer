import { useEffect, useRef } from 'preact/hooks';
import type { ApiClient } from '../api/client';
import type { TaskDetail } from '../api/types';
import { TaskDetails } from './TaskDetails';

export interface IssuePageProps {
  client: ApiClient;
  id: string;
  onChanged?: (detail: TaskDetail) => void;
  /** Back, Escape: the list or the board again. */
  onBack: () => void;
  /** The accessible name of the way back, such as "Back to the board". */
  backLabel?: string;
}

/**
 * The full issue page: the issue fills the content area beside the sidebar,
 * with no overlay, the article on the left and the properties rail on the
 * right. It takes the focus when it opens; Escape goes back unless a menu
 * inside it is open, which closes first.
 */
export function IssuePage({ client, id, onChanged, onBack, backLabel }: IssuePageProps) {
  const root = useRef<HTMLElement>(null);
  useEffect(() => {
    root.current?.focus();
  }, [id]);
  return (
    <section
      class="issue-page"
      aria-label="Issue"
      tabIndex={-1}
      ref={root}
      onKeyDown={(event) => {
        if (event.key !== 'Escape') return;
        event.preventDefault();
        onBack();
      }}
    >
      <TaskDetails key={id} client={client} id={id} onChanged={onChanged} onBack={onBack} backLabel={backLabel} variant="full" />
    </section>
  );
}
