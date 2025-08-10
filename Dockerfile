FROM postgres:17
USER postgres
WORKDIR /docker-entrypoint-initdb.d
COPY ./sql/*/. .