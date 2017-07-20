package caching

type Node struct {
	Name     string   `json:"name" redis:"-"`
	Leaf     bool     `json:"leaf" redis:"leaf"`
	Size     int64    `json:"size" redis:"size"`
	Children []string `json:"children" redis:"-"`
}

type TreeBuilder interface {
	GetNode(string) (*Node, error)
	AddNode(*Node) error
	AddChild(*Node, string) error
	Clear()
}

type TreeUpdater interface {
	Version() (string, error)
	VersionNext() (string, error)
	IncrVersion() error
	UpdateReaderVersion() error
	UpdateNodes([]*Node) error
	ReadNode(string) (*Node, error)
	ReadFlameMap() ([]string, error)
	Cleanup(string) error
	WriteLock(string, string, uint64) (bool, error)
	ReleaseLock(string, string) (bool, error)
	Close() error
}

type TreeBuilderUpdater interface {
	TreeBuilder
	TreeUpdater
}
