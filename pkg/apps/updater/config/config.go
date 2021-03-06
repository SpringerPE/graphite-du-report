package config

import (
	"strings"
)

type UpdaterConfig struct {
	Servers        []string
	BindAddress    string
	BindPort       string
	RootName       string
	RedisAddr      string
	RedisPasswd    string
	UpdateRoutines int
	BulkUpdates    int
	BulkScans      int
	StoreChildren  bool
}

func ParseServerList(servers string) []string {
	s := strings.Split(servers, ",")

	for index, name := range s {
		s[index] = strings.TrimSpace(name)
	}

	return s
}

func DefaultUpdaterConfig() *UpdaterConfig {
	return &UpdaterConfig{
		Servers:        []string{"127.0.0.1:8080"},
		BindAddress:    "127.0.0.1",
		BindPort:       "6061",
		RootName:       "root",
		RedisAddr:      "localhost:6379",
		RedisPasswd:    "",
		UpdateRoutines: 10,
		BulkUpdates:    100,
		BulkScans:      10,
		StoreChildren:  false,
	}
}
