import pathlib, tempfile, subprocess, json, hashlib, os
binary='/private/tmp/lufy-block3'; source=pathlib.Path('/private/tmp/lufy-inline-20261026')
root=pathlib.Path(tempfile.mkdtemp(prefix='lufy-evidence-smoke-'));target=root/'project';target.mkdir()
def call(*args,code=0,data=None,cwd=source):
 p=subprocess.run([binary,*args],cwd=cwd,text=True,capture_output=True,input=data)
 assert p.returncode==code,(args,p.returncode,p.stdout,p.stderr)
 return p.stdout

def snap(path):return {str(p.relative_to(path)):('directory' if p.is_dir() else hashlib.sha256(p.read_bytes()).hexdigest()) for p in path.rglob('*')}
call('install','--target',str(target),'--yes','--delegation','off');call('verify','--target',str(target),'--deep','--quiet')
assert (target/'.agents/skills/lufy-check/SKILL.md').is_file() and (target/'.lufy/contracts/evidence.md').is_file()
assert not (target/'.lufy/contracts/result-contract.md').exists()
start=json.loads(call('lifecycle','codex',data=json.dumps({'cwd':str(target),'hook_event_name':'SessionStart','session_id':'PRIVATE-SESSION','last_assistant_message':'PRIVATE-MESSAGE'})))
assert start['continue'] and 'PRIVATE-' not in json.dumps(start) and not (target/'.lufy/runtime').exists()
before=snap(target)
for name in ('SubagentStart','SubagentStop','Stop','SessionEnd'):
 out=json.loads(call('lifecycle','codex',data=json.dumps({'cwd':str(target),'hook_event_name':name,'last_assistant_message':'PRIVATE-CANARY'})))
 assert out['continue'] and out['suppressOutput'] and snap(target)==before
# A writing result: editorial comparison and literal factual value observed by the fixture.
original='El plazo es de 7 días.';rewritten='El plazo establecido es de 7 días.'
assert '7 días' in original and '7 días' in rewritten and rewritten.endswith('.')
checks={'checks':[{'name':'editorial','result':'passed','evidence':'Borrador comparado con la solicitud; gramática y sentido conservados.'},{'name':'factual','result':'passed','evidence':'La cifra literal 7 días se conserva.'}]}
evidence=root/'writing.json';evidence.write_text(json.dumps(checks))
assert json.loads(call('evidence','inspect','--file',str(evidence)))['conclusion']=='supported'
args=['--target',str(target),'--run','run-smoke','--review','review-one','--json']
call('review','start',*args)
first=json.loads(call('review','complete',*args,'--result','passed','--evidence',str(evidence)))
second=json.loads(call('review','complete',*args,'--result','passed','--evidence',str(evidence)))
assert first['status']=='recorded' and second['status']=='duplicate_noop'
assert 'Borrador comparado' not in json.dumps(snap(target))
critical=root/'critical.json';critical.write_text(json.dumps(dict(checks,findings=[{'severity':'critical','summary':'PRIVATE-CRITICAL'}])))
assert json.loads(call('evidence','inspect','--file',str(critical)))['conclusion']=='failed'
before=snap(target);call('review','complete',*args,'--result','passed','--evidence',str(critical),code=4);assert snap(target)==before
for result in ('inconclusive','not_applicable'):
 record={'checks':[{'name':'build','result':result,('reason' if result=='not_applicable' else 'evidence'):'No corresponde a esta redacción' if result=='not_applicable' else 'No terminó la comprobación'}]}
 path=root/(result+'.json');path.write_text(json.dumps(record));assert json.loads(call('evidence','inspect','--file',str(path)))['conclusion']==result
 call('review','complete',*args,'--result','passed','--evidence',str(path),code=4)
call('result','normalize','--file',str(evidence),code=2)
# Only explicit synthetic commits, in an isolated fixture; candidate verification stays read-only.
repo=root/'candidate';repo.mkdir()
def git(*args):
 env={k:v for k,v in os.environ.items() if not k.upper().startswith('GIT_')};env.update({'GIT_CONFIG_GLOBAL':os.devnull,'GIT_CONFIG_SYSTEM':os.devnull,'GIT_AUTHOR_NAME':'Synthetic','GIT_AUTHOR_EMAIL':'synthetic@example.invalid','GIT_COMMITTER_NAME':'Synthetic','GIT_COMMITTER_EMAIL':'synthetic@example.invalid'})
 return subprocess.check_output(['git','-C',str(repo),'-c','core.hooksPath='+os.devnull,*args],env=env,text=True,stderr=subprocess.DEVNULL).strip()
git('init','--initial-branch=fixture')
for name in ('work.md','evidence.md','delivery.md'):
 dest=repo/'.lufy/contracts'/name;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes((source/'.lufy/contracts'/name).read_bytes())
(repo/'document.txt').write_text(original);git('add','.');git('commit','-m','baseline');base=git('rev-parse','HEAD')
(repo/'document.txt').write_text(rewritten);git('add','document.txt');git('commit','-m','candidate')
before=snap(repo);binding=json.loads(call('candidate','capture','--target',str(repo),'--base',base,'--evidence',str(evidence)));assert snap(repo)==before
assert binding['schema_version']=='candidate-evidence/v2' and 'role' not in binding
sidecar=root/'binding.json';sidecar.write_text(json.dumps(binding))
verify=['candidate','verify','--target',str(repo),'--base',base,'--evidence',str(evidence),'--binding',str(sidecar),'--json']
call(*verify);git('commit','--allow-empty','-m','new identity');before=snap(repo);call(*verify,code=4);assert snap(repo)==before
# No body, finding text or transcript is persisted in the explicit ledger.
for path in (target/'.lufy/runtime').rglob('*'):
 if path.is_file():assert all(x not in path.read_text() for x in ('PRIVATE-','Borrador comparado','gramática y sentido'))
print(json.dumps({'status':'passed','root':str(root),'checks':['inline install and selected verification capability','deep verify','orientation without automatic ledger','ignored lifecycle without writes','editorial and factual observations','review idempotency','critical finding blocks success without writes','inconclusive is not success','not applicable is not success','retired result command rejected','candidate evidence bound without role','new commit invalidates evidence','Git observation is read only','ledger excludes raw content']},indent=2))
