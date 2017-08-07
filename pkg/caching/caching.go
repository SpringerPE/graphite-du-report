package caching

import (
	"encoding/json"
	"strings"
)

type Node struct {
	Name     string   `json:"name" redis:"-"`
	Leaf     bool     `json:"-" redis:"leaf"`
	Size     int64    `json:"value" redis:"size"`
	Children []*Node `json:"children" redis:"-"`
}

func (node *Node) MarshalJSON() ([]byte, error) {
	type Alias Node

	nameElements := strings.Split(node.Name, ".")
	lastName := nameElements[len(nameElements)-1]

	if node.Leaf {
		return json.Marshal(&struct {
			Name string `json:"name"`
			Size int64 `json:"value"`
		}{
			Name: lastName,
			Size: node.Size,
		})
	} else {
		return json.Marshal(&struct {
			Name string `json:"name"`
			Children []*Node `json:"children"`
		}{
			Name: lastName,
			Children: node.Children,
		})
	}
}

type TreeBuilder interface {
	GetNode(string) (*Node, error)
	AddNode(*Node) error
	AddChild(*Node, *Node) error
	Clear()
}

type Locker interface {
	WriteLock(string, string, uint64) (bool, error)
	ReleaseLock(string, string) (bool, error)
}

type TreeReader interface {
	ReadNode(string) (*Node, error)
	ReadFlameMap() ([]string, error)
	ReadJsonTree() ([]byte, error)
}

type TreeUpdater interface {
	Version() (string, error)
	IncrVersion() error
	UpdateReaderVersion() error
	UpdateNodes([]*Node) error
	UpdateJson(*Node) error
	Cleanup(string) error
	Close() error
}

type TreeBuilderUpdater interface {
	TreeBuilder
	TreeUpdater
}
