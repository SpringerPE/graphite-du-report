package caching

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

type RedisCaching struct {
	Pool *redis.Pool
}

func (r *RedisCaching) SetNode(node *Node) error {
	conn := r.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET",
		node.Name,
		"leaf", node.Leaf,
		"size", node.Size,
	)
	for _, child := range node.Children {
		_, err := conn.Do("SADD",
			node.Name+":children",
			child)
		if err != nil {
			fmt.Println(err)
		}
	}
	return err
}

func (r *RedisCaching) AddChild(node *Node, child string) error {
	conn := r.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("SADD", node.Name+":children", child)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (r *RedisCaching) GetNode(key string) (*Node, error) {

	node := &Node{Name: key}

	conn := r.Pool.Get()
	defer conn.Close()
	reply, err := redis.Values(conn.Do("HGETALL", key))

	if len(reply) == 0 {
		return nil, nil
	}

	if err != nil {
		fmt.Printf("error:%v , reply:%v", err, reply)
	}

	if err := redis.ScanStruct(reply, node); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("REPLY %#v\n", node)
	if node != nil {
		reply, err = redis.Values(conn.Do("SMEMBERS", key+":children"))
		var children []string
		if err := redis.ScanSlice(reply, &children); err != nil {
			fmt.Println(err)
		}

		node.Children = children
	}
	node.Name = key
	return node, err
}

func NewRedisCaching() Caching {
	return &RedisCaching{
		Pool: newPool("localhost:6379"),
	}
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
