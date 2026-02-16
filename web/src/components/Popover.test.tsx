import { cleanup, fireEvent, render, screen } from '@testing-library/preact';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Popover } from './Popover';

/** Opens a 280 px panel under a trigger at left..right in a window of the given width. */
function open(width: number, left: number, right: number, align: 'start' | 'end' = 'start') {
  vi.stubGlobal('innerWidth', width);
  vi.stubGlobal('innerHeight', 844);
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({ left, right, top: 0, bottom: 24, width: right - left, height: 24, x: left, y: 0, toJSON: () => ({}) });
  vi.spyOn(HTMLElement.prototype, 'offsetWidth', 'get').mockReturnValue(280);
  render(
    <Popover label="Filters" align={align} trigger={(props) => <button type="button" {...props}>Filter</button>}>
      {() => <span>content</span>}
    </Popover>,
  );
  fireEvent.click(screen.getByRole('button', { name: 'Filter' }));
  const panel = screen.getByRole('menu', { name: 'Filters' });
  const result = { end: panel.classList.contains('popover--end'), shift: panel.style.getPropertyValue('--popover-shift') };
  cleanup();
  vi.restoreAllMocks();
  return result;
}

describe('Popover', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('opens under the trigger where it fits', () => {
    expect(open(1440, 308, 369)).toEqual({ end: false, shift: '0px' });
    expect(open(1440, 1000, 1100, 'end')).toEqual({ end: true, shift: '0px' });
  });

  it('flips to the trigger’s right edge when the window ends first', () => {
    expect(open(1440, 1300, 1340)).toEqual({ end: true, shift: '0px' });
  });

  it('moves inside a narrow window when neither edge of the trigger leaves room', () => {
    // At 390 px the trigger spans 160..198: aligned to its right edge the panel would start at -82.
    expect(open(390, 160, 198)).toEqual({ end: true, shift: '90px' });
    expect(open(390, 10, 60, 'end')).toEqual({ end: false, shift: '0px' });
  });
});
