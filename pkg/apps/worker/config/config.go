package config

type WorkerConfig struct {
	BindAddress string
	BindPort    string
	RootName    string
	RedisAddr   string
	RedisPasswd string
}

func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		BindAddress: "127.0.0.1",
		BindPort:    "6060",
		RootName:    "root",
		RedisAddr:   "localhost:6379",
		RedisPasswd: "",
	}
}
