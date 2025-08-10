#!/usr/bin/env bash

# DAILY SCRIPT TO UPDATE DATABASE WITH NEW GAMES. RUNS CLI IN "daily" MODE
DIR=/home/jdeto/go/github.com/jdetok/bball-etl-cli
LOGD=$DIR/z_log_d
# cd into this dir
cd $DIR

# run nightly etl
export PATH=$PATH:/usr/local/go/bin

# RUN CLI PROCESS IN DAILY MODE
go run ./cli -mode daily

# TODO: open the log file (most recent in z_log) and append to it
LOGF=$LOGD/$(ls $LOGD -t | head -n 1)

echo "attempting to run sp_nightly_call() from call.sql at $(date)..." | tee -a $LOGF
# call procedures: change container name as needed
docker exec -i dev-pgbball psql -U postgres -d bballdev < ./dly.sql 2>&1 | tee -a $LOGF
# docker exec -i devpg psql -U postgres -d bball < ./call.sql

echo "finished running sp_nightly_call()" | tee -a $LOGF
echo "script complete at $(date)" | tee -a $LOGF