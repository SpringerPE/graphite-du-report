package main

import (
	"fmt"
	"log"
	"time"
	"net/http"
	_ "net/http/pprof"
	"encoding/json"

	"github.com/SpringerPE/graphite-du-report/reporter"
	"github.com/SpringerPE/graphite-du-report/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serverList = kingpin.Flag("servers", "comma separated list of the graphite servers").Default("127.0.0.1:8080").OverrideDefaultFromEnvar("SERVERS").String()
	bindAddress = kingpin.Flag("bind-address", "bind address for this server").Default("127.0.0.1").OverrideDefaultFromEnvar("BIND_ADDRESS").String()
	bindPort = kingpin.Flag("bind-port", "bind port for this server").Default("6060").OverrideDefaultFromEnvar("BIND_PORT").String()
)

func getDetails(w http.ResponseWriter, r *http.Request, node *reporter.Node) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&node)
}

func getNodeSize(w http.ResponseWriter, r *http.Request, node *reporter.Node) {
	path := r.URL.Query().Get("path")
	size, _ := node.GetNodeSize(path)
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func getOrgSize(w http.ResponseWriter, r *http.Request, node *reporter.Node) {
	//path := r.URL.Query().Get("org")
	size, _ := node.GetOrgTotalUsage([]string{"root.carbon", "root.carbon"})
	w.Write([]byte(fmt.Sprintf("%d", size)))
}

func populateDetails(ips []string) *reporter.Node {
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails(ips, "", fetcher)
	root := &reporter.Node{Name: "root", Children: map[string]*reporter.Node{}}
	reporter.ConstructTree(root, response)
	reporter.UpdateSize(root)
	return root
}

func main() {
	kingpin.Parse()

	sList := config.ParseServerList(*serverList)
	config := &config.Config{Servers: sList, BindAddress: *bindAddress, BindPort: *bindPort}

	root := populateDetails(config.Servers)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		getDetails(w, r, root)
	})

	http.HandleFunc("/size", func(w http.ResponseWriter, r *http.Request) {
		getNodeSize(w, r, root)
	})

	http.HandleFunc("/org_size", func(w http.ResponseWriter, r *http.Request) {
		getOrgSize(w, r, root)
	})

	log.Println(http.ListenAndServe(config.BindAddress+":"+config.BindPort, nil))
}
