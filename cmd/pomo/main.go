package main

import (
	"fmt"
	"log"
	"os"

	"pomo.local/internal/cli"
)

// it is filled by -ldflags="-X main.version=$(VERSION)"" in makefile
var version string = "dev"

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}

	if cli.ParseVersionFlag() {
		fmt.Println(version)
		return
	}

	switch os.Args[1] {
	case "start":
		if err := cli.ParseStart(os.Args[2:]).Run(); err != nil {
			log.Fatalf("Error running \"pomo start\": %v", err)
		}
	case "rest":
		if err := cli.ParseRest(os.Args[2:]).Run(); err != nil {
			log.Fatalf("Error running \"pomo start\": %v", err)
		}
	case "notify":
		if err := cli.ParseNotify(os.Args[2:]).Run(); err != nil {
			log.Fatalf("Error running pomo notify: %v", err)
		}
	case "auth":
		if err := cli.ParseAuth(os.Args[2:]).Run(); err != nil {
			log.Fatalf("Failed to auth: %v", err)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command!")
		cli.PrintUsage()
		os.Exit(1)
	}
}
