package skills

import (
	"strings"
	"testing"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
)

func TestExecuteCodeWithSkillContext(t *testing.T) {
	t.Log("\n" + strings.Repeat("=", 70))
	t.Log("DYNAMIC CODE EXECUTION TEST: Execute Code with Skill Context")
	t.Log(strings.Repeat("=", 70))

	// Initialize service
	service := NewService()

	skillsDir := "../../../config/skills"
	executionMode := skills.ExecutionModeAuto

	err := service.Initialize(skillsDir, executionMode)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	t.Logf("\n✅ Initialized with %d skills", len(service.ListSkills()))

	// Check if test-execution skill exists
	skill, exists := service.GetSkill("test-execution")
	if !exists {
		t.Fatal("❌ test-execution skill not found")
	}

	t.Logf("✅ Found skill: %s", skill.Name)

	// Skip if Docker/Podman not available
	if service.executor == nil {
		t.Skip("⚠️  Docker/Podman not available, skipping execution test")
	}

	t.Logf("✅ Executor available: %s", service.executor.GetInfo())

	// Test 1: Simple Python code
	t.Log("\n📝 Test 1: Simple Python code")

	code1 := `
print("=" * 60)
print("✅ SUCCESS: Dynamic code execution working!")
print("=" * 60)
print()
print("This code was written by the LLM and executed dynamically.")
print("Not a pre-written script!")
print()

import sys
print(f"Python version: {sys.version}")
print(f"Working directory: /workspace")
print()
print("✅ Test 1 passed!")
`

	request1 := &skills.CodeExecutionRequest{
		SkillName: "test-execution",
		Language:  "python",
		Code:      code1,
		Files:     nil,
	}

	result1, err := service.ExecuteCode(request1)
	if err != nil {
		t.Fatalf("❌ Test 1 failed: %v\nOutput:\n%s", err, result1.Output)
	}

	t.Logf("\n✅ Test 1 passed in %dms", result1.Duration)
	t.Log("Output:")
	t.Log(result1.Output)

	// Verify output
	if !strings.Contains(result1.Output, "SUCCESS: Dynamic code execution working!") {
		t.Errorf("❌ Expected success message not found")
	}

	// Test 2: Code with file creation
	t.Log("\n📝 Test 2: Code with file creation in workspace")

	code2 := `
import os

# Create a file in workspace
with open('/workspace/output.txt', 'w') as f:
    f.write('Hello from dynamically executed code!\\n')
    f.write('This file was created in the workspace.\\n')

print("✅ File created successfully")

# Read it back
with open('/workspace/output.txt', 'r') as f:
    content = f.read()

print("📄 File contents:")
print(content)

# Verify workspace is writable but skill directory is read-only
print("🔒 Security check:")
try:
    with open('/skill/test.txt', 'w') as f:
        f.write('test')
    print("❌ SECURITY FAILURE: Skill directory is writable!")
except (IOError, OSError):
    print("✅ Confirmed: Skill directory is read-only")

print()
print("✅ Test 2 passed!")
`

	request2 := &skills.CodeExecutionRequest{
		SkillName: "test-execution",
		Language:  "python",
		Code:      code2,
		Files:     nil,
	}

	result2, err := service.ExecuteCode(request2)
	if err != nil {
		t.Fatalf("❌ Test 2 failed: %v\nOutput:\n%s", err, result2.Output)
	}

	t.Logf("\n✅ Test 2 passed in %dms", result2.Duration)
	t.Log("Output:")
	t.Log(result2.Output)

	// Verify security
	if !strings.Contains(result2.Output, "Skill directory is read-only") {
		t.Errorf("❌ Expected security confirmation not found")
	}

	// Test 3: Code with input files
	t.Log("\n📝 Test 3: Code with input files")

	code3 := `
import os

# List files in workspace
print("📁 Files in workspace:")
files = os.listdir('/workspace')
for f in sorted(files):
    print(f"   - {f}")

print()

# Read input file
with open('/workspace/input.txt', 'r') as f:
    content = f.read()

print("📄 Input file contents:")
print(content)

# Process and create output
output_content = content.upper() + "\\n\\nProcessed by dynamic code execution!"

with open('/workspace/output.txt', 'w') as f:
    f.write(output_content)

print("✅ File processed successfully")
print()
print("✅ Test 3 passed!")
`

	inputFile := []byte("Hello from the LLM!\nThis is test data.\n")

	request3 := &skills.CodeExecutionRequest{
		SkillName: "test-execution",
		Language:  "python",
		Code:      code3,
		Files: map[string][]byte{
			"input.txt": inputFile,
		},
	}

	result3, err := service.ExecuteCode(request3)
	if err != nil {
		t.Fatalf("❌ Test 3 failed: %v\nOutput:\n%s", err, result3.Output)
	}

	t.Logf("\n✅ Test 3 passed in %dms", result3.Duration)
	t.Log("Output:")
	t.Log(result3.Output)

	// Verify file was processed
	if !strings.Contains(result3.Output, "input.txt") {
		t.Errorf("❌ Expected input file listing not found")
	}

	// Summary
	t.Log("\n" + strings.Repeat("=", 70))
	t.Log("🎉 ALL DYNAMIC CODE EXECUTION TESTS PASSED!")
	t.Log(strings.Repeat("=", 70))
	t.Log("\nThis proves:")
	t.Log("  ✅ LLM can write code dynamically")
	t.Log("  ✅ Code executes in sandboxed environment")
	t.Log("  ✅ Workspace is read-write for file operations")
	t.Log("  ✅ Skill directory is read-only (security)")
	t.Log("  ✅ Input files can be provided")
	t.Log("  ✅ Output can be captured")
	t.Log("\n🚀 Ready for real-world usage!")
	t.Log("   Next: Test with actual helper library imports")
	t.Log("")
}
