package main

import (
	"log"

	"github.com/Azure/git2grid/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
