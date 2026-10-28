from pathlib import Path
import importlib.util,json,tarfile
W=Path(__file__).resolve().parent;D=Path('/Users/adrianrojas/.local/share/lufy-ai/backups/retirement-20261028')
spec=importlib.util.spec_from_file_location('recovery',W/'replace_personal.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
base=m.clone('revision-interruption');r=m.Replacement(base);assert r.apply()=='completed'
with tarfile.open(D/'prepared.tar.gz') as tar:prior=tar.extractfile('lufy-ai').read()
key='global/lufy-ai';path=r.path(key);path.write_bytes(prior)
before=json.loads((W/'revision-preimages.json').read_text());assert m.sha(path)==before[key]['sha256']
r.state['accepted_postimages']={key:[before[key]['sha256'],m.P['binary_sha256']]};r.persist();r=m.Replacement(base);r.restore()
assert m.snapshot_files(base)==json.loads((W/'snapshot-files.json').read_text())
result={'joint_restore_with_known_partial_revision':'passed'};(W/'revision-rehearsal-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
