// The shell styles come first: a page's own stylesheet then wins a tie.
import './styles/base.css';
import { render } from 'preact';
import { App } from './app';
import { applyTheme, readTheme } from './theme';

// Nothing of the hub may live under a plain address (maintainer.localhost,
// localhost, a loopback literal): another account can serve that origin while
// the hub is down. The service clears it with Clear-Site-Data; this is the
// second line for a browser that ignores the header. Earlier releases kept the
// expanded owners of the sidebar there. With access = "open" in service.toml
// the service marks the page and the plain address works as the hub.
const openAccess = document.documentElement.dataset.access === 'open';
const plainAddress = !openAccess && !/^[0-9a-f]{32}\.maintainer\.localhost$/.test(location.hostname);
try {
  if (plainAddress) {
    localStorage.clear();
    sessionStorage.clear();
  } else localStorage.removeItem('maintainer.sidebar.owners');
} catch {
  // storage is a convenience
}
// `maintainer service open` visits the plain address to clear it, then
// expects the page to go back to its redirect page.
if (plainAddress && new URLSearchParams(location.search).get('back') === '1' && history.length > 1) {
  history.back();
} else {
  // The stored theme before the first paint, so a light choice does not flash dark.
  applyTheme(readTheme());
  const root = document.getElementById('app');
  if (root) render(<App />, root);
}
