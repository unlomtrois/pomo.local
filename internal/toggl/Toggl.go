package toggl

import "pomo.local/internal/pomo"

type Toggl struct{}

func InitToggl() *Toggl {
	panic("todo")
}

func (t *Toggl) Save(pomodoro *pomo.Pomodoro) {
	panic("todo")
}

// func (p *Pomodoro) SaveInToggl(token string, workspaceId int, userId int) error {
// 	entry := toggl.NewTogglEntry(p.Title, p.StartTime, p.StopTime, userId, workspaceId)
// 	if err := entry.Save(token, workspaceId); err != nil {
// 		return fmt.Errorf("Error saving entry: %v", err)
// 	}
// 	return nil
// }
