package scheduler

import "time"

// Task describes a scheduled notification command to run at a specific time.
type Task struct {
	ID        string    `json:"id"`
	ExecuteAt time.Time `json:"executed_at"`
	Binary    string    `json:"bin"`
	Args      []string  `json:"args"`
}
