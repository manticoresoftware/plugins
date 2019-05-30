# fetch() function example
manticore.conf - minimal config which can be used to demonstrate how it works
mysqldump.sql - MySQL dump

## load the dump
Change hostname, login, password to yours:
```
mysql -h 127.0.0.1 -u test test < mysqldump.sql
```

## prepare config
In manticore.conf change to yours:
```
  sql_host = localhost
  sql_user = test
  sql_pass =
  sql_db = test
```

## build index
```
indexer -c manticore.conf --all
```

## run searchd
```
searchd -c manticore.conf
```

## Log in to Manticore / Sphinx:
```
mysql -P9307 -h0
```
Create new function:
```
mysql> create function fetch returns string soname 'fetch.so';
```

## Run the query:
```
mysql> select id, snippet(fetch(id), 'ghi') snippet from idx where match('ghi');
+------+--------------------+
| id   | snippet            |
+------+--------------------+
|    1 | abc def <b>ghi</b> |
|    2 | <b>ghi</b> jkl mno |
+------+--------------------+
2 rows in set (0.00 sec)
```
