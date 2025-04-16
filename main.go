/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"nsl-graph/cmd"
	_ "nsl-graph/cmd/diagram"
	_ "nsl-graph/cmd/modify"
	_ "nsl-graph/cmd/print"
	_ "nsl-graph/cmd/root"
)

func main() {
	cmd.Execute()
}
