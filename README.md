# godb - a simple Go ORM

godb is a simple Go ORM. It contains a simple SQL query builder and manages mapping between SQL and structs.

Initially godb was a learning project. The purpose was to learn Go by doing real and usable stuff. But it could be useful for somebody else.

WARNING : it is still a young project and the public API could change.

## Features

* Row queries builder.
* Maps structs with tables.
* Manages nested structs.
* Manages direct SELECT, INSERT, UPDATE and DELETE on structs and slices.
* Logs SQL queries and durations.
* Manage prepared statements and cache it during transactions.
* Could by used with
  * SQLite
  * PostgreSQL
  * MySQL / MariaDB
  * MS SQL Server
  * other compatible database if you write an adapter.

godb does not manage relationship.

## Install

TODO

## Tests

godb tests use GoConvey and SQLite :

```
go get github.com/smartystreets/goconvey
go get github.com/mattn/go-sqlite3
```

SQLite tests are done with in memory database.

You can run tests with others databases, see below.

To run tests, go into the godb directory and executes `go test ./...`


### PostgreSQL

Install the driver and set the `GODB_POSTGRESQL` environment variable with the PostgreSQL connection string.

```
go get github.com/lib/pq
GODB_POSTGRESQL="your connection string" go test ./...
```

### MySQL / MariaDB

Install the driver and set the `GODB_MYSQL` environment variable with the MySQL connection string.

```
go get github.com/go-sql-driver/mysql
GODB_MYSQL="your connection string" go test ./...
```

### MS SQL Server

Install the driver and set the `GODB_MSSQL` environment variable with the SQL Server connection string.

```
go get github.com/denisenkom/go-mssqldb
GODB_MSSQL="your connection string" go test ./...
```

## Examples

TODO

# Licence

Released under the MIT License, see LICENSE.txt for more informations.
