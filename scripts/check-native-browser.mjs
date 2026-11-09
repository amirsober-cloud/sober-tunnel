// Run beside check-native-panel.py inside its disposable container.
import { readFile, writeFile } from 'node:fs/promises';
import { spawn } from 'node:child_process';
import assert from 'node:assert/strict';

const cfg=JSON.parse(await readFile('/etc/sober-tunnel/webui.json','utf8'));
const origin=`http://127.0.0.1:${cfg.port}`, base=origin+'/'+cfg.base_path.replace(/^\/+|\/+$/g,'');
const chrome=spawn('chromium',['--headless=new','--no-sandbox','--disable-gpu','--disable-dev-shm-usage','--no-proxy-server','--remote-debugging-port=19222','--user-data-dir=/tmp/sober-native-browser','about:blank'],{stdio:'ignore'});
const sleep=ms=>new Promise(resolve=>setTimeout(resolve,ms));
let socket, seq=0;const waiting=new Map(), errors=[];
async function send(method,params={}) {
  const id=++seq;
  return new Promise((resolve,reject)=>{
    const timer=setTimeout(()=>{waiting.delete(id);reject(new Error('CDP timeout: '+method));},30000);
    waiting.set(id,{resolve:value=>{clearTimeout(timer);resolve(value);},reject});
    socket.send(JSON.stringify({id,method,params}));
  });
}
async function evaluate(expression) {
  const r=await send('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});
  if(r.exceptionDetails)throw new Error(r.exceptionDetails.exception?.description||r.exceptionDetails.text);
  return r.result?.value;
}
async function until(expression) {
  const deadline=Date.now()+30000;
  while(Date.now()<deadline){if(await evaluate(`Boolean(${expression})`))return;await sleep(200);}
  const state=await evaluate("JSON.stringify({route:location.hash,heading:document.querySelector('h1')?.textContent,notice:document.querySelector('#mimic-result')?.textContent})");
  throw new Error('UI condition timed out: '+expression+'; '+state+'; exceptions: '+errors.join(','));
}
try {
  let target;
  for(let n=0;n<100;n++){
    try {target=(await(await fetch('http://127.0.0.1:19222/json/list')).json()).find(t=>t.type==='page');if(target)break;}catch{}
    await sleep(200);
  }
  assert(target,'Chromium did not start');
  socket=new WebSocket(target.webSocketDebuggerUrl);
  await new Promise((resolve,reject)=>{socket.addEventListener('open',resolve,{once:true});socket.addEventListener('error',reject,{once:true});});
  socket.addEventListener('message',event=>{
    const message=JSON.parse(event.data);
    if(message.method==='Runtime.exceptionThrown')errors.push(message.params.exceptionDetails.text);
    const entry=waiting.get(message.id);if(!entry)return;waiting.delete(message.id);
    if(message.error)entry.reject(new Error(message.error.message));else entry.resolve(message.result);
  });
  await send('Page.enable');await send('Runtime.enable');
  await send('Emulation.setDeviceMetricsOverride',{width:1440,height:1100,deviceScaleFactor:1,mobile:false});
  const navigation=await send('Page.navigate',{url:base+'/login'});assert(!navigation.errorText,navigation.errorText);
  await until("document.querySelector('#pw')");
  await evaluate(`document.querySelector('#pw').value=${JSON.stringify(cfg.password)};document.querySelector('form').requestSubmit();true`);
  await until("document.querySelector('#panel') && !document.querySelector('#panel').hidden");
  await evaluate("location.hash='/tunnels';true");
  await until("document.querySelector('#mimic-fleet .sb')");
  await evaluate("document.querySelector('[data-act=add]').click();true");
  await until("document.querySelector('.mimic-engine-choice button')");
  await until("!document.querySelector('.mimic-engine-choice').closest('.dlg').classList.contains('priming')");
  await evaluate("document.querySelector('.mimic-engine-choice button').click();true");
  await until("document.querySelector('#mimic-fields [data-flag=public-ip]') && !document.querySelector('#mimic-operation').disabled");
  await evaluate(`(()=>{const f=document.querySelector('#mimic-form');const set=(n,v)=>f.querySelector('[data-flag="'+n+'"]').value=v;document.querySelector('#mimic-name').value='browser-preview-exit';set('mimics','tls');set('tls-port','18443');set('http-decoy-port','0');set('public-ip','203.0.113.10');f.requestSubmit();return true;})()`);
  await until("document.querySelector('#mimic-result').textContent.includes('Preview validated')");
  await evaluate(`(()=>{const op=document.querySelector('#mimic-operation');op.value='pairing';op.dispatchEvent(new Event('change'));document.querySelector('#mimic-fields [data-flag="public-ip"]').value='203.0.113.10';document.querySelector('#mimic-form').requestSubmit();return true;})()`);
  await until("document.querySelector('#mimic-result').textContent.includes('Pairing token exported')");
  await evaluate(`(()=>{window.__pairToken=document.querySelector('#mimic-output').textContent.trim();const op=document.querySelector('#mimic-operation');op.value='setup-iran';op.dispatchEvent(new Event('change'));document.querySelector('#mimic-name').value='browser-preview-hub';const f=document.querySelector('#mimic-form');const set=(n,v)=>f.querySelector('[data-flag="'+n+'"]').value=v;set('alias','exit');set('socks-port','11081');set('token',window.__pairToken);f.requestSubmit();return true;})()`);
  await until("document.querySelector('#mimic-result').textContent.includes('Preview validated') && JSON.parse(document.querySelector('#mimic-config').value).role==='iran'");
  assert(await evaluate(`(()=>{const n=JSON.parse(document.querySelector('#mimic-config').value).foreign_nodes[0];return n.socks_bind==='127.0.0.1'&&n.local_socks_port===11081&&n.endpoints[0].target==='203.0.113.10:18443';})()`),'Pairing preview mismatch');
  await evaluate("location.hash='/tunnels/camouflage?name=test-hub';true");
  await until("document.querySelector('#mimic-name').value==='test-hub' && !document.querySelector('#mimic-xui').disabled");
  await evaluate("document.querySelector('#mimic-xui').click();true");
  await until("document.querySelector('#mimic-xui-output').textContent.includes('outbounds')");
  assert(await evaluate("JSON.parse(document.querySelector('#mimic-xui-output').textContent).outbounds[0].settings.servers[0].port===11081"),'x-ui browser export port mismatch');
  if(process.env.SOBER_TEST_XRAY){
    const button="[...document.querySelectorAll('.mimic-page button')].find(b=>b.textContent.includes('existing Xray'))";
    await until(`!(${button}).disabled`);
    await evaluate(`(${button}).click();true`);
    await until("document.querySelector('#mimic-result').textContent.includes('Existing Xray configuration check finished')");
    assert(await evaluate("document.querySelector('#mimic-output').textContent.includes('✓ legacy_outbound') && document.querySelector('#mimic-output').textContent.includes('✓ modern_outbound')"),'Existing x-ui Xray check failed in UI');
  }
  await until("!document.querySelector('#mimic-run').disabled");
  await evaluate("document.querySelector('#mimic-run').click();true");
  await until("document.querySelector('#mimic-job').textContent.startsWith('Finished')");
  assert(await evaluate("document.querySelector('#mimic-test-output').textContent.includes('OK')"),'Probe output missing from UI');
  await evaluate("document.querySelector('#mimic-new').click();delete window.__pairToken;true");
  assert.equal(errors.length,0,'Browser runtime errors');
  if(process.env.SOBER_SCREENSHOT){const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:false});await writeFile(process.env.SOBER_SCREENSHOT,Buffer.from(shot.data,'base64'));}
  console.log('NATIVE_BROWSER_OK: Tunnels engine choice, native forms, pairing round-trip, x-ui JSON, async diagnostic UI, zero runtime exceptions');
} finally {
  if(socket?.readyState===WebSocket.OPEN){try{await send('Browser.close');}catch{}socket.close();}
  chrome.kill();
}
