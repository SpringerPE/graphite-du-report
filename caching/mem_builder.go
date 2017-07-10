package caching

import (
	"fmt"
)

type MemBuilder struct {
	nodes map[string]*Node
}

func NewMemBuilder() TreeBuilder {
	return &MemBuilder{
		nodes: make(map[string]*Node),
	}
}

func (r *MemBuilder) AddNode(node *Node) error {
	r.nodes[node.Name] = node
	return nil
}

func (r *MemBuilder) AddChild(node *Node, child string) (err error) {
	if cachedNode, ok := r.nodes[node.Name]; ok {
		cachedNode.Children = append(cachedNode.Children, child)
	} else {
		err = fmt.Errorf("Node %s not present in memory", node.Name)
	}
	return err
}

func (r *MemBuilder) GetNode(key string) (*Node, error) {
	var err error

	if node, ok := r.nodes[key]; ok {
		return node, nil
	} else {
		err = fmt.Errorf("Node %s not present in memory", key)
		return nil, err
	}
}
