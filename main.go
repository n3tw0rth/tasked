package main

import (
	"fmt"
	"os"

	"github.com/n3tw0rth/tasked/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
