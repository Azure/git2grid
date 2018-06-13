package main

import (
	"log"

	"github.com/Azure/gittogrid/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
