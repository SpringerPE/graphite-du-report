package config

type RendererConfig struct {
	BindAddress string
	BindPort    string
}

func DefaultRendererConfig() *RendererConfig {
	return &RendererConfig{
		BindAddress: "127.0.0.1",
		BindPort:    "6062",
	}
}
