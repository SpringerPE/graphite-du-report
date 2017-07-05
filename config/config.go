package config

import (
	"strings"
)

type Config struct {
	Servers     []string
	BindAddress string
	BindPort    string
	RootName    string
	RedisAddr   string
}

func ParseServerList(servers string) []string {
	s := strings.Split(servers, ",")

	for index, name := range s {
		s[index] = strings.TrimSpace(name)
	}

	return s
}

func DefaultConfig() *Config {
	return &Config{
		Servers:     []string{"127.0.0.1:8080"},
		BindAddress: "127.0.0.1",
		BindPort:    "6060",
		RootName:    "root",
		RedisAddr:   "localhost:6379",
	}
}
