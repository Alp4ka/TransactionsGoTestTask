package main

import (
	"server/pkg/commands"
)

func main() {
	if err := commands.MainCmd.Execute(); err != nil {
		panic(err.Error())
	}
}
