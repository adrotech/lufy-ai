from pathlib import Path
import subprocess,json,hashlib,tarfile,shutil,os
W=Path(__file__).resolve().parent;B=W/'lufy-ai';T=W/'lifecycle/project';T.mkdir(parents=True)
logs=[];results={}
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def run(*args,code=0,binary=B):
 p=subprocess.run([str(binary),*map(str,args)],cwd=W,text=True,capture_output=True)
 logs.append({'command':[str(binary),*map(str,args)],'code':p.returncode,'stdout':p.stdout,'stderr':p.stderr})
 (W/'lifecycle-commands.json').write_text(json.dumps(logs,ensure_ascii=False,indent=2)+'\n')
 assert p.returncode==code,(args,p.returncode,p.stdout,p.stderr)
 return p.stdout
foreign={'.codex/local-notes.md':'personal\n','.lufy/notes/decision.md':'keep\n','AGENTS.md':'# Instrucciones propias\n\n','docs/user.md':'contenido propio\n'}
for rel,body in foreign.items():p=T/rel;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(body)
before={str(p.relative_to(T)):sha(p) for p in T.rglob('*') if p.is_file()}
run('install','--target',T,'--dry-run','--yes')
assert before=={str(p.relative_to(T)):sha(p) for p in T.rglob('*') if p.is_file()}
results['fresh_preview_read_only']='passed'
run('install','--target',T,'--yes');run('verify','--target',T,'--deep','--quiet')
state=json.loads((T/'.lufy/managed-state/install-state.json').read_text());assert state['schemaVersion']==4 and len(state['assets'])==16 and 'methodology' not in json.dumps(state)
managed={a['targetRel']:sha(T/a['targetRel']) for a in state['assets']}
run('install','--target',T,'--yes');run('sync','--target',T,'--yes')
assert managed=={rel:sha(T/rel) for rel in managed}
for rel in ['.opencode','openspec','.lufy/sdd','.codex/agents']:assert not (T/rel).exists()
results['fresh_install_repeat_and_sync']='passed'
run('backup','--target',T)
backup=sorted((T/'.lufy/managed-state/backups').glob('*/manifest.json'))[-1].parent
m=json.loads((backup/'manifest.json').read_text());assert m['schemaVersion']==2 and m['runtimeStateSchema']==4
assert '.lufy/managed-state/install-state.json' in [f['path'] for f in m['files']]
run('uninstall','--target',T,'--yes')
for rel in managed:assert not (T/rel).exists(),rel
for rel,body in foreign.items():
 actual=(T/rel).read_text()
 assert (actual.strip()==body.strip() if rel=='AGENTS.md' else actual==body),rel
results['uninstall_preserves_foreign_files']='passed'
run('restore','--target',T,'--backup',backup,'--yes');run('verify','--target',T,'--deep','--quiet')
assert managed=={rel:sha(T/rel) for rel in managed}
assert (T/'AGENTS.md').read_text().count('<!-- LUFY:BEGIN codex-harness -->')==1
results['current_backup_restores_installation']='passed'
old=W/'old-config';(old/'.lufy/config').mkdir(parents=True);(old/'.lufy/config/project.yaml').write_text('schema_version: 1\ntool: codex\n')
prior={str(p.relative_to(old)):sha(p) for p in old.rglob('*') if p.is_file()}
run('install','--target',old,'--yes',code=1)
assert prior=={str(p.relative_to(old)):sha(p) for p in old.rglob('*') if p.is_file()}
results['retired_config_rejected_without_writes']='passed'
# Upgrade sólo sobre una copia, mediante artifacts locales de la generación actual.
binary=W/'upgrade-copy/lufy-ai';binary.parent.mkdir();shutil.copy2(B,binary)
v='v0.0.0-inline-b6';release=W/'releases'/v;release.mkdir(parents=True);name='lufy-ai_'+v+'_darwin_arm64.tar.gz';archive=release/name
with tarfile.open(archive,'w:gz') as tar:tar.add(B,arcname='lufy-ai',recursive=False)
header=run('version','--runtime-contract');checksum=release/('lufy-ai_'+v+'_checksums.txt');checksum.write_text(header+sha(archive)+'  '+name+'\n')
subprocess.run(['shasum','-a','256','-c',checksum.name],cwd=release,check=True,capture_output=True)
run('upgrade','--to',v,'--base-url','file://'+str(W/'releases'),binary=binary);assert sha(binary)==sha(B)
run('sync','--target',T,'--yes',binary=binary);run('verify','--target',T,'--deep','--quiet',binary=binary)
results['upgrade_current_and_sync']='passed'
checksum.write_text(checksum.read_text().replace('state=4 backup=2','state=3'))
run('upgrade','--to',v,'--base-url','file://'+str(W/'releases'),code=1,binary=binary);assert sha(binary)==sha(B)
results['upgrade_old_generation_blocked']='passed'
(W/'upgrade-result.json').write_text(json.dumps({k:v for k,v in results.items() if k.startswith('upgrade')},indent=2)+'\n')
(W/'lifecycle-result.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results,indent=2))
