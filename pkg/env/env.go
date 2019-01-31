package env

import (
	"log"
	"os"
)

func MustGetEnv(k string) string {
	v, ok := os.LookupEnv(k)
	if !ok {
		log.Fatalf("env var '%s' not set", k)
	}

	return v
}
