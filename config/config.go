package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	CountWorkers int
	QueueSize    int
}

func LoadConfig() Config {

	var cfg = Config{CountWorkers: 4, QueueSize: 64}

	if val := os.Getenv("COUNT_WORKERS"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			cfg.CountWorkers = v
		} else {
			log.Printf("некорректное значение WORKERS=%s, используется default=%d", val, cfg.CountWorkers)
		}
	}
	if val := os.Getenv("QUEUE_SIZE"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			cfg.QueueSize = v
		} else {
			log.Printf("некорректное значение QUEUE_SIZE=%s, используется default=%d", val, cfg.QueueSize)
		}
	}

	return cfg
}
