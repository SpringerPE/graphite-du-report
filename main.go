package main

import (
	"log"
	"time"
	"net/http"
	_ "net/http/pprof"
	"encoding/json"

	"github.com/SpringerPE/graphite-du-report/reporter"
)

func getDetails(w http.ResponseWriter, r *http.Request) {
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails([]string{"127.0.0.1"}, "", fetcher)
	root := &reporter.Node{Name: "root", Children: map[string]*reporter.Node{}}
	reporter.ConstructTree(root, response)
	reporter.UpdateSize(root)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&root)
}

func main() {
	http.HandleFunc("/", getDetails)
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
