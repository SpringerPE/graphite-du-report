# Description
A tool to expose and visualise disk usage for a graphite cluster.

## Sources of inspiration

- https://github.com/Civil/ch-flamegraphs
- https://redis.io/topics/distlock

## Structure of this repositories

Executables can be found under the `cmd` folder.
The `pkg` folder contains both shared packages, while application code is located under
the `pkg/apps` subfolder

## Components
The application is split into four main components:
- the `updater` fetches and process the data coming from a graphite cluster, `/metrics/details` endpoint.
- the `worker` offers api and data endpoints accessible to web clients and renderers
- the `renderer` renders the raw disk usage data into a SVG image
- the `visualiser` visualises the raw disk usage data via an html document

## Api endpoints
### updater
The `updater` currently exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|-----------------|
|POST|`/populate`|None|Fetches data from graphite and populates the redis backend|
|DELETE|`/cleanup`|None|Cleanups old metrics|

### worker
The `worker` exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|------------|
|GET|`/`|None|information about the supported endpoints|
|GET|`[/{worker.BasePath}]/size`|`path`: dot separated string|Gets the current usage in bytes for the metrics under `path`|
|GET|`[/{worker.BasePath}]/folded`|None|Generates a folded format representation of the current disk usage|
|GET|`[/{worker.BasePath}]/json`|None|Generates a json tree representation of the current disk usage|

### renderer
The `renderer` exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|------------|
|GET|`[/{renderer.BasePath}]/flame`|None|Returns a renderered flame image|

#### How to contribute your own renderer

At the moment only one renderer implementation is offered, which produces a `flame` svg
graph starting from a folded data representation.

The `worker` process is in charge of serving the raw binary data to be 
rendered.

Additional renderers can be developed as simple web apps, implementing a `GET`
endpoint and returning an html snippet containing the rendered image or object.

### visualiser
The `visualiser` exposes:

|Method|Endpoint|Parameters|Usage|
|----------|----------|---------------|------------|
|GET|`[/{visualiser.BasePath}]/flame`|None|Returns a html page which contains the rendered disk usage stats|

## Requirements
The current implementation makes use of a single Redis instance as a data/caching backend

```
Updater  --------->    Redis  <----------    Worker  <------------ Renderer(s) <---------- Visualiser
```

The `worker` , `renderer` and `visualiser` processes need to be served from the same base url
endpoint. 

In particular the `worker` should be served under the `/` context path,
while the `renderer` should be served under the `/renderer` context path and the `visualiser` from the `/visualise` context path ie.:

- graphite-du.example.com (`worker`)
- graphite-du.example.com/renderer (`renderer`)
- graphite-du.example.com/visualiser (`visualiser`)

There are no limitations regarding the `updater`

## Configuration
The configuration of the `updater`, `worker`, `renderer` and `visualiser` processes relies on kingpin. The following `cli` 
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
|base-path|BASE_PATH|worker|base context path for this component|

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
|bind-port|PORT|6062|binding port for the process|
|base-path|BASE_PATH|renderer|base context path for this component|

in order to run the renderer:
```
#RUN THE RENDERER
./renderer
```

#### visualiser

|Param|Env|Default|Meaning|
|------|------|-------|-------|
|profiling|ENABLE_PPROF|false|enable pprof profiling|
|bind-address|BIND_ADDRESS|0.0.0.0|binding address for the process|
|bind-port|PORT|6063|binding port for the process|
|base-path|BASE_PATH|visualiser|base context path for this component|
|renderer-path|RENDERER_PATH|renderer|base context path for the renderer component|

in order to run the visualiser:
```
#RUN THE VISUALISER
./visualiser
```

## Installation
`go install github.com/SpringerPE/graphite-du-report/cmd/{worker,updater,renderer,visualiser}`, 
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
able to generate a well-formed details response. Just run `go run cmd/carbonserver_test/main.go`
