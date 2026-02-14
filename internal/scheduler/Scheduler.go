package scheduler

type Scheduler interface {
	Schedule(task Task) error
	Cancel(taskID string) error
}
