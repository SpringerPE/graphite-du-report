# Description
A tool to monitor disk usage for a graphite stack.

## Sources of inspiration

- https://github.com/Civil/ch-flamegraphs
- https://redis.io/topics/distlock

## Components
The application is divided into two main components:
- `updater` fetches and process the data coming from a graphite cluster, `/metrics/details` endpoint.
- `worker` offers api endpoints accessible to clients

## Api endpoints
### updater
The `updater` currently exposes 2 endpoints

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|-----------------|
|POST|`/populate`|None|Fetches data from graphite and populates the redis backend|
|DELETE|`/cleanup`|None|Cleanup old metrics

### worker
The `worker` exposes:
|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|------------|
|GET|`/size`|`path`: dot separated string|Get the current usage in bytes for `path`|
|DELETE|`/flame`|None|Generate a flame graph representing the disk space|


## Requirements
The current implementation makes use of a single Redis instance as a data/caching backend

```
Updater  --------->    Redis  <----------    Worker
```

## Configuration
The configuration of both the `updater` and the `worker` processes relies on kingpin. The following `cli` parameters and `envs` are defined:

|Param|Env|Component|Default|Meaning|
|------|------|---------|-------|-------|
|role|ROLE|BOTH|worker|if worker run a `worker` process otherwise run a `updater` process|
|profiling|ENABLE_PPROF|BOTH|false|enable pprof profiling|
|servers|GRAPHITE_SERVERS|updater|127.0.0.1:8080|comma separated list of graphite carbonserver endpoint, exposing `/metrics/details` endpoint|
|bind-address|BIND_ADDRESS|BOTH|0.0.0.0|binding address for the process|
|bind-port|PORT|BOSH|6061|binding port for the process|
|root-name|ROOT_NAME|BOTH|root|name for the root of the filesystem tree|
|redis-addr|REDIS_ADDR|BOTH|localhost:6379|address and port for the redis datastore
|redis-passwd|REDIS_PASSWD|BOTH|password|password to access the redis datastore|
|num-update-routines|UPDATE_ROUTINES|updater|10|num of concurrent update routines|
|num-bulk-updates|BULK_UPDATES|updater|100|num of concurrent bulk operations for redis|
|num-bulk-scans|BULK_SCANS|updater|100|num of bulk scans for redis|

A usage example for the two processes:
```
#RUN THE UPDATER
./graphite-du-report --servers localhost:8080 --root-name root --redis-addr localhost:6379 --role updater --redis-passwd password
```
```
#RUN THE WORKER
./graphite-du-report --root-name root --redis-addr localhost:6379 --role worker --redis-passwd password
```
## Installation
`go install github.com/SpringerPE/graphite-du-report`, the dependencies are vendored in a `vendor` folder.

They have been generated using the [dep](https://github.com/golang/dep) tool.

## Run the tests
First install `ginkgo` and `gomega`:

```
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
```

then simply run ```ginkgo -r```

## Creating a testing environment
In order to run the `graphite-du-report` locally it is needed to provide two main dependencies:
- a redis installation
- a carbonserver, implementing the `metrics/details` endpoint. The directory `test` contains a mock carbonserver able to generate a well-formed details response. Just run `go run test/details_server.go`
