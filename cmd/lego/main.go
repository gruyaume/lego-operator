package main

import (
	"os"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/lego-operator/internal/charm"
)

func main() {
	err := charm.Configure()
	if err != nil {
		goops.LogErrorf("could not configure charm: %v", err)
		os.Exit(1)
	}
}
