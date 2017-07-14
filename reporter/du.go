package reporter

import (
	"github.com/SpringerPE/graphite-du-report/caching"
	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
	_ "net/http/pprof"
	"strings"
	"sync"
)

type Tree struct {
	RootName       string
	UpdateRoutines int
	BulkUpdates    int
	nodes          map[string]*caching.Node
	builder        caching.TreeBuilder
	updater        caching.TreeUpdater
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
	tree := &Tree{
		RootName:       rootName,
		nodes:          nodes,
		builder:        builder,
		updater:        updater,
		BulkUpdates:    100,
		UpdateRoutines: 10,
	}
	err := tree.AddNode(rootName, root)
	return tree, err
}

func (tree *Tree) SetNumUpdateRoutines(num int) {
	tree.UpdateRoutines = num
}

func (tree *Tree) SetNumBulkUpdates(num int) {
	tree.BulkUpdates = num
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

//Calculates the disk usage in terms of number of files contained
func (tree *Tree) UpdateSize(root *caching.Node) {
	updateOps := make(chan *caching.Node)

	var wg sync.WaitGroup
	for w := 1; w <= tree.UpdateRoutines; w++ {
		wg.Add(1)
		go tree.updateNode(updateOps, &wg)
	}

	tree.updateSize(root, nil, updateOps)
	close(updateOps)
	wg.Wait()
}

func (tree *Tree) updateNode(updateOps chan *caching.Node, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := []*caching.Node{}
	for {
		node, ok := <-updateOps
		if ok {
			nodes = append(nodes, node)
			if len(nodes) > tree.BulkUpdates {
				tree.updater.UpdateNodes(nodes)
				nodes = nodes[:0]
			}
		} else {
			if len(nodes) > 0 {
				tree.updater.UpdateNodes(nodes)
			}
			break
		}
	}
}

//Calculates the disk usage in terms of number of files contained
func (tree *Tree) updateSize(root *caching.Node, parent *caching.Node, updateOps chan *caching.Node) {
	//if it is a leaf its size is already given
	for _, child := range root.Children {
		node, err := tree.GetNode(root.Name + "." + child)
		if err != nil {
		}
		tree.updateSize(node, root, updateOps)
	}
	updateOps <- root
}

/* Takes in input a protocol buffer object representing a details response
 * from graphite, and returns a structure representing the metrics tree.
 *
 * The generated tree will be rooted to a node named from the
 * tree.RootName user defined variable
 */
func ConstructTree(tree *Tree, details *pb.MetricDetailsResponse) error {
	alreadyVisited := []*caching.Node{}
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

		alreadyVisited = []*caching.Node{root}

		for currentIndex := 0; currentIndex <= leafIndex; currentIndex++ {
			currentName := strings.Join(parts[0:currentIndex+1], ".")

			if val, _ := tree.GetNodeFromRoot(currentName); val != nil {
				alreadyVisited = append(alreadyVisited, val)
				continue
			}

			if currentIndex == leafIndex {
				for _, node := range alreadyVisited {
					node.Leaf = false
					node.Size += data.Size_

				}
				break
			}

			currentNode := &caching.Node{
				Name:     tree.RootName + "." + currentName,
				Children: []string{},
				Leaf:     true,
				Size:     int64(0),
			}
			tree.AddNode(currentName, currentNode)
			tree.AddChild(alreadyVisited[len(alreadyVisited)-1], parts[currentIndex])

			alreadyVisited = append(alreadyVisited, currentNode)
		}
	}
	return nil
}
