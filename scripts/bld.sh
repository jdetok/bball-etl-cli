#!/usr/bin/env bash

# prod dir:
# DIR=/home/jdeto/go/github.com/jdetok/bball-etl-cli

# test dir
DIR=/Users/jdeto/dev/go/src/github.com/jdetok/bball-etl-cli

LOGD=$DIR/z_log_b
DB=bballtst
DB_USER=postgres
CONTAINER=tstpgbball
SCRIPT=./scripts/bld/bld.sql
GOBIN=./bin/cliv2
GORUN="go run ./cli"

# cd into project dir
cd $DIR
# build & run container
docker compose up --build -d || { echo "docker compose up failed"; exit 1; }

echo "docker compose up succesful"

# use health check to wait for database creation
until [ "$(docker inspect -f '{{.State.Health.Status}}' $CONTAINER)" = "healthy" ]; do
    echo "waiting for container to return healthy status before continuing"
    sleep 1
done

# fetch & load to database
# ./bin/cliv2 -env test -mode build || { echo \
$GORUN -env test -mode build || { echo \
    "go etl process failed, compose down & exit"; \
    docker compose down --rmi all; exit 1; }
echo "go etl process successful"

# call procedures here
docker exec -i $CONTAINER psql -U $DB_USER -d $DB < $SCRIPT || \
    { echo "error calling procedures, compose down & exit"; \
        docker compose down --rmi all; exit 1; }

exit 0