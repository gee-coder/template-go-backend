package config

import "os"

func getConfigPath() string {
	if path := os.Getenv("APP_CONFIG"); path != "" {
		return path
	}
	return "configs/config.local.yaml"
}

