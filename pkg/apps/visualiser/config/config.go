package config

type VisualiserConfig struct {
	BindAddress     string `json:"bind_address"`
	BindPort        string `json:"bind_port"`
	TemplatesFolder string `json: "templates_folder"`
}

func DefaultVisualiserConfig() *VisualiserConfig {
	return &VisualiserConfig{
		BindAddress:     "127.0.0.1",
		BindPort:        "6063",
		TemplatesFolder: "./templates",
	}
}
