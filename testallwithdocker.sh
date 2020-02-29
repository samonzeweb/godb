#!/bin/bash

# Allow at least 4 Go to docker on OSX, SQL Server is greedy

MARIADB_PASSWORD=NotSoStr0ngPassword
export GODB_MYSQL="root:$MARIADB_PASSWORD@/godb?parseTime=true"

POSTGRESQL_PASSWORD=NotSoStr0ngPassword
export GODB_POSTGRESQL="postgres://postgres:$POSTGRESQL_PASSWORD@localhost/postgres?sslmode=disable"

SQLSERVER_PASSWORD=NotSoStr0ngP@ssword
export GODB_MSSQL="Server=localhost;Database=godb;User Id=sa;Password=$SQLSERVER_PASSWORD"

STARTLOOP_SLEEP=2
STARTLOOP_MAXITERATIONS=10

star_containers() {
    docker run --name mariadb -e "MYSQL_ROOT_PASSWORD=$MARIADB_PASSWORD" -p 3306:3306 -d mariadb:latest
    docker run --name postgresql -e "POSTGRES_PASSWORD=$POSTGRESQL_PASSWORD" -p 5432:5432 -d postgres:latest
    docker run --name sqlserver -e "ACCEPT_EULA=Y" -e "SA_PASSWORD=$SQLSERVER_PASSWORD" -p 1433:1433 -d mcr.microsoft.com/mssql/server:2019-latest
}

stop_containers() {
    docker stop mariadb postgresql sqlserver
    docker rm mariadb postgresql sqlserver
}

wait_db() {
    NAME=$1
    CMDCHECK=$2

    COUNT=0
    until ( $CMDCHECK >& /dev/null); do
        echo "$NAME is starting..."
        sleep $STARTLOOP_SLEEP
        COUNT=$((COUNT+1))
        if (( $COUNT == $STARTLOOP_MAXITERATIONS )); then
            echo "$NAME take too long time to start."
            return 1
        fi
    done
}

setup_mariadb() {
    wait_db "MariaDB" "docker exec -it mariadb mysql -uroot -p$MARIADB_PASSWORD -h127.0.0.1 -e exit" \
    || return 1

    docker exec -it mariadb mysql -uroot -p$MARIADB_PASSWORD -e "create database godb;"
}

setup_postgresql() {
    wait_db "PostgreSQL" "docker exec -it postgresql psql -Upostgres -c \\q" \
    || return 1
}

setup_sqlserver() {
    wait_db "SQLServer" "docker exec -i sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P $SQLSERVER_PASSWORD -q exit" \
    || return 1

    docker exec -i sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P NotSoStr0ngP@ssword <<-EOF
      create database godb;
      go
      alter database godb set READ_COMMITTED_SNAPSHOT ON;
      go
      exit
EOF
}

# Start all containers
star_containers || exit 1
echo Containers are starting...
sleep 5

# Wait for and setup each DB
setup_postgresql || stop_containers || exit 1
setup_sqlserver || stop_containers || exit 1
setup_mariadb || stop_containers || exit 1
echo Containers are started, DB are ready.

# Install dependencies (while containers are starting)
go mod download

# Let's test (without cache) !
go clean -testcache
go test -v ./...
testresult=$?

# Cleanup
stop_containers

# Display and return a clear test status
echo ----------
if [ $testresult -eq 0 ]; then
  echo OK
else
  echo FAIL
fi
exit $testresult