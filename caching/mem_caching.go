package caching

import (
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

func (r *MemCaching) AddChild(node *Node, child string) error {
	version, _ := r.Version()
	cachedNode := r.nodes[version+":"+node.Name]
	cachedNode.Children = append(cachedNode.Children, child)
	return nil
}

func (r *MemCaching) GetNode(key string) (*Node, error) {
	version, _ := r.Version()
	return r.nodes[version+":"+key], nil
}
