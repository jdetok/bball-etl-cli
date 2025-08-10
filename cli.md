# CLI flags

## mode selector
- ### -mode
    - dly
        - etl for games that took place the previous day 
        - uesd in scripts/dly/dly.sh called nightly in cronjob
    - bld
        - etl for all nba/wnba games since 1970
        - used in scripts/bld/bld.sh to build postgres db
    - custom (not yet built)

## env selector
- ### -env
    - prod
        - loads env vars from .env
    - dev
        - loads env vars from .envdev 
    - test
        - loads env vars from .envtst
        - only used for testing scripts in this repo
