package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/SpringerPE/graphite-du-report/pkg/logging"
	"github.com/SpringerPE/graphite-du-report/pkg/helper"

	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/config"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/controller"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/reporter"

	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"

)

var (
	profiling         = kingpin.Flag("profiling", "enable profiling via pprof").Default("false").OverrideDefaultFromEnvar("ENABLE_PPROF").Bool()
	bindAddress       = kingpin.Flag("bind-address", "bind address for this server").Default("0.0.0.0").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort          = kingpin.Flag("bind-port", "bind port for this server").Default("6062").OverrideDefaultFromEnvar("PORT").String()
	rootName          = kingpin.Flag("root-name", "name for the root of the tree").Default("root").OverrideDefaultFromEnvar("ROOT_NAME").String()
	redisAddr         = kingpin.Flag("redis-addr", "bind address for the redis instance").Default("localhost:6379").OverrideDefaultFromEnvar("REDIS_ADDR").String()
	redisPasswd       = kingpin.Flag("redis-passwd", "password for redis").Default("").OverrideDefaultFromEnvar("REDIS_PASSWD").String()
	retrieveChildren  = kingpin.Flag("retrieve-children", "whether node children info should be retrieved from the cache").Default("false").OverrideDefaultFromEnvar("RETRIEVE_CHILDREN").Bool()
	basePath = kingpin.Flag("base-path", "base path for this component").Default("worker").OverrideDefaultFromEnvar("BASE_PATH").String()
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)
}

func attachStatic(router *mux.Router) {
	router.PathPrefix("/worker/static/").Handler(http.StripPrefix("/worker/static", http.FileServer(http.Dir("./assets/worker/static"))))
}

func main() {
	runWorker()
}

func runWorker() {
	kingpin.Parse()

	baseFolder, _ := os.Getwd()
	templateFolder := filepath.Join(baseFolder, "assets/worker/static/templates/*")

	workerConfig := &config.WorkerConfig{
		BindPort:    *bindPort,
		RootName:    *rootName,
		RedisAddr:   *redisAddr,
		RedisPasswd: *redisPasswd,
		RetrieveChildren: *retrieveChildren,
		TemplatesFolder: templateFolder,
		BasePath: *basePath,
	}

	jsonWorkerConfig, err := json.Marshal(workerConfig)
	if err != nil {
		panic("cannot marshal worker configuration into valid json")
	}

	treeReaderConfig, err := reporter.NewRedisTreeReaderConfig(jsonWorkerConfig)
	if err != nil {
		panic("cannot generate a new tree reader config from the worker config")
	}
	var treeReaderFactory reporter.TreeReaderFactory
	treeReaderFactory = reporter.NewRedisTreeReaderFactory(treeReaderConfig)
	worker, _ := controller.NewWorker(workerConfig, treeReaderFactory)

	router := mux.NewRouter()
	if *profiling {
		attachProfiler(router)
	}

	attachStatic(router)

	router.HandleFunc("/", worker.HandleRoot).Methods("GET").Name("Home")
	router.HandleFunc(filepath.Join("/", workerConfig.BasePath, "size"), worker.HandleNodeSize).Methods("GET").Name("Size")
	router.HandleFunc(filepath.Join("/", workerConfig.BasePath, "folded"), worker.HandleFoldedData).Methods("GET").Name("Folder")
	router.HandleFunc(filepath.Join("/", workerConfig.BasePath, "json"), helper.MakeGzipHandler(worker.HandleJsonData)).Methods("GET").Name("Json")

	srv := &http.Server{
		Handler: router,
		Addr:    workerConfig.BindAddress + ":" + workerConfig.BindPort,
	}
	logging.LogStd(fmt.Sprintf("%s", srv.ListenAndServe()))
}
