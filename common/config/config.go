package config

import "github.com/ariadata/rabbit-connect/common/cipher"

type Config struct {
	LocalAddr  string
	ServerAddr string
	CIDR       string
	Key        string
	ServerMode bool
}

func (config *Config) Init() {
	cipher.GenerateKey(config.Key)
}
