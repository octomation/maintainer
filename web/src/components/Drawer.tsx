import type { ComponentChildren } from 'preact';
import { useEffect, useRef } from 'preact/hooks';
import { Icon } from './Icon';

export interface DrawerProps {
  open: boolean;
  /** The accessible name of the dialog. */
  title: string;
  /** What the header shows; the title when absent. */
  heading?: ComponentChildren;
  /** Actions in the header before Close, such as Open on GitHub. */
  actions?: ComponentChildren;
  onClose: () => void;
  children: ComponentChildren;
  /** Where the focus goes on close when the element that opened the drawer is gone. */
  restoreTo?: () => HTMLElement | null;
  /** Wide for an issue page with its properties, narrow for a list of properties. */
  size?: 'wide' | 'narrow';
}

const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), summary, [tabindex]:not([tabindex="-1"])';

/**
 * The peek panel: a modal panel from the right on the menu surface, 8 px
 * from the viewport's edges, with a 48 px header. It takes the focus and
 * keeps it, makes everything else inert, locks the page scroll, closes on
 * Escape or the backdrop and gives the focus back.
 */
export function Drawer({ open, title, heading, actions, onClose, children, restoreTo, size = 'wide' }: DrawerProps) {
  const panel = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (!open) return;
    const previous = document.activeElement as HTMLElement | null;
    const element = panel.current;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.stopPropagation();
        onClose();
        return;
      }
      if (event.key === 'Tab' && element) {
        const focusable = Array.from(element.querySelectorAll<HTMLElement>(FOCUSABLE)).filter((node) => node.offsetParent !== null || node === document.activeElement);
        if (focusable.length === 0) {
          event.preventDefault();
          element.focus();
          return;
        }
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        const active = document.activeElement;
        if (event.shiftKey && (active === first || active === element)) {
          event.preventDefault();
          last.focus();
        } else if (!event.shiftKey && active === last) {
          event.preventDefault();
          first.focus();
        } else if (!element.contains(active)) {
          event.preventDefault();
          first.focus();
        }
      }
    };
    document.addEventListener('keydown', onKey);
    // Everything beside the drawer and its ancestors is inert while it is open.
    const made: Element[] = [];
    for (let node: Element | null = element; node && node !== document.body; node = node.parentElement) {
      for (const sibling of Array.from(node.parentElement?.children ?? [])) {
        if (sibling === node || sibling.hasAttribute('inert') || sibling.classList.contains('peek__backdrop')) continue;
        sibling.setAttribute('inert', '');
        made.push(sibling);
      }
    }
    const overflow = document.documentElement.style.overflow;
    document.documentElement.style.overflow = 'hidden';
    element?.focus();
    return () => {
      document.removeEventListener('keydown', onKey);
      for (const node of made) node.removeAttribute('inert');
      document.documentElement.style.overflow = overflow;
      // The opener may be gone, as a card is after a triage moved it: a fallback takes the focus.
      const target = previous?.isConnected ? previous : restoreTo?.();
      target?.focus?.();
    };
  }, [open, onClose, restoreTo]);
  if (!open) return null;
  return (
    <>
      <div class="peek__backdrop" onClick={onClose} data-testid="drawer-backdrop" />
      <div class={`peek peek--${size}`} role="dialog" aria-modal="true" aria-label={title} tabIndex={-1} ref={panel}>
        <div class="peek__header">
          <h2 class="peek__title">{heading ?? title}</h2>
          {actions}
          <button type="button" class="icon-button" onClick={onClose} aria-label="Close" title="Close (Escape)">
            <Icon name="x" />
          </button>
        </div>
        <div class="peek__body">{children}</div>
      </div>
    </>
  );
}
