package reporter

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"io/ioutil"
	"net/http"
	_ "net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/caching"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type limiter chan struct{}

func (l limiter) enter() { l <- struct{}{} }
func (l limiter) leave() { <-l }

func newLimiter(l int) limiter {
	return make(chan struct{}, l)
}

var errTimeout = fmt.Errorf("Max tries exceeded")

type Tree struct {
	Root   *caching.Node
	nodes  map[string]*caching.Node
	cacher caching.Cahing
}

func (tree *Tree) AddNode(key string, node *caching.Node) {
	key = tree.Root.Name + "." + key
	tree.nodes[key] = node
}

func (tree *Tree) All() map[string]*caching.Node {
	return tree.nodes
}

func (tree *Tree) GetNodeFromRoot(key string) (*caching.Node, bool) {
	key = tree.Root.Name + "." + key
	return tree.GetNode(key)
}

func (tree *Tree) GetNode(key string) (*caching.Node, bool) {

	if node, ok := tree.nodes[key]; ok {
		return node, true
	}

	return nil, false
}

func NewTree(rootName string, cacher caching.Cacher) *Tree {
	root := &caching.Node{Name: rootName, Leaf: false, Size: int64(0), Children: []string{}}
	nodes := map[string]*caching.Node{rootName: root}
	return &Tree{Root: root, nodes: nodes, cacher: cacher}
}

func ConstructTree(tree *Tree, details *pb.MetricDetailsResponse) {

	var lastVisited *caching.Node

	for metric, data := range details.Metrics {
		parts := strings.Split(metric, ".")
		leafIndex := len(parts) - 1

		lastVisited = tree.Root

		for currentIndex := 0; currentIndex <= leafIndex; currentIndex++ {
			currentName := strings.Join(parts[0:currentIndex+1], ".")
			if val, ok := tree.GetNodeFromRoot(currentName); ok {
				lastVisited = val
				continue
			}

			isLeaf := false
			size := int64(0)

			if currentIndex == leafIndex {
				isLeaf = true
				size = data.Size_
			}

			currentNode := &caching.Node{
				Name:     tree.Root.Name + "." + currentName,
				Children: []string{},
				Leaf:     isLeaf,
				Size:     size,
			}

			tree.AddNode(currentName, currentNode)

			lastVisited.Children = append(lastVisited.Children, parts[currentIndex])
			lastVisited = currentNode
		}
	}
}

type Fetcher interface {
	FetchData(url string) (*pb.MetricDetailsResponse, error)
}

type DataFetcher struct {
	http.Client
	Retries int
}

func NewDataFetcher(timeout time.Duration, retries int) Fetcher {
	return &DataFetcher{Client: http.Client{Timeout: timeout * time.Second}, Retries: retries}
}

func (fetcher *DataFetcher) FetchData(url string) (*pb.MetricDetailsResponse, error) {
	var metricsResponse pb.MetricDetailsResponse
	var response *http.Response
	var err error
	tries := 1

retry:
	if tries > fetcher.Retries {
		log.Println("Tries exceeded while trying to fetch data")
		return nil, errTimeout
	}
	response, err = fetcher.Get(url)
	if err != nil {
		log.Println("Error during communication with client")
		tries++
		time.Sleep(300 * time.Millisecond)
		goto retry
	} else {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println("Error while reading client's response")
			tries++
			time.Sleep(300 * time.Millisecond)
			goto retry
		}

		err = metricsResponse.Unmarshal(body)
		if err != nil || len(metricsResponse.Metrics) == 0 {
			log.Println("Error while parsing client's response")
			tries++
			time.Sleep(300 * time.Millisecond)
			goto retry
		}
	}

	return &metricsResponse, nil
}

//Calculates the disk usage in terms of number of files contained
func (tree *Tree) UpdateSize(root *caching.Node) (size int64) {
	size = 0
	//if it is a file its size is 1
	if root.Leaf || len(root.Children) == 0 {
		return root.Size
	}

	for _, child := range root.Children {
		node, _ := tree.GetNode(root.Name + "." + child)
		size += tree.UpdateSize(node)
	}

	root.Size = size
	tree.cacher.SetNode(root)
	return size
}

func (tree *Tree) GetNodeSize(path string) (int64, error) {
	size := int64(0)

	/*	if node, ok := tree.GetNode(path); ok {
			size = node.Size
		}
	*/
	size
	return size, nil
}

func (tree *Tree) GetOrgTotalUsage(paths []string) (int64, error) {
	size := int64(0)

	for _, path := range paths {
		s, _ := tree.GetNodeSize(path)
		size += int64(s)
	}

	return size, nil
}

func GetDetails(ips []string,
	cluster string,
	fetcher Fetcher) *pb.MetricDetailsResponse {

	response := &pb.MetricDetailsResponse{
		Metrics: make(map[string]*pb.MetricDetails),
	}
	responses := make([]*pb.MetricDetailsResponse, len(ips))
	fetchingLimiter := newLimiter(1)

	var wg sync.WaitGroup
	for idx, ip := range ips {
		wg.Add(1)
		go func(i int, ip string) {
			fetchingLimiter.enter()
			defer fetchingLimiter.leave()
			defer wg.Done()
			url := "http://" + ip + "/metrics/details/?format=protobuf"
			data, err := fetcher.FetchData(url)
			if err != nil {
				log.Println("timeout during fetching details")
				return
			}
			responses[i] = data
		}(idx, ip)
	}
	wg.Wait()

	maxCount := uint64(1)
	metricsReplicationCounter := make(map[string]uint64)
	for idx := range responses {
		if responses[idx] == nil {
			continue
		}

		response.FreeSpace += responses[idx].FreeSpace
		response.TotalSpace += responses[idx].TotalSpace

		for m, v := range responses[idx].Metrics {
			if r, ok := response.Metrics[m]; ok {
				metricsReplicationCounter[m]++
				if metricsReplicationCounter[m] > maxCount {
					maxCount = metricsReplicationCounter[m]
				}
				if v.ModTime > r.ModTime {
					r.ModTime = v.ModTime
				}
				if v.Size_ > r.Size_ {
					r.Size_ = v.Size_
				}
			} else {
				response.Metrics[m] = v
			}
		}
	}

	response.FreeSpace /= uint64(maxCount)
	response.TotalSpace /= uint64(maxCount)

	return response
}
