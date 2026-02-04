package messages

import (
	"encoding/json"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
)

// TasksGetParams represents parameters for tasks/get
type TasksGetParams struct {
	TaskID string `json:"taskId"`
}

// TasksGetResult represents the result of tasks/get
type TasksGetResult struct {
	Task domain.TaskMetadata `json:"task"`
}

// TasksResultParams represents parameters for tasks/result
type TasksResultParams struct {
	TaskID string                 `json:"taskId"`
	Meta   map[string]interface{} `json:"_meta,omitempty"`
}

// TasksListParams represents parameters for tasks/list
type TasksListParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// TasksListResult represents the result of tasks/list
type TasksListResult struct {
	Tasks      []domain.TaskMetadata `json:"tasks"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

// TasksCancelParams represents parameters for tasks/cancel
type TasksCancelParams struct {
	TaskID string `json:"taskId"`
}

// TasksCancelResult represents the result of tasks/cancel
type TasksCancelResult struct {
	Task domain.TaskMetadata `json:"task"`
}

// ParseTasksGetParams parses tasks/get parameters
func ParseTasksGetParams(params json.RawMessage) (*TasksGetParams, error) {
	var p TasksGetParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ParseTasksResultParams parses tasks/result parameters
func ParseTasksResultParams(params json.RawMessage) (*TasksResultParams, error) {
	var p TasksResultParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ParseTasksListParams parses tasks/list parameters
func ParseTasksListParams(params json.RawMessage) (*TasksListParams, error) {
	var p TasksListParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ParseTasksCancelParams parses tasks/cancel parameters
func ParseTasksCancelParams(params json.RawMessage) (*TasksCancelParams, error) {
	var p TasksCancelParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
