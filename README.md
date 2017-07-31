# Description
A tool to expose and visualise disk usage for a graphite stack.

## Sources of inspiration

- https://github.com/Civil/ch-flamegraphs
- https://redis.io/topics/distlock

## Components
The application is divided into three main components:
- `updater` fetches and process the data coming from a graphite cluster, `/metrics/details` endpoint.
- `worker` offers api endpoints accessible to clients
- `renderer` renders the raw usage data into shareable graphs 

## Api endpoints
### updater
The `updater` currently exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|-----------------|
|POST|`/populate`|None|Fetches data from graphite and populates the redis backend|
|DELETE|`/cleanup`|None|Cleanup old metrics|

### worker
The `worker` exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|------------|
|GET|`/`|None|information about the supported endpoints|
|GET|`/size`|`path`: dot separated string|Get the current usage in bytes for `path`|
|GET|`/flame`|None|Generate a html page including the flame graph|
|GET|`/folded`|None|Retrieves a representation of the metrics in folded format|

### renderer
The `renderer` exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|------------|
|GET|`/render/flame`|None|Return a renderered flame image|

## Requirements
The current implementation makes use of a single Redis instance as a data/caching backend

```
Updater  --------->    Redis  <----------    Worker  <------------ Renderer(s)
```

## Configuration
The configuration of the `updater`, `worker` and `renderer` processes relies on kingpin. The following `cli` 
parameters and `envs` are defined:

#### updater

|Param|Env|Default|Meaning|
|------|------|-------|-------|
|profiling|ENABLE_PPROF|false|enable pprof profiling|
|servers|GRAPHITE_SERVERS|127.0.0.1:8080|comma separated list of graphite carbonserver endpoint, exposing `/metrics/details` endpoint|
|bind-address|BIND_ADDRESS|0.0.0.0|binding address for the process|
|bind-port|PORT|6061|binding port for the process|
|root-name|ROOT_NAME|root|name for the root of the filesystem tree|
|redis-addr|REDIS_ADDR|localhost:6379|address and port for the redis datastore
|redis-passwd|REDIS_PASSWD|password|password to access the redis datastore|
|num-update-routines|UPDATE_ROUTINES|10|num of concurrent update routines|
|num-bulk-updates|BULK_UPDATES|100|num of concurrent bulk operations for redis|
|num-bulk-scans|BULK_SCANS|100|num of bulk scans for redis|

in order to run the updater:

```
#RUN THE UPDATER
./updater --servers localhost:8080 --root-name root --redis-addr localhost:6379 --redis-passwd password
```

#### worker

|Param|Env|Default|Meaning|
|------|------|-------|-------|
|profiling|ENABLE_PPROF|false|enable pprof profiling|
|bind-address|BIND_ADDRESS|0.0.0.0|binding address for the process|
|bind-port|PORT|6062|binding port for the process|
|root-name|ROOT_NAME|root|name for the root of the filesystem tree|
|redis-addr|REDIS_ADDR|localhost:6379|address and port for the redis datastore
|redis-passwd|REDIS_PASSWD|password|password to access the redis datastore|
|num-update-routines|UPDATE_ROUTINES|10|num of concurrent update routines|
|num-bulk-updates|BULK_UPDATES|100|num of concurrent bulk operations for redis|
|num-bulk-scans|BULK_SCANS|100|num of bulk scans for redis|

in order to run the worker:
```
#RUN THE WORKER
./worker --root-name root --redis-addr localhost:6379 --redis-passwd password
```

#### renderer

|Param|Env|Default|Meaning|
|------|------|-------|-------|
|profiling|ENABLE_PPROF|false|enable pprof profiling|
|bind-address|BIND_ADDRESS|0.0.0.0|binding address for the process|
|bind-port|PORT|6063|binding port for the process|

in order to run the renderer:
```
#RUN THE RENDERER
./renderer
```

## Installation
`go install github.com/SpringerPE/graphite-du-report/{worker,updater,renderer}`, 
the dependencies are vendored in a `vendor` folder.

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
- a carbonserver, implementing the `metrics/details` endpoint. The directory `test` contains a mock carbonserver 
able to generate a well-formed details response. Just run `go run test/details_server.go`
