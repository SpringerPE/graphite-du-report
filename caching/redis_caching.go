package caching

import (
	"fmt"
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
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (r *RedisCaching) Version() (string, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := redis.String(conn.Do("GET", "version"))
	if err != nil {
		fmt.Println(err)
	}
	return version, err
}

func (r *RedisCaching) SetNode(node *Node) error {
	conn := r.Pool.Get()
	defer conn.Close()

	version, _ := r.Version()
	versionedName := version + ":" + node.Name

	_, err := conn.Do("HMSET", versionedName, "leaf", node.Leaf, "size", node.Size)
	for _, child := range node.Children {
		_, err := conn.Do("SADD", versionedName+":children", child)
		if err != nil {
			fmt.Println(err)
		}
	}
	return err
}

func (r *RedisCaching) AddChild(node *Node, child string) error {
	conn := r.Pool.Get()
	defer conn.Close()

	version, _ := r.Version()
	versionedName := version + ":" + node.Name

	_, err := conn.Do("SADD", versionedName+":children", child)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (r *RedisCaching) GetNode(key string) (*Node, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	node := &Node{Name: key}

	version, _ := r.Version()
	key = version + ":" + key

	reply, err := redis.Values(conn.Do("HGETALL", key))

	if err != nil {
		fmt.Printf("error:%v , reply:%v", err, reply)
	}

	if len(reply) == 0 {
		return nil, nil
	}

	if err := redis.ScanStruct(reply, node); err != nil {
		fmt.Println(err)
	}
	fmt.Println(key + ":children")
	reply, err = redis.Values(conn.Do("SMEMBERS", key+":children"))
	var children []string
	if err := redis.ScanSlice(reply, &children); err != nil {
		fmt.Println(err)
	}

	node.Children = children
	return node, err
}

func NewRedisCaching() Caching {
	cacher := &RedisCaching{
		Pool: newPool("localhost:6379"),
	}
	return cacher
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	pool *redis.Pool
)
