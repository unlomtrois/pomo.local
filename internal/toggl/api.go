package toggl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UTCTime struct {
	time.Time
}

func (t *UTCTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Format(time.RFC3339))
}

type TogglEntry struct { // not full
	Billable          bool    `json:"billable"`
	CreatedWith       string  `json:"created_with"`
	Description       string  `json:"description"`
	Duration          int     `json:"duration,omitempty"`
	ProjectId         int     `json:"project_id,omitempty"`
	SharedWithUserIds []int   `json:"shared_with_user_ids,omitempty"`
	Start             UTCTime `json:"start"`
	Stop              UTCTime `json:"stop"`
	UserId            int     `json:"user_id"`
	WorkspaceId       int     `json:"workspace_id"`
}

func NewTogglEntry(description string, start time.Time, stop time.Time, userId int, workspaceId int) *TogglEntry {
	return &TogglEntry{
		CreatedWith: "pomo.local (https://github.com/unlomtrois/pomo.local)",
		Description: description,
		Start:       UTCTime{start},
		Stop:        UTCTime{stop},
		UserId:      userId,
		WorkspaceId: workspaceId,
	}
}

func (entry *TogglEntry) Save(token string, workspaceId int) error {
	entryJson, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("Error marshalling entry: %v", err)
	}

	url := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries", workspaceId)
	reqBody := bytes.NewBuffer(entryJson)
	req, err := http.NewRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(token, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	quotaRemaining := resp.Header.Get("X-Toggl-Quota-Remaining")
	fmt.Printf("Quota remaining: %s requests\n", quotaRemaining)

	quotaResetsIn := resp.Header.Get("X-Toggl-Quota-Resets-In")
	fmt.Printf("Quota resets in: %s seconds\n", quotaResetsIn)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response: %v", err)
	}

	fmt.Printf("Response: %s\n", string(respBody))
	return nil
}

func CurrentEntry(token string) ([]byte, error) {
	url := "https://api.track.toggl.com/api/v9/me/time_entries/current"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.SetBasicAuth(token, "api_token")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}
	return body, nil
}
