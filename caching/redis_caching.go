package caching

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"regexp"
	"strings"
	"time"
)

type RedisCaching struct {
	Pool      *redis.Pool
	BulkScans int
}

func (r *RedisCaching) SetNumBulkScans(num int) {
	r.BulkScans = num
}

func (r *RedisCaching) Cleanup(rootName string) error {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.Version()
	if err != nil {
		return err
	}
	rxp, _ := regexp.Compile(fmt.Sprintf("%s:%s*", version, rootName))
	rxp_folded, _ := regexp.Compile(fmt.Sprintf("%s:%s*", version, "folded"))

	if err != nil {
		return err
	}

	var (
		cursor int64
		items  []string
	)

	for {
		values, err := redis.Values(conn.Do("SCAN", cursor, "count", 5000))
		if err != nil {
			return err
		}

		values, err = redis.Scan(values, &cursor, &items)
		if err != nil {
			return err
		}

		conn.Send("MULTI")
		for _, x := range items {
			if rxp.MatchString(x) {
				continue
			}
			if rxp_folded.MatchString(x) {
				continue
			}
			if strings.HasPrefix(x, "version") {
				continue
			}
			conn.Send("DEL", x)
		}
		_, err = conn.Do("EXEC")
		if err != nil {
			return err
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}

func (r *RedisCaching) IncrVersion() error {
	conn := r.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("INCR", "version.next")
	return err
}

func (r *RedisCaching) UpdateReaderVersion() error {
	conn := r.Pool.Get()
	defer conn.Close()

	next_version, err := redis.String(conn.Do("GET", "version.next"))
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", "version", next_version)
	return err
}

func (r *RedisCaching) Version() (string, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := redis.String(conn.Do("GET", "version"))
	return version, err
}

func (r *RedisCaching) VersionNext() (string, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := redis.String(conn.Do("GET", "version.next"))
	return version, err
}

func (r *RedisCaching) UpdateNodes(nodes []*Node) error {
	version, err := r.VersionNext()
	if err != nil {
		return err
	}

	conn := r.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	for _, node := range nodes {
		versionedName := version + ":" + node.Name
		conn.Send("HMSET", versionedName, "leaf", node.Leaf, "size", node.Size)
		conn.Send("HMSET", version+":folded", node.Name, node.Size)
		for _, child := range node.Children {
			conn.Send("SADD", versionedName+":children", child)
		}
	}
	_, err = conn.Do("EXEC")

	return err
}

func (r *RedisCaching) AddChild(node *Node, child string) error {
	version, err := r.VersionNext()
	if err != nil {
		return err
	}

	conn := r.Pool.Get()
	defer conn.Close()

	versionedName := version + ":" + node.Name

	_, err = conn.Do("SADD", versionedName+":children", child)
	return err
}

func (r *RedisCaching) ReadFlameMap() (map[string]int64, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.Version()
	if err != nil {
		return nil, err
	}

	return redis.Int64Map(conn.Do("HGETALL", version+":folded"))
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

func NewRedisCaching(address string, passwd string) *RedisCaching {
	cacher := &RedisCaching{
		Pool:      newPool(address, passwd),
		BulkScans: 10,
	}
	return cacher
}

func newPool(addr string, passwd string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		Wait:        true,
		MaxActive:   10,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				addr,
				redis.DialPassword(passwd),
				redis.DialConnectTimeout(10*time.Second),
			)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

var (
	pool *redis.Pool
)
