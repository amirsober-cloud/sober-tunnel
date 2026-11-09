/* "Enjoying sober-tunnel?" and "Support sober-tunnel".
 * Static screens; the addresses copy on click. */

import { brand } from '../brand.js';
import { openScreen } from '../ui/screen.js';
import { oops } from '../ui/toast.js';

/* Both dialogs are asks, so every button either opens a link or ends the ask.
   "Do not ask again" is remembered here the way the panel remembers it. */
function links(root, close) {
  const open = url => window.open(url, '_blank', 'noopener');
  root.querySelectorAll('button').forEach(b => {
    const t = b.textContent.trim().toLowerCase();
    if (t.includes('github')) b.addEventListener('click', () => open(brand.repository));
    else if (t.includes('telegram')) {
      b.hidden = !brand.telegram;
      if (brand.telegram) b.addEventListener('click', () => open(brand.telegram));
    }
    else if (t === 'later' || t === 'close') b.addEventListener('click', close);
    else if (t === 'do not ask again') b.addEventListener('click', () => {
      try { localStorage.setItem('bp_star', 'never'); } catch (e) {}
      close();
    });
  });
}

/* The version on the About line, read from the same place the rest of the
   panel reads it. */
function stampVersion(root) {
  const el = root.querySelector('#bpver');
  if (!el) return;
  try {
    const v = document.documentElement.dataset.version || '';
    if (v) el.textContent = v;
  } catch (e) { /* the notice stands without it */ }
}

export const starView = ctx => openScreen('support', {
  pick: '.dlg.star0',
  bind: (root, close) => { links(root, close); stampVersion(root); ctx.setTeardown(close); },
}).catch(oops);

export const supportView = ctx => openScreen('support', {
  pick: '.dlg.sup0',
  bind: (root, close) => { links(root, close); stampVersion(root); ctx.setTeardown(close); },
}).catch(oops);
