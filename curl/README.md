## Introduction

Curl is demo UDF function for Manticore written in go lang.

Manticore Search is an open source search server designed to be fast, scalable and with powerful and accurate full-text search capabilities. It is a fork of popular search engine Sphinx.

## Installation
First, clone or download the repo.
Go into folder 'curl', and run

```
go build -buildmode=c-shared -o goudf.so .
```

### Plugging

Place built `goudf.so` somewhere in the system. Add param `plugin_dir = /path/to/udf/dir` to your
config into 'common' section.

### Usage

```mysql
  CREATE FUNCTION curl RETURNS STRING SONAME 'goudf.so';
  SELECT curl('https://yandex.ru/robots.txt');
```

## Source
Example is written on Go with CGo package. Base file `udfhelpers.go` contains C binding skeleton of the
library and some helper functions for converting/transferring data between go and C.

`curl.go` contains all the rest necessary for UDF to be functional,
that is 'curl_ini' and 'curl' functions.
