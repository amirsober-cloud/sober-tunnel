#!/usr/bin/env python3
"""Fetch a checksum-verified official Xray into a private TEST directory only."""
import hashlib
import io
import json
from pathlib import Path
import re
import sys
import urllib.request
import zipfile

destination = Path(sys.argv[1]).resolve()
destination.mkdir(parents=True, exist_ok=True, mode=0o700)

def fetch(url):
    request = urllib.request.Request(url, headers={"User-Agent": "sober-tunnel-integration-tests"})
    with urllib.request.urlopen(request, timeout=90) as response:
        return response.read()

release = json.loads(fetch("https://api.github.com/repos/XTLS/Xray-core/releases/latest"))
asset = next(a for a in release["assets"] if a["name"] == "Xray-linux-64.zip")
digest = asset.get("digest") or ""
if digest.startswith("sha256:"):
    expected = digest.split(":", 1)[1]
else:
    checksum = fetch(asset["browser_download_url"] + ".dgst").decode()
    expected = re.search(r"(?<![a-fA-F0-9])[a-fA-F0-9]{64}(?![a-fA-F0-9])", checksum).group(0).lower()
archive = fetch(asset["browser_download_url"])
if hashlib.sha256(archive).hexdigest() != expected:
    raise SystemExit("Official Xray archive checksum mismatch")
with zipfile.ZipFile(io.BytesIO(archive)) as bundle:
    # Extract only the binary, without trusting archive directory names.
    executable = bundle.read("xray")
(destination / "xray").write_bytes(executable)
(destination / "xray").chmod(0o755)
metadata = {"tag": release["tag_name"], "archive_sha256": expected, "source": asset["browser_download_url"]}
(destination / "provenance.json").write_text(json.dumps(metadata, indent=2) + "\n")
print("Verified official Xray for isolated tests:", release["tag_name"], expected)
