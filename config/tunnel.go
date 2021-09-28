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
type RedisConfigConfig struct {
	Host            string
	Username        string
	Password        string
	RedisConfigName string
	PortNum         int
}

func initRedisConfigConfig() RedisConfigConfig {
	host := os.Getenv("DB_HOST")
	if host == "" {
		//panic("Config Error: DB_HOST must be set")
	}
	username := os.Getenv("DB_USER")
	if username == "" {
		//panic("Config Error: DB_USER must be set")
	}
	password := os.Getenv("DB_PASS")
	if password == "" {
		//panic("Config Error: DB_PASS must be set")
	}
	databaseName := os.Getenv("DB_NAME")
	if databaseName == "" {
		//panic("Config Error: DB_NAME must be set")
	}
	portNumStr := os.Getenv("DB_PORT")
	if portNumStr == "" {
		//panic("Config Error: DB_PORT must be set")
	}
	portNumInt, err := strconv.Atoi(portNumStr)
	if err != nil {
		//panic("Config Error: DB_PORT must be valid integer")
	}
	return RedisConfigConfig{
		Host:            host,
		Username:        username,
		Password:        password,
		RedisConfigName: databaseName,
		PortNum:         portNumInt,
	}
}

type GlobalConfig struct {
	AppConfig
	RedisConfigConfig
}

var Config GlobalConfig

// ----------------------------------------------------

func NewConfig() (*GlobalConfig, error) {
	tmp := GlobalConfig{
		AppConfig:         initAppConfig(),
		RedisConfigConfig: initRedisConfigConfig(),
	}
	Config = tmp
	return &Config, nil
}
