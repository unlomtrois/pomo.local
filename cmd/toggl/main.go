package main

import (
	"flag"
	"log"
	"os"
	"pomodoro/internal/toggl"
)

func main() {
	var token string
	var workspaceId string
	var userId string

	flag.StringVar(&token, "token", "", "--token <token>")
	flag.StringVar(&workspaceId, "workspace", "", "--workspace <workspaceId>")
	flag.StringVar(&userId, "user", "", "--user <userId>")
	flag.Parse()

	if token == "" {
		log.Fatalln("please add token")
		os.Exit(1)
	}

	if err := toggl.AddEntry(token, workspaceId, userId); err != nil {
		log.Fatalf("Error adding entry: %v", err)
		os.Exit(1)
	}
}
