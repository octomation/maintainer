import { useEffect, useState } from 'preact/hooks';

/** Below 900 px the sidebar is an overlay and the issue page opens over the list. */
export const NARROW = '(max-width: 899px)';
/** Below 600 px the controls are touch-sized and the top bar takes a second row. */
export const TOUCH = '(max-width: 599px)';

/** Whether a media query matches now; false where matchMedia is missing. */
export function matches(query: string): boolean {
  return typeof matchMedia === 'function' && matchMedia(query).matches;
}

/** A media query as state: it follows the viewport. */
export function useMediaQuery(query: string): boolean {
  const [on, setOn] = useState(() => matches(query));
  useEffect(() => {
    if (typeof matchMedia !== 'function') return;
    const media = matchMedia(query);
    const update = () => setOn(media.matches);
    update();
    media.addEventListener('change', update);
    return () => media.removeEventListener('change', update);
  }, [query]);
  return on;
}
