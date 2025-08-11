# dev notes

## bypassing cache when updating golib: 
`GOPRIVATE=github.com/jdetok/* go get -x github.com/jdetok/golib@latest`
`GOPRIVATE=github.com/jdetok/* go get -x github.com/jdetok/bball-etl-go@latest`

`docker compose -f devcompose.yaml up --build`
`docker compose -f devcompose.yaml down --rmi all`