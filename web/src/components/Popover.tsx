import type { ComponentChildren, JSX } from 'preact';
import { useEffect, useId, useLayoutEffect, useRef, useState } from 'preact/hooks';

export interface TriggerProps {
  ref: (element: HTMLButtonElement | null) => void;
  'aria-expanded': boolean;
  'aria-controls': string;
  'aria-haspopup': 'menu' | 'dialog' | 'true';
  onClick: () => void;
}

export interface PopoverProps {
  /** Renders the button that opens the panel; spread the props on it. */
  trigger: (props: TriggerProps) => JSX.Element;
  /** The panel's content; `close(true)` closes and returns the focus to the trigger. */
  children: (close: (focusTrigger?: boolean) => void) => ComponentChildren;
  /** The role of the panel: a menu of items or a group of controls. */
  role?: 'menu' | 'group' | 'dialog';
  /** The accessible name of the panel. */
  label: string;
  /** Align the panel's right edge with the trigger's; it also flips when it would leave the viewport. */
  align?: 'start' | 'end';
  class?: string;
  /** Opens the panel from outside, such as a chip that opens its dimension. */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  disabled?: boolean;
  /** Focus the panel itself on open rather than its first control, as a panel of settings does. */
  focusPanel?: boolean;
}

/** Arrow keys move between the items of a menu; Home and End jump. */
function moveFocus(panel: HTMLElement, event: KeyboardEvent) {
  const items = Array.from(panel.querySelectorAll<HTMLElement>('[role^="menuitem"]:not([disabled]), [data-menu-item]:not([disabled])'));
  if (items.length === 0) return;
  const index = items.indexOf(document.activeElement as HTMLElement);
  let next = -1;
  if (event.key === 'ArrowDown') next = index < 0 ? 0 : (index + 1) % items.length;
  else if (event.key === 'ArrowUp') next = index <= 0 ? items.length - 1 : index - 1;
  else if (event.key === 'Home') next = 0;
  else if (event.key === 'End') next = items.length - 1;
  if (next >= 0) {
    event.preventDefault();
    items[next].focus();
  }
}

/**
 * A floating panel under its trigger: a menu or a group of controls on the
 * menu surface. It takes the focus when it opens, closes on Escape (the
 * focus returns to the trigger) and on a click outside; the Escape is
 * handled in the capture phase, so a drawer or a pane around it stays open.
 */
export function Popover({ trigger, children, role = 'menu', label, align = 'start', class: cls, open: controlled, onOpenChange, disabled, focusPanel = false }: PopoverProps) {
  const [inner, setInner] = useState(false);
  const open = controlled ?? inner;
  const setOpen = (value: boolean) => {
    setInner(value);
    onOpenChange?.(value);
  };
  const id = useId();
  const anchor = useRef<HTMLDivElement>(null);
  const button = useRef<HTMLButtonElement | null>(null);
  const panel = useRef<HTMLDivElement>(null);
  const [end, setEnd] = useState(align === 'end');

  useLayoutEffect(() => {
    if (!open) return;
    const rect = anchor.current?.getBoundingClientRect();
    const width = panel.current?.offsetWidth ?? 0;
    if (rect && typeof innerHeight === 'number') {
      panel.current?.style.setProperty('--popover-room', `${Math.max(160, innerHeight - rect.bottom - 12)}px`);
      if (align === 'start') setEnd(width > 0 && rect.left + width > innerWidth - 8);
      else setEnd(!(width > 0 && rect.right - width < 8));
    }
    if (focusPanel) {
      panel.current?.focus();
      return;
    }
    // A menu opens on its checked item, as a property menu shows the current value.
    const first = panel.current?.querySelector<HTMLElement>('[data-autofocus], [role^="menuitem"][aria-checked="true"]:not([disabled])') ?? panel.current?.querySelector<HTMLElement>('[role^="menuitem"]:not([disabled]), input, select, button:not([disabled])');
    first?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const close = (back: boolean) => {
      setOpen(false);
      if (back) button.current?.focus();
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.stopPropagation();
        event.preventDefault();
        close(true);
        return;
      }
      if (role === 'menu' && panel.current?.contains(document.activeElement)) moveFocus(panel.current, event);
    };
    const onPointer = (event: MouseEvent) => {
      if (anchor.current && !anchor.current.contains(event.target as Node)) close(false);
    };
    document.addEventListener('keydown', onKey, true);
    document.addEventListener('mousedown', onPointer);
    return () => {
      document.removeEventListener('keydown', onKey, true);
      document.removeEventListener('mousedown', onPointer);
    };
  }, [open]);

  const close = (focusTrigger = false) => {
    setOpen(false);
    if (focusTrigger) button.current?.focus();
  };

  return (
    <div class={`popover-anchor${cls ? ` ${cls}` : ''}`} ref={anchor}>
      {trigger({
        ref: (element) => {
          button.current = element;
        },
        'aria-expanded': open,
        'aria-controls': id,
        'aria-haspopup': role === 'menu' ? 'menu' : role === 'dialog' ? 'dialog' : 'true',
        onClick: () => {
          if (!disabled) setOpen(!open);
        },
      })}
      {open && (
        <div class={`popover${end ? ' popover--end' : ''}`} id={id} ref={panel} role={role} aria-label={label} tabIndex={focusPanel ? -1 : undefined}>
          {children(close)}
        </div>
      )}
    </div>
  );
}
