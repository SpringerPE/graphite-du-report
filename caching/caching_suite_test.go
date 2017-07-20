package caching_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/SpringerPE/graphite-du-report/caching"

	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"

	"testing"
)

func TestCaching(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Caching Suite")
}

type MockPool struct {
	conn *redigomock.Conn
}

func newMockPool(c *redigomock.Conn) Pool {
	return &MockPool{
		conn: c,
	}
}

func (p *MockPool) Get() redis.Conn {
	return p.conn
}

func (p *MockPool) Close() error {
	return nil
}
