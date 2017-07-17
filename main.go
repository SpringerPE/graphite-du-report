package main

import (
	"log"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"
	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/controller"
	"github.com/SpringerPE/graphite-du-report/reporter"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	role              = kingpin.Flag("role", "either worker or updater").Default("worker").OverrideDefaultFromEnvar("ROLE").String()
	profiling         = kingpin.Flag("profiling", "enable profiling via pprof").Default("false").OverrideDefaultFromEnvar("ENABLE_PPROF").Bool()
	serverList        = kingpin.Flag("servers", "comma separated list of the graphite servers").Default("127.0.0.1:8080").OverrideDefaultFromEnvar("GRAPHITE_SERVERS").String()
	bindAddress       = kingpin.Flag("bind-address", "bind address for this server").Default("0.0.0.0").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort          = kingpin.Flag("bind-port", "bind port for this server").Default("6061").OverrideDefaultFromEnvar("PORT").String()
	rootName          = kingpin.Flag("root-name", "name for the root of the tree").Default("root").OverrideDefaultFromEnvar("ROOT_NAME").String()
	redisAddr         = kingpin.Flag("redis-addr", "bind address for the redis instance").Default("localhost:6379").OverrideDefaultFromEnvar("REDIS_ADDR").String()
	redisPasswd       = kingpin.Flag("redis-passwd", "password for redis").Default("").OverrideDefaultFromEnvar("REDIS_PASSWD").String()
	numUpdateRoutines = kingpin.Flag("num-update-routines", "number of concurrent update routines").Default("10").OverrideDefaultFromEnvar("UPDATE_ROUTINES").Int()
	numBulkUpdates    = kingpin.Flag("num-bulk-updates", "number of concurrent bulk node updates").Default("100").OverrideDefaultFromEnvar("BULK_UPDATES").Int()
	numBulkScans      = kingpin.Flag("num-bulk-scans", "number of concurrent bulk node scans").Default("100").OverrideDefaultFromEnvar("BULK_SCANS").Int()
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func main() {
	kingpin.Parse()

	if *role == "updater" {
		runUpdater()
	} else {
		runWorker()
	}
}

func runWorker() {
	kingpin.Parse()

	config := &config.WorkerConfig{
		BindPort:    *bindPort,
		RootName:    *rootName,
		RedisAddr:   *redisAddr,
		RedisPasswd: *redisPasswd,
	}

	reader := caching.NewRedisCaching(config.RedisAddr, config.RedisPasswd)
	treeReader, _ := reporter.NewTreeReader(config.RootName, reader)
	worker, _ := controller.NewWorker(treeReader, config)

	r := mux.NewRouter()
	if *profiling {
		attachProfiler(r)
	}

	r.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		worker.HandleNodeSize(w, r)
	}).Methods("GET").Name("Size")

	r.HandleFunc("/flame", func(w http.ResponseWriter, r *http.Request) {
		worker.HandleFlame(w, r)
	}).Methods("GET").Name("Flame")

	srv := &http.Server{
		Handler: r,
		Addr:    config.BindAddress + ":" + config.BindPort,
	}
	log.Println(srv.ListenAndServe())
}

func runUpdater() {
	sList := config.ParseServerList(*serverList)
	config := &config.UpdaterConfig{
		Servers: sList, BindAddress: *bindAddress,
		BindPort:       *bindPort,
		RootName:       *rootName,
		RedisAddr:      *redisAddr,
		RedisPasswd:    *redisPasswd,
		UpdateRoutines: *numUpdateRoutines,
		BulkUpdates:    *numBulkUpdates,
		BulkScans:      *numBulkScans,
	}
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	builder := caching.NewMemBuilder()
	reader := caching.NewRedisCaching(config.RedisAddr, config.RedisPasswd)
	reader.SetNumBulkScans(config.BulkScans)
	tree, _ := reporter.NewTree(config.RootName, builder, reader)
	tree.SetNumUpdateRoutines(config.UpdateRoutines)
	tree.SetNumBulkUpdates(config.BulkUpdates)

	up, _ := controller.NewUpdater(tree, fetcher, config)
	r := mux.NewRouter()
	if *profiling {
		attachProfiler(r)
	}

	r.HandleFunc("/populate", func(w http.ResponseWriter, r *http.Request) {
		up.PopulateDetails(w, r)
	}).Methods("POST").Name("Populate")

	r.HandleFunc("/cleanup", func(w http.ResponseWriter, r *http.Request) {
		up.Cleanup(w, r)
	}).Methods("DELETE").Name("Cleanup")

	srv := &http.Server{
		Handler: r,
		Addr:    config.BindAddress + ":" + config.BindPort,
	}
	log.Println(srv.ListenAndServe())
}
