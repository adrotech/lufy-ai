from pathlib import Path
import shutil,importlib.util,json,subprocess,stat
W=Path(__file__).resolve().parent;backup=Path('/Users/adrianrojas/.local/share/lufy-ai/backups/inline-20261027');proof=W/'standalone-proof';proof.mkdir()
# Run exclusively from durable resources, without candidate/payload/temp-workspace references.
for name in ['snapshot.tar.gz','snapshot-files.json','snapshot-dirs.json','prepared.tar.gz','plan.json','replace_personal.py']:shutil.copy2(backup/name,proof/name)
spec=importlib.util.spec_from_file_location('personal_replacement_standalone',proof/'replace_personal.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m);m.resources()
base=m.clone('full-restore');r=m.Replacement(base);assert r.apply()=='completed'
for key in ['global/lufy-ai','project/tools/lufy-cli-go/bin/lufy-ai','project/tools/lufy-cli-go/lufy-ai']:assert m.sha(r.path(key))==m.P['binary_sha256']
assert json.loads((r.root/'.lufy/managed-state/install-state.json').read_text())['schemaVersion']==3
assert r.state['data_saved']==m.DATA
# Model a validation failure after the entire candidate and derived projections were installed.
r.state['phase']='interrupted';r.persist();r=m.Replacement(base);assert r.restore()=='restored'
expected=json.loads((proof/'snapshot-files.json').read_text());assert m.snapshot_files(base)==expected
for key,mode in json.loads((proof/'snapshot-dirs.json').read_text()).items():assert (base/key).is_dir() and stat.S_IMODE((base/key).stat().st_mode)==mode,key
old=subprocess.run([str(base/'global/lufy-ai'),'version'],capture_output=True,text=True,check=True);assert 'v0.6.25' in old.stdout
old=subprocess.run([str(base/'global/lufy-ai'),'status','--target',str(r.root),'--json'],cwd=proof,capture_output=True,text=True,check=True);assert json.loads(old.stdout)['schemaVersion']==2
result={'resources_loaded_from_durable_backup':'passed','candidate_binaries_config_manifest_and_projections_replaced_before_failure':'passed','all_754_files_restored_with_modes':'passed','all_269_directories_restored_with_modes':'passed','previous_cli_reads_restored_state':'passed'}
(W/'full-restore-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
