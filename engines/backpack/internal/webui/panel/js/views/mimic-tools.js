import { $, esc, copyText } from '../lib/dom.js';
import * as api from '../api.js';
import { go } from '../router.js';

const labels = {
  'setup-foreign': 'Create foreign exit', 'setup-iran': 'Create Iran hub',
  'add-node': 'Add foreign node to this hub', 'edit-node': 'Edit foreign node',
  'remove-node': 'Remove foreign node', 'edit-foreign': 'Edit foreign exit / rotate token',
  pairing: 'Export pairing token', probe: 'Probe all protocol endpoints',
  test: 'Full connection test: exit IP, sites, DNS and TUN',
  speedtest: 'Download / upload speed test', 'check-ip': 'Check this server’s IP reputation',
};
const common = new Set(['alias', 'persona', 'mimics', 'domain', 'acme-email', 'public-ip', 'token', 'target-ip', 'target', 'socks-port', 'profile', 'node', 'mimic', 'seconds', 'dir']);
const choices = {
  persona: ['auto','cpanel','directadmin','devops'], profile: ['balanced','high-speed'],
  'egress-mode': ['ipv4','ipv6','dual'], 'decoy-style': ['apache','directadmin'], dir: ['both','down','up'],
};

export function mimicView(ctx) {
  let closed=false, current='', node='', catalog={}, poll=null, busy=false, job=null;
  const host=$('#view');
  ctx.setTeardown(()=>{closed=true; clearTimeout(poll);});
  host.innerHTML=`<section class="mimic-page">
    <header><a href="#/tunnels">← Tunnels</a><h1>Camouflage / Hedioum</h1><p>Create both ends, pair them, then connect x-ui to the hub’s local SOCKS5 bridge.</p></header>
    <label>Server<select id="mimic-node"><option value="">This server</option></select></label>
    <div class="mimic-layout"><section><h2>Camouflage tunnels</h2><div id="mimic-list">Loading…</div><button id="mimic-refresh">Refresh</button><button id="mimic-new">New tunnel</button><a href="#/servers">Manage servers</a></section>
    <section><h2 id="mimic-heading">New tunnel</h2><label>Instance name<input id="mimic-name" maxlength="48" placeholder="foreign-main" autocomplete="off"></label>
      <label>Setup / edit operation<select id="mimic-operation"></select></label>
      <form id="mimic-form"><div id="mimic-fields" class="mimic-fields"></div><details><summary>Advanced protocol, pool, SOCKS, TUN and gateway settings</summary><div id="mimic-advanced" class="mimic-fields"></div></details><button type="submit">Build validated preview</button></form>
      <p class="mimic-help">New exits use the native Persona defaults. Set individual protocol ports under Advanced if another service owns them. A pairing token fills the hub’s foreign IP, protocol ports and persona. SOCKS binds to 127.0.0.1 by default.</p>
      <details id="mimic-preview"><summary>Configuration preview / advanced JSON editor</summary><textarea id="mimic-config" spellcheck="false" rows="18"></textarea></details>
      <div class="mimic-actions"><button id="mimic-save">Validate and save</button><button id="mimic-preflight">Preflight checks</button><button data-mimic-action="start">Start</button><button data-mimic-action="stop">Stop</button><button data-mimic-action="restart">Restart</button><button id="mimic-logs">Logs</button><button data-mimic-action="delete">Delete</button></div>
      <p id="mimic-result" role="status" aria-live="polite"></p><pre id="mimic-output" tabindex="0"></pre><button id="mimic-copy">Copy result</button>
      <section><h3>Diagnostics</h3><p>Tests run on the selected server and use the saved configuration. A completed run may contain failed checks; read each result.</p><label>Diagnostic<select id="mimic-diagnostic"></select></label><div id="mimic-test-fields" class="mimic-fields"></div><button id="mimic-run">Run diagnostic</button><p id="mimic-job" role="status"></p><pre id="mimic-test-output" tabindex="0"></pre></section>
      <section><h3>Connect x-ui / 3x-ui</h3><p>On the Iran hub, add a SOCKS outbound to 127.0.0.1 and this node’s SOCKS port. Route the chosen inbound tags to that outbound, keeping domestic/direct rules ahead of it. Use remote DNS / AsIs. Both processes must share the same network namespace; Docker bridge containers have their own 127.0.0.1.</p><button id="mimic-xui">Generate x-ui settings</button><label>Xray settings format<select id="mimic-xui-format"><option value="legacy_outbound">3x-ui / servers[]</option><option value="modern_outbound">Current Xray / address + port</option></select></label><pre id="mimic-xui-output" tabindex="0"></pre><button id="mimic-xui-copy">Copy x-ui settings</button></section>
      <details><summary>Host operations</summary><p>These apply to the selected server. Review ports before changing OpenSSH. Each action requires its own confirmation.</p><button data-host="move-ssh">Move OpenSSH to saved decoy port</button><button data-host="open-firewall">Open saved exit listen ports</button><button data-host="bbr">Enable BBR</button></details>
    </section></div></section>`;
  const get=id=>$('#'+id,host);
  const message=text=>{if(!closed)get('mimic-result').textContent=text;};
  const output=text=>{if(!closed)get('mimic-output').textContent=text;};
  async function task(fn) {
    if(busy||closed)return; busy=true;
    host.querySelectorAll('button,select,input,textarea').forEach(el=>el.disabled=true);
    try {await fn();} catch(e){message(e.message);}
    finally {busy=false;if(!closed)host.querySelectorAll('button,select,input,textarea').forEach(el=>el.disabled=false);}
  }
  const name=()=>get('mimic-name').value.trim();
  const saved=()=>{if(!current)throw new Error('Save or select a tunnel first.');return current;};
  function field(f) {
    let input; const id='flag-'+f.name;
    if(f.type==='bool')input=`<select data-flag="${esc(f.name)}" id="${id}"><option value="">Keep native default / current value</option><option value="true">Enabled</option><option value="false">Disabled</option></select>`;
    else if(choices[f.name])input=`<select data-flag="${esc(f.name)}" id="${id}"><option value="">Native default / current value</option>${choices[f.name].map(v=>`<option>${esc(v)}</option>`).join('')}</select>`;
    else input=`<input data-flag="${esc(f.name)}" id="${id}" type="${f.type==='int'?'number':'text'}" autocomplete="off" placeholder="${esc(f.default||'Native default / current value')}">`;
    return `<label>${esc(f.name.replaceAll('-',' '))}${input}<small>${esc(f.help)}</small></label>`;
  }
  function form(command,target,advanced) {
    const flags=catalog[command]||[];
    get(target).innerHTML=flags.filter(f=>!advanced||common.has(f.name)).map(field).join('');
    if(advanced)get(advanced).innerHTML=flags.filter(f=>!common.has(f.name)).map(field).join('');
  }
  function values(container) {
    const flags={};get(container).querySelectorAll('[data-flag]').forEach(el=>{if(el.value!=='')flags[el.dataset.flag]=el.value;});return flags;
  }
  function operation(command) {get('mimic-operation').value=command;form(command,'mimic-fields','mimic-advanced');}
  function fresh() {
    clearTimeout(poll);job=null;current='';get('mimic-name').value='';get('mimic-name').readOnly=false;
    get('mimic-heading').textContent='New tunnel';get('mimic-config').value='';get('mimic-job').textContent='';get('mimic-test-output').textContent='';get('mimic-xui-output').textContent='';xui=[];xuiNode.replaceChildren();output('');message('');operation('setup-foreign');
  }
  async function select(instance) {
    const result=await api.mimicSettings(instance,node);if(closed)return;
    clearTimeout(poll);job=null;current=result.name;get('mimic-name').value=current;get('mimic-name').readOnly=true;
    get('mimic-heading').textContent=current;get('mimic-config').value=JSON.stringify(result.config,null,2);
    operation(result.config.role==='foreign'?'edit-foreign':'edit-node');
    if(result.config.role==='iran') {
      const alias=host.querySelector('#mimic-fields [data-flag="alias"]');if(alias)alias.value=result.config.foreign_nodes?.[0]?.alias||'';
    }
    message('Loaded saved configuration. Build a preview to change it; stop the tunnel before saving.');output('');
    get('mimic-test-output').textContent='';get('mimic-job').textContent='';get('mimic-xui-output').textContent='';
    xui=[];xuiNode.replaceChildren();
    get('mimic-diagnostic').value=result.config.role==='foreign'?'check-ip':'probe';form(get('mimic-diagnostic').value,'mimic-test-fields');
    try {const previous=JSON.parse(sessionStorage.getItem('sober-camouflage-job'));if(previous?.name===current&&previous.node===node){job=previous;pollJob();}}catch{}
  }
  async function load() {
    const items=await api.mimicInstances(node);if(closed)return;
    get('mimic-list').innerHTML=items.length?items.map(i=>`<button class="mimic-instance" data-name="${esc(i.name)}"><strong>${esc(i.name)}</strong><span>${esc(i.role)} · ${i.active?'Running':'Stopped'}</span></button>`).join(''):'<p>No camouflage tunnels yet.</p>';
    get('mimic-list').querySelectorAll('[data-name]').forEach(el=>el.addEventListener('click',()=>task(()=>select(el.dataset.name))));
  }
  get('mimic-form').addEventListener('submit',event=>{event.preventDefault();task(async()=>{
    const command=get('mimic-operation').value;
    const req={name:name(),action:'prepare',command,flags:values('mimic-form')};
    if(!command.startsWith('setup-')&&get('mimic-config').value.trim())req.config=JSON.parse(get('mimic-config').value);
    const result=await api.mimicAction(req,node);if(closed)return;
    get('mimic-config').value=JSON.stringify(result.config,null,2);get('mimic-preview').open=true;output(result.output||'');
    message(command==='pairing'?'Pairing token exported. Treat it as a password.':'Preview validated. Review it, then save. No running configuration changed.');
  });});
  get('mimic-operation').addEventListener('change',()=>operation(get('mimic-operation').value));
  get('mimic-diagnostic').addEventListener('change',()=>form(get('mimic-diagnostic').value,'mimic-test-fields'));
  get('mimic-new').addEventListener('click',fresh);
  get('mimic-refresh').addEventListener('click',()=>task(load));
  get('mimic-save').addEventListener('click',()=>task(async()=>{
    const config=JSON.parse(get('mimic-config').value);await api.mimicAction({name:name(),action:'save',config},node);if(closed)return;
    current=name();get('mimic-name').readOnly=true;message('Saved. Run preflight checks before starting.');await load();
  }));
  async function preflight() {
    const checks=await api.mimicAction({name:saved(),action:'preflight'},node);
    output(checks.map(c=>`${c.ok?'✓':'✗'} ${c.name}: ${c.detail}`).join('\n'));return checks;
  }
  get('mimic-preflight').addEventListener('click',()=>task(async()=>{const c=await preflight();message(c.every(x=>x.ok)?'Preflight passed at this moment.':'Review the failed checks before starting.');}));
  host.querySelectorAll('[data-mimic-action]').forEach(el=>el.addEventListener('click',()=>task(async()=>{
    const action=el.dataset.mimicAction, instance=saved();
    if(action==='delete'&&!window.confirm(`Delete ${instance} and its configuration?`))return;
    if(action==='start'&&!(await preflight()).every(c=>c.ok))throw new Error('Preflight found a conflict. Review the results.');
    await api.mimicAction({name:instance,action},node);if(closed)return;
    if(action==='delete')fresh();message(`Service ${action} completed.`);await load();
  })));
  get('mimic-logs').addEventListener('click',()=>task(async()=>{const r=await api.mimicLogs(saved(),node);output(r.logs||'No log entries.');}));
  get('mimic-copy').addEventListener('click',()=>copyText(get('mimic-output').textContent));
  async function pollJob() {
    if(closed||!job)return;
    try {
      const result=await api.mimicAction({action:'job',name:job.name,job_id:job.id},job.node);if(closed||!job||result.id!==job.id)return;
      get('mimic-test-output').textContent=result.output||'Waiting for output…';
      get('mimic-job').textContent=result.done?(result.error?'Finished with error: '+result.error:'Finished. Review individual checks above.'):'Running '+labels[result.command]+'…';
      if(!result.done)poll=setTimeout(pollJob,1500);
    }catch(e){if(!closed)get('mimic-job').textContent=e.message;}
  }
  get('mimic-run').addEventListener('click',()=>task(async()=>{
    const result=await api.mimicAction({name:saved(),action:'diagnose',command:get('mimic-diagnostic').value,flags:values('mimic-test-fields')},node);
    if(closed)return;clearTimeout(poll);job={id:result.id,name:current,node};try{sessionStorage.setItem('sober-camouflage-job',JSON.stringify(job));}catch{}get('mimic-test-output').textContent='';await pollJob();
  }));
  let xui=[];
  const xuiNode=document.createElement('select');xuiNode.id='mimic-xui-node';
  const xuiLabel=document.createElement('label');xuiLabel.textContent='Foreign node';xuiLabel.append(xuiNode);
  get('mimic-xui-output').before(xuiLabel);
  function renderXui() {
    const format=get('mimic-xui-format').value;
    const row=xui.find(row=>row.alias===xuiNode.value);
    get('mimic-xui-output').textContent=row?JSON.stringify({outbounds:[row[format]],routing:{domainStrategy:row.domain_strategy,rules:[row.routing_rule]}},null,2):'';
  }
  get('mimic-xui').addEventListener('click',()=>task(async()=>{xui=await api.mimicAction({name:saved(),action:'integration'},node);if(!closed){xuiNode.innerHTML=xui.map(row=>`<option value="${esc(row.alias)}">${esc(row.alias)} — ${esc(row.address)}:${row.port}</option>`).join('');renderXui();message('Merge into x-ui Xray settings. Replace the inbound tag and review existing routing rules before applying.');}}));
  get('mimic-xui-format').addEventListener('change',renderXui);
  xuiNode.addEventListener('change',renderXui);
  get('mimic-xui-copy').addEventListener('click',()=>copyText(get('mimic-xui-output').textContent));
  host.querySelectorAll('[data-host]').forEach(el=>el.addEventListener('click',()=>task(async()=>{
    const confirmations={'move-ssh':'MOVE SSH','open-firewall':'OPEN PORTS',bbr:'ENABLE BBR'},command=el.dataset.host;
    const confirm=window.prompt(`Apply ${el.textContent} on ${node||'this server'}? Type ${confirmations[command]} to continue.`);
    if(confirm!==confirmations[command])return;
    const result=await api.mimicAction({name:saved(),action:'host-action',command,confirm},node);output(result.output);message('Host operation finished. Review its output.');
  })));
  get('mimic-node').addEventListener('change',()=>{node=get('mimic-node').value;fresh();task(load);});
  task(async()=>{
    catalog=await api.mimicOptions();if(closed)return;
    get('mimic-operation').innerHTML=['setup-foreign','setup-iran','add-node','edit-node','remove-node','edit-foreign','pairing'].map(c=>`<option value="${c}">${labels[c]}</option>`).join('');
    get('mimic-diagnostic').innerHTML=['probe','test','speedtest','check-ip'].map(c=>`<option value="${c}">${labels[c]}</option>`).join('');
    fresh();form('probe','mimic-test-fields');
    try {const r=await api.nodesCached();if(!closed)for(const n of r.nodes||[]){const opt=document.createElement('option');opt.value=n.name;opt.textContent=n.name;get('mimic-node').append(opt);}}catch{}
    if(closed)return;await load();
    if(ctx.query?.get('name'))await select(ctx.query.get('name'));
  });
}
