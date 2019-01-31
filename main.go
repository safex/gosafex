package main

import (
	"github.com/safex/gosafex/cmd"
	"github.com/safex/gosafex/internal/log"
	"github.com/safex/gosafex/pkg/config"
)

func init() {
	config.LoadDefault("GOSAFEX")
	log.LoadDefault()
}

func main() {
	cmd.Execute()
}
