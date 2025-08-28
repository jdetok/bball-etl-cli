#!/usr/bin/env bash

# RUN BBALL ETL CLI IN "daily" MODE - UPDATE DB WITH NEW GAME FROM PREV DAY
DIR=/home/jdeto/go/github.com/jdetok/bball-etl-cli
LOGD=z_log_d
LOGF=$LOGD/dly_etl_$(date +'%m%d%y_%H%M%S').log
ETL=bin/cli
PG_DKR=pgbball
PG_DB=bball
PROC=scripts/dly.sql

# add go binary to path
export PATH=$PATH:/usr/local/go/bin

# cd into this dir
cd $DIR

# create log file 
echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ DAILY BBALL ETL STARTED
++ $(date)
++ LOGFILE: $LOGF
" >> $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ RUNNING GO ETL CLI APPLICATION
" >> $LOGF

# RUN CLI PROCESS IN DAILY MODE, PASS LOG FILE
./$ETL -env prod -mode daily -logf $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ GO ETL CLI APPLICATION RAN SUCCESSFULLY
++ $(date)
++ ATTEMPTING NIGHTLY POSTGRES PROCEDURES TO UPDATE API TABLES
" >> $LOGF

# call procedures: change container name as needed
docker exec -i $PG_DKR psql -U postgres -d $PG_DB < ./$PROC 2>&1 | tee -a $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ POSTGRES PROCEDURES COMPLETE
++ $(date)
" >> $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ DAILY BBALL ETL COMPLETE - ATTEMPTING TO EMAIL LOG
++ $(date)
" >> $LOGF

# email log
./$ETL -mode email -logf $LOGF

echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++ EMAIL SENT - SCRIPT COMPLETE
++ $(date)" >> $LOGF