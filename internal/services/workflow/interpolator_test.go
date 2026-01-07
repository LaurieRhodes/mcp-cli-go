package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		variables map[string]string
		want      string
		wantErr   bool
	}{
		{
			name: "simple variable",
			text: "Hello {{name}}",
			variables: map[string]string{
				"name": "World",
			},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name: "multiple variables",
			text: "{{greeting}} {{name}}!",
			variables: map[string]string{
				"greeting": "Hello",
				"name":     "World",
			},
			want:    "Hello World!",
			wantErr: false,
		},
		{
			name: "step result",
			text: "Based on: {{analyze}}",
			variables: map[string]string{
				"analyze": "The data shows...",
			},
			want:    "Based on: The data shows...",
			wantErr: false,
		},
		{
			name: "environment variable",
			text: "Project: {{env.PROJECT}}",
			variables: map[string]string{
				"env.PROJECT": "my-project",
			},
			want:    "Project: my-project",
			wantErr: false,
		},
		{
			name: "missing variable",
			text: "Hello {{name}}",
			variables: map[string]string{
				"other": "value",
			},
			want:    "Hello {{name}}",
			wantErr: true,
		},
		{
			name: "variable with whitespace",
			text: "Hello {{ name }}",
			variables: map[string]string{
				"name": "World",
			},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:      "no variables",
			text:      "Hello World",
			variables: map[string]string{},
			want:      "Hello World",
			wantErr:   false,
		},
		{
			name: "multiple occurrences",
			text: "{{name}} says hi to {{name}}",
			variables: map[string]string{
				"name": "Alice",
			},
			want:    "Alice says hi to Alice",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpolator()
			for k, v := range tt.variables {
				interp.Set(k, v)
			}

			got, err := interp.Interpolate(tt.text)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSetStepResult(t *testing.T) {
	interp := NewInterpolator()

	interp.SetStepResult("step1", "result1")

	value, ok := interp.GetVariable("step1")
	assert.True(t, ok)
	assert.Equal(t, "result1", value)
}

func TestSetEnv(t *testing.T) {
	interp := NewInterpolator()

	env := map[string]string{
		"PROJECT": "my-project",
		"ENV":     "production",
	}

	interp.SetEnv(env)

	project, ok := interp.GetVariable("env.PROJECT")
	assert.True(t, ok)
	assert.Equal(t, "my-project", project)

	envVal, ok := interp.GetVariable("env.ENV")
	assert.True(t, ok)
	assert.Equal(t, "production", envVal)
}

func TestHasVariable(t *testing.T) {
	interp := NewInterpolator()
	interp.Set("exists", "value")

	assert.True(t, interp.HasVariable("exists"))
	assert.False(t, interp.HasVariable("not_exists"))
}

func TestClone(t *testing.T) {
	original := NewInterpolator()
	original.Set("key1", "value1")
	original.Set("key2", "value2")

	clone := original.Clone()

	// Check clone has same values
	val, ok := clone.GetVariable("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Modify clone
	clone.Set("key3", "value3")

	// Original should not have new value
	assert.False(t, original.HasVariable("key3"))

	// Clone should have it
	assert.True(t, clone.HasVariable("key3"))
}

func TestClear(t *testing.T) {
	interp := NewInterpolator()
	interp.Set("key1", "value1")
	interp.Set("key2", "value2")

	assert.True(t, interp.HasVariable("key1"))
	assert.True(t, interp.HasVariable("key2"))

	interp.Clear()

	assert.False(t, interp.HasVariable("key1"))
	assert.False(t, interp.HasVariable("key2"))
}
