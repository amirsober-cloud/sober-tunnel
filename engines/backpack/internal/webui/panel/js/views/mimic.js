import { $, esc } from '../lib/dom.js';
import * as api from '../api.js';

// This editor retains the complete native configuration rather than dropping
// advanced fields when a form is saved. Credentials stay in this admin view.
export function mimicView(ctx) {
  let closed = false, current = '', node = '';
  ctx.setTeardown(() => { closed = true; });
  const host = $('#view');
  host.innerHTML = `<section class="mimic-page">
    <header><h1>Camouflage tunnels</h1><p>Dynamic pools, protocol camouflage, SOCKS5, UDP, TUN and gateway routing.</p></header>
    <label>Server<select id="mimic-node"><option value="">This server</option></select></label>
    <div class="mimic-layout"><section><h2>Instances</h2><div id="mimic-list" aria-live="polite">Loading…</div><button id="mimic-refresh">Refresh</button><a href="#/servers">Manage servers</a></section>
    <section><h2 id="mimic-heading">New instance</h2>
      <label>Name<input id="mimic-name" maxlength="48" placeholder="foreign-main" autocomplete="off"></label>
      <label>Template<select id="mimic-template"><option value="foreign">Foreign exit</option><option value="iran">Iran hub</option></select></label>
      <button id="mimic-new">New from template</button>
      <label>Configuration (JSON)<textarea id="mimic-config" spellcheck="false" rows="24" aria-describedby="mimic-help"></textarea></label>
      <p id="mimic-help">Use the same auth_token at both ends. Stop an instance before editing. Saving validates the configuration without opening ports or changing SSH/firewall rules. TUN and gateway settings apply when you start the instance.</p>
      <div class="mimic-actions"><button id="mimic-save">Validate and save</button><button data-mimic-action="start">Start</button><button data-mimic-action="stop">Stop</button><button data-mimic-action="restart">Restart</button><button id="mimic-logs">Logs</button><button data-mimic-action="delete">Delete</button></div>
      <p id="mimic-result" role="status"></p><pre id="mimic-log"></pre>
    </section></div></section>`;
  const message = text => { if (!closed) $('#mimic-result').textContent = text; };
  let busy = false;
  async function task(fn) {
    if (busy) return; busy = true;
    host.querySelectorAll('button, #mimic-node').forEach(b => { b.disabled = true; });
    try { await fn(); } catch (e) { message(e.message); }
    finally { busy = false; if (!closed) host.querySelectorAll('button, #mimic-node').forEach(b => { b.disabled = false; }); }
  }
  function fresh() {
    current = ''; $('#mimic-name').value = ''; $('#mimic-name').readOnly = false;
    $('#mimic-heading').textContent = 'New instance'; $('#mimic-log').textContent = '';
    const raw = new Uint8Array(32); crypto.getRandomValues(raw);
    const token = Array.from(raw, v => v.toString(16).padStart(2, '0')).join('');
    const cfg = $('#mimic-template').value === 'foreign' ? {
      role: 'foreign', auth_token: token, foreign_listen_port: 8443,
      mimics: [{ type: 'tls', port: 8443 }], http_decoy_port: -1,
      decoy_port: 2022, egress_ip_mode: 'ipv4', decoy_style: 'apache',
    } : {
      role: 'iran', foreign_nodes: [{ alias: 'foreign-main', target_ip: 'YOUR_FOREIGN_IP', target_port: 8443,
        endpoints: [{ target: 'YOUR_FOREIGN_IP:8443', mimic: 'tls' }],
        local_socks_port: 1081, socks_bind: '127.0.0.1', auth_token: 'COPY_TOKEN_FROM_FOREIGN',
        min_connections: 3, max_connections: 10, bandwidth_limit_mbps: 8, bandwidth_jitter_mbps: 2 }],
    };
    $('#mimic-config').value = JSON.stringify(cfg, null, 2); message('');
  }
  async function load() {
    const items = await api.mimicInstances(node); if (closed) return;
    $('#mimic-list').innerHTML = items.length ? items.map(i => `<button class="mimic-instance" data-name="${esc(i.name)}"><strong>${esc(i.name)}</strong><span>${esc(i.role)} · ${i.active ? 'Running' : 'Stopped'}</span></button>`).join('') : '<p>No camouflage instances yet.</p>';
    $('#mimic-list').querySelectorAll('[data-name]').forEach(b => b.addEventListener('click', () => task(async () => {
      const i = await api.mimicSettings(b.dataset.name, node); if (closed) return;
      current = i.name; $('#mimic-name').value = current; $('#mimic-name').readOnly = true;
      $('#mimic-heading').textContent = current; $('#mimic-config').value = JSON.stringify(i.config, null, 2);
      $('#mimic-log').textContent = ''; message('');
    })));
  }
  $('#mimic-new').addEventListener('click', fresh);
  $('#mimic-refresh').addEventListener('click', () => task(load));
  $('#mimic-save').addEventListener('click', () => task(async () => {
    const name = $('#mimic-name').value.trim(), config = JSON.parse($('#mimic-config').value);
    await api.mimicAction({ name, action: 'save', config }, node); if (closed) return;
    current = name; $('#mimic-name').readOnly = true; message('Configuration validated and saved. Start the instance when ready.'); await load();
  }));
  host.querySelectorAll('[data-mimic-action]').forEach(b => b.addEventListener('click', () => task(async () => {
    if (!current) throw new Error('Save or select an instance first.');
    const action = b.dataset.mimicAction;
    if (action === 'delete' && !window.confirm(`Delete ${current} and its configuration?`)) return;
    await api.mimicAction({ name: current, action }, node); if (closed) return;
    if (action === 'delete') fresh();
    message(action === 'delete' ? 'Instance deleted.' : `Service ${action} completed. Refresh to check its current state.`); await load();
  })));
  $('#mimic-logs').addEventListener('click', () => task(async () => {
    if (!current) throw new Error('Select an instance first.');
    const result = await api.mimicLogs(current, node); if (!closed) $('#mimic-log').textContent = result.logs || 'No log entries.';
  }));
  $('#mimic-node').addEventListener('change', () => { node = $('#mimic-node').value; fresh(); task(load); });
  fresh(); task(async () => {
    let result;
    try { result = await api.nodesCached(); } catch { result = { nodes: [] }; }
    if (closed) return;
    for (const n of result.nodes || []) {
      const option = document.createElement('option'); option.value = n.name; option.textContent = n.name;
      $('#mimic-node').append(option);
    }
    await load();
  });
}
