"""Revisión puntual del candidato instalado; conserva el snapshot conjunto original."""
from pathlib import Path
import json,os,hashlib,shutil,subprocess,importlib.util
W=Path(__file__).resolve().parent;D=Path('/Users/adrianrojas/.local/share/lufy-ai/backups/retirement-20261028');R=Path('/Users/adrianrojas/Desktop/projects/lufy-ai');G=Path('/Users/adrianrojas/.local/bin/lufy-ai')
plan=json.loads((W/'plan.json').read_text());before=json.loads((W/'revision-preimages.json').read_text());journal=json.loads((D/'journal.json').read_text());assert journal['phase']=='completed'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def actual(key):return G if key=='global/lufy-ai' else R/key.removeprefix('project/')
def atomic(p,body,mode):
 q=p.with_name(p.name+'.revision-new');q.write_bytes(body);os.chmod(q,mode);os.replace(q,p)
for name in ['rehearsal-result.json','lifecycle-result.json']:
 data=json.loads((W/name).read_text());assert data and all(v=='passed' for v in data.values())
for op in plan['operations']:
 p=actual(op['key']);assert p.resolve()==p and sha(p)==before[op['key']]['sha256'],('drift',str(p))
lock=R/'.lufy/managed-state/.lock';fd=os.open(lock,os.O_CREAT|os.O_EXCL|os.O_WRONLY,0o600);os.write(fd,json.dumps({'operation':'inline-retirement-final','pid':os.getpid()}).encode())
try:
 revision=D/'revision-1';revision.mkdir(mode=0o700)
 for name in ['plan.json','prepared.tar.gz','journal.json','replace_personal.py']:shutil.copy2(D/name,revision/name)
 # Registrar ambos postimages exactos antes de escribir permite recuperar una revisión interrumpida.
 journal.setdefault('accepted_postimages',{})
 for op in plan['operations']:
  journal['accepted_postimages'][op['key']]=list({before[op['key']]['sha256'],op['sha256']})
 atomic(D/'journal.json',(json.dumps(journal,ensure_ascii=False,indent=2)+'\n').encode(),0o600)
 for name in ['plan.json','prepared.tar.gz','replace_personal.py','rehearsal-result.json','lifecycle-result.json','upgrade-result.json','revision-preimages.json','revision-build.json','update_live.py']:
  atomic(D/name,(W/name).read_bytes(),0o600)
 try:
  changed=0
  for op in plan['operations']:
   p=actual(op['key']);assert sha(p)==before[op['key']]['sha256'],('drift bajo lock',str(p))
   q=W/'payload'/op['key'];assert sha(q)==op['sha256']
   if sha(p)!=op['sha256']:atomic(p,q.read_bytes(),op['mode']);changed+=1
  for args in [('verify','--target',str(R),'--deep','--quiet'),('memory','validate','--target',str(R),'--json')]:
   proc=subprocess.run([str(G),*args],cwd=W,capture_output=True,text=True);assert proc.returncode==0,(args,proc.stderr)
  for op in plan['operations']:journal['postimages'][op['key']]=sha(actual(op['key']))
  atomic(D/'journal.json',(json.dumps(journal,ensure_ascii=False,indent=2)+'\n').encode(),0o600)
  result={'status':'completed','revised_files':changed,'binary_sha256':sha(G),'original_joint_snapshot_preserved':True,'interrupted_revision_recoverable_by_exact_postimages':True}
  atomic(D/'revision-result.json',(json.dumps(result,indent=2)+'\n').encode(),0o600);(W/'revision-result.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
 except BaseException:
  spec=importlib.util.spec_from_file_location('recovery',W/'replace_personal.py');mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod);mod.Replacement(R,live=True).restore();raise
finally:
 os.close(fd);lock.unlink()
