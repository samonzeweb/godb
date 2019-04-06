#!/bin/bash

# Allow at least 4 Go to docker on OSX, SQL Server is greedy

# Start containers
docker run --name mariadb -e 'MYSQL_ROOT_PASSWORD=NotSoStr0ngPassword' -p 3306:3306 -d mariadb:latest
docker run --name postgresql -e 'POSTGRES_PASSWORD=NotSoStr0ngPassword' -p 5432:5432 -d postgres:latest
docker run --name sqlserver -e 'ACCEPT_EULA=Y' -e 'SA_PASSWORD=NotSoStr0ngP@ssword' -p 1433:1433 -d microsoft/mssql-server-linux:latest

# Install dependencies (while containers are starting)
go mod download

# If the containers take too long time to start, the script
# wait for the given seconds args before using them.
if [ ! -z "$1" ] && [ $1 -gt 0 ]; then
	echo "Waiting $1 more seconds"
  sleep $1
fi

# MariaDB setup
docker exec -it mariadb mysql -uroot -pNotSoStr0ngPassword -hlocalhost -e "create database godb;"
export GODB_MYSQL="root:NotSoStr0ngPassword@/godb?parseTime=true"

# PostgreSQL setup
export GODB_POSTGRESQL="postgres://postgres:NotSoStr0ngPassword@localhost/postgres?sslmode=disable"

# SQL Server setup
docker exec -i sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P NotSoStr0ngP@ssword <<-EOF
  create database godb;
  go
  alter database godb set READ_COMMITTED_SNAPSHOT ON;
  go
  exit
EOF
export GODB_MSSQL="Server=localhost;Database=godb;User Id=sa;Password=NotSoStr0ngP@ssword"

# Let's test (without cache) !
go clean -testcache
go test -v ./...
testresult=$?

# Cleanup
docker stop mariadb postgresql sqlserver
docker rm mariadb postgresql sqlserver

echo ----------
if [ $testresult -eq 0 ]; then
  echo OK
else
  echo FAIL
fi
exit $testresult