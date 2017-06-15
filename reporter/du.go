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

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type limiter chan struct{}

func (l limiter) enter() { l <- struct{}{} }
func (l limiter) leave() { <-l }

func newLimiter(l int) limiter {
	return make(chan struct{}, l)
}

var errTimeout = fmt.Errorf("Max tries exceeded")

type Node struct {
	Name     string `json: "name"`
	Leaf     bool `json: "leaf"`
	Size     int64 `json: "size"`
	Children map[string]*Node `json: "children"`
}

func ConstructTree(root *Node, details *pb.MetricDetailsResponse) {
	for metric, data := range details.Metrics {
		currentNode := root
		parts := strings.Split(metric, ".")
		leafIndex := len(parts) - 1
		for index, part := range parts {
			if val, ok := currentNode.Children[part]; ok {
				currentNode = val
				continue
			}
			isLeaf := false
			size := int64(0)
			if index == leafIndex {
				isLeaf = true
				size = data.Size_
			}
			currentNode.Children[part] = &Node{
				Name:     part,
				Children: map[string]*Node{},
				Leaf:     isLeaf,
				Size:     size,
			}
			currentNode = currentNode.Children[part]
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
func UpdateSize(root *Node) (size int64) {
	size = 0
	//if it is a file its size is 1
	if root.Leaf {
		return root.Size
	}

	for _, child := range root.Children {
		size += UpdateSize(child)
	}

	root.Size = size
	return size
}

func Visit(name string, root *Node) {
	if name != "" {
		name += "."
	}
	name += root.Name

	for _, child := range root.Children {
		Visit(name, child)
	}

	fmt.Printf("Folder: %s Size: %d\n", name, root.Size)

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
