package scheduler

import "time"

type Task struct {
	ID        string    `json:"id"`
	ExecuteAt time.Time `json:"executed_at"`
	Binary    string    `json:"bin"`
	Args      []string  `json:"args"`
}
