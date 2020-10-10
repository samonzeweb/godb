#!/usr/bin/env bash

COMPOSE_FILE="docker-compose-test.yml"

MARIADB_USER=godb
MARIADB_PASSWORD=godb
export GODB_MYSQL="$MARIADB_USER:$MARIADB_PASSWORD@/godb?parseTime=true"

POSTGRESQL_USER=godb
POSTGRESQL_PASSWORD=godb
export GODB_POSTGRESQL="postgres://$POSTGRESQL_USER:$POSTGRESQL_PASSWORD@localhost/godb?sslmode=disable"

SQLSERVER_USER=sa
SQLSERVER_PASSWORD=NotSoStr0ngP@ssword
export GODB_MSSQL="Server=127.0.0.1;Database=godb;User Id=$SQLSERVER_USER;Password=$SQLSERVER_PASSWORD"

STARTLOOP_SLEEP=2
STARTLOOP_MAXITERATIONS=30

start_containers() {
    docker-compose -f "$COMPOSE_FILE" up -d
}

stop_containers() {
    docker-compose -f "$COMPOSE_FILE" down
}

stop_containers_and_exit() {
    docker-compose -f "$COMPOSE_FILE" down
    exit 1
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
    wait_db "MariaDB" "docker exec -it godb_mariadb_1 mysql -u$MARIADB_USER -p$MARIADB_PASSWORD -h127.0.0.1 -e exit" \
    || return 1

    #docker exec -it mariadb mysql -uroot -p$MARIADB_PASSWORD -e "create database godb;"
}

setup_postgresql() {
    wait_db "PostgreSQL" "docker exec -it godb_postgresql_1 psql -U$POSTGRESQL_USER -c \\q" \
    || return 1
}

setup_sqlserver() {
    wait_db "SQLServer" "docker exec -i godb_sqlserver_1 /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P $SQLSERVER_PASSWORD -q exit" \
    || return 1

    docker exec -i godb_sqlserver_1 /opt/mssql-tools/bin/sqlcmd -S localhost -U $SQLSERVER_USER -P NotSoStr0ngP@ssword <<-EOF
      create database godb;
      go
      alter database godb set READ_COMMITTED_SNAPSHOT ON;
      go
      exit
EOF
}



# Start all containers
start_containers || stop_containers_and_exit
echo Containers are starting...

# Install dependencies (while containers are starting)
echo Fetch Go dependencies
go mod download

# Wait for and setup each DB
echo Wait until DB are ready
setup_postgresql || stop_containers_and_exit
setup_mariadb || stop_containers_and_exit
setup_sqlserver || stop_containers_and_exit
echo Containers are started, DB are ready.

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