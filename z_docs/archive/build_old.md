# guide to build & load bball postgres database
this file details instructions for building the bball database
# bld.sh: startup script
the bld.sh script, or some variation of it, shoudl be used to build the docker container, run the Go program to insert nba data into the intake schema, call the sp_build() stored procedure to populate other schemas with the data in intake
## 1. build docker image & compose up
- ** NOTE: the port used in the compose file MUST match to PG_PORT in .env file for script to work
- dev/testing: `docker compose -f devcompose.yaml up --build -d`
    - compose down if necessary: `docker compose -f devcompose.yaml down --rmi all` 
    - uses `devcompose.yaml` file
        - can be whatever file name .yaml - just has to match
- final: `docker compose up --build -d`
    - compose down if necessary: `docker compose down --rmi all` 
    - uses `compose.yaml` file
## 2. wait for container to pass health check
- the docker compose up cmd returns once the container successfully starts 
    - the .sql files in `/docker-entrypoint-initdb.d` run once the container starts though, so the database is not ready to accept connections until these scripts complete. the unil...do... block achieves this
- health check configured in compose file: 
    ```docker
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]
      interval: 2s
      timeout: 2s
      retries: 10
    ```
- wait until "healthy" is returned
    ```docker
    until [ "$(docker inspect -f '{{.State.Health.Status}}' devpg)" = "healthy" ]; do
        echo "waiting for container to return healthy status before continuing"
        sleep 1
    done
    ```
## 3. run go ETL program
- run the Go program to fetch nba data by season from stats.nba.com & insert it into the tables in the intake schema
- `go run ./etl`
- the two vars below can be edited in etl/main.go to change which seasons are fetched
```go
var st string = "1970"
var en string = time.Now().Format("2006") // current year
```
- will fetch all seasons from `st` (start year) through `en` (end year)
    - by default, `en` is set to the current year
    - ETL process is run from LATEST season to OLDEST season    
    
## 4. call sp_build() procedure to run lg, stats, and api procedures. these all insert data from the intake tables into their final tables

- `docker exec -i devpg psql -U postgres -d bball < ./call.sql`
- ./call.sql contains only a call to `sp_build()`
- ### sp_build() calls:
    1. lg.sp_szn_load()
        - load seasons into lg.szn
    2. lg.sp_team_all_load()
        - load teams into lg.team
    3. stats.sp_tbox()
        - load team box scores into stats.tbox
    4. lg.sp_plr_all_load()
        - load players into lg.plr
    5. INSERT INTO lg.plr
        - before continuing, an insert statement must be executed to fix an issue with WNBA player Angel Robinson having two separate player ids
    6. stats.sp_pbox()
        - load player box scores into stats.pbox
    7. api.sp_plr_agg()
        - load aggregate season/career player stats into api.plr_agg
