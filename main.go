package main

import (
	"log"

	"github.com/Azure/git2grid/server"
)

func main() {
	app := server.App()
	
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
