package caching

import (
	"fmt"
	"strconv"
)

type MemCaching struct {
	nodes   map[string]*Node
	version int
}

func NewMemCaching() Caching {
	return &MemCaching{
		nodes:   make(map[string]*Node),
		version: 0,
	}
}

func (r *MemCaching) IncrVersion() error {
	r.version = r.version + 1
	return nil
}

func (r *MemCaching) Version() (string, error) {
	return strconv.Itoa(r.version), nil
}

func (r *MemCaching) SetNode(node *Node) error {
	version, _ := r.Version()
	r.nodes[version+":"+node.Name] = node
	return nil
}

func (r *MemCaching) AddChild(node *Node, child string) (err error) {
	version, _ := r.Version()
	if cachedNode, ok := r.nodes[version+":"+node.Name]; ok {
		cachedNode.Children = append(cachedNode.Children, child)
	} else {
		err = fmt.Errorf("Node %s not present in memory", node.Name)
	}
	return err
}

func (r *MemCaching) GetNode(key string) (*Node, error) {
	var err error

	version, _ := r.Version()
	if node, ok := r.nodes[version+":"+key]; ok {
		return node, nil
	} else {
		err = fmt.Errorf("Node %s not present in memory", key)
		return nil, err
	}
}
