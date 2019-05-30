all : fetch.so

fetch.so : fetch.c
	gcc fetch.c -shared -o fetch.so -L/usr/local/mysql/lib -lmysqlclient -g -fPIC
