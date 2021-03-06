package caching

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"encoding/json"
)

var ErrLockMismatch = errors.New("key is locked with a different secret")

const incrScript = `
local v = redis.call("GET", KEYS[1])
if v == false
then
	return redis.call("SET", KEYS[1], 0)
else
	return redis.call("SET", KEYS[1], math.mod(v+1, ARGV[1]))
end
`

const lockScript = `
local v = redis.call("GET", KEYS[1])
if v == false or v == ARGV[1]
then
	return redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2]) and 1
else
	return 0
end
`

const unlockScript = `
local v = redis.call("GET",KEYS[1])
if v == false then
	return 1
elseif v == ARGV[1] then
	return redis.call("DEL",KEYS[1])
else
	return 0
end
`

type Pool interface {
	Get() redis.Conn
	Close() error
}

type RedisCaching struct {
	Pool          Pool
	BulkScans     int
	StoreChildren bool
}

func NewRedisCaching(address string, passwd string, storeChildren bool) *RedisCaching {
	cacher := &RedisCaching{
		Pool:          newPool(address, passwd),
		BulkScans:     10,
		StoreChildren: storeChildren,
	}
	return cacher
}

func newPool(addr string, passwd string) Pool {
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

func (r *RedisCaching) Close() error {
	return r.Pool.Close()
}

// writeLock attempts to grab a redis lock. The error returned is safe to ignore
// if all you care about is whether or not the lock was acquired successfully.
func (r *RedisCaching) WriteLock(name, secret string, ttl uint64) (bool, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	script := redis.NewScript(1, lockScript)
	resp, err := redis.Int(script.Do(conn, name, secret, int64(ttl)))
	if err != nil {
		return false, err
	}
	if resp == 0 {
		return false, ErrLockMismatch
	}
	return true, nil
}

// writeLock releases the redis lock
func (r *RedisCaching) ReleaseLock(name, secret string) (bool, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	script := redis.NewScript(1, unlockScript)
	resp, err := redis.Int(script.Do(conn, name, secret))
	if err != nil {
		return false, err
	}
	if resp == 0 {
		return false, ErrLockMismatch
	}
	return true, nil
}

func (r *RedisCaching) SetNumBulkScans(num int) {
	r.BulkScans = num
}

func (r *RedisCaching) Cleanup(rootName string) error {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.version(conn)
	if err != nil {
		return err
	}

	isGeneratedKey, err := regexp.Compile(
		fmt.Sprintf("(?P<Version>[0-9]+):(%s|folded|json)", rootName))

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
		//scan all the generated keys and delete those with an old version number
		for _, x := range items {
			result := isGeneratedKey.FindStringSubmatch(x)
			if len(result) == 3 {
				if result[1] != version {
					conn.Send("DEL", x)
				}
			}

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

	script := redis.NewScript(1, incrScript)
	_, err := script.Do(conn, "version.next", 100)
	return err
}

func (r *RedisCaching) UpdateReaderVersion() error {
	conn := r.Pool.Get()
	defer conn.Close()

	nextVersion, err := r.versionNext(conn)
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", "version", nextVersion)
	return err
}

func (r *RedisCaching) Version() (string, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.version(conn)
	return version, err
}

func (r *RedisCaching) version(conn redis.Conn) (string, error) {
	version, err := redis.String(conn.Do("GET", "version"))
	return version, err
}

func (r *RedisCaching) versionNext(conn redis.Conn) (string, error) {
	version, err := redis.String(conn.Do("GET", "version.next"))
	return version, err
}

func (r *RedisCaching) UpdateJson(root *Node) error {

	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.versionNext(conn)
	if err != nil {
		return err
	}

	jsonTree, err := json.Marshal(root)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", version+":json", jsonTree)
	return err
}


/* Stores a list of nodes and their children (if StoreChildren is set to true) into the redis
datastore.

Nodes are stored in a dictionary whose entries look like: "version:node-name": {"leaf", bool, "size", int64)}
Children are stored in a list: "version:node-name:children": ["name1", "name2"]

This method stores a folded data representation suitable for direct use when creating flame graphs.
 */
func (r *RedisCaching) UpdateNodes(nodes []*Node) error {
	var nodeEntry string

	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.versionNext(conn)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	for _, node := range nodes {
		versionedName := version + ":" + node.Name
		conn.Send("HMSET", versionedName, "leaf", node.Leaf, "size", node.Size)

		//save in folded format as well for flame graphs
		entryName := strings.Replace(node.Name, ".", ";", -1)
		if node.Leaf == true {
			nodeEntry = fmt.Sprintf("%s %d", entryName, node.Size)
			conn.Send("LPUSH", version+":folded", nodeEntry)
		}

		if r.StoreChildren {
			for _, child := range node.Children {
				conn.Send("SADD", versionedName+":children", child.Name)
			}
		}
	}
	_, err = conn.Do("EXEC")

	return err
}

func (r *RedisCaching) AddChild(node *Node, child string) error {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.versionNext(conn)
	if err != nil {
		return err
	}

	versionedName := version + ":" + node.Name

	_, err = conn.Do("SADD", versionedName+":children", child)
	return err
}

func (r *RedisCaching) ReadJsonTree() ([]byte, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.version(conn)
	if err != nil {
		return nil, err
	}

	return redis.Bytes(conn.Do("GET", version+":json"))
}

func (r *RedisCaching) ReadFlameMap() ([]string, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.version(conn)
	if err != nil {
		return nil, err
	}

	return redis.Strings(conn.Do("LRANGE", version+":folded", 0, -1))
}

func (r *RedisCaching) ReadNode(key string) (*Node, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	version, err := r.version(conn)
	if err != nil {
		return nil, err
	}

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

	return node, nil
}

