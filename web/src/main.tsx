// The shell styles come first: a page's own stylesheet then wins a tie.
import './styles/base.css';
import { render } from 'preact';
import { App } from './app';
import { applyTheme, readTheme } from './theme';

// The stored theme before the first paint, so a light choice does not flash dark.
applyTheme(readTheme());
const root = document.getElementById('app');
if (root) render(<App />, root);
