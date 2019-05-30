# MySQL UDF (plugin) for Manticore Search / Sphinx Search
This plugin implements function fetch() which will go to your mysql table from which you (supposedly, but not required) have built the index and will fetch the title so you don't have
 to do it from your app. The output of the function can then be used in other functions, e.g. SNIPPET()

## Configure credentials to your MySQL
In fetch.c update mysql hostname, user, password and db name:
```
char * g_szHost         = "localhost";
char * g_szUser         = "test";
char * g_szPwd          = "";
char * g_szDb           = "test";
```

### Update the query which will be used to fetch the data from mysql:
```
char * g_szQuery        = "SELECT title FROM wiki WHERE id=%"PRIi64;
```

## Build the plugin
```
make
```

### Install the plugin
Log in to Manticore / Sphinx:
```
mysql -P9307 -h0
```
Create new function:
```
create function fetch returns string soname 'fetch.so';
```

## Here's an example of how it works (you can find all the sources in directory 'test'):
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
