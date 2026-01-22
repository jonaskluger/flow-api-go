package flowapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Entity represents a generic Flow entity
type Entity map[string]interface{}

// FindEntities searches for entities of a given type with optional filters
func (c *Client) FindEntities(entityType string, filters interface{}, fields []string) ([]Entity, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build request body
	body := make(map[string]interface{})
	if filters != nil {
		body["filters"] = filters
	} else {
		// Empty filters array if none provided
		body["filters"] = []interface{}{}
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Construct URL with fields parameter
	url := fmt.Sprintf("%s/api/%s/entity/%s/_search", c.baseURL, c.apiVersion, entityType)
	if len(fields) > 0 {
		url += "?fields=" + strings.Join(fields, ",")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd+shotgun.api3_array+json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			ID            int                    `json:"id"`
			Type          string                 `json:"type"`
			Attributes    map[string]interface{} `json:"attributes"`
			Relationships map[string]interface{} `json:"relationships"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to Entity slice
	entities := make([]Entity, len(result.Data))
	for i, item := range result.Data {
		entity := Entity{
			"id":   item.ID,
			"type": item.Type,
		}
		// Merge attributes
		for k, v := range item.Attributes {
			entity[k] = v
		}
		// Merge relationships
		for k, v := range item.Relationships {
			entity[k] = v
		}
		entities[i] = entity
	}

	return entities, nil
}

// GetEntity retrieves a single entity by ID
func (c *Client) GetEntity(entityType string, id int, fields []string) (Entity, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Construct URL with fields parameter
	url := fmt.Sprintf("%s/api/%s/entity/%s/%d", c.baseURL, c.apiVersion, entityType, id)
	if len(fields) > 0 {
		url += "?fields=" + strings.Join(fields, ",")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ID            int                    `json:"id"`
			Type          string                 `json:"type"`
			Attributes    map[string]interface{} `json:"attributes"`
			Relationships map[string]interface{} `json:"relationships"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to Entity
	entity := Entity{
		"id":   result.Data.ID,
		"type": result.Data.Type,
	}
	// Merge attributes
	for k, v := range result.Data.Attributes {
		entity[k] = v
	}
	// Merge relationships
	for k, v := range result.Data.Relationships {
		entity[k] = v
	}

	return entity, nil
}

// CreateEntity creates a new entity
func (c *Client) CreateEntity(entityType string, data map[string]interface{}) (Entity, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	bodyJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/%s/entity/%s", c.baseURL, c.apiVersion, entityType)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ID            int                    `json:"id"`
			Type          string                 `json:"type"`
			Attributes    map[string]interface{} `json:"attributes"`
			Relationships map[string]interface{} `json:"relationships"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to Entity
	entity := Entity{
		"id":   result.Data.ID,
		"type": result.Data.Type,
	}
	// Merge attributes
	for k, v := range result.Data.Attributes {
		entity[k] = v
	}
	// Merge relationships
	for k, v := range result.Data.Relationships {
		entity[k] = v
	}

	return entity, nil
}

func (c *Client) GetUserByLogin(login string) (Entity, error) {
	filters := []interface{}{
		[]interface{}{"login", "is", login},
	}

	users, err := c.FindEntities("human_users", filters, []string{"id", "name", "login", "email"})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with login: %s", login)
	}

	return users[0], nil
}

func (c *Client) GetUserByName(name string) (Entity, error) {
	filters := []interface{}{
		[]interface{}{"name", "is", name},
	}

	users, err := c.FindEntities("human_users", filters, []string{"id", "name", "login", "email"})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with name: %s", name)
	}

	return users[0], nil
}

func (c *Client) GetShots(projectID int, fields []string) ([]Entity, error) {
	var filters interface{}

	if projectID > 0 {
		filters = []interface{}{
			[]interface{}{"project", "is", map[string]interface{}{
				"type": "Project",
				"id":   projectID,
			}},
		}
	}

	if len(fields) == 0 {
		fields = []string{"code", "description", "sg_status_list"}
	}

	return c.FindEntities("shots", filters, fields)
}

func (c *Client) GetTasksForShot(shotID int, fields []string) ([]Entity, error) {
	filters := []interface{}{
		[]interface{}{"entity", "is", map[string]interface{}{
			"type": "Shot",
			"id":   shotID,
		}},
	}

	if len(fields) == 0 {
		fields = []string{"content", "sg_status_list", "task_assignees"}
	}

	return c.FindEntities("tasks", filters, fields)
}

func (c *Client) GetTasksForUser(userID int, fields []string) ([]Entity, error) {
	filters := []interface{}{
		[]interface{}{"task_assignees", "is", map[string]interface{}{
			"type": "HumanUser",
			"id":   userID,
		}},
	}

	if len(fields) == 0 {
		fields = []string{"content", "entity", "sg_status_list", "project"}
	}

	return c.FindEntities("tasks", filters, fields)
}

func (c *Client) GetUserShotTasks(userID int, shotID int, fields []string) ([]Entity, error) {
	filters := []interface{}{
		[]interface{}{"entity", "is", map[string]interface{}{
			"type": "Shot",
			"id":   shotID,
		}},
		[]interface{}{"task_assignees", "is", map[string]interface{}{
			"type": "HumanUser",
			"id":   userID,
		}},
	}

	if len(fields) == 0 {
		fields = []string{"content", "sg_status_list", "task_assignees"}
	}

	return c.FindEntities("tasks", filters, fields)
}

func (c *Client) GetShotsForUser(userID int, fields []string) ([]Entity, error) {
	// First get all tasks for the user
	tasks, err := c.GetTasksForUser(userID, []string{"entity"})
	if err != nil {
		return nil, err
	}

	// Extract unique shot IDs
	shotIDs := make(map[int]bool)
	for _, task := range tasks {
		if entity, ok := task["entity"].(map[string]interface{}); ok {
			// Check if relationship is wrapped in a "data" field
			if data, ok := entity["data"].(map[string]interface{}); ok {
				entity = data
			}

			if entityType, ok := entity["type"].(string); ok && entityType == "Shot" {
				if id, ok := entity["id"].(float64); ok {
					shotIDs[int(id)] = true
				} else if id, ok := entity["id"].(int); ok {
					shotIDs[id] = true
				}
			}
		}
	}

	if len(shotIDs) == 0 {
		return []Entity{}, nil
	}

	// Convert map keys to slice
	ids := make([]int, 0, len(shotIDs))
	for id := range shotIDs {
		ids = append(ids, id)
	}

	// Get shot details
	filters := []interface{}{
		[]interface{}{"id", "in", ids},
	}

	if len(fields) == 0 {
		fields = []string{"code", "description", "sg_status_list", "project"}
	}

	return c.FindEntities("shots", filters, fields)
}

// NewClientFromEnv creates a new client using environment variables
// It will automatically try to load a .env file from common locations
func NewClientFromEnv() (*Client, error) {
	// Try to load .env file (silently fail if not found)
	tryLoadEnv()

	siteURL := getEnv("FLOW_SITE_URL")
	scriptName := getEnv("FLOW_SCRIPT_NAME")
	scriptKey := getEnv("FLOW_SCRIPT_KEY")

	if siteURL == "" {
		return nil, fmt.Errorf("FLOW_SITE_URL environment variable is required")
	}
	if scriptName == "" {
		return nil, fmt.Errorf("FLOW_SCRIPT_NAME environment variable is required")
	}
	if scriptKey == "" {
		return nil, fmt.Errorf("FLOW_SCRIPT_KEY environment variable is required")
	}

	return NewClient(Config{
		SiteURL:    siteURL,
		ScriptName: scriptName,
		ScriptKey:  scriptKey,
	})
}

// tryLoadEnv attempts to load .env files from common locations
func tryLoadEnv() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")
}

// getEnv gets an environment variable
func getEnv(key string) string {
	return os.Getenv(key)
}
