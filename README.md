# pqstream

pqstream is a program that streams changes out of a postgres database with the intent of populating other systems and enabling stream processing of data sets.

## installation

```sh
go get -u github.com/tmc/pqstream/cmd/{pqs,pqsd}
```

## basic usage

connect the agent:

```sh
pqsd -connect postgresql://user:pass@host/dbname
```

connect the cli:
```sh
pqs
```

at this point you will see streams of database operations rendered to stdout
