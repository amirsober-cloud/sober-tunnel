#!/usr/bin/env python3
"""Run ONLY in a disposable Linux container; uses /etc and /usr/local/bin.

Copy build/sober-tunnel and build/sober-tunnel-mimic to /usr/local/bin first.
The systemctl stub starts real tunnel processes without a systemd daemon.
"""
import http.cookiejar
import base64
import json
import os
from pathlib import Path
import re
import secrets
import signal
import subprocess
import sys
import time
import urllib.parse
import urllib.request


def systemctl():
    args = sys.argv[2:]
    if args[0] == "daemon-reload":
        return
    unit = args[-1]
    if not re.fullmatch(r"sober-mimic-[A-Za-z0-9_-]+\.service", unit):
        raise SystemExit(1)
    state = Path("/tmp/sober-service-state")
    state.mkdir(exist_ok=True)
    pidfile = state / unit
    pid = int(pidfile.read_text()) if pidfile.exists() else None
    active = False
    if pid:
        try:
            os.kill(pid, 0)
            active = True
        except ProcessLookupError:
            pass
    if args[0] == "is-active":
        print("active" if active else "inactive")
        raise SystemExit(0 if active else 3)
    if args[0] in ("disable", "restart") and active:
        os.kill(pid, signal.SIGTERM)
        time.sleep(0.3)
        active = False
    if args[0] in ("enable", "restart") and not active:
        name = unit[len("sober-mimic-"):-len(".service")]
        with (state / (unit + ".log")).open("ab") as output:
            process = subprocess.Popen(
                ["/usr/local/bin/sober-tunnel", "--mimic-config",
                 "/etc/sober-tunnel/mimic/" + name + ".json"],
                stdin=subprocess.DEVNULL, stdout=output, stderr=output,
                start_new_session=True)
        pidfile.write_text(str(process.pid))


def main():
    if not Path("/.dockerenv").exists():
        raise SystemExit("This test must run in a disposable Docker container")
    script = str(Path(__file__).resolve())
    stub = Path("/usr/local/bin/systemctl")
    stub.write_text("#!/bin/sh\nexec python3 " + script + " --systemctl \"$@\"\n")
    stub.chmod(0o755)
    config_dir = Path("/etc/sober-tunnel")
    config_dir.mkdir(mode=0o700, exist_ok=True)
    password = secrets.token_hex(24)
    prefix = secrets.token_hex(12)
    panel_config = config_dir / "webui.json"
    panel_config.write_text(json.dumps({"port": 17778, "password": password,
                                        "base_path": prefix, "https": False}))
    panel_config.chmod(0o600)
    origin = "http://127.0.0.1:17778"
    base = origin + "/" + prefix + "/"
    jar = http.cookiejar.CookieJar()
    client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar))

    def api(payload):
        request = urllib.request.Request(base + "api/mimic", json.dumps(payload).encode(),
                                         {"Content-Type": "application/json", "Origin": origin})
        with client.open(request, timeout=30) as response:
            return json.load(response)

    def diagnostic(command):
        job = api({"name": "test-hub", "action": "diagnose", "command": command,
                   "flags": {"node": "exit", **({"seconds": "1"} if command == "speedtest" else {})}})
        for _ in range(160):
            result = api({"name": "test-hub", "action": "job", "job_id": job["id"]})
            if result["done"]:
                assert not result.get("error"), result.get("error")
                return result["output"]
            time.sleep(0.25)
        raise AssertionError("diagnostic did not finish")

    with Path("/tmp/sober-panel-check.log").open("wb") as output:
        panel = subprocess.Popen(["/usr/local/bin/sober-tunnel", "--webui"], stdout=output, stderr=output)
    try:
        for _ in range(100):
            try:
                client.open(base + "login", timeout=1).close()
                break
            except OSError:
                time.sleep(0.2)
        request = urllib.request.Request(base + "login", urllib.parse.urlencode({"password": password}).encode(),
                                         {"Origin": origin})
        client.open(request, timeout=10).close()
        exit_preview = api({"action": "prepare", "name": "test-exit", "command": "setup-foreign",
                            "flags": {"mimics": "tls", "tls-port": "18443", "http-decoy-port": "0",
                                      "public-ip": "127.0.0.1"}})
        assert not (config_dir / "mimic/test-exit.json").exists(), "preview changed live config"
        api({"action": "save", "name": "test-exit", "config": exit_preview["config"]})
        pairing = api({"action": "prepare", "name": "test-exit", "command": "pairing",
                       "flags": {"public-ip": "127.0.0.1"}})["output"].strip()
        hub_preview = api({"action": "prepare", "name": "test-hub", "command": "setup-iran",
                           "flags": {"alias": "exit", "socks-port": "11081", "token": pairing}})
        api({"action": "save", "name": "test-hub", "config": hub_preview["config"]})
        for name in ("test-exit", "test-hub"):
            checks = api({"action": "preflight", "name": name})
            assert all(check["ok"] for check in checks), checks
            api({"action": "start", "name": name})
        time.sleep(1)
        probe = diagnostic("probe")
        assert "OK" in probe and "FAIL" not in probe, probe
        speed = diagnostic("speedtest")
        assert "download" in speed and "upload" in speed and "FAILED" not in speed, speed
        rpc_request = base64.b64encode(json.dumps({"op": "mimic", "body": {"name": "test-hub", "action": "diagnose",
                                                  "command": "probe", "flags": {"node": "exit"}}}).encode()).decode() + "\n"
        rpc = subprocess.run(["/usr/local/bin/sober-tunnel", "node", "exec", "-"],
                             input=rpc_request,
                             text=True, capture_output=True, check=True, timeout=15)
        reply = json.loads(base64.b64decode(rpc.stdout.strip().splitlines()[-1]))
        assert reply["ok"], reply.get("err")
        for _ in range(100):
            result = api({"action": "job", "name": "test-hub", "job_id": reply["body"]["id"]})
            if result["done"]:
                assert not result.get("error") and "OK" in result["output"], "worker failed after node RPC exited"
                break
            time.sleep(0.2)
        else:
            raise AssertionError("detached RPC diagnostic did not complete")
        integrations = api({"action": "integration", "name": "test-hub"})
        assert integrations[0]["address"] == "127.0.0.1" and integrations[0]["port"] == 11081
        code = subprocess.check_output(["curl", "-fsS", "--max-time", "30", "--socks5-hostname",
                                        "127.0.0.1:11081", "https://example.com", "-o", "/dev/null",
                                        "-w", "%{http_code}"], text=True)
        assert code == "200", code
        if os.environ.get("SOBER_BROWSER_CHECK") == "1":
            subprocess.run(["node", str(Path(__file__).with_name("check-native-browser.mjs"))], check=True, timeout=180)
        print("NATIVE_PANEL_E2E_OK: pairing, save, preflight, real engines, async probe/speed, x-ui export, SOCKS HTTPS 200")
    finally:
        panel.terminate()
        panel.wait(timeout=10)


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--systemctl":
        systemctl()
    else:
        main()
