import io
import json
from pathlib import Path
import tarfile
import tempfile
import subprocess
import unittest
from unittest.mock import patch

import prepare as p

HEAD = 'a' * 40
BASE = 'b' * 40
RELEASE = 'registry/hasura/example/releases/v1.2.3/connector-packaging.json'
METADATA = b'version: v1\npackagingDefinition:\n  type: ManagedDockerBuild\n'


def archive(metadata=METADATA, extra=None):
    buf = io.BytesIO()
    with tarfile.open(fileobj=buf, mode='w:gz') as tar:
        item = tarfile.TarInfo('./' + p.METADATA)
        item.size = len(metadata)
        tar.addfile(item, io.BytesIO(metadata))
        if extra is not None:
            tar.addfile(extra, io.BytesIO(b''))
    return buf.getvalue()


def bundle(out):
    data = archive()
    files = {RELEASE: json.dumps({'version': 'v1.2.3', 'uri': 'https://github.com/hasura/example/releases/download/v1.2.3/example.tar.gz', 'checksum': {'type': 'sha256', 'value': p.digest(data)}}).encode(), p.package_path(RELEASE): data, p.metadata_path(RELEASE): METADATA}
    for name, raw in files.items():
        p.write(out, name, raw)
    p.write(out, 'manifest.json', json.dumps({'pr': '42', 'before': None, 'head': HEAD, 'base': BASE, 'sha256': {n: p.digest(raw) for n, raw in files.items()}}).encode())
    p.write(out, 'changed-files.json', json.dumps({'added_files': [RELEASE], 'modified_files': [], 'deleted_files': []}).encode())


class StagingInputTests(unittest.TestCase):
    def test_exact_identity(self):
        p.validate_identity('42', HEAD)
        for pr, sha in [('0', HEAD), ('42', 'main'), ('42', 'a' * 39)]:
            with self.assertRaises(p.InputError):
                p.validate_identity(pr, sha)

    def test_release_locations(self):
        self.assertEqual('github', p.url_kind('https://github.com/hasura/example/releases/download/v1/example.tar.gz'))
        self.assertEqual('gcs', p.url_kind('https://storage.googleapis.com/staging-connector-platform-registry/ndc-jdbc-assets/example/example.tar.gz'))
        for url in ['http://github.com/hasura/example/releases/download/v1/example.tar.gz', 'https://github.com/other/example/releases/download/v1/a.tgz', 'https://github.com/hasura/example/raw/main/a.tgz']:
            with self.assertRaises(p.InputError):
                p.url_kind(url)

    def test_archive_entry_policy(self):
        self.assertEqual(METADATA, p.validate_archive(archive()))
        for kind in [tarfile.SYMTYPE, tarfile.LNKTYPE, tarfile.FIFOTYPE]:
            entry = tarfile.TarInfo('extra')
            entry.type = kind
            with self.assertRaises(p.InputError):
                p.validate_archive(archive(extra=entry))
        with self.assertRaises(p.InputError):
            p.validate_archive(archive(extra=tarfile.TarInfo('./' + p.METADATA)))
        with patch.object(p, 'MAX_UNPACKED', 100):
            with self.assertRaises(p.InputError):
                p.validate_archive(archive())

    def test_staging_bundle_verification(self):
        with tempfile.TemporaryDirectory() as name, patch.object(p, 'git', return_value=BASE.encode()):
            out = Path(name)
            bundle(out)
            p.verify('42', HEAD, out)
            with self.assertRaises(p.InputError):
                p.verify('42', 'c' * 40, out)
            p.write(out, 'unexpected', b'other')
            with self.assertRaises(p.InputError):
                p.verify('42', HEAD, out)
            (out / 'unexpected').unlink()
            (out / p.metadata_path(RELEASE)).write_bytes(b'not the archive metadata')
            with self.assertRaises(p.InputError):
                p.verify('42', HEAD, out)

    def test_protected_main_push_range_uses_merged_blobs_only(self):
        with tempfile.TemporaryDirectory() as name:
            root = Path(name)
            def cmd(*args):
                return subprocess.check_output(['git', '-c', 'user.name=Test', '-c', 'user.email=test@example.com', *args], cwd=root)
            cmd('init', '-q')
            path = Path('registry/hasura/example/README.md')
            (root / path).parent.mkdir(parents=True)
            (root / path).write_text('before')
            cmd('add', '.'); cmd('commit', '-qm', 'before')
            before = cmd('rev-parse', 'HEAD').decode().strip()
            (root / path).write_text('merged registry data')
            (root / path.parent / 'test.sh').write_text('# not a publication file')
            cmd('add', '.'); cmd('commit', '-qm', 'after')
            after = cmd('rev-parse', 'HEAD').decode().strip()
            (root / path).write_text('untracked worktree data')
            with patch.object(p, 'git', side_effect=cmd):
                out = root / 'bundle'
                p.prepare(None, after, out, before=before)
                self.assertEqual('merged registry data', (out / path).read_text())
                self.assertFalse((out / path.parent / 'test.sh').exists())
                p.verify(None, after, out, before=before)
                with self.assertRaises(p.InputError):
                    p.verify(None, after, out, before='d' * 40)
                with self.assertRaises(p.InputError):
                    p.prepare(None, after, root / 'bad', before='0' * 40)
                with self.assertRaises(p.InputError):
                    p.prepare(None, before, root / 'bad', before=after)

    def test_new_connector_paths_are_data_only(self):
        self.assertIsNone(p.OVERVIEW.fullmatch('registry/hasura/example/tests/script.sh'))
        self.assertIsNone(p.OVERVIEW.fullmatch('registry/hasura/example/../README.md'))
        self.assertEqual({'metadata.json'}, {Path(x).name for x in ['registry/hasura/example/metadata.json'] if p.OVERVIEW.fullmatch(x)})


if __name__ == '__main__':
    unittest.main()
