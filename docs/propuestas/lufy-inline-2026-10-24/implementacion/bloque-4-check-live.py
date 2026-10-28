from pathlib import Path
import json,hashlib,subprocess,os
W=Path(__file__).resolve().parent;R=Path('/Users/adrianrojas/Desktop/projects/lufy-ai');B=Path('/Users/adrianrojas/.local/bin/lufy-ai');P=json.loads((W/'plan.json').read_text());results={};outputs={}
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def run(args,cwd=W):
 p=subprocess.run(list(map(str,args)),cwd=cwd,text=True,capture_output=True,env=dict(os.environ,GIT_OPTIONAL_LOCKS='0',GIT_NO_LAZY_FETCH='1'));outputs[str(args)]=dict(code=p.returncode,stdout=p.stdout,stderr=p.stderr);assert p.returncode==0,(args,p.stderr);return p.stdout
# Check pre-existing user work before recording any new documentation/memory.
protected=json.loads((W/'protected-files.json').read_text())
for rel,h in protected.items():assert sha(R/rel)==h,rel
results['protected_existing_files']=len(protected)
git=json.loads((W/'git-before.json').read_text())
for name in ['HEAD','index']:assert sha(R/'.git'/name)==git[name],name
assert run(['git','--no-optional-locks','-c','core.fsmonitor=false','branch','--show-current'],R).strip()==git['branch']
results['git_head_index_branch_unchanged']='passed'
for b in [B,R/'tools/lufy-cli-go/bin/lufy-ai',R/'tools/lufy-cli-go/lufy-ai']:assert sha(b)==P['binary_sha256'],str(b)
results['three_executables_same_candidate']='passed'
outputs['version']=run([B,'version'])
# Exercise embedded source and repository source independently; both must remain current.
for source in [W,R]:run([B,'sync','--target',R,'--yes'],source)
report=json.loads(run([B,'verify','--target',R,'--deep','--json']));outputs['verify']=report
report=json.loads(run([B,'status','--target',R,'--json']));outputs['status']=report
assert report['schemaVersion']==3 and report['execution']=={'mode':'inline','delegation':'off'} and report['drifted']==0 and report['missing']==0
assert len(json.loads((R/'.lufy/managed-state/install-state.json').read_text())['assets'])==16
for op in P['operations']:
 if op['sha256'] is None and op['key'].startswith('project/'):assert not (R/op['key'].removeprefix('project/')).exists(),op['key']
results['embedded_and_repository_sync_without_retired_assets']='passed'
run([R/'scripts/install.sh','--target',R,'--dry-run'],R);results['strict_wrapper_current_local_binary']='passed'
outputs['skills']=json.loads(run([B,'skills','status','--target',R,'--json']))
outputs['memory']=json.loads(run([B,'memory','validate','--target',R,'--json']))
for rel,h in protected.items():assert sha(R/rel)==h,rel
for name in ['HEAD','index']:assert sha(R/'.git'/name)==git[name],name
results['protected_files_unchanged_after_sync']='passed'
backup=Path('/Users/adrianrojas/.local/share/lufy-ai/backups/inline-20261027');assert json.loads((backup/'journal.json').read_text())['phase']=='completed'
for name in ['snapshot.tar.gz','prepared.tar.gz','plan.json','replace_personal.py']:assert sha(backup/name)==sha(W/name),name
results['durable_backup_and_completed_journal']='passed'
(W/'live-result.json').write_text(json.dumps(results,ensure_ascii=False,indent=2)+'\n');(W/'live-commands.json').write_text(json.dumps(outputs,ensure_ascii=False,indent=2)+'\n');print(json.dumps(results,indent=2))
