package caching

import ()

type FakeCaching struct {
	nodes map[string]*Node
}

func NewFakeCaching() Caching {
	return &FakeCaching{
		nodes: make(map[string]*Node),
	}
}

func (r *FakeCaching) SetNode(node *Node) error {
	r.nodes[node.Name] = node
	return nil
}

func (r *FakeCaching) AddChild(node *Node, child string) error {
	cachedNode := r.nodes[node.Name]
	cachedNode.Children = append(cachedNode.Children, child)
	return nil
}

func (r *FakeCaching) GetNode(key string) (*Node, error) {
	return r.nodes[key], nil
}
