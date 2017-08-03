package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/pkg/logging"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/updater/config"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/updater/controller"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
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
	storeChildren      = kingpin.Flag("store-children", "whether node children info should be stored in the cache").Default("false").OverrideDefaultFromEnvar("STORE_CHILDREN").Bool()
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func main() {
	runUpdater()
}

func runUpdater() {
	kingpin.Parse()

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
		StoreChildren:  *storeChildren,
	}

	updater := controller.NewUpdater(config)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	router.HandleFunc("/", updater.HandleRoot).Methods("GET").Name("Home")
	router.HandleFunc("/populate", updater.PopulateDetails).Methods("POST").Name("Populate")
	router.HandleFunc("/cleanup", updater.Cleanup).Methods("DELETE").Name("Cleanup")

	srv := &http.Server{
		Handler: router,
		Addr:    config.BindAddress + ":" + config.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}
