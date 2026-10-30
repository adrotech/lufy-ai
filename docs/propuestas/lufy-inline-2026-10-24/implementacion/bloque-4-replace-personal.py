"""Procedimiento puntual del snapshot personal; no es un migrador del runtime."""
from pathlib import Path
import os,sys,json,hashlib,shutil,subprocess,tarfile,stat
W=Path(__file__).resolve().parent;P=json.loads((W/'plan.json').read_text());REAL=Path(P['target']);GLOBAL=Path(P['global_binary']);BIN=W/'lufy-ai';DATA=['.lufy/context','.lufy/skill-registry.json']
assert str(REAL)=='/Users/adrianrojas/Desktop/projects/lufy-ai' and str(GLOBAL)=='/Users/adrianrojas/.local/bin/lufy-ai'
NAME='inline-20261027';REAL_BACKUP=Path('/Users/adrianrojas/.local/share/lufy-ai/backups')/NAME

def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def write_json(path,data):
 path.parent.mkdir(parents=True,exist_ok=True);tmp=path.with_name(path.name+'.tmp');tmp.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n');os.replace(tmp,path)
def safe(key):
 assert key.startswith(('project/','global/')) and '..' not in Path(key).parts,key
 return key
class Replacement:
 def __init__(self,base,live=False):
  self.live=live;self.base=Path(base)
  if live:assert self.base==REAL
  else:assert str(self.base).startswith(str(W)+'/rehearsal-')
  self.root=REAL if live else self.base/'project'
  self.backup=REAL_BACKUP if live else self.base/'journal'
  self.journal=self.backup/'journal.json';self.saved=self.backup/'physical';self.state={'phase':'prepared','data_saved':[],'started':[],'done':[]}
  if self.journal.exists():self.state=json.loads(self.journal.read_text())
  self.state.setdefault('postimages',{})
 def path(self,key):
  safe(key)
  if key.startswith('project/'):return self.root/key.removeprefix('project/')
  return GLOBAL if self.live else self.base/key
 def payload(self,op):
  if op['sha256'] is None:return None
  q=W/'payload'/op['key'];assert sha(q)==op['sha256'],('payload changed',op['key'])
  body=q.read_bytes()
  if op['key']=='project/.lufy/managed-state/install-state.json':
   data=json.loads(body);data['targetRoot']=str(self.root.resolve());body=(json.dumps(data,ensure_ascii=False,indent=2)+'\n').encode()
  return body
 def current(self,p):
  if p.is_symlink() or p.exists() and not p.is_file():raise RuntimeError('destino no regular: '+str(p))
  return sha(p) if p.exists() else None
 def expected(self,op):
  b=self.payload(op);return hashlib.sha256(b).hexdigest() if b is not None else None
 def persist(self):write_json(self.journal,self.state)
 def preflight(self):
  assert sha(BIN)==P['binary_sha256'],'binario preparado cambió'
  for op in P['operations']:
   key=op['key'];p=self.path(key);assert p.resolve()==p,('ruta simbólica',key)
   value=self.current(p);old=P['preimages'][key];before=old['sha256'] if old else None;after=self.expected(op)
   if key not in self.state['started']:
    if value!=before:raise RuntimeError('drift; no se sustituyó la instalación: '+key)
   elif value not in (before,after) and not (value is None and (self.saved/key).is_file()):
    raise RuntimeError('conflicto en sustitución parcial: '+key)
 def command(self,*args):
  # No HOME replacement, credentials, network release or model invocation.
  env=dict(os.environ,GIT_OPTIONAL_LOCKS='0',GIT_NO_LAZY_FETCH='1')
  p=subprocess.run([str(BIN),*map(str,args)],cwd=W,env=env,text=True,capture_output=True)
  log=self.backup/'commands.jsonl'
  with log.open('a') as f:f.write(json.dumps({'args':list(map(str,args)),'code':p.returncode,'stdout':p.stdout,'stderr':p.stderr},ensure_ascii=False)+'\n')
  if p.returncode:raise RuntimeError('falló comprobación '+str(args)+': '+p.stderr[:300])
  return p.stdout
 def save_data(self):
  for rel in DATA:
   if rel in self.state['data_saved']:continue
   p=self.root/rel;q=self.saved/'data'/rel
   self.state['phase']='saving_projections';self.persist()
   if p.exists():q.parent.mkdir(parents=True,exist_ok=True);os.replace(p,q)
   self.state['data_saved'].append(rel);self.persist()
 def verify(self):
  self.command('verify','--target',self.root,'--deep','--quiet')
  self.command('memory','validate','--target',self.root,'--json')
  body=(self.root/'AGENTS.md').read_text();assert body.count('<!-- LUFY:BEGIN codex-harness -->')==1
  report=json.loads(self.command('status','--target',self.root,'--json'));assert report['schemaVersion']==3 and report['execution']=={'mode':'inline','delegation':'off'}
  assert report['drifted']==0 and report['missing']==0
 def apply(self,fail_after=None):
  if self.state['phase']=='completed':self.verify();return 'already_completed'
  if self.state['phase']=='restored':raise RuntimeError('snapshot ya restaurado; no reutilizar esta instancia')
  self.preflight();self.backup.mkdir(parents=True,exist_ok=True);os.chmod(self.backup,0o700)
  self.state['phase']='applying';self.persist();applied=0
  for op in P['operations']:
   key=op['key'];p=self.path(key);after=self.expected(op)
   if key in self.state['done']:assert self.current(p)==after;continue
   if key not in self.state['started']:self.state['started'].append(key);self.persist()
   saved=self.saved/key
   if self.current(p)==after and saved.exists():
    self.state['done'].append(key);self.persist();continue
   if p.exists() and not saved.exists():
    saved.parent.mkdir(parents=True,exist_ok=True);assert saved.parent.stat().st_dev==p.parent.stat().st_dev,'backup físico requiere el mismo filesystem';os.replace(p,saved)
   body=self.payload(op)
   if body is not None:
    p.parent.mkdir(parents=True,exist_ok=True);tmp=p.with_name(p.name+'.inline-new');tmp.write_bytes(body);os.chmod(tmp,op['mode']);os.replace(tmp,p)
   self.state['postimages'][key]=self.current(p);self.state['done'].append(key);self.persist();applied+=1
   if fail_after and applied==fail_after:
    self.state['phase']='interrupted';self.persist();raise RuntimeError('interrupción de ensayo')
  self.save_data();self.state['phase']='validating';self.persist()
  # Prune only empty retired skill/agent folders, never user documents.
  for rel in ['.agents/skills','.codex/agents','.codex/instructions']:
   folder=self.root/rel
   if folder.exists():
    for p in sorted(folder.rglob('*'),key=lambda p:len(p.parts),reverse=True):
     if p.is_dir():
      try:p.rmdir()
      except OSError:pass
    if rel!='.agents/skills':
     try:folder.rmdir()
     except OSError:pass
  self.command('skills','ensure','--target',self.root,'--json')
  self.command('context','build','--target',self.root,'--json')
  self.verify();self.state['phase']='completed';self.persist();return 'completed'
 def restore(self):
  if self.state['phase']=='restored':return 'already_restored'
  for key in self.state['started']:
   p=self.path(key);current=self.current(p);old=P['preimages'][key];before=old['sha256'] if old else None;after=self.state['postimages'][key] if key in self.state['postimages'] else self.expected(next(o for o in P['operations'] if o['key']==key))
   if current not in (before,after,None):raise RuntimeError('drift posterior; no se sobrescribe al restaurar: '+key)
  # Restore physical preimages to preserve file attributes and modes as well as bytes.
  for key in reversed(self.state['started']):
   p=self.path(key);saved=self.saved/key
   if saved.exists():
    if p.exists():p.unlink()
    p.parent.mkdir(parents=True,exist_ok=True);os.replace(saved,p)
   elif P['preimages'][key] is None and p.exists():p.unlink()
  for rel in self.state['data_saved']:
   p=self.root/rel;q=self.saved/'data'/rel
   if p.is_dir():shutil.rmtree(p)
   elif p.exists():p.unlink()
   if q.exists():p.parent.mkdir(parents=True,exist_ok=True);os.replace(q,p)
  self.state['phase']='restored';self.persist();return 'restored'

def clone(name):
 base=W/('rehearsal-'+name);assert not base.exists();base.mkdir()
 with tarfile.open(W/'snapshot.tar.gz') as tar:
  for member in tar.getmembers():assert not member.name.startswith('/') and '..' not in Path(member.name).parts
  tar.extractall(base)
 return base

def snapshot_files(base):
 expected=json.loads((W/'snapshot-files.json').read_text());result={}
 for key in expected:
  p=base/key;result[key]={'sha256':sha(p),'mode':stat.S_IMODE(p.stat().st_mode)} if p.is_file() else None
 return result

def rehearse():
 results={};expected=json.loads((W/'snapshot-files.json').read_text())
 conflict=clone('conflict-v4');r=Replacement(conflict);key=P['operations'][0]['key'];p=r.path(key);p.write_bytes(p.read_bytes()+b'\nDRIFT\n');before=snapshot_files(conflict)
 try:r.apply();raise AssertionError('drift accepted')
 except RuntimeError as err:assert 'drift' in str(err)
 assert before==snapshot_files(conflict);assert not r.journal.exists();results['drift_blocks_before_writes']='passed'
 rollback=clone('rollback-v4');r=Replacement(rollback)
 try:r.apply(fail_after=7);raise AssertionError('interruption missing')
 except RuntimeError as err:assert 'interrupción' in str(err)
 assert r.state['phase']=='interrupted';r=Replacement(rollback);r.restore();assert snapshot_files(rollback)==expected
 old=subprocess.run([str(rollback/'global/lufy-ai'),'version'],text=True,capture_output=True);assert old.returncode==0 and 'v0.6.25' in old.stdout
 restored=subprocess.run([str(rollback/'global/lufy-ai'),'status','--target',str(rollback/'project'),'--json'],cwd=W,text=True,capture_output=True);assert restored.returncode==0 and json.loads(restored.stdout)['schemaVersion']==2
 results['joint_restore_after_interruption']='passed'
 resume=clone('resume-v4');r=Replacement(resume)
 try:r.apply(fail_after=11)
 except RuntimeError as err:assert 'interrupción' in str(err)
 r=Replacement(resume);assert r.apply()=='completed';first=snapshot_files(resume);assert r.apply()=='already_completed';assert first==snapshot_files(resume)
 r.command('sync','--target',r.root,'--yes');r.verify()
 for op in P['operations']:
  if op['sha256'] is None:assert not r.path(op['key']).exists(),op['key']
 results['sync_without_retired_assets']='passed'
 assert sha(resume/'global/lufy-ai')==P['binary_sha256'];results['resume_without_duplicates']='passed';results['deep_verify_and_memory']='passed'
 # A later user edit blocks restoration rather than being overwritten.
 key='project/.codex/lufy-agent-mapping.md';p=r.path(key);body=p.read_bytes();p.write_bytes(body+b'\nUSER-WORK\n');before=snapshot_files(resume)
 try:r.restore();raise AssertionError('new work overwritten')
 except RuntimeError as err:assert 'drift posterior' in str(err)
 assert before==snapshot_files(resume);p.write_bytes(body);results['later_drift_blocks_restore']='passed'
 (W/'rehearsal-result.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results,indent=2))

def resources():
 if not BIN.exists() or not (W/'payload').exists():
  with tarfile.open(W/'prepared.tar.gz') as tar:
   for m in tar.getmembers():assert '..' not in Path(m.name).parts and not m.name.startswith('/') and (m.name in ('lufy-ai','payload') or m.name.startswith('payload/'))
   tar.extractall(W)
 assert sha(BIN)==P['binary_sha256']
def install_live():
 resources()
 result=json.loads((W/'rehearsal-result.json').read_text());assert all(v=='passed' for v in result.values())
 upgrade=json.loads((W/'upgrade-result.json').read_text());assert all(v=='passed' for v in upgrade.values())
 r=Replacement(REAL,live=True);r.preflight()
 lock=REAL/'.lufy/managed-state/.lock'
 if lock.exists():
  try:
   prior=json.loads(lock.read_text());assert prior['operation']=='lufy-inline-personal-20261027';os.kill(prior['pid'],0)
  except ProcessLookupError:lock.unlink()
  except (ValueError,KeyError,AssertionError):raise RuntimeError('lock ajeno; no se modifica')
  else:raise RuntimeError('sustitución ya está en ejecución')
 fd=os.open(lock,os.O_CREAT|os.O_EXCL|os.O_WRONLY,0o600)
 os.write(fd,json.dumps({'operation':'lufy-inline-personal-20261027','pid':os.getpid()}).encode())
 try:
  r.preflight()
  REAL_BACKUP.mkdir(parents=True,mode=0o700,exist_ok=True);os.chmod(REAL_BACKUP,0o700)
  for name in ['snapshot.tar.gz','snapshot-files.json','snapshot-dirs.json','prepared.tar.gz','plan.json','preview.md','rehearsal-result.json','upgrade-result.json','replace_personal.py']:
   src=W/name;dst=REAL_BACKUP/name
   if src!=dst and (not dst.exists() or sha(src)!=sha(dst)):
    temp=dst.with_name(dst.name+'.new');shutil.copy2(src,temp);os.replace(temp,dst)
  try:print(json.dumps({'live':r.apply(),'backup':str(REAL_BACKUP)},indent=2))
  except BaseException:
   r=Replacement(REAL,live=True);r.restore();raise
 finally:
  os.close(fd);lock.unlink()
if __name__=='__main__':
 mode=sys.argv[1] if len(sys.argv)>1 else ''
 if mode=='rehearse':rehearse()
 elif mode=='apply':install_live()
 elif mode=='restore-live':
  resources();print(Replacement(REAL,live=True).restore())
 else:raise SystemExit('usar rehearse, apply o restore-live; sólo este snapshot personal')
