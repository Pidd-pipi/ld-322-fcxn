package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv     string `env:"APP_ENV" envDefault:"development"`
	ServerPort int    `env:"SERVER_PORT" envDefault:"8080"`
	DBHost     string `env:"DB_HOST" envDefault:"127.0.0.1"`
	DBPort     int    `env:"DB_PORT" envDefault:"3306"`
	DBName     string `env:"DB_NAME" envDefault:"greenhouse"`
	DBUser     string `env:"DB_USER" envDefault:"greenhouse"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"greenhouse_pwd"`
	RedisHost  string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort  int    `env:"REDIS_PORT" envDefault:"6379"`
	JWTSecret  string `env:"JWT_SECRET" envDefault:"development-secret-change-me"`
}

func Load() (Config, error) { return env.ParseAs[Config]() }
func (c Config) MySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
func (c Config) RedisAddress() string { return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort) }
