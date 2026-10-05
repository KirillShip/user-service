package config

import (
	"log"
	"os"
)

type Config struct {
	Port         string
	Address      string
	DatabaseDSN  string
	JWTSecretkey []byte
}

func LoadConfig() *Config {
	dsn, exist := os.LookupEnv("DatabaseDSN")
	if !exist {
		log.Println("Error in loading DatabaseDSN config")
		return nil
	}
	jwtKey, exist := os.LookupEnv("JWTSecretkey")
	if !exist {
		log.Println("Error in loading JWTSecretkey config")
		return nil
	}
	return &Config{
		Port:         GetEnv("Port", "8080"),
		Address:      GetEnv("Address", "localhost"),
		DatabaseDSN:  dsn,
		JWTSecretkey: []byte(jwtKey),
	}
}

func GetEnv(key string, defaultValue string) string {
	val, exist := os.LookupEnv(key)
	if !exist {
		val = defaultValue
	}
	return val
}
