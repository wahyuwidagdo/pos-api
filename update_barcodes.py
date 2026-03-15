import psycopg2
import time

conn = psycopg2.connect("dbname=pos user=postgres password=postgres host=127.0.0.1 port=5432")
cur = conn.cursor()

cur.execute("SELECT id FROM products WHERE barcode IS NULL OR barcode = '';")
rows = cur.fetchall()

for row in rows:
    new_barcode = f"8{int(time.time() * 1000000)}{row[0]}"
    cur.execute("UPDATE products SET barcode = %s WHERE id = %s", (new_barcode, row[0]))

conn.commit()
cur.close()
conn.close()
