package toggl

import "pomo.local/internal/pomo"

// Toggl integrates session saving with the Toggl Track service.
type Toggl struct{}

// InitToggl initializes the Toggl integration.
func InitToggl() *Toggl {
	panic("todo")
}

// Save persists the session to Toggl Track.
func (t *Toggl) Save(_ *pomo.Session) {
	panic("todo")
}

// func (p *Pomodoro) SaveInToggl(token string, workspaceId int, userId int) error {
// 	entry := toggl.NewTogglEntry(p.Title, p.StartTime, p.StopTime, userId, workspaceId)
// 	if err := entry.Save(token, workspaceId); err != nil {
// 		return fmt.Errorf("Error saving entry: %v", err)
// 	}
// 	return nil
// }
