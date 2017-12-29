# pqstream

pqstream is a program that streams changes out of a postgres database with the intent of populating other systems and enabling stream processing of data sets.

[![ci status](https://circleci.com/gh/tmc/pqstream.svg?style=shield)](https://circleci.com/gh/tmc/workflows/pqstream/tree/master) 
[![go report card](https://goreportcard.com/badge/github.com/tmc/pqstream)](https://goreportcard.com/report/github.com/tmc/pqstream)
[![coverage](https://codecov.io/gh/tmc/pqstream/branch/master/graph/badge.svg)](https://codecov.io/gh/tmc/pqstream)

## installation

```sh
$ go get -u github.com/tmc/pqstream/cmd/{pqs,pqsd}
```

## basic usage

create an example database:

```sh
$ createdb dbname
# echo "create table notes (id serial, created_at timestamp, note text)" | psql dbname
```

connect the agent:

```sh
$ pqsd -connect postgresql://user:pass@host/dbname
```

connect the cli:
```sh
$ pqs
```

at this point you will see streams of database operations rendered to stdout:


(in a psql shell):

```sql
dbname=# insert into notes values (default, default, 'here is a sample note');
INSERT 0 1
dbname=# insert into notes values (default, default, 'here is a sample note');
INSERT 0 1
dbname=# update notes set note = 'here is an updated note' where id=1;
UPDATE 1
dbname=# delete from notes where id = 1;
DELETE 1
dbname=#
```

our client should now show our operations:
```sh
$ pqs
{"schema":"public","table":"notes","op":"INSERT","id":"1","payload":{"created_at":null,"id":1,"note":"here is a sample note"}}
{"schema":"public","table":"notes","op":"INSERT","id":"2","payload":{"created_at":null,"id":2,"note":"here is a sample note"}}
{"schema":"public","table":"notes","op":"UPDATE","id":"1","payload":{"created_at":null,"id":1,"note":"here is an updated note"},"changes":{"note":"here is a sample note"}}
{"schema":"public","table":"notes","op":"DELETE","id":"1","payload":{"created_at":null,"id":1,"note":"here is an updated note"}}
```


## field redaction

If there's a need to prevent sensitive fields (i.e. PII) from being exported the `redactions` flag can be used with `pqsd`:


```sh
$ pqsd -connect postgresql://user:pass@host/dbname -redactions='{"public":{"users":["first_name","last_name","email"]}}'
```

The `redactions` is encoded in [JSON](http://json.org/) and conforms to the following layout: 
``` json
'{"schema":{"table":["field1","field2"]}}'`
```
