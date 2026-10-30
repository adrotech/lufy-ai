"""Comprobaciones propias del revisor en SQLite en memoria; no ejecuta código de modelos."""
import csv,json,sqlite3
from pathlib import Path
root=Path('/private/tmp/lufy-inline-families-20261027-g/analyze-join/fixture')
db=sqlite3.connect(':memory:')
db.executescript('CREATE TABLE sales(sale_id INTEGER, amount INTEGER); CREATE TABLE lines(line_id INTEGER, sale_id INTEGER, product TEXT);')
with (root/'sales.csv').open() as f:db.executemany('INSERT INTO sales VALUES(:sale_id,:amount)',list(csv.DictReader(f)))
with (root/'lines.csv').open() as f:db.executemany('INSERT INTO lines VALUES(:line_id,:sale_id,:product)',list(csv.DictReader(f)))
queries={'join': 'SELECT SUM(s.amount) FROM sales s JOIN lines l ON s.sale_id=l.sale_id','sales':'SELECT SUM(amount) FROM sales','exists':'SELECT SUM(s.amount) FROM sales s WHERE EXISTS(SELECT 1 FROM lines l WHERE l.sale_id=s.sale_id)'}
observed={name:db.execute(query).fetchone()[0] for name,query in queries.items()};assert observed=={'join':200,'sales':100,'exists':100}
db.execute('INSERT INTO sales VALUES(2,100)');db.execute("INSERT INTO lines VALUES(3,2,'C')")
assert db.execute(queries['sales']).fetchone()[0]==200
assert db.execute(queries['exists']).fetchone()[0]==200
assert db.execute('SELECT SUM(DISTINCT amount) FROM sales').fetchone()[0]==100
print(json.dumps({'fixture':observed,'equal_amount_counterexample':'SUM(DISTINCT amount) pierde una venta distinta; EXISTS conserva 200','fixture_files_modified':False},ensure_ascii=False,indent=2))
