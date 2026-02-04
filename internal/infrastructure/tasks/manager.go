package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain"
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
)

// Manager handles task lifecycle and storage
type Manager struct {
	tasks         map[string]*domain.Task
	mu            sync.RWMutex
	defaultTTL    time.Duration
	maxTTL        time.Duration
	pollInterval  int64 // in milliseconds
	cleanupTicker *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewManager creates a new task manager
func NewManager(defaultTTL, maxTTL time.Duration, pollInterval int64) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		tasks:        make(map[string]*domain.Task),
		defaultTTL:   defaultTTL,
		maxTTL:       maxTTL,
		pollInterval: pollInterval,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start cleanup routine
	m.cleanupTicker = time.NewTicker(1 * time.Minute)
	go m.cleanupExpiredTasks()

	return m
}

// CreateTask creates a new task
func (m *Manager) CreateTask(requestMethod string, requestParams interface{}, requestedTTL int64) (*domain.Task, error) {
	taskID, err := generateTaskID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate task ID: %w", err)
	}

	// Determine TTL
	ttl := m.defaultTTL
	if requestedTTL > 0 {
		requestedDuration := time.Duration(requestedTTL) * time.Millisecond
		if requestedDuration < m.maxTTL {
			ttl = requestedDuration
		} else {
			ttl = m.maxTTL
		}
	}

	now := time.Now()
	task := &domain.Task{
		ID:            taskID,
		Status:        domain.TaskStatusWorking,
		StatusMessage: "Task is being processed",
		CreatedAt:     now,
		LastUpdatedAt: now,
		ExpiresAt:     now.Add(ttl),
		RequestMethod: requestMethod,
		Cancel:        make(chan struct{}),
		Done:          make(chan struct{}),
		ResultChan:    make(chan domain.TaskResult, 1),
	}

	m.mu.Lock()
	m.tasks[taskID] = task
	m.mu.Unlock()

	logging.Info("Created task %s for %s (TTL: %s)", taskID, requestMethod, ttl)
	return task, nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(taskID string) (*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// GetTaskMetadata retrieves task metadata by ID
func (m *Manager) GetTaskMetadata(taskID string) (domain.TaskMetadata, error) {
	task, err := m.GetTask(taskID)
	if err != nil {
		return domain.TaskMetadata{}, err
	}

	return task.GetMetadata(m.pollInterval), nil
}

// WaitForResult waits for a task to complete and returns its result
// This implements the blocking behavior for tasks/result
func (m *Manager) WaitForResult(taskID string, timeout time.Duration) (interface{}, error) {
	task, err := m.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// If already in terminal state, return immediately
	if task.Status.IsTerminal() {
		if task.Error != nil {
			return nil, task.Error
		}
		return task.Result, nil
	}

	// Wait for completion or timeout
	ctx, cancel := context.WithTimeout(m.ctx, timeout)
	defer cancel()

	select {
	case <-task.Done:
		if task.Error != nil {
			return nil, task.Error
		}
		return task.Result, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for task result")

	case <-m.ctx.Done():
		return nil, fmt.Errorf("task manager shutting down")
	}
}

// CancelTask cancels a task
func (m *Manager) CancelTask(taskID string) error {
	task, err := m.GetTask(taskID)
	if err != nil {
		return err
	}

	// Can only cancel non-terminal tasks
	if task.Status.IsTerminal() {
		return fmt.Errorf("task %s is already in terminal state: %s", taskID, task.Status)
	}

	// Signal cancellation
	close(task.Cancel)

	// Update status
	m.mu.Lock()
	task.SetCancelled()
	m.mu.Unlock()

	logging.Info("Canceled task %s", taskID)
	return nil
}

// ListTasks returns a paginated list of tasks
func (m *Manager) ListTasks(cursor string, limit int) ([]domain.TaskMetadata, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all tasks
	var allTasks []*domain.Task
	for _, task := range m.tasks {
		allTasks = append(allTasks, task)
	}

	// Sort by creation time (newest first)
	// TODO: Implement proper sorting

	// Apply pagination
	start := 0
	if cursor != "" {
		// Decode cursor to find start position
		// For now, simple implementation without cursor
	}

	end := start + limit
	if end > len(allTasks) {
		end = len(allTasks)
	}

	// Convert to metadata
	var result []domain.TaskMetadata
	for i := start; i < end; i++ {
		result = append(result, allTasks[i].GetMetadata(m.pollInterval))
	}

	// Generate next cursor if there are more results
	nextCursor := ""
	if end < len(allTasks) {
		nextCursor = fmt.Sprintf("%d", end)
	}

	return result, nextCursor, nil
}

// DeleteTask removes a task from storage
func (m *Manager) DeleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tasks, taskID)
	logging.Debug("Deleted task %s", taskID)
	return nil
}

// cleanupExpiredTasks removes expired tasks
func (m *Manager) cleanupExpiredTasks() {
	for {
		select {
		case <-m.cleanupTicker.C:
			m.mu.Lock()
			now := time.Now()
			for id, task := range m.tasks {
				if now.After(task.ExpiresAt) {
					logging.Debug("Cleaning up expired task %s", id)
					delete(m.tasks, id)
				}
			}
			m.mu.Unlock()

		case <-m.ctx.Done():
			m.cleanupTicker.Stop()
			return
		}
	}
}

// Close shuts down the task manager
func (m *Manager) Close() {
	m.cancel()

	// Cancel all pending tasks
	m.mu.Lock()
	for id, task := range m.tasks {
		if !task.Status.IsTerminal() {
			close(task.Cancel)
			task.SetCancelled()
			logging.Info("Canceled task %s during shutdown", id)
		}
	}
	m.mu.Unlock()
}

// GetPollInterval returns the suggested poll interval in milliseconds
func (m *Manager) GetPollInterval() int64 {
	return m.pollInterval
}

// generateTaskID generates a cryptographically random task ID
func generateTaskID() (string, error) {
	// Generate 16 random bytes (128 bits)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Format as UUID-style string
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(bytes[0:4]),
		hex.EncodeToString(bytes[4:6]),
		hex.EncodeToString(bytes[6:8]),
		hex.EncodeToString(bytes[8:10]),
		hex.EncodeToString(bytes[10:16]),
	), nil
}

// GetTaskCount returns the number of tasks (for monitoring)
func (m *Manager) GetTaskCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tasks)
}

// GetTaskStats returns task statistics
func (m *Manager) GetTaskStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total":          0,
		"working":        0,
		"completed":      0,
		"failed":         0,
		"canceled":      0,
		"input_required": 0,
	}

	for _, task := range m.tasks {
		stats["total"]++
		stats[string(task.Status)]++
	}

	return stats
}
