// Package toggl provides the Toggl Track API client for saving time entries.
package toggl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UTCTime wraps time.Time to marshal as RFC3339 in JSON.
type UTCTime struct {
	time.Time
}

// MarshalJSON encodes the time as an RFC3339 string.
func (t *UTCTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(time.RFC3339))
}

// Entry represents a Toggl Track time entry (partial fields).
type Entry struct {
	Billable          bool    `json:"billable"`
	CreatedWith       string  `json:"created_with"`
	Description       string  `json:"description"`
	Duration          int     `json:"duration,omitempty"`
	ProjectID         int     `json:"project_id,omitempty"`
	SharedWithUserIDs []int   `json:"shared_with_user_ids,omitempty"`
	Start             UTCTime `json:"start"`
	Stop              UTCTime `json:"stop"`
	UserID            int     `json:"user_id"`
	WorkspaceID       int     `json:"workspace_id"`
}

// NewEntry constructs a Entry for the given session details.
func NewEntry(description string, start time.Time, stop time.Time, userID int, workspaceID int) *Entry {
	return &Entry{
		CreatedWith: "pomo.local (https://github.com/unlomtrois/pomo.local)",
		Description: description,
		Start:       UTCTime{start},
		Stop:        UTCTime{stop},
		UserID:      userID,
		WorkspaceID: workspaceID,
	}
}

// Save posts the time entry to the Toggl Track API.
func (entry *Entry) Save(token string, workspaceID int) error {
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error marshalling entry: %v", err)
	}

	url := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries", workspaceID)
	reqBody := bytes.NewBuffer(entryJSON)
	req, err := http.NewRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(token, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	quotaRemaining := resp.Header.Get("X-Toggl-Quota-Remaining")
	fmt.Printf("Quota remaining: %s requests\n", quotaRemaining)

	quotaResetsIn := resp.Header.Get("X-Toggl-Quota-Resets-In")
	fmt.Printf("Quota resets in: %s seconds\n", quotaResetsIn)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	fmt.Printf("Response: %s\n", string(respBody))
	return nil
}

// CurrentEntry fetches the currently running Toggl time entry.
func CurrentEntry(token string) ([]byte, error) {
	url := "https://api.track.toggl.com/api/v9/me/time_entries/current"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.SetBasicAuth(token, "api_token")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}
	return body, nil
}
