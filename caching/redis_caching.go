package caching

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedisCaching struct {
	Pool *redis.Pool
}

func (r *RedisCaching) IncrVersion() error {
	conn := r.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("INCR", "version")
	return err
}

func (r *RedisCaching) Version() (string, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := redis.String(conn.Do("GET", "version"))
	return version, err
}

func (r *RedisCaching) UpdateNode(node *Node) error {
	version, err := r.Version()
	if err != nil {
		return err
	}

	conn := r.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	versionedName := version + ":" + node.Name

	conn.Send("HMSET", versionedName, "leaf", node.Leaf, "size", node.Size)
	for _, child := range node.Children {
		conn.Send("SADD", versionedName+":children", child)
	}
	_, err = conn.Do("EXEC")

	return err
}

func (r *RedisCaching) AddChild(node *Node, child string) error {
	version, err := r.Version()
	if err != nil {
		return err
	}

	conn := r.Pool.Get()
	defer conn.Close()

	versionedName := version + ":" + node.Name

	_, err = conn.Do("SADD", versionedName+":children", child)
	return err
}

func (r *RedisCaching) ReadNode(key string) (*Node, error) {
	version, err := r.Version()
	if err != nil {
		return nil, err
	}

	conn := r.Pool.Get()
	defer conn.Close()

	node := &Node{Name: key}
	key = version + ":" + key

	reply, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}
	if len(reply) == 0 {
		return nil, nil
	}

	if err = redis.ScanStruct(reply, node); err != nil {
		return nil, err
	}
	reply, err = redis.Values(conn.Do("SMEMBERS", key+":children"))
	var children []string
	if err := redis.ScanSlice(reply, &children); err != nil {
		return nil, err
	}

	node.Children = children
	return node, nil
}

func NewRedisCaching(address string) TreeUpdater {
	cacher := &RedisCaching{
		Pool: newPool(address),
	}
	return cacher
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		Wait:        true,
		MaxActive:   10,
		IdleTimeout: 5 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	pool *redis.Pool
)
