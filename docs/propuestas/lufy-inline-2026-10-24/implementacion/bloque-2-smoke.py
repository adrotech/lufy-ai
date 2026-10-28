import json, pathlib, subprocess, tempfile, hashlib, shutil
binary='/private/tmp/lufy-block2'
source='/private/tmp/lufy-inline-20261026'
root=pathlib.Path(tempfile.mkdtemp(prefix='lufy-inline-smoke-'))
target=root/'project'; target.mkdir()
def call(*args, code=0, cwd=source):
 p=subprocess.run([binary,*args],cwd=cwd,text=True,capture_output=True)
 if p.returncode!=code:raise AssertionError((args,p.returncode,p.stdout,p.stderr))
 return p.stdout

def snapshot(path):return {str(p.relative_to(path)): ('dir' if p.is_dir() else hashlib.sha256(p.read_bytes()).hexdigest()) for p in path.rglob('*')}
call('install','--target',str(target),'--yes','--delegation','off')
manifest=target/'.lufy/managed-state/install-state.json'; config=target/'.lufy/config/project.yaml'
st=json.loads(manifest.read_text())
assert st['schemaVersion']==3 and st['execution']=={'mode':'inline','delegation':'off'}
for a in st['assets']:
 assert not any(x in a['targetRel'] for x in ('.opencode','openspec','sdd-workflow','.codex/agents'))
assert set(json.loads((target/'.codex/hooks.json').read_text())['hooks'])=={'SessionStart'}
for cmd in ('info','status','verify'):
 report=json.loads(call(cmd,'--target',str(target),'--json'))
 assert report['execution']==st['execution'] and not any('methodologyByTier' in k for k in report)
state_before=manifest.read_bytes();call('install','--target',str(target),'--yes');assert manifest.read_bytes()==state_before
# drift must not mutate; explicit policy changes must persist even when no asset changed
config.write_text(config.read_text().replace('delegation: "off"','delegation: on_request')+'\ncustom_rule: preserve\n')
before=snapshot(target);call('sync','--target',str(target),'--yes',code=1);assert snapshot(target)==before
call('sync','--target',str(target),'--yes','--delegation','on_request')
assert json.loads(manifest.read_text())['execution']['delegation']=='on_request', 'sync did not persist the selected policy'
assert 'custom_rule: preserve' in config.read_text()
call('verify','--target',str(target),'--quiet')
for path,body in [('.lufy/config/project.yaml','schema_version: 1\ntool: codex\n'),('.lufy/config/project.yaml','schema_version: 2\nmethodology_by_tier: null\n'),('.lufy/managed-state/install-state.json','{"schemaVersion":2}')]:
 old=root/('old-'+str(len(list(root.iterdir()))));old.mkdir();p=old/path;p.parent.mkdir(parents=True);p.write_text(body);before=snapshot(old)
 call('install','--target',str(old),'--yes',code=1);assert snapshot(old)==before
# A checkout with only the current selected sources is recognized without historical hosts.
minimal=root/'inline-source';minimal.mkdir()
for rel in [a['targetRel'] for a in st['assets']]+['AGENTS.md','tools/lufy-cli-go/go.mod']:
 dest=minimal/rel;dest.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(pathlib.Path(source)/rel,dest)
local=root/'from-inline-source';local.mkdir()
output=call('install','--target',str(local),'--yes',cwd=minimal)
observed=next(line.removeprefix('Source root: ') for line in output.splitlines() if line.startswith('Source root: '))
assert pathlib.Path(observed).resolve()==minimal.resolve(), output
call('verify','--target',str(local),'--quiet',cwd=minimal)
# Embedded distribution, outside the source checkout
embedded=root/'embedded';embedded.mkdir()
output=call('install','--target',str(embedded),'--yes',cwd=root)
assert 'embedded' in output and json.loads((embedded/'.lufy/managed-state/install-state.json').read_text())['execution']['delegation']=='auto'
call('verify','--target',str(embedded),'--quiet',cwd=root)
call('uninstall','--target',str(target),'--yes');call('install','--target',str(target),'--yes');call('verify','--target',str(target),'--quiet')
print(json.dumps({'status':'passed','root':str(root),'checks':['fresh install','current schemas','selected assets','single orientation hook','info/status/verify','idempotency','drift without writes','policy persistence without asset changes','custom values preserved','old models rejected without writes','source discovery without historical hosts','embedded install','uninstall/reinstall']},indent=2))
