package main

import (
	"time"

	"github.com/SpringerPE/graphite-du-report/reporter"
)

func main() {
	fetcher := reporter.NewDataFetcher(120*time.Second, 3)
	response := reporter.GetDetails([]string{"127.0.0.1"}, "", fetcher)
	root := &reporter.Node{Name: "root", Children: map[string]*reporter.Node{}}
	reporter.ConstructTree(root, response)
	reporter.Count(root)
	//fmt.Printf("%v", *root)
	reporter.Visit("", root)
}
