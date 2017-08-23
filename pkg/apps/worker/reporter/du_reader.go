package reporter

import (
	_ "net/http/pprof"

	"encoding/json"

	"github.com/SpringerPE/graphite-du-report/pkg/caching"
)

type TreeReaderFactory interface {
	CreateTreeReader() *TreeReader
}

type RedisTreeReaderConfig struct {
	RootName    string `json:"root_name"`
	RedisAddr   string `json:"redis_addr"`
	RedisPasswd string `json:"redis_passwd"`
	RetrieveChildren bool `json:"retrieve_children"`
}

func NewRedisTreeReaderConfig(jsonConf []byte) (*RedisTreeReaderConfig, error) {
	treeReaderConfig := &RedisTreeReaderConfig{}
	err := json.Unmarshal(jsonConf, treeReaderConfig)
	if err != nil {
		return nil, err
	}
	return treeReaderConfig, nil
}

type RedisTreeReaderFactory struct {
	config *RedisTreeReaderConfig
}

func NewRedisTreeReaderFactory(config *RedisTreeReaderConfig) *RedisTreeReaderFactory {
	return &RedisTreeReaderFactory{config: config}
}

func (factory *RedisTreeReaderFactory) CreateTreeReader() *TreeReader {

	reader := caching.NewRedisCaching(factory.config.RedisAddr, factory.config.RedisPasswd, factory.config.RetrieveChildren)
	treeReader, _ := NewTreeReader(factory.config.RootName, reader)

	return treeReader
}

type TreeReader struct {
	RootName string
	reader   caching.TreeReader
}

//Constructor for Tree object
func NewTreeReader(rootName string, reader caching.TreeReader) (*TreeReader, error) {
	tree := &TreeReader{RootName: rootName, reader: reader}
	return tree, nil
}

func (tree *TreeReader) ReadNode(key string) (*caching.Node, error) {
	node, err := tree.reader.ReadNode(key)
	return node, err
}

func (tree *TreeReader) ReadFlameMap() ([]string, error) {
	return tree.reader.ReadFlameMap()
}

func (tree *TreeReader) ReadJsonTree() ([]byte, error) {
	return tree.reader.ReadJsonTree()
}


func (tree *TreeReader) GetNodeSize(path string) (int64, error) {
	size := int64(0)
	node, err := tree.ReadNode(path)
	if err != nil {
		return size, err
	}
	size = node.Size
	return size, nil
}
