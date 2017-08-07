package reporter

import (
	"fmt"
	_ "net/http/pprof"
	"strings"
	"sync"

	"github.com/SpringerPE/graphite-du-report/pkg/caching"
	"github.com/SpringerPE/graphite-du-report/pkg/logging"

	pb "github.com/go-graphite/carbonzipper/carbonzipperpb3"
)

type Tree struct {
	RootName       string
	UpdateRoutines int
	BulkUpdates    int
	nodes          map[string]*caching.Node `json:"tree"`
	builder        caching.TreeBuilder
	updater        caching.TreeUpdater
	locker         caching.Locker
}

//Constructor for Tree object
func NewTree(rootName string, builder caching.TreeBuilder, updater caching.TreeUpdater,
	locker caching.Locker) (*Tree, error) {
	root := &caching.Node{
		Name:     rootName,
		Leaf:     false,
		Size:     int64(0),
		Children: []*caching.Node{},
	}

	nodes := map[string]*caching.Node{rootName: root}
	tree := &Tree{
		RootName:       rootName,
		nodes:          nodes,
		builder:        builder,
		updater:        updater,
		locker:         locker,
		BulkUpdates:    100,
		UpdateRoutines: 10,
	}
	err := tree.AddNode(rootName, root)
	return tree, err
}

func (tree *Tree) WriteLock(name, secret string, ttl uint64) (bool, error) {
	return tree.locker.WriteLock(name, secret, ttl)
}

func (tree *Tree) ReleaseLock(name, secret string) (bool, error) {
	return tree.locker.ReleaseLock(name, secret)
}

func (tree *Tree) SetNumUpdateRoutines(num int) {
	tree.UpdateRoutines = num
}

func (tree *Tree) SetNumBulkUpdates(num int) {
	tree.BulkUpdates = num
}

func (tree *Tree) Cleanup(rootName string) error {
	return tree.updater.Cleanup(rootName)
}

func (tree *Tree) IncrVersion() error {
	return tree.updater.IncrVersion()
}

func (tree *Tree) UpdateReaderVersion() error {
	return tree.updater.UpdateReaderVersion()
}

func (tree *Tree) AddNode(key string, node *caching.Node) error {
	return tree.builder.AddNode(node)
}

func (tree *Tree) AddChild(node *caching.Node, child *caching.Node) error {
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

func (tree *Tree) Persist() error {
	// Get the root node
	root, err := tree.GetNode(tree.RootName)
	if err != nil {
		return fmt.Errorf("%s: %v", "cannot get the root node", err)
	}
	// Increase the version for the newly build tree
	err = tree.IncrVersion()
	if err != nil {
		return fmt.Errorf("%s: %v", "cannot incr version", err)
	}
	// Save the tree to the persistent datastore
	// TODO: generate and handle errors here too
	tree.persistTree(root)
	tree.updater.UpdateJson(root)
	//Update the reader version to the current version upon succeeding
	err = tree.UpdateReaderVersion()
	if err != nil {
		return fmt.Errorf("%s: %v", "cannot incr version", err)
	}
	logging.LogStd(fmt.Sprintf("%s", "Tree initialisation finished"))

	logging.LogStd(fmt.Sprintf("%s", "Cleaning up old versions..."))
	tree.Cleanup(tree.RootName)
	if err != nil {
		return fmt.Errorf("%s: %v", "error while cleaning up old versions", err)
	}
	logging.LogStd(fmt.Sprintf("%s", "Cleaning up finished"))
	return nil
}

//Calculates the disk usage in terms of number of files contained
func (tree *Tree) persistTree(root *caching.Node) {
	updateOps := make(chan *caching.Node)

	var wg sync.WaitGroup
	for w := 1; w <= tree.UpdateRoutines; w++ {
		wg.Add(1)
		go tree.persistNode(updateOps, &wg)
	}

	tree.persist(root, nil, updateOps)
	close(updateOps)
	wg.Wait()
}

//Calculates the disk usage in terms of number of files contained
func (tree *Tree) persist(root *caching.Node, parent *caching.Node, updateOps chan *caching.Node) {
	//if it is a leaf its size is already given
	for _, child := range root.Children {
		node, err := tree.GetNode(child.Name)
		//TODO: handle this somehow
		if err != nil {
		}
		tree.persist(node, root, updateOps)
	}
	updateOps <- root
}

func (tree *Tree) persistNode(updateOps chan *caching.Node, wg *sync.WaitGroup) {
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

/* Takes in input a protocol buffer object representing a details response
 * from graphite, and returns a structure representing the metrics tree.
 *
 * The generated tree will be rooted to a node named from the
 * tree.RootName user defined variable
 */
func (tree *Tree) ConstructTree(details *pb.MetricDetailsResponse) error {
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
				for index, node := range alreadyVisited {
					if index != len(alreadyVisited)-1 {
						node.Leaf = false
					}
					node.Size += data.Size_

				}
				break
			}

			currentNode := &caching.Node{
				Name:     tree.RootName + "." + currentName,
				Children: []*caching.Node{},
				Leaf:     true,
				Size:     int64(0),
			}

			tree.AddNode(currentName, currentNode)
			tree.AddChild(alreadyVisited[len(alreadyVisited)-1], currentNode)

			alreadyVisited = append(alreadyVisited, currentNode)
		}
	}

	return nil
}
