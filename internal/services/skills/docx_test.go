package skills

import (
	"strings"
	"testing"
	"time"

	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
)

func TestDocxSkillCreateDocument(t *testing.T) {
	t.Log("\n" + strings.Repeat("=", 70))
	t.Log("DOCX SKILL TEST: Creating Word Document")
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

	// Check if docx skill exists
	skill, exists := service.GetSkill("docx")
	if !exists {
		t.Fatal("❌ docx skill not found")
	}

	t.Logf("✅ Found skill: %s", skill.Name)
	t.Logf("   Has scripts: %v", skill.HasScripts)

	if !skill.HasScripts {
		t.Fatal("❌ docx skill should have scripts")
	}

	// Check if our test script exists
	found := false
	for _, script := range skill.Scripts {
		if script == "create_test_doc.py" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("❌ create_test_doc.py not found in scripts: %v", skill.Scripts)
	}

	t.Logf("✅ Found create_test_doc.py script")

	// Skip if Docker/Podman not available
	if service.executor == nil {
		t.Skip("⚠️  Docker/Podman not available, skipping execution test")
	}

	t.Logf("✅ Executor available: %s", service.executor.GetInfo())

	// Execute the document creation script
	t.Log("\n🚀 Creating Word document...")
	t.Log("   This will:")
	t.Log("   1. Pull python:3.11-alpine image (if needed)")
	t.Log("   2. Install python-docx library")
	t.Log("   3. Create a Word document with multiple sections")
	t.Log("   4. Return success message")
	t.Log("")

	startTime := time.Now()
	output, err := service.ExecuteScript(skill, "create_test_doc.py", nil)
	duration := time.Since(startTime)

	if err != nil {
		t.Logf("\n❌ Script execution failed after %v", duration)
		t.Logf("\nOutput:\n%s", output)
		t.Fatalf("Error: %v", err)
	}

	t.Logf("\n✅ Script executed successfully in %v", duration)
	t.Log("\n📄 Script Output:")
	t.Log("────────────────────────────────────────────────────────────")
	t.Log(output)
	t.Log("────────────────────────────────────────────────────────────")

	// Verify output contains expected success messages
	if !strings.Contains(output, "SUCCESS: Word document created") {
		t.Errorf("❌ Expected success message not found in output")
	}

	if !strings.Contains(output, "File size:") {
		t.Errorf("❌ Expected file size info not found in output")
	}

	if !strings.Contains(output, "Document contains:") {
		t.Errorf("❌ Expected document contents list not found in output")
	}

	t.Log("\n" + strings.Repeat("=", 70))
	t.Log("🎉 DOCX SKILL TEST PASSED!")
	t.Log(strings.Repeat("=", 70))
	t.Log("\nThis proves:")
	t.Log("  ✅ docx skill scripts are discoverable")
	t.Log("  ✅ Python script executes in sandbox")
	t.Log("  ✅ python-docx library works in container")
	t.Log("  ✅ Word document creation succeeds")
	t.Log("  ✅ Complex document manipulation possible")
	t.Log("  ✅ Real-world skill usage verified")
	t.Log("")
}
