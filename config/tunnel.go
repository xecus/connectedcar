package config

import (
	"os"
	"strconv"
)

// アプリ関連設定
type AppConfig struct {
	Env                 string
	SshServerListenPort string
	SentryDsn           string
	SentryEnv           string
}

func initAppConfig() AppConfig {
	sshServerListenPort := os.Getenv("SSH_SERVER_LISTEN_PORT")
	if sshServerListenPort == "" {
		sshServerListenPort = ":2222"
	}
	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn == "" {
		sentryDsn = ""
	}
	sentryEnv := os.Getenv("SENTRY_ENV")
	if sentryEnv == "" {
		sentryEnv = "dev"
	}
	return AppConfig{
		Env:                 "local",
		SshServerListenPort: sshServerListenPort,
		SentryDsn:           sentryDsn,
		SentryEnv:           sentryEnv,
	}
}

// データベース関連設定
type RedisConfig struct {
	Addr     string
	Password string
	Database int
}

func initRedisConfig() RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		//panic("Config Error: DB_HOST must be set")
	}
	password := os.Getenv("REDIS_PASSWORDS")
	if password == "" {
		//panic("Config Error: DB_PASS must be set")
	}
	databaseStr := os.Getenv("REDIS_DATABASE")
	if databaseStr == "" {
		//panic("Config Error: DB_PORT must be set")
	}
	databaseInt, err := strconv.Atoi(databaseStr)
	if err != nil {
		//panic("Config Error: DB_PORT must be valid integer")
	}
	return RedisConfig{
		Addr:     addr,
		Password: password,
		Database: databaseInt,
	}
}

type GlobalConfig struct {
	AppConfig
	RedisConfig
}

var Config GlobalConfig

// ----------------------------------------------------

func NewConfig() (*GlobalConfig, error) {
	tmp := GlobalConfig{
		AppConfig:   initAppConfig(),
		RedisConfig: initRedisConfig(),
	}
	Config = tmp
	return &Config, nil
}
