package caching

import (
	"fmt"
	"strconv"
	"sync"
)

type MemUpdater struct {
	nodes   map[string]*Node
	version int
	mutex   *sync.RWMutex
}

func NewMemUpdater() TreeUpdater {
	return &MemUpdater{
		nodes:   make(map[string]*Node),
		version: 0,
		mutex:   &sync.RWMutex{},
	}
}

func (r *MemUpdater) IncrVersion() error {
	r.version = r.version + 1
	return nil
}

func (r *MemUpdater) Version() (string, error) {
	return strconv.Itoa(r.version), nil
}

func (r *MemUpdater) UpdateNode(node *Node) error {
	version, _ := r.Version()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.nodes[version+":"+node.Name] = node
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
