package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

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

func getNodeSize(w http.ResponseWriter, r *http.Request, tree *reporter.TreeReader) {
	path := r.URL.Query().Get("path")
	size, _ := tree.GetNodeSize(path)
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func visit(w http.ResponseWriter, r *http.Request, tree *reporter.TreeReader, config *config.Config) {
	root, _ := tree.ReadNode(config.RootName)
	flame := []string{}
	tree.Visit(root, &flame)
	w.Write([]byte(fmt.Sprintf("%s", strings.Join(flame, "\n"))))
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

	reader := caching.NewRedisCaching(config.RedisAddr)
	treeReader, _ := reporter.NewTreeReader(config.RootName, reader)
	http.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		getNodeSize(w, r, treeReader)
	})

	http.HandleFunc("/visit", func(w http.ResponseWriter, r *http.Request) {
		visit(w, r, treeReader, config)
	})

	log.Println(http.ListenAndServe(config.BindAddress+":"+config.BindPort, nil))
}
