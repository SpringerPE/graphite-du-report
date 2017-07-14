package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"
	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/logging"
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

func errorResponse(w http.ResponseWriter, msg string, err error) {
	logging.LogError(msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, msg)
}

func getNodeSize(w http.ResponseWriter, r *http.Request, tree *reporter.TreeReader) {
	path := r.URL.Query().Get("path")
	size, err := tree.GetNodeSize(path)
	if err != nil {
		errorResponse(w, "failed getting the node size", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%d", size))
}

func flame(w http.ResponseWriter, r *http.Request, tree *reporter.TreeReader, config *config.WorkerConfig) {
	flame, err := tree.ReadFlameMap()
	if err != nil {
		errorResponse(w, "failed reading the root node", err)
		return
	}
	fmt.Println("Tree visit completed")
	w.WriteHeader(http.StatusOK)
	for k, v := range flame {
		fmt.Fprintf(w, fmt.Sprintf("%s %d\n", strings.Replace(k, ".", ";", -1), v))
	}
}

func populateDetails(w http.ResponseWriter, r *http.Request, config *config.UpdaterConfig) {
	logging.LogStd(fmt.Sprintf("%s", "Tree initialisation started"))
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails(config.Servers, "", fetcher)
	builder := caching.NewMemBuilder()
	updater := caching.NewRedisCaching(config.RedisAddr, config.RedisPasswd)

	updater.SetNumBulkScans(config.BulkScans)

	tree, err := reporter.NewTree(config.RootName, builder, updater)
	if err != nil {
		errorResponse(w, "cannot instanciate new reporter tree", err)
		return
	}
	tree.SetNumUpdateRoutines(config.UpdateRoutines)
	tree.SetNumBulkUpdates(config.BulkUpdates)

	logging.LogStd(fmt.Sprintf("%s", "Tree building started"))
	reporter.ConstructTree(tree, response)
	logging.LogStd(fmt.Sprintf("%s", "Tree building finished"))
	root, err := tree.GetNode(config.RootName)
	if err != nil {
		errorResponse(w, "cannot get the root node", err)
		return
	}
	err = updater.IncrVersion()
	if err != nil {
		errorResponse(w, "cannot incr version", err)
		return
	}
	tree.UpdateSize(root)
	logging.LogStd(fmt.Sprintf("%s", "Tree initialisation finished"))
	err = updater.UpdateReaderVersion()
	if err != nil {
		errorResponse(w, "couldn't update current version to match next", err)
		return
	}

	logging.LogStd(fmt.Sprintf("%s", "Cleaning up old versions..."))
	updater.Cleanup(config.RootName)
	logging.LogStd(fmt.Sprintf("%s", "Cleaning up finished"))
	if err != nil {
		errorResponse(w, "error while cleaning up old versions", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	return
}

func cleanup(w http.ResponseWriter, r *http.Request, config *config.UpdaterConfig) {
	logging.LogStd(fmt.Sprintf("%s", "Tree cleanup started"))
	updater := caching.NewRedisCaching(config.RedisAddr, config.RedisPasswd)

	err := updater.Cleanup(config.RootName)
	if err != nil {
		errorResponse(w, "failed cleaning up", err)
		return
	}
	logging.LogStd(fmt.Sprintf("%s", "cleanup finished"))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, fmt.Sprintf("%s", "OK"))
}

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

	r := mux.NewRouter()
	if *profiling {
		attachProfiler(r)
	}

	r.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		getNodeSize(w, r, treeReader)
	}).Methods("GET").Name("Size")

	r.HandleFunc("/flame", func(w http.ResponseWriter, r *http.Request) {
		flame(w, r, treeReader, config)
	}).Methods("GET").Name("Visit")

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
	r := mux.NewRouter()
	if *profiling {
		attachProfiler(r)
	}

	r.HandleFunc("/populate", func(w http.ResponseWriter, r *http.Request) {
		populateDetails(w, r, config)
	}).Methods("POST").Name("Populate")

	r.HandleFunc("/cleanup", func(w http.ResponseWriter, r *http.Request) {
		cleanup(w, r, config)
	}).Methods("DELETE").Name("Cleanup")

	srv := &http.Server{
		Handler: r,
		Addr:    config.BindAddress + ":" + config.BindPort,
	}
	log.Println(srv.ListenAndServe())
}
