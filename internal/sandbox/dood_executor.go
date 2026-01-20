package sandbox

import (
	"context"
	"fmt"
	"io"
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

// DooDockerExecutor uses Docker API directly with socket mount (for containerized deployments)
type DooDockerExecutor struct {
	config ExecutorConfig
	client *docker.Client
}

// NewDooDockerExecutor creates a new Docker-out-of-Docker executor
// Works with both Docker and Podman sockets
func NewDooDockerExecutor(config ExecutorConfig) (*DooDockerExecutor, error) {
	// Try multiple socket locations (Docker and Podman)
	socketPaths := []string{
		"unix:///var/run/docker.sock",                                    // Standard Docker
		fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", os.Getuid()), // Podman rootless
		"unix:///run/podman/podman.sock",                                 // Podman rootful
	}
	
	var client *docker.Client
	var err error
	
	for _, socketPath := range socketPaths {
		client, err = docker.NewClient(socketPath)
		if err == nil {
			// Test if socket is actually working
			if _, err := client.Version(); err == nil {
				return &DooDockerExecutor{
					config: config,
					client: client,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to Docker/Podman socket: %w", err)
}

// IsAvailable checks if Docker socket is accessible
func (d *DooDockerExecutor) IsAvailable() bool {
	_, err := d.client.Version()
	return err == nil
}

// ExecutePython runs a Python script using Docker API
func (d *DooDockerExecutor) ExecutePython(ctx context.Context, skillDir, scriptPath string, args []string) (string, error) {
	image := d.config.GetImageForSkill(skillDir)
	return d.executeInContainer(ctx, skillDir, image, "python", scriptPath, args)
}

// ExecuteBash runs a Bash script using Docker API
func (d *DooDockerExecutor) ExecuteBash(ctx context.Context, skillDir, scriptPath string, args []string) (string, error) {
	return d.executeInContainer(ctx, skillDir, "alpine:latest", "sh", scriptPath, args)
}

// executeInContainer handles the actual container execution
func (d *DooDockerExecutor) executeInContainer(
	ctx context.Context,
	skillDir string,
	image string,
	interpreter string,
	scriptPath string,
	args []string,
) (string, error) {
	// Pull image if not present
	if err := d.ensureImage(ctx, image); err != nil {
		return "", fmt.Errorf("failed to ensure image: %w", err)
	}

	// Build command
	cmd := []string{interpreter, "/skill/" + scriptPath}
	cmd = append(cmd, args...)

	// Create container
	pidsLimit := int64(100)
	container, err := d.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:           image,
			Cmd:             cmd,
			NetworkDisabled: true,
			Memory:          256 * 1024 * 1024, // 256MB
			WorkingDir:      "/skill",
		},
		HostConfig: &docker.HostConfig{
			Binds:          []string{
				fmt.Sprintf("%s:/skill:ro", skillDir),
				fmt.Sprintf("%s:/outputs:rw", d.config.OutputsDir),
			},
			ReadonlyRootfs: true,
			PidsLimit:      &pidsLimit,
			SecurityOpt:    []string{"no-new-privileges"},
			CapDrop:        []string{"ALL"},
		},
		Context: ctx,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Ensure container cleanup
	defer func() {
		d.client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
	}()

	// Start container
	if err := d.client.StartContainer(container.ID, nil); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for completion with timeout
	resultCh := make(chan error, 1)
	go func() {
		exitCode, err := d.client.WaitContainer(container.ID)
		if err != nil {
			resultCh <- err
			return
		}
		if exitCode != 0 {
			resultCh <- fmt.Errorf("script exited with code %d", exitCode)
			return
		}
		resultCh <- nil
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		d.client.StopContainer(container.ID, 1)
		return "", fmt.Errorf("execution timeout after %v", d.config.Timeout)
	case err := <-resultCh:
		if err != nil {
			// Get logs even on error
			logs, _ := d.getContainerLogs(container.ID)
			return logs, err
		}
	}

	// Get output
	output, err := d.getContainerLogs(container.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}

	return output, nil
}

// ensureImage pulls an image if it doesn't exist locally
func (d *DooDockerExecutor) ensureImage(ctx context.Context, image string) error {
	// Check if image exists
	_, err := d.client.InspectImage(image)
	if err == nil {
		return nil // Image exists
	}

	// Pull image
	return d.client.PullImage(docker.PullImageOptions{
		Repository: image,
		Context:    ctx,
	}, docker.AuthConfiguration{})
}

// getContainerLogs retrieves logs from a container
func (d *DooDockerExecutor) getContainerLogs(containerID string) (string, error) {
	var stdout, stderr []byte
	
	// Create pipes for output
	stdoutPipe := &bytesWriter{buf: &stdout}
	stderrPipe := &bytesWriter{buf: &stderr}

	err := d.client.Logs(docker.LogsOptions{
		Container:    containerID,
		OutputStream: stdoutPipe,
		ErrorStream:  stderrPipe,
		Stdout:       true,
		Stderr:       true,
	})

	output := string(stdout)
	if len(stderr) > 0 {
		output += "\n" + string(stderr)
	}

	return output, err
}

// bytesWriter implements io.Writer by writing to a byte slice
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// GetInfo returns information about the Docker/Podman daemon
func (d *DooDockerExecutor) GetInfo() string {
	env, err := d.client.Version()
	if err != nil {
		return "Docker/Podman (DooD, version unknown)"
	}
	version := env.Get("Version")
	return fmt.Sprintf("Docker/Podman %s (DooD via socket)", version)
}

// ExecutePythonCode runs Python code with dual mount support
// workspaceDir: read-write workspace for files and code execution
// skillLibsDir: read-only skill directory for importing helper libraries
func (d *DooDockerExecutor) ExecutePythonCode(ctx context.Context, workspaceDir, skillLibsDir, scriptPath string, args []string) (string, error) {
	image := d.config.GetImageForSkill(skillLibsDir)
	return d.executeCodeInContainer(ctx, workspaceDir, skillLibsDir, image, "python", scriptPath, args)
}

// ExecuteBashCode runs Bash code with dual mount support
// workspaceDir: read-write workspace for files and code execution
// skillLibsDir: read-only skill directory (for future bash libraries)
func (d *DooDockerExecutor) ExecuteBashCode(ctx context.Context, workspaceDir, skillLibsDir, scriptPath string, args []string) (string, error) {
	image := d.config.GetImageForSkill(skillLibsDir)
	return d.executeCodeInContainer(ctx, workspaceDir, skillLibsDir, image, "bash", scriptPath, args)
}

// executeCodeInContainer handles container execution with dual mounts
func (d *DooDockerExecutor) executeCodeInContainer(
	ctx context.Context,
	workspaceDir string,
	skillLibsDir string,
	image string,
	interpreter string,
	scriptPath string,
	args []string,
) (string, error) {
	// Pull image if not present
	if err := d.ensureImage(ctx, image); err != nil {
		return "", fmt.Errorf("failed to ensure image: %w", err)
	}

	// Build command
	cmd := []string{interpreter, scriptPath}
	cmd = append(cmd, args...)

	// Create container with dual mounts
	pidsLimit := int64(100)
	networkMode := d.config.GetNetworkModeForSkill(skillLibsDir)
	container, err := d.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      image,
			Cmd:        cmd,
			WorkingDir: "/workspace",
			Env:        []string{"PYTHONPATH=/skill"},
			Memory:     256 * 1024 * 1024, // 256MB
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/workspace:rw", workspaceDir),  // Read-write workspace
				fmt.Sprintf("%s:/skill:ro", skillLibsDir),      // Read-only skill libs,
				fmt.Sprintf("%s:/outputs:rw", d.config.OutputsDir), // Persistent outputs directory
			},
			ReadonlyRootfs: false, // Can't be read-only with /tmp needed
			Tmpfs:          map[string]string{"/tmp": "rw,exec,size=100m"},
			PidsLimit:      &pidsLimit,
			SecurityOpt:    []string{"no-new-privileges"},
			CapDrop:        []string{"ALL"},
			NetworkMode:    networkMode, // Configurable per skill
		},
		Context: ctx,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Ensure container cleanup
	defer func() {
		d.client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
	}()

	// Start container
	if err := d.client.StartContainer(container.ID, nil); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for completion with timeout
	resultCh := make(chan error, 1)
	go func() {
		exitCode, err := d.client.WaitContainer(container.ID)
		if err != nil {
			resultCh <- err
			return
		}
		if exitCode != 0 {
			resultCh <- fmt.Errorf("code exited with code %d", exitCode)
			return
		}
		resultCh <- nil
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		d.client.StopContainer(container.ID, 1)
		return "", fmt.Errorf("execution timeout after %v", d.config.Timeout)
	case err := <-resultCh:
		if err != nil {
			// Get logs even on error
			logs, _ := d.getContainerLogs(container.ID)
			return logs, err
		}
	}

	// Get output
	output, err := d.getContainerLogs(container.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}

	return output, nil
}

var _ io.Writer = (*bytesWriter)(nil) // Ensure interface compliance

