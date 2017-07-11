package reporter

import (
	"fmt"
	_ "net/http/pprof"
	"strings"

	"github.com/SpringerPE/graphite-du-report/caching"
)

type TreeReader struct {
	RootName string
	reader   caching.TreeUpdater
}

//Constructor for Tree object
func NewTreeReader(rootName string, reader caching.TreeUpdater) (*TreeReader, error) {
	tree := &TreeReader{RootName: rootName, reader: reader}
	return tree, nil
}

func (tree *TreeReader) ReadNode(key string) (*caching.Node, error) {
	node, err := tree.reader.ReadNode(key)
	return node, err
}

//Calculates the disk usage in terms of number of files contained
func (tree *TreeReader) Visit(root *caching.Node, doc *[]string) {
	//if it is a leaf its size is already given
	if root.Leaf || len(root.Children) == 0 {
		*doc = append(*doc, fmt.Sprintf("%s %d", strings.Replace(root.Name, ".", ";", -1), root.Size))
		return
	}

	for _, child := range root.Children {
		node, err := tree.ReadNode(root.Name + "." + child)
		if err != nil {
			fmt.Println(err)
		}
		tree.Visit(node, doc)
	}

	//tree.updater.Visit(root)
}

func (tree *TreeReader) GetNodeSize(path string) (int64, error) {
	size := int64(0)
	node, _ := tree.ReadNode(path)
	size = node.Size
	return size, nil
}
