package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/pkg/logging"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/config"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/controller"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	profiling         = kingpin.Flag("profiling", "enable profiling via pprof").Default("false").OverrideDefaultFromEnvar("ENABLE_PPROF").Bool()
	bindAddress       = kingpin.Flag("bind-address", "bind address for this server").Default("0.0.0.0").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort          = kingpin.Flag("bind-port", "bind port for this server").Default("6061").OverrideDefaultFromEnvar("PORT").String()
	rootName          = kingpin.Flag("root-name", "name for the root of the tree").Default("root").OverrideDefaultFromEnvar("ROOT_NAME").String()
	redisAddr         = kingpin.Flag("redis-addr", "bind address for the redis instance").Default("localhost:6379").OverrideDefaultFromEnvar("REDIS_ADDR").String()
	redisPasswd       = kingpin.Flag("redis-passwd", "password for redis").Default("").OverrideDefaultFromEnvar("REDIS_PASSWD").String()
	retrieveChildren  = kingpin.Flag("retrieve-children", "whether node children info should be retrieved from the cache").Default("false").OverrideDefaultFromEnvar("RETRIEVE_CHILDREN").Bool()
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func attachStatic(router *mux.Router) {
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./assets/worker/static"))))
}

func main() {
	runWorker()
}

func runWorker() {
	kingpin.Parse()

	config := &config.WorkerConfig{
		BindPort:    *bindPort,
		RootName:    *rootName,
		RedisAddr:   *redisAddr,
		RedisPasswd: *redisPasswd,
		RetrieveChildren: *retrieveChildren,
	}

	worker, _ := controller.NewWorker(config)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	attachStatic(router)

	router.HandleFunc("/", worker.HandleRoot).Methods("GET").Name("Home")
	router.HandleFunc("/size", worker.HandleNodeSize).Methods("GET").Name("Size")
	router.HandleFunc("/folded", worker.HandleFoldedData).Methods("GET").Name("Folder")
	router.HandleFunc("/flame", worker.HandleFlame).Methods("GET").Name("Flame")

	srv := &http.Server{
		Handler: router,
		Addr:    config.BindAddress + ":" + config.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}
