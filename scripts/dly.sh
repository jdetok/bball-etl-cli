#!/usr/bin/env bash

# add go binary to path
export PATH=$PATH:/usr/local/go/bin

# DAILY SCRIPT TO UPDATE DATABASE WITH NEW GAMES. RUNS CLI IN "daily" MODE
DIR=/home/jdeto/go/github.com/jdetok/bball-etl-cli
# LOGD=$DIR/z_log_d
LOGD=z_log_d
LOGF=$LOGD/dly_etl_$(date +'%m%d%y_%H%M%S').log

# cd into this dir
cd $DIR

# create log file 
# touch $LOGF
echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ DAILY BBALL ETL STARTED
++ $(date)
++ LOGFILE: $LOGF
" >> $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ RUNNING GO ETL CLI APPLICATION
" >> $LOGF

# RUN CLI PROCESS IN DAILY MODE, PASS LOG FILE
./bin/logf_test -env prod -mode daily -logf $LOGF
# ./bin/cliv2 -env prod -mode daily

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ GO ETL CLI APPLICATION RAN SUCCESSFULLY
++ $(date)
++ ATTEMPTING NIGHTLY POSTGRES PROCEDURES TO UPDATE API TABLES
" >> $LOGF

# echo "attempting to run sp_nightly_call() from call.sql at $(date)..." | tee -a $LOGF
# call procedures: change container name as needed
docker exec -i pgbball psql -U postgres -d bball < ./scripts/dly.sql 2>&1 | tee -a $LOGF
# docker exec -i devpg psql -U postgres -d bball < ./call.sql

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ POSTGRES PROCEDURES COMPLETE
++ $(date)
" >> $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ DAILY BBALL ETL COMPLETE
++ $(date)
" >> $LOGF