package main

import (
	// Flag package for command-line parsing
	"flag"
	"fmt"
)

func main() {
	// Define and parse command-line flag for level number
	// Order - flag name, default value, description
	level := flag.Int("level", 0, "Level number to display")
	flag.Parse()

	fmt.Println("Go Infra Quest: Level", *level, ". GG Noobs!")
}
