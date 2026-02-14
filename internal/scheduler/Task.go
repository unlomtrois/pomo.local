package scheduler

import "time"

type Task struct {
	ID        string
	ExecuteAt time.Time
	Binary    string
	Args      []string
}
