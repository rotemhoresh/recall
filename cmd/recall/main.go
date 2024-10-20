package main

import (
	"fmt"

	"github.com/rotemhoresh/recall/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Println(err)
	}
}
