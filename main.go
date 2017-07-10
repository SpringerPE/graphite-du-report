package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/SpringerPE/graphite-du-report/caching"

	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/reporter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serverList  = kingpin.Flag("servers", "comma separated list of the graphite servers").Default("127.0.0.1:8080").OverrideDefaultFromEnvar("SERVERS").String()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("127.0.0.1").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort    = kingpin.Flag("bind-port", "bind port for this server").Default("6060").OverrideDefaultFromEnvar("BIND_PORT").String()
	rootName    = kingpin.Flag("root-name", "name for the root of the tree").Default("root").OverrideDefaultFromEnvar("ROOT_NAME").String()
	redisAddr   = kingpin.Flag("redis-addr", "bind address for the redis instance").Default("localhost:6379").OverrideDefaultFromEnvar("REDIS_ADDR").String()
)

func getNodeSize(w http.ResponseWriter, r *http.Request, tree *reporter.Tree) {
	path := r.URL.Query().Get("path")
	size, _ := tree.GetNodeSize(path)
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func getOrgSize(w http.ResponseWriter, r *http.Request, tree *reporter.Tree) {
	size, _ := tree.GetOrgTotalUsage([]string{"root.carbon", "root.carbon"})
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func populateDetails(config *config.Config) *reporter.Tree {
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails(config.Servers, "", fetcher)
	builder := caching.NewMemBuilder()
	updater := caching.NewRedisCaching(config.RedisAddr)

	tree, err := reporter.NewTree(config.RootName, builder, updater)
	if err != nil {
		fmt.Println(err)
	}

	reporter.ConstructTree(tree, response)
	root, _ := tree.GetNode(config.RootName)
	tree.UpdateSize(root)
	fmt.Println("Tree initialisation finished")
	return tree
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

	tree := populateDetails(config)

	http.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		getNodeSize(w, r, tree)
	})

	http.HandleFunc("/org_size", func(w http.ResponseWriter, r *http.Request) {
		getOrgSize(w, r, tree)
	})

	log.Println(http.ListenAndServe(config.BindAddress+":"+config.BindPort, nil))
}
