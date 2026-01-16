package workflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/config"
)

// WorkflowWorkerPool manages concurrent step execution with dependency awareness
type WorkflowWorkerPool struct {
	maxWorkers int
	semaphore  chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex

	// Thread-safe result storage
	stepResults map[string]string
	stepErrors  map[string]error
	completed   map[string]bool

	// Coordination
	notifyCompletion chan string

	// Error handling
	errorPolicy    string // cancel_all, complete_running, continue
	cancelFunc     context.CancelFunc
	acceptingWork  bool
	workMu         sync.Mutex

	// Execution context
	orchestrator *Orchestrator
	
	// Observability (Phase 3)
	bufferedLogger *BufferedLogger
	timeline       *ExecutionTimeline
}

// NewWorkerPool creates a new workflow worker pool
func NewWorkerPool(maxWorkers int, errorPolicy string, orchestrator *Orchestrator) *WorkflowWorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 3 // Default from assessment
	}
	
	if errorPolicy == "" {
		errorPolicy = "cancel_all" // Safe default
	}

	return &WorkflowWorkerPool{
		maxWorkers:       maxWorkers,
		semaphore:        make(chan struct{}, maxWorkers),
		stepResults:      make(map[string]string),
		stepErrors:       make(map[string]error),
		completed:        make(map[string]bool),
		notifyCompletion: make(chan string, 100), // Buffered to prevent blocking
		errorPolicy:      errorPolicy,
		acceptingWork:    true,
		orchestrator:     orchestrator,
		bufferedLogger:   NewBufferedLogger(),
		timeline:         NewExecutionTimeline(maxWorkers),
	}
}

// SetCancelFunc sets the cancel function for context cancellation
func (p *WorkflowWorkerPool) SetCancelFunc(cancel context.CancelFunc) {
	p.cancelFunc = cancel
}

// SubmitStep submits a step for execution in the worker pool
func (p *WorkflowWorkerPool) SubmitStep(ctx context.Context, step *config.StepV2) error {
	p.workMu.Lock()
	if !p.acceptingWork {
		p.workMu.Unlock()
		return fmt.Errorf("worker pool no longer accepting work due to previous error")
	}
	p.workMu.Unlock()

	p.wg.Add(1)
	
	// Acquire worker slot (blocks if pool is full)
	select {
	case p.semaphore <- struct{}{}:
		// Got a slot
	case <-ctx.Done():
		p.wg.Done()
		return ctx.Err()
	}

	go func(s *config.StepV2) {
		defer p.wg.Done()
		defer func() { <-p.semaphore }() // Release slot

		// Record timeline start
		p.timeline.RecordStepStart(s.Name)
		p.bufferedLogger.StartStep(s.Name)

		// Execute the step (stores result in orchestrator.stepResults internally)
		err := p.orchestrator.executeStep(ctx, s)

		// Record timeline end
		p.timeline.RecordStepEnd(s.Name)
		p.bufferedLogger.EndStep(s.Name)

		// Get result from orchestrator (thread-safe read)
		var result string
		if err == nil {
			p.orchestrator.stepResultsMu.RLock()
			result = p.orchestrator.stepResults[s.Name]
			p.orchestrator.stepResultsMu.RUnlock()
		}

		// Store result (thread-safe)
		p.mu.Lock()
		if err == nil {
			p.stepResults[s.Name] = result
			p.completed[s.Name] = true
		} else {
			p.stepErrors[s.Name] = err
			p.completed[s.Name] = false
		}
		p.mu.Unlock()

		// Handle error according to policy
		if err != nil {
			p.handleError(s.Name, err)
		}

		// Notify completion
		select {
		case p.notifyCompletion <- s.Name:
		case <-ctx.Done():
			// Context cancelled, don't block
		}
	}(step)

	return nil
}

// SubmitLoop submits a loop for execution in the worker pool
func (p *WorkflowWorkerPool) SubmitLoop(ctx context.Context, loop *config.LoopV2) error {
	p.workMu.Lock()
	if !p.acceptingWork {
		p.workMu.Unlock()
		return fmt.Errorf("worker pool no longer accepting work due to previous error")
	}
	p.workMu.Unlock()

	p.wg.Add(1)
	
	// Acquire worker slot (blocks if pool is full)
	select {
	case p.semaphore <- struct{}{}:
		// Got a slot
	case <-ctx.Done():
		p.wg.Done()
		return ctx.Err()
	}

	go func(l *config.LoopV2) {
		defer p.wg.Done()
		defer func() { <-p.semaphore }() // Release slot

		// Record timeline start
		p.timeline.RecordStepStart(l.Name)
		p.bufferedLogger.StartStep(l.Name)

		// Execute the loop
		err := p.orchestrator.executeLoop(ctx, l)

		// Record timeline end
		p.timeline.RecordStepEnd(l.Name)
		p.bufferedLogger.EndStep(l.Name)

		// Get result from orchestrator (thread-safe read)
		var result string
		if err == nil {
			p.orchestrator.stepResultsMu.RLock()
			result = p.orchestrator.stepResults[l.Name]
			p.orchestrator.stepResultsMu.RUnlock()
		}

		// Store result (thread-safe)
		p.mu.Lock()
		if err == nil {
			p.stepResults[l.Name] = result
			p.completed[l.Name] = true
		} else {
			p.stepErrors[l.Name] = err
			p.completed[l.Name] = false
		}
		p.mu.Unlock()

		// Handle error according to policy
		if err != nil {
			p.handleError(l.Name, err)
		}

		// Notify completion
		select {
		case p.notifyCompletion <- l.Name:
		case <-ctx.Done():
			// Context cancelled, don't block
		}
	}(loop)

	return nil
}

// handleError processes step errors according to error policy
func (p *WorkflowWorkerPool) handleError(stepName string, err error) {
	p.orchestrator.logger.Error("Step %s failed: %v", stepName, err)

	switch p.errorPolicy {
	case "cancel_all":
		// Cancel all in-flight steps
		p.orchestrator.logger.Info("Error policy: cancel_all - Cancelling all in-flight steps")
		p.workMu.Lock()
		p.acceptingWork = false
		p.workMu.Unlock()
		
		if p.cancelFunc != nil {
			p.cancelFunc()
		}

	case "complete_running":
		// Let running steps finish, but don't start new ones
		p.orchestrator.logger.Info("Error policy: complete_running - Allowing in-flight steps to complete")
		p.workMu.Lock()
		p.acceptingWork = false
		p.workMu.Unlock()

	case "continue":
		// Keep going (fault-tolerant mode)
		p.orchestrator.logger.Info("Error policy: continue - Continuing workflow execution")
		// No action needed - keep accepting work
	}
}

// Wait waits for all workers to complete
func (p *WorkflowWorkerPool) Wait() {
	p.wg.Wait()
	close(p.notifyCompletion)
}

// GetResult returns the result for a step (thread-safe)
func (p *WorkflowWorkerPool) GetResult(stepName string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result, ok := p.stepResults[stepName]
	return result, ok
}

// GetError returns the error for a step (thread-safe)
func (p *WorkflowWorkerPool) GetError(stepName string) (error, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	err, ok := p.stepErrors[stepName]
	return err, ok
}

// IsCompleted checks if a step has completed (thread-safe)
func (p *WorkflowWorkerPool) IsCompleted(stepName string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.completed[stepName]
}

// GetAllResults returns all step results (thread-safe)
func (p *WorkflowWorkerPool) GetAllResults() map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	results := make(map[string]string)
	for k, v := range p.stepResults {
		results[k] = v
	}
	return results
}

// GetAllErrors returns all step errors (thread-safe)
func (p *WorkflowWorkerPool) GetAllErrors() map[string]error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	errors := make(map[string]error)
	for k, v := range p.stepErrors {
		errors[k] = v
	}
	return errors
}

// GetCompleted returns the completed status map (thread-safe)
func (p *WorkflowWorkerPool) GetCompleted() map[string]bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	completed := make(map[string]bool)
	for k, v := range p.completed {
		completed[k] = v
	}
	return completed
}
