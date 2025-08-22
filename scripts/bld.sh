#!/usr/bin/env bash
DIR=/home/jdeto/go/github.com/jdetok/bball-etl-cli
LOGD=$DIR/z_log_b
# cd into this dir
cd $DIR
# build & run container
docker compose up --build -d || { echo "docker compose up failed"; exit 1; }

echo "docker compose up succesful"

# use health check to wait for database creation
until [ "$(docker inspect -f '{{.State.Health.Status}}' tstpgbball)" = "healthy" ]; do
    echo "waiting for container to return healthy status before continuing"
    sleep 1
done

# fetch & load to database
./bin/cliv2 -env test -mode build || { echo \
    "go etl process failed, compose down & exit"; \
    docker compose down --rmi all; exit 1; }
echo "go etl process successful"

# call procedures here
docker exec -i tstpgbball psql -U postgres -d bballtst < ./scripts/bld/bld.sql || \
    { echo "error calling procedures, compose down & exit"; \
        docker compose down --rmi all; exit 1; }

exit 0