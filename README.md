# pqstream

pqstream is a program that streams changes out of a postgres database with the intent of populating other systems and enabling stream processing of data sets.

![ci status](https://circleci.com/gh/tmc/pqstream.svg?style=shield) ![report card](https://goreportcard.com/badge/github.com/tmc/pqstream)

## installation

```sh
go get -u github.com/tmc/pqstream/cmd/{pqs,pqsd}
```

## basic usage

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
a=> insert into notes values (DEFAULT, DEFAULT, 'here is an example note');
INSERT 0 1
a=> delete from notes where id=1;
DELETE 1
```

our client should now show our operations:
```sh
$ pqs
{"schema":"public","table":"notes","op":"INSERT","payload":{"created_at":"2017-09-04T01:11:34.65629","id":1,"notes":"here is an example note"}}
{"schema":"public","table":"notes","op":"DELETE","payload":{"created_at":"2017-09-04T01:11:34.65629","id":1,"notes":"here is an example note"}}
```
