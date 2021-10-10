package adapter

import "github.com/xecus/connectedcar/config"

type KvsRwInterface interface {
	Init(globalConfig *config.GlobalConfig)
	Write(key, value string) error
	Read(key string) (string, error)
}
