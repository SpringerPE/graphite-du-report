package config

import (
	"strings"
)

type Config struct {
	Servers []string
	BindAddress string
	BindPort string
}

func ParseServerList(servers string) []string {
	return strings.Split(servers, ",")
}

func DefaultConfig() *Config {
	return &Config {
		Servers: []string{"127.0.0.1:8080"},
		BindAddress: "127.0.0.1",
		BindPort: "6060",
	}
}
