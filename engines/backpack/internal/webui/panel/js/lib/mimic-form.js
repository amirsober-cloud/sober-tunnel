export function savedFlags(config, alias) {
 const foreign=config.role==='foreign';
 const node=(config.foreign_nodes||[]).find(n=>n.alias===alias)||(config.foreign_nodes||[])[0]||{};
 const source=foreign?config:node;
 const keys={'alias':'alias','persona':'persona','domain':'domain','acme-email':'acme_email','decoy-style':'decoy_style','decoy-port':'decoy_port','http-decoy-port':'http_decoy_port','egress-mode':'egress_ip_mode','egress-bind-ip':'egress_bind_ip','token':'auth_token','target-ip':'target_ip','socks-port':'local_socks_port','socks-bind':'socks_bind','min':'min_connections','max':'max_connections','bw':'bandwidth_limit_mbps','jitter':'bandwidth_jitter_mbps','tun':'tun_enabled','tun-name':'tun_name','tun-addr':'tun_addr','dns':'dns_enabled','gateway':'gateway_enabled','gateway-iface':'gateway_iface'};
 const result={};
 for(const [flag,key] of Object.entries(keys))if(source[key]!==undefined)result[flag]=String(source[key]);
 for(const flag of ['tun','dns','gateway'])if(!foreign)result[flag]=String(Boolean(source[keys[flag]]));
 if(!foreign)result['gateway-lan']=(node.gateway_lan||[]).join(',');
 const endpoints=foreign?(config.mimics||[]):(node.endpoints?.length?node.endpoints:[{mimic:'ssh',target:`${node.target_ip}:${node.target_port}`}]);
 result.mimics=[...new Set(endpoints.map(e=>e.type||e.mimic))].join(',');
 const ports={ssh:foreign?'listen-port':'ssh-port',tls:'tls-port',smtp:'smtp-port',imap:'imap-port',smtps:'smtps-port',imaps:'imaps-port',directadmin:'directadmin-port'};
 for(const e of endpoints){const type=e.type||e.mimic,port=foreign?e.port:Number(String(e.target||'').split(':').at(-1));if(ports[type]&&Number.isFinite(port))result[ports[type]]=String(port);}
 if(!foreign && endpoints[0]?.target){const target=endpoints[0].target;result['target-ip']=target.slice(0,target.lastIndexOf(':')).replace(/^\[|\]$/g,'');}
 result['tls-servername']=endpoints.find(e=>e.server_name)?.server_name||'';
 return result;
}
export function pairingFlags(text) {
 const value=text.trim();
 if(/^[a-f0-9]{32}$/i.test(value))return null;
 const token=JSON.parse(atob(value.replaceAll('-','+').replaceAll('_','/')));
 if(token.v!==2||!token.ip||!token.eps||!/^[a-f0-9]{32}$/i.test(token.auth||''))throw new Error('Paste the full Pairing token exported by the foreign server.');
 const flags={'target-ip':token.ip,mimics:Object.keys(token.eps).join(','),'tls-servername':token.sni||''};
 for(const [type,port] of Object.entries(token.eps)){
 if(!Number.isInteger(port)||port<1||port>65535)throw new Error('Invalid pairing port');
 flags[type+'-port']=String(port);
 }
 return flags;
}