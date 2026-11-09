import { esc } from '../lib/dom.js';
import * as api from '../api.js';

// Mounted inside Tunnels, alongside the existing transport cards.
export function mountMimicFleet(host) {
  let closed=false,timer;
  async function load() {
    try {
      const items=await api.mimicInstances();if(closed)return;
      host.innerHTML=`<div class="sech2"><h3>Camouflage / Hedioum</h3><span>${items.length}</span><span class="sp"></span><a class="sb primary" href="#/tunnels/camouflage">Add camouflage tunnel</a></div>`+
        (items.length?'<div class="mimic-fleet">'+items.map(i=>`<a class="tl-card" href="#/tunnels/camouflage?name=${encodeURIComponent(i.name)}"><strong>${esc(i.name)}</strong><p>${esc(i.role)} · ${i.active?'Running':'Stopped'}</p><span>Manage · Pairing · Tests · x-ui</span></a>`).join('')+'</div>':'<p>Create an Iran SOCKS5 hub or a foreign exit from Add tunnel.</p>');
    }catch(e){if(!closed)host.textContent='Camouflage: '+e.message;}
    if(!closed)timer=setTimeout(load,10000);
  }
  load();return()=>{closed=true;clearTimeout(timer);};
}
