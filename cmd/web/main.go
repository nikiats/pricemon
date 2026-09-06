package main

import (
	"log"
	"net/http"

	"pricemon/internal/web"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", web.StaticHandler()))
}
