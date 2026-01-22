package toggl

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func AddEntry(token string, workspaceId string, userId string) error {
	entry := fmt.Appendf(nil, `{
		"billable":false,
		"created_with":"pomo.local",
		"description": "4 from pomo.local",
		"duration": null,
		"project_id": null,
		"shared_with_user_ids":[],
		"start":"2026-01-22T8:00:00Z",
		"stop":"2026-01-22T9:00:00Z",
		"user_id": %s,
		"workspace_id": %s
	}`, userId, workspaceId)

	url := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%s/time_entries", workspaceId)
	reqBody := bytes.NewBuffer(entry)

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

	quotaRemaining := resp.Header.Get("X-Toggl-Quota-Remaining")
	fmt.Printf("Quota remaining: %s requests\n", quotaRemaining)

	quotaResetsIn := resp.Header.Get("X-Toggl-Quota-Resets-In")
	fmt.Printf("Quota resets in: %s seconds\n", quotaResetsIn)

	defer resp.Body.Close()
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
