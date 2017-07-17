package reporter_test

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/SpringerPE/graphite-du-report/caching"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type FakeDataFetcher struct {
	Responses map[string]*pb.MetricDetailsResponse
}

func NewFakeDataFetcher() *FakeDataFetcher {
	return &FakeDataFetcher{make(map[string]*pb.MetricDetailsResponse)}
}

func (fetcher *FakeDataFetcher) FetchData(url string) (*pb.MetricDetailsResponse, error) {
	if val, ok := fetcher.Responses[url]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("No data to fetch")
}

type MockUpdater struct {
	nodes       map[string]*caching.Node
	flame       []string
	version     int
	versionNext int
	mutex       *sync.RWMutex
}

func NewMockUpdater() caching.TreeUpdater {
	return &MockUpdater{
		nodes:       make(map[string]*caching.Node),
		version:     0,
		versionNext: 0,
		mutex:       &sync.RWMutex{},
	}
}

func (r *MockUpdater) Cleanup(name string) error {
	return nil
}

func (r *MockUpdater) IncrVersion() error {
	r.versionNext += 1
	return nil
}

func (r *MockUpdater) UpdateReaderVersion() error {
	r.version = r.versionNext
	return nil
}

func (r *MockUpdater) Version() (string, error) {
	return strconv.Itoa(r.version), nil
}

func (r *MockUpdater) VersionNext() (string, error) {
	return strconv.Itoa(r.versionNext), nil
}

func (r *MockUpdater) UpdateNodes(nodes []*caching.Node) error {
	versionNext, _ := r.VersionNext()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, node := range nodes {
		r.nodes[versionNext+":"+node.Name] = node
	}
	return nil
}

//TODO
func (r *MockUpdater) ReadFlameMap() ([]string, error) {
	return nil, nil
}

func (r *MockUpdater) ReadNode(key string) (*caching.Node, error) {
	var err error
	version, _ := r.Version()
	if node, ok := r.nodes[version+":"+key]; ok {
		return node, nil
	} else {
		err = fmt.Errorf("Node %s not present in memory", key)
		return nil, err
	}
}
