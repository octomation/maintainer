import { fireEvent, render, screen, waitFor } from '@testing-library/preact';
import { describe, expect, it, vi } from 'vitest';
import { Drawer } from './Drawer';

function Page({ open, onClose, restoreTo, opener = true }: { open: boolean; onClose: () => void; restoreTo?: () => HTMLElement | null; opener?: boolean }) {
  return (
    <div>
      {opener && <button type="button">Behind</button>}
      <button type="button">Fallback</button>
      <Drawer open={open} title="A row" heading={<span>row</span>} actions={<a href="https://example.test">Open</a>} onClose={onClose} restoreTo={restoreTo}>
        <button type="button">First</button>
        <button type="button">Last</button>
        <details>
          <summary>More</summary>
          <p>Details</p>
        </details>
      </Drawer>
    </div>
  );
}

describe('Drawer', () => {
  it('takes the focus, keeps it inside, makes the rest inert, locks the scroll and gives the focus back', async () => {
    const onClose = vi.fn();
    const { rerender } = render(<Page open={false} onClose={onClose} />);
    const behind = screen.getByRole('button', { name: 'Behind' });
    behind.focus();
    rerender(<Page open onClose={onClose} />);
    const dialog = await screen.findByRole('dialog', { name: 'A row' });
    expect(dialog.getAttribute('aria-modal')).toBe('true');
    expect(dialog.textContent).toContain('row');
    expect(document.activeElement).toBe(dialog);
    expect(behind.hasAttribute('inert')).toBe(true);
    expect(document.documentElement.style.overflow).toBe('hidden');
    // Tab cycles inside: from the last control to the first, and back with Shift.
    // The disclosure's summary is the last stop of the cycle: Tab wraps from it, Shift-Tab wraps to it.
    const first = screen.getByRole('link', { name: 'Open' });
    const summary = screen.getByText('More');
    summary.focus();
    expect(document.activeElement).toBe(summary);
    fireEvent.keyDown(document, { key: 'Tab' });
    expect(document.activeElement).toBe(first);
    fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });
    expect(document.activeElement).toBe(summary);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
    rerender(<Page open={false} onClose={onClose} />);
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(behind.hasAttribute('inert')).toBe(false);
    expect(document.documentElement.style.overflow).toBe('');
    expect(document.activeElement).toBe(behind);
  });

  it('falls back to the given element when the opener is gone on close', async () => {
    const onClose = vi.fn();
    const fallback = () => screen.getByRole('button', { name: 'Fallback' });
    const { rerender } = render(<Page open={false} onClose={onClose} restoreTo={fallback} />);
    screen.getByRole('button', { name: 'Behind' }).focus();
    rerender(<Page open onClose={onClose} restoreTo={fallback} />);
    await screen.findByRole('dialog');
    rerender(<Page open={false} onClose={onClose} restoreTo={fallback} opener={false} />);
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Fallback' }));
  });

  it('closes from the backdrop and the Close button', () => {
    const onClose = vi.fn();
    render(<Page open onClose={onClose} />);
    fireEvent.click(screen.getByTestId('drawer-backdrop'));
    fireEvent.click(screen.getByRole('button', { name: 'Close' }));
    expect(onClose).toHaveBeenCalledTimes(2);
  });
});
