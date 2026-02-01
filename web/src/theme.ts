export type Theme = 'system' | 'light' | 'dark';

const KEY = 'maintainer.theme';
/** The order of the switch: dark is the default, then light, then the system's. */
const THEMES: Theme[] = ['dark', 'light', 'system'];

/** Reads the theme the user picked; dark when none or storage fails. */
export function readTheme(): Theme {
  try {
    const stored = localStorage.getItem(KEY);
    return THEMES.includes(stored as Theme) ? (stored as Theme) : 'dark';
  } catch {
    return 'dark';
  }
}

/** Applies a theme to <html> as data-theme; dark, the default, is not stored. */
export function applyTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme);
  try {
    if (theme === 'dark') localStorage.removeItem(KEY);
    else localStorage.setItem(KEY, theme);
  } catch {
    // storage is a convenience
  }
}

export function nextTheme(theme: Theme): Theme {
  return THEMES[(THEMES.indexOf(theme) + 1) % THEMES.length];
}
