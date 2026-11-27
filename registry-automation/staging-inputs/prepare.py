"""Build and verify a bounded, data-only bundle for the staging publisher.

The prepare job has no publication credentials and never checks out a PR tree.
The publish job runs verify offline on the artifact from the same workflow run.
"""
import argparse
import hashlib
import io
import json
import gzip
import os
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys
import tarfile
import urllib.parse
import urllib.request
import xml.etree.ElementTree as ET

SLUG = r"[a-z0-9][a-z0-9_-]{0,77}"
VER = r"v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.+-]+)?"
ROOT = rf"registry/{SLUG}/{SLUG}"
OVERVIEW = re.compile(rf"{ROOT}/(?:metadata\.json|README\.md|logo\.(?:png|svg))\Z", re.ASCII)
RELEASE = re.compile(rf"{ROOT}/releases/({VER})/connector-packaging\.json\Z", re.ASCII)
SHA = re.compile(r"[0-9a-f]{40}\Z", re.ASCII)
HASH = re.compile(r"[0-9a-fA-F]{64}\Z", re.ASCII)
BUCKET = "staging-connector-platform-registry"
RELEASE_HOSTS = {"release-assets.githubusercontent.com", "objects.githubusercontent.com", "github-releases.githubusercontent.com"}
MAX_PACKAGE = 32 * 1024 * 1024
MAX_UNPACKED = 96 * 1024 * 1024
MAX_DATA = 2 * 1024 * 1024
MAX_BUNDLE = 192 * 1024 * 1024
MAX_DIFF = 400
METADATA = ".hasura-connector/connector-metadata.yaml"


def fail(message):
    raise InputError(message)


def unique_pairs(pairs):
    d = {}
    for k, v in pairs:
        if k in d:
            fail("duplicate JSON field")
        d[k] = v
    return d


def read_json(raw):
    return json.loads(raw, object_pairs_hook=unique_pairs,
                      parse_constant=lambda _: fail("invalid JSON number"))


def validate_identity(pr, head):
    if not re.fullmatch(r"[1-9][0-9]{0,8}", pr, re.ASCII) or not SHA.fullmatch(head):
        fail("a positive PR number and full lowercase head SHA are required")


def git(*args):
    return subprocess.check_output(["git", "-c", "core.hooksPath=/dev/null", *args], timeout=60)


def read_blob(tree, name):
    entries = git("ls-tree", "-z", tree, "--", name).split(b"\0")
    if len(entries) != 2 or not entries[0]:
        fail("expected exactly one registry blob")
    raw_meta, _, raw_name = entries[0].partition(b"\t")
    mode, kind, sha = raw_meta.decode("ascii").split()
    if (mode, kind) != ("100644", "blob") or raw_name != name.encode("ascii") or not SHA.fullmatch(sha):
        fail("registry input is not a regular non-executable file")
    if int(git("cat-file", "-s", sha)) > MAX_DATA:
        fail("registry input exceeds the size limit")
    return git("cat-file", "blob", sha)


def url_kind(url):
    u = urllib.parse.urlsplit(url)
    if len(url) > 1000 or any(ord(c) < 33 or ord(c) > 126 for c in url) or u.scheme != "https" or u.netloc != u.hostname or u.query or u.fragment:
        fail("package URL must be a plain HTTPS release URL")
    if u.hostname == "github.com" and re.fullmatch(r"/(?:hasura|singlestore-labs)/[A-Za-z0-9_.-]+/releases/download/[A-Za-z0-9_.+%-]+/[A-Za-z0-9_.+%-]+\.(?:tar\.gz|tgz)", u.path):
        return "github"
    if u.hostname == "storage.googleapis.com" and re.fullmatch(rf"/{BUCKET}/ndc-jdbc-assets(?:/[A-Za-z0-9_.+-]+)+\.(?:tar\.gz|tgz)", u.path) and not {".", ".."}.intersection(u.path.split("/")):
        return "gcs"
    fail("package source is not on the release allowlist")


class InputError(ValueError):
    pass


class SafeRedirect(urllib.request.HTTPRedirectHandler):
    max_repeats = 2
    max_redirections = 3

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        start = urllib.parse.urlsplit(req.full_url)
        target = urllib.parse.urlsplit(newurl)
        if start.hostname not in RELEASE_HOSTS | {"github.com"} or target.scheme != "https" or target.netloc != target.hostname or target.hostname not in RELEASE_HOSTS:
            fail("package fetch encountered an unapproved redirect")
        return super().redirect_request(req, fp, code, msg, headers, newurl)


def download(url):
    url_kind(url)
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), SafeRedirect())
    req = urllib.request.Request(url, headers={"User-Agent": "ndc-hub-staging-validate"})
    with opener.open(req, timeout=20) as response:
        if response.status != 200:
            fail("package fetch failed")
        data = response.read(MAX_PACKAGE + 1)
    if not data or len(data) > MAX_PACKAGE:
        fail("package exceeds compressed size limit")
    return data


def validate_archive(data):
    if not data or len(data) > MAX_PACKAGE:
        fail("archive exceeds compressed size limit")
    size = 0
    names = set()
    found = False
    with gzip.GzipFile(fileobj=io.BytesIO(data)) as stream:
        unpacked = stream.read(MAX_UNPACKED + 1)
    if len(unpacked) > MAX_UNPACKED:
        fail("archive exceeds expanded size limit")
    metadata = None
    with tarfile.open(fileobj=io.BytesIO(unpacked), mode="r:") as archive:
        for i, member in enumerate(archive, 1):
            name = member.name.removeprefix("./").rstrip("/")
            if name in ("", ".") and member.isdir():
                continue
            if i > 10000 or not name or len(name) > 500 or not name.isascii() or name.startswith("/") or "\\" in name or any(ord(c) < 32 or ord(c) == 127 for c in name) or set(name.split("/")) & {"", ".", ".."} or name in names:
                fail("invalid or duplicate archive path")
            names.add(name)
            if not (member.isfile() or member.isdir()) or member.issparse() or (member.isdir() and member.size):
                fail("unsupported archive entry type")
            size += member.size
            if size > MAX_UNPACKED or member.size < 0:
                fail("archive exceeds expanded size limit")
            if name == METADATA:
                if not member.isfile() or not 0 < member.size <= MAX_DATA:
                    fail("invalid package metadata entry")
                found = True
                metadata = archive.extractfile(member).read(MAX_DATA + 1)
    if not found or not metadata or len(metadata) > MAX_DATA:
        fail("missing or oversized connector metadata in archive")
    return metadata


def validate_data(name, raw):
    if not raw or len(raw) > MAX_DATA:
        fail("invalid registry file size")
    if name.endswith("/metadata.json"):
        d = read_json(raw)
        overview = d.get("overview") if isinstance(d, dict) else None
        if not isinstance(overview, dict) or overview.get("namespace") != name.split("/")[1]:
            fail("connector namespace does not match its metadata")
    if name.endswith("/README.md"):
        raw.decode("utf-8")
    if name.endswith("/logo.png") and not raw.startswith(b"\x89PNG\r\n\x1a\n"):
        fail("invalid PNG logo")
    if name.endswith("/logo.svg"):
        if re.search(rb"<!\s*(?:DOCTYPE|ENTITY)", raw, re.I):
            fail("SVG DTDs are not accepted")
        root = ET.fromstring(raw)
        if root.tag.split("}")[-1] != "svg":
            fail("invalid SVG logo")
        for item in root.iter():
            if item.tag.split("}")[-1].lower() in {"script", "foreignobject", "iframe"}:
                fail("active content is not accepted in SVG logos")
            for k, v in item.attrib.items():
                attr = k.split("}")[-1].lower()
                if attr.startswith("on") or (attr == "href" and not v.startswith("#")):
                    fail("only local SVG references are accepted")
    match = RELEASE.fullmatch(name)
    if match:
        d = read_json(raw)
        checksum = d.get("checksum") if isinstance(d, dict) else None
        if not isinstance(checksum, dict) or checksum.get("type") != "sha256" or not isinstance(checksum.get("value"), str) or not HASH.fullmatch(checksum["value"]) or d.get("version") != match[1] or not isinstance(d.get("uri"), str):
            fail("invalid connector release metadata or SHA-256")
        url_kind(d["uri"])
        return d["uri"], checksum["value"].lower()
    return None


def package_path(name):
    return "packages/" + name.removeprefix("registry/").removesuffix("/connector-packaging.json") + "/package.tgz"


def metadata_path(name):
    return package_path(name).removesuffix("package.tgz") + "connector-metadata.yaml"


def digest(data):
    return hashlib.sha256(data).hexdigest()


def write(out, name, data):
    p = out / name
    p.parent.mkdir(parents=True, exist_ok=True)
    with p.open("xb") as f:
        f.write(data)


def prepare(pr, head, out):
    base = git("rev-parse", "HEAD").decode().strip()
    ref = f"refs/pull/{pr}/head"
    git("fetch", "--no-tags", "--no-recurse-submodules", "origin", ref)
    if git("rev-parse", "FETCH_HEAD").decode().strip() != head:
        fail("the reviewed PR head changed")
    ancestor = git("merge-base", base, head).decode().strip()
    r = git("diff", "--name-status", "-z", "--no-renames", ancestor, head, "--", "registry/").split(b"\0")
    if r[-1] != b"" or len(r) % 2 != 1 or (len(r) - 1) // 2 > MAX_DIFF:
        fail("invalid or oversized registry change set")
    out.mkdir(parents=True, mode=0o700)
    added, modified, manifest, total = [], [], {}, 0

    def add(name, raw):
        nonlocal total
        total += len(raw)
        if total > MAX_BUNDLE:
            fail("bundle exceeds size limit")
        write(out, name, raw)
        manifest[name] = digest(raw)

    for i in range(0, len(r) - 1, 2):
        status, name = r[i].decode("ascii"), r[i+1].decode("utf-8")
        is_release = RELEASE.fullmatch(name)
        if not (OVERVIEW.fullmatch(name) or is_release):
            continue  # Test fixtures, scripts and all non-publication paths are never copied.
        if status not in {"A", "M"} or (is_release and status != "A"):
            fail("publication paths cannot be removed or change type, and package definitions are immutable")
        if git("ls-tree", "-z", ancestor, "--", name) != git("ls-tree", "-z", base, "--", name):
            fail("main changed a publication path since the PR diverged; rebase and review again")
        raw = read_blob(head, name)
        pkg = validate_data(name, raw)
        add(name, raw)
        (added if status == "A" else modified).append(name)
        if pkg:
            url, expected = pkg
            data = download(url)
            if digest(data) != expected:
                fail("connector package SHA-256 mismatch")
            metadata = validate_archive(data)
            add(package_path(name), data)
            add(metadata_path(name), metadata)
    # New connector publication reads its README even if absent from the diff.
    # Reject incomplete entries rather than silently reading any other tree.
    for name in added:
        if name.endswith("/metadata.json"):
            prefix = name.removesuffix("metadata.json")
            if prefix + "README.md" not in added or not any(prefix + logo in added for logo in ("logo.png", "logo.svg")):
                fail("new connector publication requires a new README and logo")
    write(out, "changed-files.json", json.dumps({"added_files": added, "modified_files": modified, "deleted_files": []}, sort_keys=True).encode())
    write(out, "manifest.json", json.dumps({"pr": pr, "head": head, "base": base, "sha256": manifest}, sort_keys=True).encode())
    verify(pr, head, out)


def bundle_file(out, name, max_size):
    if name.startswith("/") or set(name.split("/")) & {"", ".", ".."}:
        fail("invalid bundle path")
    p = out
    for part in PurePosixPath(name).parts:
        p /= part
        if p.is_symlink():
            fail("bundle symlinks are forbidden")
    if not p.is_file() or p.stat().st_size > max_size:
        fail("missing or oversized bundle file")
    with p.open("rb") as f:
        return f.read(max_size + 1)


def verify(pr, head, out):
    base = git("rev-parse", "HEAD").decode().strip()
    if out.is_symlink() or not out.is_dir():
        fail("invalid bundle directory")
    m = read_json(bundle_file(out, "manifest.json", MAX_DATA))
    d = read_json(bundle_file(out, "changed-files.json", MAX_DATA))
    if not isinstance(m, dict) or set(m) != {"pr", "head", "base", "sha256"} or (m["pr"], m["head"], m["base"]) != (pr, head, base) or not isinstance(m["sha256"], dict):
        fail("bundle must match the exact PR, head and trusted workflow commit")
    if not isinstance(d, dict) or set(d) != {"added_files", "modified_files", "deleted_files"} or not all(isinstance(v, list) for v in d.values()) or d["deleted_files"] or sum(len(v) for v in d.values()) > MAX_DIFF:
        fail("invalid bundle change manifest")
    names, total = set(), 0
    for status in ("added_files", "modified_files"):
        for name in d[status]:
            release = isinstance(name, str) and RELEASE.fullmatch(name)
            if not isinstance(name, str) or name in names or not (OVERVIEW.fullmatch(name) or release) or (release and status != "added_files"):
                fail("invalid publication path in bundle")
            raw = bundle_file(out, name, MAX_DATA)
            total += len(raw)
            names.add(name)
            if m["sha256"].get(name) != digest(raw):
                fail("registry file digest mismatch")
            pkg = validate_data(name, raw)
            if pkg:
                pn = package_path(name)
                data = bundle_file(out, pn, MAX_PACKAGE)
                total += len(data)
                names.add(pn)
                if digest(data) != pkg[1] or m["sha256"].get(pn) != pkg[1]:
                    fail("bundle archive digest mismatch")
                mn = metadata_path(name)
                metadata = bundle_file(out, mn, MAX_DATA)
                names.add(mn)
                total += len(metadata)
                if not metadata or m["sha256"].get(mn) != digest(metadata) or metadata != validate_archive(data):
                    fail("package metadata digest mismatch")
    if total > MAX_BUNDLE:
        fail("bundle exceeds size limit")
    if names != set(m["sha256"]):
        fail("bundle declares unexpected files")
    actual = set()
    for parent, directories, files in os.walk(out, followlinks=False):
        for entry in directories + files:
            if (Path(parent) / entry).is_symlink():
                fail("bundle symlinks are forbidden")
        actual.update(str((Path(parent) / f).relative_to(out)) for f in files)
    if actual != names | {"manifest.json", "changed-files.json"}:
        fail("bundle contains undeclared files")
    print(f"Validated {len(d['added_files']) + len(d['modified_files'])} publication inputs for the exact PR head.")


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--pr", required=True)
    p.add_argument("--head", required=True)
    p.add_argument("--output", required=True, type=Path)
    p.add_argument("--verify", action="store_true")
    args = p.parse_args()
    validate_identity(args.pr, args.head)
    (verify if args.verify else prepare)(args.pr, args.head, args.output)


if __name__ == "__main__":
    try:
        main()
    except InputError as e:
        # Only application-authored validation messages may reach Action logs.
        sys.exit(str(e))
    except Exception as e:
        sys.exit(f"Staging input validation failed ({type(e).__name__}).")
