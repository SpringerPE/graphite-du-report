package caching

import (
	"fmt"
	"strconv"
	"sync"
)

type MemUpdater struct {
	nodes       map[string]*Node
	version     int
	versionNext int
	mutex       *sync.RWMutex
}

func NewMemUpdater() TreeUpdater {
	return &MemUpdater{
		nodes:       make(map[string]*Node),
		version:     0,
		versionNext: 0,
		mutex:       &sync.RWMutex{},
	}
}

func (r *MemUpdater) Cleanup(name string) error {
	return nil
}

func (r *MemUpdater) IncrVersion() error {
	r.version = r.version + 1
	return nil
}

func (r *MemUpdater) UpdateReaderVersion() error {
	r.version = r.versionNext
	return nil
}

func (r *MemUpdater) Version() (string, error) {
	return strconv.Itoa(r.version), nil
}

func (r *MemUpdater) VersionNext() (string, error) {
	return strconv.Itoa(r.versionNext), nil
}

func (r *MemUpdater) UpdateNodes(nodes []*Node) error {
	version, _ := r.Version()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, node := range nodes {
		r.nodes[version+":"+node.Name] = node
	}
	return nil
}

/*func (r *MemCaching) AddChild(node *Node, child string) (err error) {
	version, _ := r.Version()
	if cachedNode, ok := r.nodes[version+":"+node.Name]; ok {
		cachedNode.Children = append(cachedNode.Children, child)
	} else {
		err = fmt.Errorf("Node %s not present in memory", node.Name)
	}
	return err
}*/

//TODO
func (r *MemUpdater) ReadFlameMap() (map[string]int64, error) {
	return nil, nil
}

func (r *MemUpdater) ReadNode(key string) (*Node, error) {
	var err error

	version, _ := r.Version()
	if node, ok := r.nodes[version+":"+key]; ok {
		return node, nil
	} else {
		err = fmt.Errorf("Node %s not present in memory", key)
		return nil, err
	}
}
