package main

import (
	"flag"
	"log"
	"os"
	"pomodoro/internal/toggl"
	"strconv"
	"time"
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

	workspaceIdInt, err := strconv.Atoi(workspaceId)
	if err != nil {
		log.Fatalf("Error converting workspaceId to int: %v", err)
		os.Exit(1)
	}
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		log.Fatalf("Error converting userId to int: %v", err)
		os.Exit(1)
	}

	entry := toggl.NewTogglEntry("7 from pomo.local", time.Now(), time.Now().Add(25*time.Minute), userIdInt, workspaceIdInt)
	if err := entry.Save(token, workspaceIdInt); err != nil {
		log.Fatalf("Error saving entry: %v", err)
		os.Exit(1)
	}
}
