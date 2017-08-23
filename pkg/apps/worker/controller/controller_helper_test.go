package controller_test

import (
	"github.com/SpringerPE/graphite-du-report/pkg/caching"
	"github.com/SpringerPE/graphite-du-report/pkg/apps/worker/reporter"
	"fmt"
	"strings"
)

type MockTreeReaderFactory struct {}

func (factory MockTreeReaderFactory) CreateTreeReader() *reporter.TreeReader {
	treeCaching := &MockCachingReader{
		Nodes: map[string]*caching.Node{
			"root": &caching.Node{Name: "root", Leaf: true, Size: int64(10), Children: []*caching.Node{}},
		},
	}
	treeReader, _ := reporter.NewTreeReader("root", treeCaching)

	return treeReader
}

type TreeReaderFactory interface {
	CreateTreeReader() reporter.TreeReader
}

type MockCachingReader struct {
	Nodes map[string]*caching.Node
}


func (mtr *MockCachingReader) ReadNode(path string) (*caching.Node, error) {
	if node,ok := mtr.Nodes[path]; ok {
		return node, nil
	}
	return nil, fmt.Errorf("node not existent")
}

func (mtr *MockCachingReader) ReadFlameMap() ([]string, error) {
	flamemap := []string{}
	for name, node := range mtr.Nodes {
		flamemap = append(flamemap,
			fmt.Sprintf("%s %d", strings.Replace(name, ".", ";", -1), node.Size))
	}
	return flamemap, nil
}

func (mtr *MockCachingReader) ReadJsonTree() ([]byte, error) {
	return []byte("{\"root\": 10}"), nil
}

func (mtr *MockCachingReader) GetNodeSize(path string) (int64, error) {
	fmt.Printf("PATH: %v", path)
	if node, ok := mtr.Nodes[path]; ok {
		return node.Size, nil
	}
	return int64(0), fmt.Errorf("node is not existent")
}