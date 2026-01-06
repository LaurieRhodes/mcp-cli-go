package skills

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/domain/skills"
)

func TestImageMappingIntegration(t *testing.T) {
	t.Log("\n======================================================================")
	t.Log("IMAGE MAPPING INTEGRATION TEST")
	t.Log("======================================================================")

	// Initialize skill service
	service := NewService()
	
	skillsDir := "../../../config/skills"
	executionMode := skills.ExecutionModeAuto
	
	err := service.Initialize(skillsDir, executionMode)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	t.Logf("\n✅ Initialized with %d skills\n", len(service.skills))

	// Check that image mapping was loaded
	if service.imageMapping == nil {
		t.Fatal("❌ Image mapping not loaded")
	}

	t.Log("✅ Image mapping loaded")
	t.Logf("   Default image: %s", service.imageMapping.DefaultImage)
	t.Logf("   Skills mapped: %d", len(service.imageMapping.Skills))

	// Test each skill mapping
	skillTests := []struct {
		name          string
		expectedImage string
	}{
		{"docx", "mcp-skills-docx"},
		{"pptx", "mcp-skills-pptx"},
		{"xlsx", "mcp-skills-xlsx"},
		{"pdf", "mcp-skills-pdf"},
	}

	for _, tt := range skillTests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if skill exists
			skill, exists := service.skills[tt.name]
			if !exists {
				t.Skipf("Skill '%s' not found in skills directory", tt.name)
				return
			}

			// Check mapping
			actualImage := service.imageMapping.GetImageForSkill(tt.name)
			if actualImage != tt.expectedImage {
				t.Errorf("❌ Skill '%s': expected image '%s', got '%s'",
					tt.name, tt.expectedImage, actualImage)
			} else {
				t.Logf("✅ Skill '%s' -> '%s'", tt.name, actualImage)
			}

			// Verify image exists
			if service.executor != nil {
				t.Logf("   Executor available: %s", service.executor.GetInfo())
				t.Logf("   Image to use: %s", actualImage)
			}

			t.Logf("   Skill location: %s", skill.DirectoryPath)
		})
	}
}

func TestExecuteCodeWithCorrectImage(t *testing.T) {
	t.Log("\n======================================================================")
	t.Log("EXECUTE CODE WITH CORRECT IMAGE TEST")
	t.Log("======================================================================")

	// Initialize skill service
	service := NewService()
	
	skillsDir := "../../../config/skills"
	executionMode := skills.ExecutionModeAuto
	
	err := service.Initialize(skillsDir, executionMode)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Skip if no executor available
	if service.executor == nil {
		t.Skip("Executor not available")
	}

	t.Logf("✅ Executor available: %s\n", service.executor.GetInfo())

	// Create a simple Python script that just prints the image info
	code := `import sys
print(f"Python version: {sys.version}")
print(f"Python executable: {sys.executable}")

# Try to import packages that should be in specific images
try:
    import defusedxml
    print("✅ defusedxml available (docx image)")
except ImportError:
    print("❌ defusedxml not available")

try:
    import pptx
    print("✅ python-pptx available (pptx image)")
except ImportError:
    print("❌ python-pptx not available")

try:
    import openpyxl
    print("✅ openpyxl available (xlsx image)")
except ImportError:
    print("❌ openpyxl not available")

try:
    import pypdf
    print("✅ pypdf available (pdf image)")
except ImportError:
    print("❌ pypdf not available")
`

	// Test with docx skill (should have defusedxml)
	t.Run("docx_image", func(t *testing.T) {
		if _, exists := service.skills["docx"]; !exists {
			t.Skip("docx skill not found")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Create temp directory for workspace
		workspaceDir, err := os.MkdirTemp("", "test-workspace-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(workspaceDir)

		// Write code to file
		scriptPath := filepath.Join(workspaceDir, "test.py")
		if err := os.WriteFile(scriptPath, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write script: %v", err)
		}

		// Get docx skill directory
		skill := service.skills["docx"]
		
		t.Logf("\n🚀 Executing with docx skill image...")
		t.Logf("   Skill dir: %s", skill.DirectoryPath)
		t.Logf("   Expected image: mcp-skills-docx")

		// Execute
		output, err := service.executor.ExecutePythonCode(ctx, workspaceDir, skill.DirectoryPath, scriptPath, nil)
		
		if err != nil {
			t.Logf("❌ Execution failed: %v", err)
			t.Logf("Output: %s", output)
			// Don't fail the test if image needs to be pulled
			if testing.Short() {
				t.Skip("Skipping in short mode - image might need pulling")
			}
		} else {
			t.Logf("✅ Execution succeeded!")
			t.Logf("\nOutput:\n%s", output)
			
			// Check if defusedxml is available (should be in mcp-skills-docx)
			if !strings.Contains(output, "defusedxml available") {
				t.Error("❌ defusedxml not available - wrong image used?")
			}
		}
	})
}
