package grifts

import (
	"github.com/Azure/gittogrid/actions"
	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
