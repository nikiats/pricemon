package main

import (
	"log"
	"os"

	"gopricemon/internal/migrations"
	"gopricemon/internal/startup"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: migrate <up|down>")
	}

	config, err := startup.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "up":
		err = migrations.Up(config.DatabaseURL)
	case "down":
		err = migrations.Down(config.DatabaseURL)
	default:
		log.Fatal("usage: migrate <up|down>")
	}
	if err != nil {
		log.Fatal(err)
	}
}
