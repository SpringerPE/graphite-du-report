package main

import (
	"github.com/SpringerPE/graphite-du-report/reporter"
)

func main() {
	response := reporter.GetDetails([]string{"127.0.0.1"}, "")
	root := &reporter.Node{Name: "root", Children: map[string]*reporter.Node{}}
	reporter.ConstructTree(root, response)
	reporter.Count(root)
	//fmt.Printf("%v", *root)
	reporter.Visit("", root)
}
