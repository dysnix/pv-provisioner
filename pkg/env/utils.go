package env

import (
	"log"
	"os"
)

func GetEnvOrPanic(key string) string {
	var value = os.Getenv(key)
	if value == "" {
		log.Panicf("Error get env value by name '%v'", key)
	}
	return value
}
