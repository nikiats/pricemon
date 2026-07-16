package main

import (
	"log"

	"gopricemon/internal/startup"
)

func main() {
	config, err := startup.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := startup.Run(*config); err != nil {
		log.Fatal(err)
	}
}
