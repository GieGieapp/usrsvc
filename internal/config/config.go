package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port      string
	PGDSN     string
	CORSAllow []string
}

func Load() Config {
	port := getenv("APP_PORT", "8080")
	dsn := os.Getenv("PG_DSN")
	if strings.TrimSpace(dsn) == "" {
		user := getenv("DB_USER", "usrsvc")
		pass := getenv("DB_PASS", "secret")
		host := getenv("DB_HOST", "localhost")
		portDB := getenv("DB_PORT", "5432")
		name := getenv("DB_NAME", "usrsvc")
		ssl := getenv("DB_SSLMODE", "disable")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, portDB, name, ssl)
	}
	var cors []string
	if s := os.Getenv("CORS_ALLOW_ORIGINS"); s != "" {
		for _, p := range strings.Split(s, ",") {
			if v := strings.TrimSpace(p); v != "" {
				cors = append(cors, v)
			}
		}
	}
	return Config{Port: port, PGDSN: dsn, CORSAllow: cors}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
