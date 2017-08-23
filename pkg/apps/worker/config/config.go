package config

type WorkerConfig struct {
	BindAddress string `json:"bind_address"`
	BindPort    string `json:"bind_port"`
	RootName    string `json:"root_name"`
	RedisAddr   string `json:"redis_addr"`
	RedisPasswd string `json:"redis_passwd"`
	RetrieveChildren bool `json:"retrieve_children"`
	TemplatesFolder string `json: "templates_folder"`
}

func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		BindAddress: "127.0.0.1",
		BindPort:    "6060",
		RootName:    "root",
		RedisAddr:   "localhost:6379",
		RedisPasswd: "",
		RetrieveChildren: false,
		TemplatesFolder: "./templates",
	}
}
