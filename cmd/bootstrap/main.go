package main

import (
	"log"

	"github.com/gee-coder/template-go-backend/internal/bootstrap"
)

func main() {
	if err := bootstrap.RunDatabaseBootstrap(); err != nil {
		log.Fatal(err)
	}
}
