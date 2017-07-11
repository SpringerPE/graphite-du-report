package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"

	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/logging"
	"github.com/SpringerPE/graphite-du-report/reporter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serverList  = kingpin.Flag("servers", "comma separated list of the graphite servers").Default("127.0.0.1:8080").OverrideDefaultFromEnvar("SERVERS").String()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("127.0.0.1").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort    = kingpin.Flag("bind-port", "bind port for this server").Default("6061").OverrideDefaultFromEnvar("BIND_PORT").String()
	rootName    = kingpin.Flag("root-name", "name for the root of the tree").Default("root").OverrideDefaultFromEnvar("ROOT_NAME").String()
	redisAddr   = kingpin.Flag("redis-addr", "bind address for the redis instance").Default("localhost:6379").OverrideDefaultFromEnvar("REDIS_ADDR").String()
)

func populateDetails(w http.ResponseWriter, r *http.Request, config *config.Config) (int, error) {
	logging.LogStd(fmt.Sprintf("%s", "Tree initialisation started"))
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails(config.Servers, "", fetcher)
	builder := caching.NewMemBuilder()
	updater := caching.NewRedisCaching(config.RedisAddr)

	tree, err := reporter.NewTree(config.RootName, builder, updater)
	if err != nil {
		logging.LogError("cannot instanciate new reporter tree", err)
		return 500, err
	}

	reporter.ConstructTree(tree, response)
	logging.LogStd(fmt.Sprintf("%s", "Tree building finished"))

	root, _ := tree.GetNode(config.RootName)
	updater.IncrVersion()
	tree.UpdateSize(root)
	logging.LogStd(fmt.Sprintf("%s", "Tree initialisation finished"))
	err = updater.UpdateReaderVersion()
	if err == nil {
		logging.LogStd(fmt.Sprintf("%s", "Cleaning up old versions..."))
		updater.Cleanup(config.RootName)
		logging.LogStd(fmt.Sprintf("%s", "Cleaning up finished"))
	}

	w.Write([]byte(fmt.Sprintf("%s", "OK")))
	return 200, nil
}

func cleanup(w http.ResponseWriter, r *http.Request, config *config.Config) (int, error) {
	logging.LogStd(fmt.Sprintf("%s", "Tree cleanup started"))
	updater := caching.NewRedisCaching(config.RedisAddr)

	err := updater.Cleanup(config.RootName)
	if err != nil {
		logging.LogError("failed cleaning up", err)
		return 500, err
	}
	logging.LogStd(fmt.Sprintf("%s", "cleanup finished"))
	return 200, nil
}

func main() {
	kingpin.Parse()

	sList := config.ParseServerList(*serverList)
	config := &config.Config{
		Servers: sList, BindAddress: *bindAddress,
		BindPort:  *bindPort,
		RootName:  *rootName,
		RedisAddr: *redisAddr,
	}

	http.HandleFunc("/populate", func(w http.ResponseWriter, r *http.Request) {
		populateDetails(w, r, config)
	})

	http.HandleFunc("/cleanup", func(w http.ResponseWriter, r *http.Request) {
		cleanup(w, r, config)
	})

	log.Println(http.ListenAndServe(config.BindAddress+":"+config.BindPort, nil))
}
