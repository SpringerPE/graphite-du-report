package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/SpringerPE/graphite-du-report/config"
	"github.com/SpringerPE/graphite-du-report/reporter"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serverList  = kingpin.Flag("servers", "comma separated list of the graphite servers").Default("127.0.0.1:8080").OverrideDefaultFromEnvar("SERVERS").String()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("127.0.0.1").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort    = kingpin.Flag("bind-port", "bind port for this server").Default("6060").OverrideDefaultFromEnvar("BIND_PORT").String()
	rootName    = kingpin.Flag("root-name", "name for the root of the tree").Default("root").OverrideDefaultFromEnvar("ROOT_NAME").String()
)

func getDetails(w http.ResponseWriter, r *http.Request, tree *reporter.Tree) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree.Root)
}

func getNodeSize(w http.ResponseWriter, r *http.Request, tree *reporter.Tree) {
	path := r.URL.Query().Get("path")
	size, _ := tree.GetNodeSize(path)
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func getOrgSize(w http.ResponseWriter, r *http.Request, tree *reporter.Tree) {
	//path := r.URL.Query().Get("org")
	size, _ := tree.GetOrgTotalUsage([]string{"root.carbon", "root.carbon"})
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func populateDetails(ips []string, rootName string) *reporter.Tree {
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails(ips, "", fetcher)
	tree := reporter.NewTree(rootName)
	reporter.ConstructTree(tree, response)
	tree.UpdateSize(tree.Root)
	return tree
}

func main() {
	kingpin.Parse()

	sList := config.ParseServerList(*serverList)
	config := &config.Config{Servers: sList, BindAddress: *bindAddress, BindPort: *bindPort}

	tree := populateDetails(config.Servers, config.RootName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getDetails(w, r, tree)
	})

	http.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		getNodeSize(w, r, tree)
	})

	http.HandleFunc("/org_size", func(w http.ResponseWriter, r *http.Request) {
		getOrgSize(w, r, tree)
	})

	log.Println(http.ListenAndServe(config.BindAddress+":"+config.BindPort, nil))
}
