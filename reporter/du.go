package reporter

import (
	_ "net/http/pprof"
	"strings"

	"github.com/SpringerPE/graphite-du-report/caching"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type Tree struct {
	RootName string
	nodes    map[string]*caching.Node
	builder  caching.TreeBuilder
	updater  caching.TreeUpdater
}

//Constructor for Tree object
func NewTree(rootName string, builder caching.TreeBuilder, updater caching.TreeUpdater) (*Tree, error) {
	root := &caching.Node{
		Name:     rootName,
		Leaf:     false,
		Size:     int64(0),
		Children: []string{},
	}

	nodes := map[string]*caching.Node{rootName: root}
	tree := &Tree{RootName: rootName, nodes: nodes, builder: builder, updater: updater}
	err := tree.IncrVersion()
	if err != nil {
		return nil, err
	}
	err = tree.AddNode(rootName, root)
	return tree, err
}

func (tree *Tree) IncrVersion() error {
	return tree.updater.IncrVersion()
}

func (tree *Tree) AddNode(key string, node *caching.Node) error {
	return tree.builder.AddNode(node)
}

func (tree *Tree) AddChild(node *caching.Node, child string) error {
	return tree.builder.AddChild(node, child)
}

func (tree *Tree) GetNodeFromRoot(key string) (*caching.Node, error) {
	key = tree.RootName + "." + key
	return tree.GetNode(key)
}

func (tree *Tree) GetNode(key string) (*caching.Node, error) {
	node, err := tree.builder.GetNode(key)
	return node, err
}

func (tree *Tree) ReadNode(key string) (*caching.Node, error) {
	node, err := tree.updater.ReadNode(key)
	return node, err
}

//Calculates the disk usage in terms of number of files contained
func (tree *Tree) UpdateSize(root *caching.Node) (size int64) {
	size = 0
	//if it is a leaf its size is already given
	if root.Leaf || len(root.Children) == 0 {
		return root.Size
	}

	for _, child := range root.Children {
		node, err := tree.GetNode(root.Name + "." + child)
		if err != nil {
		}
		size += tree.UpdateSize(node)
	}

	root.Size = size
	go tree.updater.UpdateNode(root)
	return size
}

func (tree *Tree) GetNodeSize(path string) (int64, error) {
	size := int64(0)
	node, _ := tree.ReadNode(path)
	size = node.Size
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

/* Takes in input a protocol buffer object representing a details response
 * from graphite, and returns a structure representing the metrics tree.
 *
 * The generated tree will be rooted to a node named from the
 * tree.RootName user defined variable
 */
func ConstructTree(tree *Tree, details *pb.MetricDetailsResponse) error {

	var lastVisited *caching.Node
	root, err := tree.GetNode(tree.RootName)
	if err != nil {
		return err
	}

	//cycles on all the metrics of the details response.
	//For each metric it splits the metric name into dot separated elements. Each
	//element will represent a node in the tree structure.
	//
	//All the nodes will have initial Size = 0
	for metric, data := range details.Metrics {
		parts := strings.Split(metric, ".")
		leafIndex := len(parts) - 1

		lastVisited = root

		for currentIndex := 0; currentIndex <= leafIndex; currentIndex++ {
			currentName := strings.Join(parts[0:currentIndex+1], ".")

			if val, _ := tree.GetNodeFromRoot(currentName); val != nil {
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
				Name:     tree.RootName + "." + currentName,
				Children: []string{},
				Leaf:     isLeaf,
				Size:     size,
			}

			tree.AddNode(currentName, currentNode)
			tree.AddChild(lastVisited, parts[currentIndex])

			lastVisited = currentNode
		}
	}
	return nil
}
