# hello

A simple 12-factor Go server that counts visits using redis.

## How to run

    docker build -t hello:0.2.0 .
    kind load docker-image hello:0.2.0
    kubectl apply -f k8s/
    kubectl port-forward service/hello 8080:80
    curl localhost:8080/

## Where it meets the twelve factors

| Factor | Where it meets it |
|---|---|
| I. Codebase | The whole codebase is inside this repo; this pull request is the whole change |
| II. Dependencies | go-redis is declared in `go.mod` / `go.sum`. The Dockerfile runs `go mod download`, so nothing is assumed on the machine. |
| III. Config | `PORT`, `GREETING`, `REDIS_URL` are read from the environment in `main.go` and set in `k8s/deployment.yaml`. |
| IV. Backing services | Redis is found only through `REDIS_URL`. If it's unreachable, the server responds `503`. |
| V. Build, release, run | The only image is `hello:0.2.0`. It is configured differently in Docker and Kubernetes, doesn't need a rebuild. |
| VI. Processes | The visit counter is on Redis so the count rises across both replicas and there is no state in memory. |
| VII. Port binding | The Go server serves HTTP itself on `$PORT`. No separate web server is needed. |
| VIII. Concurrency | 2 replicas behind one Service (`replicas: 2` in `k8s/service.yaml`). |
| IX. Disposability | On SIGTERM the server stops accepting connections, lets running requests finish and exits 0. |
| X. Dev/prod parity | The same `hello:0.2.0` image and the same `redis:8-alpine` image run in Docker and in kind. |
| XI. Logs | One JSON line per request on stdout, nothing on stderr. The Redis library's own messages are redirected into the same log. |
| XII. Admin processes | `reset` runs as a one-off pod from the same image (`kubectl run reset ...`). |