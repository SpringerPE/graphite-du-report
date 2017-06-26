package caching

type Node struct {
	Name     string   `json:"name" redis:"-"`
	Leaf     bool     `json:"leaf" redis:"leaf"`
	Size     int64    `json:"size" redis:"size"`
	Children []string `json:"children" redis:"-"`
}

type Caching interface {
	SetNode(*Node) error
	GetNode(string) (*Node, error)
	AddChild(*Node, string) error
}
