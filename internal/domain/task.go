package domain

import (
	"encoding/json"
	"time"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	// TaskStatusWorking indicates the task is currently being processed
	TaskStatusWorking TaskStatus = "working"

	// TaskStatusInputRequired indicates the task needs input before continuing
	TaskStatusInputRequired TaskStatus = "input_required"

	// TaskStatusCompleted indicates the task completed successfully
	TaskStatusCompleted TaskStatus = "completed"

	// TaskStatusFailed indicates the task failed
	TaskStatusFailed TaskStatus = "failed"

	// TaskStatusCancelled indicates the task was canceled
	TaskStatusCancelled TaskStatus = "canceled"
)

// IsTerminal returns true if the status is a terminal state
func (s TaskStatus) IsTerminal() bool {
	return s == TaskStatusCompleted || s == TaskStatusFailed || s == TaskStatusCancelled
}

// TaskRequest represents the task field in request parameters
type TaskRequest struct {
	// TTL is the requested time-to-live in milliseconds
	TTL int64 `json:"ttl,omitempty"`
}

// CreateTaskResult is returned when a task is created
type CreateTaskResult struct {
	Task TaskMetadata `json:"task"`
}

// TaskMetadata contains information about a task
type TaskMetadata struct {
	// TaskID is the unique identifier for the task
	TaskID string `json:"taskId"`

	// Status is the current status of the task
	Status TaskStatus `json:"status"`

	// StatusMessage provides additional context about the status
	StatusMessage string `json:"statusMessage,omitempty"`

	// CreatedAt is the ISO 8601 timestamp when the task was created
	CreatedAt string `json:"createdAt"`

	// LastUpdatedAt is the ISO 8601 timestamp of the last status update
	LastUpdatedAt string `json:"lastUpdatedAt"`

	// TTL is the time-to-live in milliseconds (actual, not requested)
	TTL int64 `json:"ttl"`

	// PollInterval is the suggested polling interval in milliseconds
	PollInterval int64 `json:"pollInterval,omitempty"`
}

// Task represents a long-running operation
type Task struct {
	// Metadata
	ID            string
	Status        TaskStatus
	StatusMessage string
	CreatedAt     time.Time
	LastUpdatedAt time.Time
	ExpiresAt     time.Time

	// Request information
	RequestMethod string
	RequestParams json.RawMessage

	// Result information
	Result      interface{}
	Error       error
	IsToolError bool // For tool calls with isError=true

	// Execution context
	Cancel     chan struct{}   // Channel to signal cancellation
	Done       chan struct{}   // Channel signaling completion
	ResultChan chan TaskResult // Channel for async result delivery
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Success bool
	Result  interface{}
	Error   error
}

// GetMetadata returns the task metadata for protocol responses
func (t *Task) GetMetadata(pollInterval int64) TaskMetadata {
	return TaskMetadata{
		TaskID:        t.ID,
		Status:        t.Status,
		StatusMessage: t.StatusMessage,
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		LastUpdatedAt: t.LastUpdatedAt.Format(time.RFC3339),
		TTL:           t.ExpiresAt.Sub(t.CreatedAt).Milliseconds(),
		PollInterval:  pollInterval,
	}
}

// UpdateStatus updates the task status and timestamp
func (t *Task) UpdateStatus(status TaskStatus, message string) {
	t.Status = status
	t.StatusMessage = message
	t.LastUpdatedAt = time.Now()
}

// SetResult sets the task result and marks it as completed
func (t *Task) SetResult(result interface{}) {
	t.Result = result
	t.UpdateStatus(TaskStatusCompleted, "Task completed successfully")
	close(t.Done)
}

// SetError sets the task error and marks it as failed
func (t *Task) SetError(err error, isToolError bool) {
	t.Error = err
	t.IsToolError = isToolError
	t.UpdateStatus(TaskStatusFailed, err.Error())
	close(t.Done)
}

// SetCancelled marks the task as canceled
func (t *Task) SetCancelled() {
	t.UpdateStatus(TaskStatusCancelled, "Task was canceled")
	close(t.Done)
}

// TaskListRequest represents a request to list tasks
type TaskListRequest struct {
	// Cursor for pagination (opaque token)
	Cursor string `json:"cursor,omitempty"`
}

// TaskListResult represents the response to a tasks/list request
type TaskListResult struct {
	// Tasks is the list of task metadata
	Tasks []TaskMetadata `json:"tasks"`

	// NextCursor is the cursor for the next page (if any)
	NextCursor string `json:"nextCursor,omitempty"`
}

// TaskGetRequest represents a request to get task status
type TaskGetRequest struct {
	TaskID string `json:"taskId"`
}

// TaskGetResult represents the response to a tasks/get request
type TaskGetResult struct {
	Task TaskMetadata `json:"task"`
}

// TaskResultRequest represents a request to get task result
type TaskResultRequest struct {
	TaskID string `json:"taskId"`
}

// TaskCancelRequest represents a request to cancel a task
type TaskCancelRequest struct {
	TaskID string `json:"taskId"`
}

// TaskCancelResult represents the response to a tasks/cancel request
type TaskCancelResult struct {
	Task TaskMetadata `json:"task"`
}

// TaskCapabilities represents task support capabilities
type TaskCapabilities struct {
	// Requests maps request types to their task support
	Requests map[string]bool `json:"requests,omitempty"`

	// List indicates if tasks/list is supported
	List bool `json:"list,omitempty"`

	// Cancel indicates if tasks/cancel is supported
	Cancel bool `json:"cancel,omitempty"`
}
