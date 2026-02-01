import type { ComponentChildren } from 'preact';
import { Icon } from './Icon';
import { Popover } from './Popover';

/**
 * The Display button of the top bar and its panel: grouping, ordering and
 * the like, as labelled rows with the control on the right.
 */
export function DisplayMenu({ children }: { children: (close: (focusTrigger?: boolean) => void) => ComponentChildren }) {
  return (
    <Popover
      role="group"
      label="Display options"
      align="end"
      focusPanel
      trigger={(props) => (
        <button type="button" {...props} class="display-button" title="Grouping and ordering">
          <Icon name="adjustments" size={14} />
          <span class="display-button__text">View</span>
          <Icon name="chevron-down" size={12} class="display-button__chevron" />
        </button>
      )}
    >
      {(close) => <div class="display-menu">{children(close)}</div>}
    </Popover>
  );
}

/** A row of the Display panel: the name on the left, the control on the right. */
export function DisplayRow({ label, children, id }: { label: string; children: ComponentChildren; id?: string }) {
  return (
    <div class="display-row">
      <span class="display-row__label" id={id}>
        {label}
      </span>
      <span class="display-row__control">{children}</span>
    </div>
  );
}

export interface DisplaySelectProps<T extends string> {
  label: string;
  value: T;
  options: { value: T; label: string }[];
  onChange: (value: T) => void;
}

/** A 24 px select on the right of a Display row. */
export function DisplaySelect<T extends string>({ label, value, options, onChange }: DisplaySelectProps<T>) {
  return (
    <DisplayRow label={label}>
      <span class="mini-select">
        <select aria-label={label} value={value} onChange={(event) => onChange((event.target as HTMLSelectElement).value as T)}>
          {options.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        <Icon name="chevron-down" size={12} class="mini-select__icon" />
      </span>
    </DisplayRow>
  );
}
