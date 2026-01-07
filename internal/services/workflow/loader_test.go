package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromBytes(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid simple workflow",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "test"
`,
			wantErr: false,
		},
		{
			name: "valid workflow with fallback",
			yaml: `
name: test
version: 1.0.0

execution:
  providers:
    - provider: anthropic
      model: claude-sonnet-4
    - provider: openai
      model: gpt-4o

steps:
  - name: step1
    run: "test"
`,
			wantErr: false,
		},
		{
			name: "missing name",
			yaml: `
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "test"
`,
			wantErr: true,
		},
		{
			name: "missing version",
			yaml: `
name: test

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "test"
`,
			wantErr: true,
		},
		{
			name: "no steps",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps: []
`,
			wantErr: true,
		},
		{
			name: "no provider",
			yaml: `
name: test
version: 1.0.0

execution: {}

steps:
  - name: step1
    run: "test"
`,
			wantErr: true,
		},
		{
			name: "step without name",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - run: "test"
`,
			wantErr: true,
		},
		{
			name: "duplicate step names",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "test"
  - name: step1
    run: "test"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow, err := loader.LoadFromBytes([]byte(tt.yaml))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workflow)
			}
		})
	}
}

func TestValidateConsensus(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid consensus",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: decision
    consensus:
      prompt: "Approve?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
`,
			wantErr: false,
		},
		{
			name: "consensus with too few providers",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: decision
    consensus:
      prompt: "Approve?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
      require: unanimous
`,
			wantErr: true,
		},
		{
			name: "consensus without prompt",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: decision
    consensus:
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: unanimous
`,
			wantErr: true,
		},
		{
			name: "invalid requirement",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: decision
    consensus:
      prompt: "Approve?"
      executions:
        - provider: anthropic
          model: claude-sonnet-4
        - provider: openai
          model: gpt-4o
      require: invalid
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loader.LoadFromBytes([]byte(tt.yaml))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDependencies(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid dependency",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "test"
  - name: step2
    run: "test"
    needs: [step1]
`,
			wantErr: false,
		},
		{
			name: "invalid dependency",
			yaml: `
name: test
version: 1.0.0

execution:
  provider: anthropic
  model: claude-sonnet-4

steps:
  - name: step1
    run: "test"
    needs: [unknown]
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loader.LoadFromBytes([]byte(tt.yaml))

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
