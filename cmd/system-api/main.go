package main

import (
	"log"

	"github.com/gee-coder/template-go-backend/internal/api"
	"github.com/gee-coder/template-go-backend/internal/bootstrap"
)

func main() {
	if err := bootstrap.RunHTTP(false, api.NewSystemRouter); err != nil {
		log.Fatal(err)
	}
}
