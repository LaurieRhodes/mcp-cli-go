package config

import (
	"testing"
)

func TestLoopV2_Validate_IterateMode(t *testing.T) {
	tests := []struct {
		name    string
		loop    LoopV2
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid iterate mode",
			loop: LoopV2{
				Mode:          "iterate",
				Items:         "{{statements}}",
				Workflow:      "process_one",
				MaxIterations: 100,
			},
			wantErr: false,
		},
		{
			name: "iterate mode missing items",
			loop: LoopV2{
				Mode:          "iterate",
				Workflow:      "process_one",
				MaxIterations: 100,
			},
			wantErr: true,
			errMsg:  "iterate mode requires 'items' field",
		},
		{
			name: "iterate mode missing workflow",
			loop: LoopV2{
				Mode:          "iterate",
				Items:         "{{array}}",
				MaxIterations: 100,
			},
			wantErr: true,
			errMsg:  "workflow field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loop.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoopV2.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("LoopV2.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoopV2_Validate_RefineMode(t *testing.T) {
	tests := []struct {
		name    string
		loop    LoopV2
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid refine mode",
			loop: LoopV2{
				Mode:          "refine",
				Until:         "Review says PASS",
				Workflow:      "improve_code",
				MaxIterations: 5,
			},
			wantErr: false,
		},
		{
			name: "refine mode missing until",
			loop: LoopV2{
				Mode:          "refine",
				Workflow:      "improve_code",
				MaxIterations: 5,
			},
			wantErr: true,
			errMsg:  "refine mode requires 'until' condition",
		},
		{
			name: "defaults to refine mode",
			loop: LoopV2{
				Until:         "Done",
				Workflow:      "test",
				MaxIterations: 5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loop.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoopV2.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("LoopV2.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
			// Check mode defaulted to refine
			if !tt.wantErr && tt.loop.Mode == "" {
				if tt.loop.Mode != "refine" {
					t.Errorf("Mode should default to 'refine', got '%s'", tt.loop.Mode)
				}
			}
		})
	}
}

func TestLoopV2_Validate_SuccessRate(t *testing.T) {
	tests := []struct {
		name    string
		loop    LoopV2
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid success rate",
			loop: LoopV2{
				Mode:           "iterate",
				Items:          "{{array}}",
				Workflow:       "test",
				MaxIterations:  100,
				MinSuccessRate: 0.95,
			},
			wantErr: false,
		},
		{
			name: "success rate below bounds",
			loop: LoopV2{
				Mode:           "iterate",
				Items:          "{{array}}",
				Workflow:       "test",
				MaxIterations:  100,
				MinSuccessRate: -0.1,
			},
			wantErr: true,
			errMsg:  "min_success_rate must be between 0.0 and 1.0, got -0.100000",
		},
		{
			name: "success rate above bounds",
			loop: LoopV2{
				Mode:           "iterate",
				Items:          "{{array}}",
				Workflow:       "test",
				MaxIterations:  100,
				MinSuccessRate: 1.5,
			},
			wantErr: true,
			errMsg:  "min_success_rate must be between 0.0 and 1.0, got 1.500000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loop.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoopV2.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("LoopV2.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoopV2_Validate_RetryConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		loop    LoopV2
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid retry configuration",
			loop: LoopV2{
				Mode:          "iterate",
				Items:         "{{array}}",
				Workflow:      "test",
				MaxIterations: 100,
				OnFailure:     "retry",
				MaxRetries:    3,
				RetryDelay:    "5s",
			},
			wantErr: false,
		},
		{
			name: "retry without max_retries",
			loop: LoopV2{
				Mode:          "iterate",
				Items:         "{{array}}",
				Workflow:      "test",
				MaxIterations: 100,
				OnFailure:     "retry",
			},
			wantErr: true,
			errMsg:  "retry mode requires max_retries >= 1",
		},
		{
			name: "invalid on_failure value",
			loop: LoopV2{
				Mode:          "iterate",
				Items:         "{{array}}",
				Workflow:      "test",
				MaxIterations: 100,
				OnFailure:     "invalid",
			},
			wantErr: true,
			errMsg:  "on_failure must be 'halt', 'continue', or 'retry', got 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loop.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoopV2.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("LoopV2.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoopExecutionResult_CheckSuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		result   LoopExecutionResult
		minRate  float64
		expected bool
	}{
		{
			name: "100% success meets 95% requirement",
			result: LoopExecutionResult{
				TotalItems: 100,
				Succeeded:  100,
				Failed:     0,
			},
			minRate:  0.95,
			expected: true,
		},
		{
			name: "96% success meets 95% requirement",
			result: LoopExecutionResult{
				TotalItems: 100,
				Succeeded:  96,
				Failed:     4,
			},
			minRate:  0.95,
			expected: true,
		},
		{
			name: "90% success fails 95% requirement",
			result: LoopExecutionResult{
				TotalItems: 100,
				Succeeded:  90,
				Failed:     10,
			},
			minRate:  0.95,
			expected: false,
		},
		{
			name: "empty result always succeeds",
			result: LoopExecutionResult{
				TotalItems: 0,
				Succeeded:  0,
				Failed:     0,
			},
			minRate:  0.95,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.CheckSuccessRate(tt.minRate)
			if result != tt.expected {
				t.Errorf("CheckSuccessRate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
