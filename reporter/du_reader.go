package reporter

import (
	_ "net/http/pprof"

	"github.com/SpringerPE/graphite-du-report/caching"
)

type TreeReader struct {
	RootName string
	reader   caching.TreeReader
}

//Constructor for Tree object
func NewTreeReader(rootName string, reader caching.TreeReader) (*TreeReader, error) {
	tree := &TreeReader{RootName: rootName, reader: reader}
	return tree, nil
}

func (tree *TreeReader) ReadNode(key string) (*caching.Node, error) {
	node, err := tree.reader.ReadNode(key)
	return node, err
}

func (tree *TreeReader) ReadFlameMap() ([]string, error) {
	return tree.reader.ReadFlameMap()
}

func (tree *TreeReader) GetNodeSize(path string) (int64, error) {
	size := int64(0)
	node, _ := tree.ReadNode(path)
	size = node.Size
	return size, nil
}
