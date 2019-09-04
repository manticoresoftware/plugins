# Manticore Search plugins and UDFs (user defined functions)

Manticore Search can be extended with help of plugins and custom functions (aka user defined functions or UDFs). Please consult with the [documentation](https://docs.manticoresearch.com/latest/html/extending.html) on how it works.

This repository contains the following open source plugins and UDFs:
  * [fetch()](./mysql_fetch) - to fetch a document by id from mysql (written in C++)
  * [curl()](./curl) - to download content from the web by url (written in Go)
