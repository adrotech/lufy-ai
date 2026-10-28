from pathlib import Path
import subprocess,json,tarfile,hashlib,shutil,os
W=Path('/private/tmp/lufy-personal-inline-20261027');B=W/'lufy-ai';target=W/'rehearsal-resume-v4/project';binary=W/'upgrade-copy-v2/lufy-ai';binary.parent.mkdir();shutil.copy2(B,binary)
v='v0.0.0-inline-test2';release=W/'releases'/v;release.mkdir(parents=True);name='lufy-ai_'+v+'_darwin_arm64.tar.gz';archive=release/name
with tarfile.open(archive,'w:gz') as tar:tar.add(B,arcname='lufy-ai',recursive=False)
header=subprocess.check_output(['go','-C','tools/lufy-cli-go','run','./cmd/release-contract'],cwd='/private/tmp/lufy-inline-20261026',env=dict(os.environ,GOCACHE='/private/tmp/lufy-inline-go-cache'),text=True)
checksum=release/('lufy-ai_'+v+'_checksums.txt');checksum.write_text(header+hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+name+'\n')
subprocess.run(['shasum','-a','256','-c',checksum.name],cwd=release,check=True,capture_output=True)
def run(*args,code=0):
 p=subprocess.run([str(binary),*map(str,args)],cwd=W,text=True,capture_output=True);assert p.returncode==code,(args,p.stdout,p.stderr);return p.stdout
baseline=hashlib.sha256(binary.read_bytes()).hexdigest();run('upgrade','--to',v,'--base-url','file://'+str(W/'releases'));assert hashlib.sha256(binary.read_bytes()).hexdigest()==baseline
run('sync','--target',target,'--yes');run('verify','--target',target,'--deep','--quiet')
plan=json.loads((W/'plan.json').read_text());state=json.loads((target/'.lufy/managed-state/install-state.json').read_text());assert len(state['assets'])==16
for op in plan['operations']:
 if op['sha256'] is None and op['key'].startswith('project/'):assert not (target/op['key'].removeprefix('project/')).exists(),op['key']
checksum.write_text(checksum.read_text().split('\n',1)[1]);run('upgrade','--to',v,'--base-url','file://'+str(W/'releases'),code=1);assert hashlib.sha256(binary.read_bytes()).hexdigest()==baseline
result={'upgrade_current_generation':'passed','sync_does_not_reinstall_retired_assets':'passed','upgrade_previous_generation_blocked_without_binary_change':'passed','checksum_tool_accepts_contract_header':'passed'}
(W/'upgrade-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
