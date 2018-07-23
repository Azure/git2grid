package grifts

import (
	"github.com/Azure/git2grid/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
